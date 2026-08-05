package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"sonora.dev/go-core/application/appsettings"
	appauth "sonora.dev/go-core/application/auth"
	"sonora.dev/go-core/domain/identity"

	"sonora.dev/backend/internal/http/middleware"
	"sonora.dev/backend/internal/http/response"
)

const (
	refreshCookieName = "sonora_refresh_token"
	stateCookieName   = "sonora_oauth_state"
	authCookiePath    = "/api/v1/auth"
)

type AuthHandler struct {
	service     *appauth.Service
	settings    *appsettings.Service
	frontendURL string
	adminURL    string
}

func NewAuthHandler(service *appauth.Service, settings *appsettings.Service, frontendURL, adminURL string) *AuthHandler {
	return &AuthHandler{service: service, settings: settings, frontendURL: frontendURL, adminURL: adminURL}
}

// Config is public (no auth) — the Login page (both apps) needs to know
// whether to show the Google button BEFORE there's any token at all
// (Sprint 14 sisipan, ADR 0012).
func (h *AuthHandler) Config(c *fiber.Ctx) error {
	values, err := h.settings.List(c.Context())
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to load config")
	}
	appName := values[appsettings.KeyAppName]
	if appName == "" {
		appName = "Sonora"
	}
	return response.OK(c, fiber.StatusOK, fiber.Map{
		"google_oauth_enabled": values[appsettings.KeyGoogleOAuthEnabled] == "true",
		"app_name":             appName,
	})
}

// GoogleLogin redirects to Google's consent screen. The state nonce is
// echoed back on the callback and checked against a short-lived cookie to
// guard against CSRF. ?app=admin also travels in the state (never as a
// separate untrusted redirect param) so the callback knows whether to
// hand off to the main frontend or the admin app — Sprint 9's Drive
// Manager needs its own login, same Google account, same Owner check.
func (h *AuthHandler) GoogleLogin(c *fiber.Ctx) error {
	if !h.settings.IsGoogleOAuthEnabled(c.Context()) {
		return response.Fail(c, fiber.StatusForbidden, "google_oauth_disabled", "Google login sedang dinonaktifkan admin")
	}

	app := "frontend"
	if c.Query("app") == "admin" {
		app = "admin"
	}

	nonce, err := randomState()
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to start login")
	}
	state := nonce + ":" + app

	c.Cookie(&fiber.Cookie{
		Name:     stateCookieName,
		Value:    state,
		Path:     authCookiePath,
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteLaxMode,
		Expires:  time.Now().Add(5 * time.Minute),
	})

	return c.Redirect(h.service.GoogleAuthURL(state), fiber.StatusFound)
}

// GoogleCallback exchanges the code, upserts the user + device + tokens,
// then hands off to whichever app started the flow: refresh token as an
// httpOnly cookie, access token in the redirect fragment (never sent to
// any server, unlike a query string).
func (h *AuthHandler) GoogleCallback(c *fiber.Ctx) error {
	if !h.settings.IsGoogleOAuthEnabled(c.Context()) {
		return response.Fail(c, fiber.StatusForbidden, "google_oauth_disabled", "Google login sedang dinonaktifkan admin")
	}

	state := c.Query("state")
	cookieState := c.Cookies(stateCookieName)
	c.ClearCookie(stateCookieName)

	if state == "" || cookieState == "" || state != cookieState {
		return c.Redirect(h.frontendURL+"/login?error=invalid_state", fiber.StatusFound)
	}
	targetURL := h.resolveTargetURL(state)

	code := c.Query("code")
	if code == "" {
		return c.Redirect(targetURL+"/login?error=missing_code", fiber.StatusFound)
	}

	deviceName := c.Get(fiber.HeaderUserAgent)
	if deviceName == "" {
		deviceName = "Unknown Device"
	}
	if len(deviceName) > 120 {
		deviceName = deviceName[:120]
	}

	_, pair, err := h.service.HandleGoogleCallback(c.Context(), code, deviceName, identity.DeviceWeb)
	if err != nil {
		return c.Redirect(targetURL+"/login?error=oauth_failed", fiber.StatusFound)
	}

	h.setRefreshCookie(c, pair.RefreshToken, pair.RefreshTokenExpiresAt)

	redirectURL := targetURL + "/auth/callback#access_token=" + pair.AccessToken +
		"&expires_in=" + expiresInSeconds(pair.AccessTokenExpiresAt) +
		"&device_id=" + pair.DeviceID.String()
	return c.Redirect(redirectURL, fiber.StatusFound)
}

