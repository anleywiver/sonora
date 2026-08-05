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
	ID uuid.UUID
	// GoogleID is "" for a user invited (Sprint 14 sisipan, ADR 0009) but
	// who hasn't logged in yet — the repository maps that to/from a real
	// NULL in the database, so the rest of the app never deals with a
	// pointer for this.
	GoogleID  string
	Email     string
	Name      string
	AvatarURL string
	Role      Role
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (u User) IsOwner() bool { return u.Role == RoleOwner }

// IsPending is true for an invited user who has never actually logged in
// via Google yet (ADR 0009) — the admin Manage Users page shows this as
// a distinct "Invited" status rather than "Active".
func (u User) IsPending() bool { return u.GoogleID == "" }

type UserRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByGoogleID(ctx context.Context, googleID string) (*User, error)
	// FindByEmail supports claiming a pending invite on first login (ADR
	// 0009) — returns ErrNotFound if no user (pending or active) has this
	// email.
	FindByEmail(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, user *User) error
	// ClaimInvite fills in a pending invite's google_id/name/avatar_url on
	// first real login — id/email/role are left untouched.
	ClaimInvite(ctx context.Context, id uuid.UUID, googleID, name, avatarURL string) error
	Update(ctx context.Context, user *User) error
	// Delete removes a user's access outright — the caller (application/
	// users) is responsible for refusing this for an Owner.
	Delete(ctx context.Context, id uuid.UUID) error
	ListAll(ctx context.Context) ([]*User, error)
	Count(ctx context.Context) (int64, error)
	// FindOwner returns the oldest Owner user — scheduled jobs that aren't
	// tied to an HTTP request (Bandcamp/cloud sync sync) attribute their
	// ingest_jobs rows to this user, per the single-owner personal
	// deployment assumption from Sprint 2 (see ADR 0004).
	FindOwner(ctx context.Context) (*User, error)
}
