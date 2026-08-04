// Package auth implements the login/session use cases: Google OAuth
// exchange, JWT access tokens, and refresh token rotation bound to a
// device. IDs are generated here (UUIDv7), never left to Postgres.
package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"sonora.dev/go-core/domain/identity"
	"sonora.dev/go-core/infrastructure/jwt"
	"sonora.dev/go-core/infrastructure/oauth"
)

type TokenPair struct {
	AccessToken           string
	AccessTokenExpiresAt  time.Time
	RefreshToken          string
	RefreshTokenExpiresAt time.Time
	// DeviceID lets the frontend know "which device am I" — needed for
	// Active Device (Sprint 8): comparing against playback_state's
	// active_device_id and passing to /ws/token. Not carried in the JWT
	// itself (access tokens are deliberately device-agnostic).
	DeviceID uuid.UUID
}

type Service struct {
	users   identity.UserRepository
	devices identity.DeviceRepository
	tokens  identity.RefreshTokenRepository
	google  *oauth.GoogleClient
	jwt     *jwt.Issuer

	refreshTTL time.Duration
}

func NewService(
	users identity.UserRepository,
	devices identity.DeviceRepository,
	tokens identity.RefreshTokenRepository,
	google *oauth.GoogleClient,
	jwtIssuer *jwt.Issuer,
	refreshTTL time.Duration,
) *Service {
	return &Service{
		users:      users,
		devices:    devices,
		tokens:     tokens,
		google:     google,
		jwt:        jwtIssuer,
		refreshTTL: refreshTTL,
	}
}

func (s *Service) GoogleAuthURL(state string) string {
	return s.google.AuthURL(state)
}

// HandleGoogleCallback finds-or-creates the user, records a new device for
// this session, and issues a fresh token pair. The very first user ever
// created becomes Owner; everyone after is Member.
func (s *Service) HandleGoogleCallback(ctx context.Context, code, deviceName string, deviceType identity.DeviceType) (*identity.User, *TokenPair, error) {
	profile, err := s.google.Exchange(ctx, code)
	if err != nil {
		return nil, nil, err
	}

	user, err := s.users.FindByGoogleID(ctx, profile.Sub)
	if errors.Is(err, identity.ErrNotFound) {
		user, err = s.createUser(ctx, profile)
		if err != nil {
			return nil, nil, err
		}
	} else if err != nil {
		return nil, nil, err
	}

	deviceID, err := uuid.NewV7()
	if err != nil {
		return nil, nil, fmt.Errorf("auth: generate device id: %w", err)
	}
	device := &identity.Device{
		ID:     deviceID,
		UserID: user.ID,
		Name:   deviceName,
		Type:   deviceType,
	}
	if err := s.devices.Create(ctx, device); err != nil {
		return nil, nil, err
	}

	pair, err := s.issueTokenPair(ctx, user, device.ID)
	if err != nil {
		return nil, nil, err
	}
	return user, pair, nil
}

func (s *Service) createUser(ctx context.Context, profile *oauth.GoogleProfile) (*identity.User, error) {
	count, err := s.users.Count(ctx)
	if err != nil {
		return nil, err
	}
	role := identity.RoleMember
	if count == 0 {
		role = identity.RoleOwner
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("auth: generate user id: %w", err)
	}
	user := &identity.User{
		ID:        id,
		GoogleID:  profile.Sub,
		Email:     profile.Email,
		Name:      profile.Name,
		AvatarURL: profile.Picture,
		Role:      role,
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// Refresh rotates the refresh token: the presented one is revoked and a new
// pair is issued, bound to the same device.
//
// A presented token that was already revoked (as opposed to merely
// expired) is a reuse signal — a rotated-out token being replayed, which
// only happens if it leaked and two parties are racing to use it. Sprint
// 12 hardening (ADR 0006): that case revokes every session the user has,
// not just this one request, since we can no longer tell which of the
// two parties is legitimate.
func (s *Service) Refresh(ctx context.Context, rawRefreshToken string) (*TokenPair, error) {
	existing, err := s.tokens.FindByHash(ctx, hashToken(rawRefreshToken))
	if err != nil {
		if errors.Is(err, identity.ErrNotFound) {
			return nil, ErrInvalidRefreshToken
		}
		return nil, err
	}
	if existing.RevokedAt != nil {
		if revokeErr := s.tokens.RevokeAllForUser(ctx, existing.UserID, time.Now()); revokeErr != nil {
			return nil, revokeErr
		}
		return nil, ErrRefreshTokenReused
	}
	if !existing.IsValid(time.Now()) {
		return nil, ErrInvalidRefreshToken
	}
	if err := s.tokens.Revoke(ctx, existing.ID, time.Now()); err != nil {
		return nil, err
	}

	user, err := s.users.FindByID(ctx, existing.UserID)
	if err != nil {
		return nil, err
	}
	return s.issueTokenPair(ctx, user, existing.DeviceID)
}

// Logout revokes the presented refresh token. Unknown/already-revoked
// tokens are treated as already logged out rather than an error.
func (s *Service) Logout(ctx context.Context, rawRefreshToken string) error {
	existing, err := s.tokens.FindByHash(ctx, hashToken(rawRefreshToken))
	if errors.Is(err, identity.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	return s.tokens.Revoke(ctx, existing.ID, time.Now())
}

func (s *Service) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	return s.tokens.RevokeAllForUser(ctx, userID, time.Now())
}

func (s *Service) Me(ctx context.Context, userID uuid.UUID) (*identity.User, error) {
	return s.users.FindByID(ctx, userID)
}

func (s *Service) ListDevices(ctx context.Context, userID uuid.UUID) ([]*identity.Device, error) {
	return s.devices.ListByUser(ctx, userID)
}

// RemoveDevice deletes the device (scoped to its owner) and revokes any
// refresh tokens bound to it, forcing that session to log out.
func (s *Service) RemoveDevice(ctx context.Context, deviceID, userID uuid.UUID) error {
	if err := s.devices.Delete(ctx, deviceID, userID); err != nil {
		return err
	}
	return s.tokens.RevokeAllForDevice(ctx, deviceID, time.Now())
}

// SetActiveDevice marks deviceID as the user's Active Device (Sprint 8
// Transfer Playback) — devices.is_active bookkeeping only; the
// authoritative active device for playback purposes is
// playback_states.active_device_id (application/playback), kept in sync
// by the caller (PlayerHandler.Transfer).
func (s *Service) SetActiveDevice(ctx context.Context, userID, deviceID uuid.UUID) error {
	return s.devices.SetActive(ctx, userID, deviceID)
}

func (s *Service) issueTokenPair(ctx context.Context, user *identity.User, deviceID uuid.UUID) (*TokenPair, error) {
	access, accessExp, err := s.jwt.Issue(user.ID, string(user.Role))
	if err != nil {
		return nil, err
	}

	rawRefresh, hash, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}
	tokenID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("auth: generate refresh token id: %w", err)
	}
	refreshExp := time.Now().Add(s.refreshTTL)
	rt := &identity.RefreshToken{
		ID:        tokenID,
		UserID:    user.ID,
		DeviceID:  deviceID,
		TokenHash: hash,
		ExpiresAt: refreshExp,
	}
	if err := s.tokens.Create(ctx, rt); err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:           access,
		AccessTokenExpiresAt:  accessExp,
		RefreshToken:          rawRefresh,
		RefreshTokenExpiresAt: refreshExp,
		DeviceID:              deviceID,
	}, nil
}
