// Package storageaccount implements the Sprint 3 minimal storage-account
// bootstrap (create + list only). See ADR 0002 — full management (health
// check, quota-aware routing, Drive Manager UI) is Sprint 9.
package storageaccount

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"sonora.dev/go-core/infrastructure/crypto"
	"sonora.dev/go-core/infrastructure/postgres/sqlc"
)

type Account struct {
	ID           uuid.UUID
	Provider     string
	Label        string
	AccountEmail string
	QuotaBytes   *int64
	UsedBytes    int64
	IsActive     bool
	HealthStatus string
}

type Service struct {
	q   *sqlc.Queries
	box *crypto.Box
}

func NewService(q *sqlc.Queries, box *crypto.Box) *Service {
	return &Service{q: q, box: box}
}

// Create registers a storage account from a Drive OAuth refresh token
// obtained out-of-band today (Sprint 9 adds an in-app consent flow). The
// refresh token is encrypted before it ever touches the database.
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

func fromRow(row sqlc.StorageAccount) *Account {
	email := ""
	if row.AccountEmail != nil {
		email = *row.AccountEmail
	}
	return &Account{
		ID:           uuid.UUID(row.ID.Bytes),
		Provider:     row.Provider,
		Label:        row.Label,
		AccountEmail: email,
		QuotaBytes:   row.QuotaBytes,
		UsedBytes:    row.UsedBytes,
		IsActive:     row.IsActive,
		HealthStatus: row.HealthStatus,
	}
}