// resolveTargetURL maps the state's embedded app name to one of exactly
// two configured URLs — never an arbitrary caller-supplied redirect, so
// this can't become an open redirect.
func (h *AuthHandler) resolveTargetURL(state string) string {
	_, app, found := strings.Cut(state, ":")
	if found && app == "admin" {
		return h.adminURL
	}
	return h.frontendURL
}

type loginAdminRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginAdmin is the admin app's credential login (Sprint 14 sisipan,
// ADR 0012) — email+password, must resolve to role Owner.
func (h *AuthHandler) LoginAdmin(c *fiber.Ctx) error {
	var req loginAdminRequest
	if err := c.BodyParser(&req); err != nil || req.Email == "" || req.Password == "" {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "email and password are required")
	}

	deviceName := c.Get(fiber.HeaderUserAgent)
	if deviceName == "" {
		deviceName = "Unknown Device"
	}
	if len(deviceName) > 120 {
		deviceName = deviceName[:120]
	}

	_, pair, err := h.service.LoginAdmin(c.Context(), req.Email, req.Password, deviceName, identity.DeviceWeb)
	if err != nil {
		return response.Fail(c, fiber.StatusUnauthorized, "unauthenticated", "email atau password salah")
	}

	h.setRefreshCookie(c, pair.RefreshToken, pair.RefreshTokenExpiresAt)
	return response.OK(c, fiber.StatusOK, fiber.Map{
		"access_token": pair.AccessToken,
		"expires_in":   int(time.Until(pair.AccessTokenExpiresAt).Seconds()),
		"device_id":    pair.DeviceID,
	})
}

type loginMemberRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login is the main app's credential login — username+password.
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req loginMemberRequest
	if err := c.BodyParser(&req); err != nil || req.Username == "" || req.Password == "" {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "username and password are required")
	}

	deviceName := c.Get(fiber.HeaderUserAgent)
	if deviceName == "" {
		deviceName = "Unknown Device"
	}
	if len(deviceName) > 120 {
		deviceName = deviceName[:120]
	}

	_, pair, err := h.service.LoginMember(c.Context(), req.Username, req.Password, deviceName, identity.DeviceWeb)
	if err != nil {
		return response.Fail(c, fiber.StatusUnauthorized, "unauthenticated", "username atau password salah")
	}

	h.setRefreshCookie(c, pair.RefreshToken, pair.RefreshTokenExpiresAt)
	return response.OK(c, fiber.StatusOK, fiber.Map{
		"access_token": pair.AccessToken,
		"expires_in":   int(time.Until(pair.AccessTokenExpiresAt).Seconds()),
		"device_id":    pair.DeviceID,
	})
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	raw := c.Cookies(refreshCookieName)
	if raw == "" {
		return response.Fail(c, fiber.StatusUnauthorized, "unauthenticated", "missing refresh token")
	}

	pair, err := h.service.Refresh(c.Context(), raw)
	if err != nil {
		h.clearRefreshCookie(c)
		if errors.Is(err, appauth.ErrRefreshTokenReused) {
			log.Printf("auth: refresh token reuse detected, all sessions revoked (request_id=%s)", c.Locals("requestid"))
			return response.Fail(c, fiber.StatusUnauthorized, "unauthenticated", "invalid or expired refresh token")
		}
		if errors.Is(err, appauth.ErrInvalidRefreshToken) {
			return response.Fail(c, fiber.StatusUnauthorized, "unauthenticated", "invalid or expired refresh token")
		}
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to refresh session")
	}

	h.setRefreshCookie(c, pair.RefreshToken, pair.RefreshTokenExpiresAt)
	return response.OK(c, fiber.StatusOK, fiber.Map{
		"access_token": pair.AccessToken,
		"expires_in":   int(time.Until(pair.AccessTokenExpiresAt).Seconds()),
		"device_id":    pair.DeviceID,
	})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	raw := c.Cookies(refreshCookieName)
	h.clearRefreshCookie(c)
	if raw == "" {
		return response.OK(c, fiber.StatusOK, fiber.Map{})
	}
	if err := h.service.Logout(c.Context(), raw); err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to logout")
	}
	return response.OK(c, fiber.StatusOK, fiber.Map{})
}

