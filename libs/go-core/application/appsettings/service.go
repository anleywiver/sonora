// Package appsettings implements the generic key-value settings store
// (Sprint 14 sisipan, ADR 0012) backing the Google OAuth runtime toggle
// and the Admin Settings page (app name, language, maintenance mode).
package appsettings

import (
	"context"
	"fmt"
)

const (
	KeyGoogleOAuthEnabled = "google_oauth_enabled"
	KeyMaintenanceMode    = "maintenance_mode"
	KeyAppName            = "app_name"
	KeyDefaultLanguage    = "default_language"
)

type Repository interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string) error
	List(ctx context.Context) (map[string]string, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) (map[string]string, error) {
	values, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("appsettings: list: %w", err)
	}
	return values, nil
}

func (s *Service) Set(ctx context.Context, key, value string) error {
	if err := s.repo.Set(ctx, key, value); err != nil {
		return fmt.Errorf("appsettings: set %s: %w", key, err)
	}
	return nil
}

// IsGoogleOAuthEnabled defaults to false (fail-safe) if the row is
// somehow missing rather than erroring the whole auth flow.
func (s *Service) IsGoogleOAuthEnabled(ctx context.Context) bool {
	v, err := s.repo.Get(ctx, KeyGoogleOAuthEnabled)
	if err != nil {
		return false
	}
	return v == "true"
}
