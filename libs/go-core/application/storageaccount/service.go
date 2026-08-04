// Package storageaccount implements storage account management: create,
// list, delete, and health check (quota-aware routing lives in
// application/ingest and application/catalog, which query
// GetActiveStorageAccount directly). Started as the Sprint 3 minimal
// bootstrap (ADR 0002); Sprint 9 adds delete + health check + the Drive
// Manager admin UI that drives them.
package storageaccount

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"sonora.dev/go-core/infrastructure/crypto"
	"sonora.dev/go-core/infrastructure/postgres/sqlc"
	"sonora.dev/go-core/infrastructure/storage"
)

var (
	ErrNotFound = errors.New("storageaccount: not found")
	// ErrInUse means the account still has files stored on it — deleting
	// it would orphan those songs, so it's rejected rather than cascaded.
	ErrInUse = errors.New("storageaccount: account still has files stored on it")
)

type Account struct {
	ID                uuid.UUID
	Provider          string
	Label             string
	AccountEmail      string
	QuotaBytes        *int64
	UsedBytes         int64
	IsActive          bool
	HealthStatus      string
	LastHealthCheckAt *time.Time
}

type Service struct {
	q                  *sqlc.Queries
	box                *crypto.Box
	googleClientID     string
	googleClientSecret string
}

func NewService(q *sqlc.Queries, box *crypto.Box, googleClientID, googleClientSecret string) *Service {
	return &Service{q: q, box: box, googleClientID: googleClientID, googleClientSecret: googleClientSecret}
}

// Create registers a storage account from a Drive OAuth refresh token
// obtained out-of-band today (no in-app consent flow yet — a personal/
// family-scale VPS deployment doesn't need one badly enough to justify
// building it). The refresh token is encrypted before it ever touches
// the database.
func (s *Service) Create(ctx context.Context, label, accountEmail, refreshToken string, quotaBytes *int64) (*Account, error) {
	encrypted, err := s.box.Encrypt(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("storageaccount: encrypt credentials: %w", err)
	}
	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("storageaccount: generate id: %w", err)
	}
	var email *string
	if accountEmail != "" {
		email = &accountEmail
	}
	row, err := s.q.CreateStorageAccount(ctx, sqlc.CreateStorageAccountParams{
		ID:                   pgtype.UUID{Bytes: id, Valid: true},
		Provider:             "google_drive",
		Label:                label,
		AccountEmail:         email,
		CredentialsEncrypted: encrypted,
		QuotaBytes:           quotaBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("storageaccount: create: %w", err)
	}
	return fromRow(row), nil
}

func (s *Service) List(ctx context.Context) ([]*Account, error) {
	rows, err := s.q.ListStorageAccounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("storageaccount: list: %w", err)
	}
	out := make([]*Account, 0, len(rows))
	for _, row := range rows {
		out = append(out, fromRow(row))
	}
	return out, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	affected, err := s.q.DeleteStorageAccount(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		// storage_files.storage_account_id is ON DELETE RESTRICT — Postgres
		// rejects this at the FK level rather than orphaning songs.
		return fmt.Errorf("%w: %v", ErrInUse, err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

// HealthCheck confirms the stored credentials still work and refreshes
// quota_bytes/used_bytes from the provider — the numbers GetActiveStorageAccount's
// quota-aware routing (application/ingest) relies on.
func (s *Service) HealthCheck(ctx context.Context, id uuid.UUID) (*Account, error) {
	row, err := s.q.GetStorageAccountByID(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return nil, ErrNotFound
	}

	refreshToken, err := s.box.Decrypt(row.CredentialsEncrypted)
	if err != nil {
		return nil, fmt.Errorf("storageaccount: decrypt credentials: %w", err)
	}

	provider := storage.NewGoogleDriveProvider(ctx, s.googleClientID, s.googleClientSecret, refreshToken)
	quota, err := provider.HealthCheck(ctx)

	healthStatus := "healthy"
	var quotaBytes *int64
	usedBytes := row.UsedBytes
	if err != nil {
		healthStatus = "down"
	} else {
		quotaBytes = &quota.LimitBytes
		usedBytes = quota.UsedBytes
	}

	if updateErr := s.q.UpdateStorageAccountHealth(ctx, sqlc.UpdateStorageAccountHealthParams{
		ID:           row.ID,
		HealthStatus: healthStatus,
		QuotaBytes:   quotaBytes,
		UsedBytes:    usedBytes,
	}); updateErr != nil {
		return nil, fmt.Errorf("storageaccount: record health check: %w", updateErr)
	}

	updated, err := s.q.GetStorageAccountByID(ctx, row.ID)
	if err != nil {
		return nil, fmt.Errorf("storageaccount: reload after health check: %w", err)
	}
	return fromRow(updated), nil
}

func fromRow(row sqlc.StorageAccount) *Account {
	email := ""
	if row.AccountEmail != nil {
		email = *row.AccountEmail
	}
	var lastHealthCheckAt *time.Time
	if row.LastHealthCheckAt.Valid {
		lastHealthCheckAt = &row.LastHealthCheckAt.Time
	}
	return &Account{
		ID:                uuid.UUID(row.ID.Bytes),
		Provider:          row.Provider,
		Label:             row.Label,
		AccountEmail:      email,
		QuotaBytes:        row.QuotaBytes,
		UsedBytes:         row.UsedBytes,
		IsActive:          row.IsActive,
		HealthStatus:      row.HealthStatus,
		LastHealthCheckAt: lastHealthCheckAt,
	}
}
