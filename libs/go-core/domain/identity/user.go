package identity

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleOwner  Role = "owner"
	RoleMember Role = "member"
)

type User struct {
	ID        uuid.UUID
	GoogleID  string
	Email     string
	Name      string
	AvatarURL string
	Role      Role
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (u User) IsOwner() bool { return u.Role == RoleOwner }

type UserRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByGoogleID(ctx context.Context, googleID string) (*User, error)
	Create(ctx context.Context, user *User) error
	Count(ctx context.Context) (int64, error)
	// FindOwner returns the oldest Owner user — scheduled jobs that aren't
	// tied to an HTTP request (Bandcamp/cloud sync sync) attribute their
	// ingest_jobs rows to this user, per the single-owner personal
	// deployment assumption from Sprint 2 (see ADR 0004).
	FindOwner(ctx context.Context) (*User, error)
}
