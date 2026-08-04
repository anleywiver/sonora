package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

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
	frontendURL string
	adminURL    string
}

func NewAuthHandler(service *appauth.Service, frontendURL, adminURL string) *AuthHandler {
	return &AuthHandler{service: service, frontendURL: frontendURL, adminURL: adminURL}
}

// GoogleLogin redirects to Google's consent screen. The state nonce is
// echoed back on the callback and checked against a short-lived cookie to
// guard against CSRF. ?app=admin also travels in the state (never as a
// separate untrusted redirect param) so the callback knows whether to
// hand off to the main frontend or the admin app — Sprint 9's Drive
// Manager needs its own login, same Google account, same Owner check.
func (h *AuthHandler) GoogleLogin(c *fiber.Ctx) error {
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

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	raw := c.Cookies(refreshCookieName)
	if raw == "" {
		return response.Fail(c, fiber.StatusUnauthorized, "unauthenticated", "missing refresh token")
	}

	pair, err := h.service.Refresh(c.Context(), raw)
	if err != nil {
		h.clearRefreshCookie(c)
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
