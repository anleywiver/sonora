package identity

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// RefreshToken stores only a hash of the token — the raw value is handed to
// the client once and never persisted. Rotated on every use, bound to a
// single device.
type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	DeviceID  uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

func (t RefreshToken) IsValid(now time.Time) bool {
	return t.RevokedAt == nil && now.Before(t.ExpiresAt)
}

type RefreshTokenRepository interface {
	FindByHash(ctx context.Context, tokenHash string) (*RefreshToken, error)
	Create(ctx context.Context, token *RefreshToken) error
	Revoke(ctx context.Context, id uuid.UUID, revokedAt time.Time) error
	RevokeAllForUser(ctx context.Context, userID uuid.UUID, revokedAt time.Time) error
	RevokeAllForDevice(ctx context.Context, deviceID uuid.UUID, revokedAt time.Time) error
	// DeleteExpired removes rows past their expiry regardless of revoked
	// state — Sprint 10 garbage collector housekeeping, not a security
	// control (that's Sprint 12).
	DeleteExpired(ctx context.Context, before time.Time) (int64, error)
}