// LogoutAll revokes every refresh token for the calling user. Restricted to
// Owner by the route's RequireRole middleware.
func (h *AuthHandler) LogoutAll(c *fiber.Ctx) error {
	h.clearRefreshCookie(c)
	if err := h.service.LogoutAll(c.Context(), middleware.UserID(c)); err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to logout all sessions")
	}
	return response.OK(c, fiber.StatusOK, fiber.Map{})
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	user, err := h.service.Me(c.Context(), middleware.UserID(c))
	if err != nil {
		if errors.Is(err, identity.ErrNotFound) {
			return response.Fail(c, fiber.StatusNotFound, "not_found", "user not found")
		}
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to load user")
	}
	return response.OK(c, fiber.StatusOK, userJSON(user))
}

// maxAvatarDataURLLen caps the profile page's "change photo" upload
// (Sprint 14 sisipan, ADR 0009) — the client resizes to a small
// thumbnail before sending, so a real thumbnail is always well under
// this; it just guards against something huge landing in the DB.
const maxAvatarDataURLLen = 300_000

type updateMeRequest struct {
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

// UpdateMe edits the caller's own name/avatar (Sprint 14 sisipan).
// avatar_url is expected to be a small data: URL (see ADR 0009 for why
// this doesn't go through the Drive storage pool) — not validated as a
// real image beyond a size cap and a data:image/ prefix check.
func (h *AuthHandler) UpdateMe(c *fiber.Ctx) error {
	var req updateMeRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid request body")
	}
	if req.AvatarURL != "" {
		if len(req.AvatarURL) > maxAvatarDataURLLen {
			return response.Fail(c, fiber.StatusBadRequest, "validation_error", "avatar image is too large")
		}
		if !strings.HasPrefix(req.AvatarURL, "data:image/") {
			return response.Fail(c, fiber.StatusBadRequest, "validation_error", "avatar_url must be a data:image/... URL")
		}
	}
	user, err := h.service.UpdateMe(c.Context(), middleware.UserID(c), req.Name, req.AvatarURL)
	if err != nil {
		if errors.Is(err, identity.ErrNotFound) {
			return response.Fail(c, fiber.StatusNotFound, "not_found", "user not found")
		}
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to update profile")
	}
	return response.OK(c, fiber.StatusOK, userJSON(user))
}

func (h *AuthHandler) setRefreshCookie(c *fiber.Ctx, value string, expiresAt time.Time) {
	c.Cookie(&fiber.Cookie{
		Name:     refreshCookieName,
		Value:    value,
		Path:     authCookiePath,
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteLaxMode,
		Expires:  expiresAt,
	})
}

func (h *AuthHandler) clearRefreshCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     authCookiePath,
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteLaxMode,
		Expires:  time.Now().Add(-time.Hour),
	})
}

func userJSON(u *identity.User) fiber.Map {
	return fiber.Map{
		"id":         u.ID,
		"email":      u.Email,
		"name":       u.Name,
		"avatar_url": u.AvatarURL,
		"role":       u.Role,
		"created_at": u.CreatedAt,
	}
}

func randomState() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func expiresInSeconds(t time.Time) string {
	seconds := int(time.Until(t).Seconds())
	if seconds < 0 {
		seconds = 0
	}
	return strconv.Itoa(seconds)
}
