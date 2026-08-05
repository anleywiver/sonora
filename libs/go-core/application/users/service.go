// Package users implements the admin Manage Users page (Sprint 14
// sisipan, ADR 0009): invite-before-first-login, list, and remove access.
// Auth itself (Google OAuth exchange, token issuance) stays in
// application/auth — this package only owns the admin-facing user
// management actions.
package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"sonora.dev/go-core/domain/identity"
)

var (
	ErrEmailTaken        = errors.New("users: a user with this email already exists")
	ErrCannotRemoveOwner = errors.New("users: cannot remove an Owner's access")
	ErrNotFound          = identity.ErrNotFound
)

type User struct {
	ID        uuid.UUID
	Name      string
	Email     string
	Role      string
	IsPending bool
	CreatedAt string
}

type Service struct {
	users identity.UserRepository
}

func NewService(users identity.UserRepository) *Service {
	return &Service{users: users}
}

func (s *Service) List(ctx context.Context) ([]User, error) {
	rows, err := s.users.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("users: list: %w", err)
	}
	out := make([]User, 0, len(rows))
	for _, u := range rows {
		out = append(out, User{
			ID:        u.ID,
			Name:      u.Name,
			Email:     u.Email,
			Role:      string(u.Role),
			IsPending: u.IsPending(),
			CreatedAt: u.CreatedAt.Format("2006-01-02"),
		})
	}
	return out, nil
}

// Invite pre-creates a Member row with no google_id yet (ADR 0009) — no
// email is actually sent (this project has no email infrastructure at
// all); the Owner is expected to reach out out-of-band (e.g. the
// WhatsApp link on the Login page). The row is claimed automatically the
// first time this email successfully logs in via Google.
func (s *Service) Invite(ctx context.Context, email, name string) (*User, error) {
	if _, err := s.users.FindByEmail(ctx, email); err == nil {
		return nil, ErrEmailTaken
	} else if !errors.Is(err, identity.ErrNotFound) {
		return nil, fmt.Errorf("users: check existing email: %w", err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("users: generate id: %w", err)
	}
	if name == "" {
		name = email
	}
	user := &identity.User{
		ID:    id,
		Email: email,
		Name:  name,
		Role:  identity.RoleMember,
		// GoogleID left "" — the repository maps that to NULL (pending).
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("users: invite: %w", err)
	}
	return &User{ID: user.ID, Name: user.Name, Email: user.Email, Role: string(user.Role), IsPending: true, CreatedAt: user.CreatedAt.Format("2006-01-02")}, nil
}

// RemoveAccess deletes a user's access outright. Refuses to remove an
// Owner (ADR 0009) — there's no scenario in a personal deployment where
// that should happen through this admin action.
func (s *Service) RemoveAccess(ctx context.Context, id uuid.UUID) error {
	user, err := s.users.FindByID(ctx, id)
	if errors.Is(err, identity.ErrNotFound) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("users: find: %w", err)
	}
	if user.IsOwner() {
		return ErrCannotRemoveOwner
	}
	if err := s.users.Delete(ctx, id); err != nil {
		return fmt.Errorf("users: delete: %w", err)
	}
	return nil
}
