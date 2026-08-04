// Package ingestsource manages Bandcamp/cloud-sync connections and syncs
// them into the same Accept -> Process ingest pipeline manual uploads use
// (Sprint 10, see docs/decisions/0004-sprint10-scheduled-jobs-and-ingest-sources.md).
// Connections are Owner-managed and global (mirrors application/storageaccount),
// not per-regular-user.
package ingestsource

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"sonora.dev/go-core/application/ingest"
	"sonora.dev/go-core/domain/identity"
	"sonora.dev/go-core/infrastructure/bandcamp"
	"sonora.dev/go-core/infrastructure/crypto"
	"sonora.dev/go-core/infrastructure/dropbox"
	"sonora.dev/go-core/infrastructure/postgres/sqlc"
)

var ErrNotFound = errors.New("ingestsource: connection not found")

type Connection struct {
	ID           uuid.UUID
	Provider     string
	Label        string
	AccountEmail string
	IsActive     bool
	LastSyncedAt *time.Time
}

type Service struct {
	q              *sqlc.Queries
	box            *crypto.Box
	users          identity.UserRepository
	ingest         *ingest.Service
	bandcampClient *bandcamp.Client
	dropboxClient  *dropbox.Client
	tmpDir         string
}

func NewService(q *sqlc.Queries, box *crypto.Box, users identity.UserRepository, ingestSvc *ingest.Service, tmpDir string) *Service {
	return &Service{
		q:              q,
		box:            box,
		users:          users,
		ingest:         ingestSvc,
		bandcampClient: bandcamp.NewClient(),
		dropboxClient:  dropbox.NewClient(),
		tmpDir:         tmpDir,
	}
}

// Connect stores a provider connection. credentialsJSON is the already-
// marshaled provider-specific credential struct (bandcamp.Credentials or
// dropbox.Credentials) — kept opaque here since storage/encryption
// doesn't care which provider it is.
func (s *Service) Connect(ctx context.Context, provider, label, accountEmail, credentialsJSON string) (*Connection, error) {
	encrypted, err := s.box.Encrypt(credentialsJSON)
	if err != nil {
		return nil, fmt.Errorf("ingestsource: encrypt credentials: %w", err)
	}
	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("ingestsource: generate id: %w", err)
	}
	var email *string
	if accountEmail != "" {
		email = &accountEmail
	}
	row, err := s.q.CreateIngestSourceConnection(ctx, sqlc.CreateIngestSourceConnectionParams{
		ID:                   toPgUUID(id),
		Provider:             provider,
		Label:                label,
		AccountEmail:         email,
		CredentialsEncrypted: encrypted,
	})
	if err != nil {
		return nil, fmt.Errorf("ingestsource: create connection: %w", err)
	}
	return fromRow(row), nil
}

func (s *Service) List(ctx context.Context) ([]*Connection, error) {
	rows, err := s.q.ListIngestSourceConnections(ctx)
	if err != nil {
		return nil, fmt.Errorf("ingestsource: list: %w", err)
	}
	out := make([]*Connection, 0, len(rows))
	for _, row := range rows {
		out = append(out, fromRow(row))
	}
	return out, nil
}

func (s *Service) Disconnect(ctx context.Context, id uuid.UUID) error {
	affected, err := s.q.DeleteIngestSourceConnection(ctx, toPgUUID(id))
	if err != nil {
		return fmt.Errorf("ingestsource: delete: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

// SyncAll runs Sync for every active connection, best-effort — one
// connection's failure (e.g. an expired Bandcamp cookie) doesn't stop the
// others. Called by the Sprint 10 scheduler and by the admin "Sync now"
// button alike.
func (s *Service) SyncAll(ctx context.Context) {
	rows, err := s.q.ListActiveIngestSourceConnections(ctx)
	if err != nil {
		log.Printf("ingestsource: sync_all: list connections: %v", err)
		return
	}
	for _, row := range rows {
		if err := s.Sync(ctx, fromPgUUID(row.ID)); err != nil {
			log.Printf("ingestsource: sync connection %s (%s) failed: %v", fromPgUUID(row.ID), row.Label, err)
		}
	}
}

func (s *Service) Sync(ctx context.Context, connectionID uuid.UUID) error {
	row, err := s.q.GetIngestSourceConnectionByID(ctx, toPgUUID(connectionID))
	if err != nil {
		return ErrNotFound
	}

	credentialsJSON, err := s.box.Decrypt(row.CredentialsEncrypted)
	if err != nil {
		return fmt.Errorf("ingestsource: decrypt credentials: %w", err)
	}

	owner, err := s.users.FindOwner(ctx)
	if err != nil {
		return fmt.Errorf("ingestsource: find owner user: %w", err)
	}

	since := time.Unix(0, 0)
	if row.LastSyncedAt.Valid {
		since = row.LastSyncedAt.Time
	}

	var syncErr error
	switch row.Provider {
	case "bandcamp":
		syncErr = s.syncBandcamp(ctx, credentialsJSON, owner.ID, since)
	case "cloud_sync":
		syncErr = s.syncDropbox(ctx, credentialsJSON, owner.ID, since)
	default:
		syncErr = fmt.Errorf("unknown provider %q", row.Provider)
	}
	if syncErr != nil {
		return syncErr
	}

	if err := s.q.UpdateIngestSourceConnectionSync(ctx, sqlc.UpdateIngestSourceConnectionSyncParams{
		ID:           row.ID,
		LastSyncedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}); err != nil {
		return fmt.Errorf("ingestsource: record sync time: %w", err)
	}
	return nil
}

func (s *Service) syncBandcamp(ctx context.Context, credentialsJSON string, ownerID uuid.UUID, since time.Time) error {
	var creds bandcamp.Credentials
	if err := json.Unmarshal([]byte(credentialsJSON), &creds); err != nil {
		return fmt.Errorf("parse bandcamp credentials: %w", err)
	}

	purchases, err := s.bandcampClient.ListNewPurchases(ctx, creds, since)
	if err != nil {
		return fmt.Errorf("list bandcamp purchases: %w", err)
	}

	for _, purchase := range purchases {
		purchase := purchase
		err := s.downloadAndAccept(ctx, ownerID, "bandcamp", func(w io.Writer) error {
			return s.bandcampClient.Download(ctx, purchase, w)
		})
		if err != nil {
			log.Printf("ingestsource: bandcamp: ingest %q failed: %v", purchase.Title, err)
		}
	}
	return nil
}

func (s *Service) syncDropbox(ctx context.Context, credentialsJSON string, ownerID uuid.UUID, since time.Time) error {
	var creds dropbox.Credentials
	if err := json.Unmarshal([]byte(credentialsJSON), &creds); err != nil {
		return fmt.Errorf("parse dropbox credentials: %w", err)
	}

	files, err := s.dropboxClient.ListNewFiles(ctx, creds, since)
	if err != nil {
		return fmt.Errorf("list dropbox files: %w", err)
	}

	for _, file := range files {
		file := file
		err := s.downloadAndAccept(ctx, ownerID, "cloud_sync", func(w io.Writer) error {
			return s.dropboxClient.Download(ctx, creds, file, w)
		})
		if err != nil {
			log.Printf("ingestsource: dropbox: ingest %q failed: %v", file.Name, err)
		}
	}
	return nil
}

// downloadAndAccept downloads into a fresh temp file (same INGEST_TMP_DIR
// the HTTP upload path uses), hashes it while writing, then hands off to
// the same Accept the manual-upload path uses. Since the scheduler runs
// inside the worker process already, a "pending" result is processed
// synchronously here instead of round-tripping through Asynq/Redis.
func (s *Service) downloadAndAccept(ctx context.Context, ownerID uuid.UUID, sourceType string, download func(io.Writer) error) error {
	if err := os.MkdirAll(s.tmpDir, 0o755); err != nil {
		return fmt.Errorf("prepare tmp dir: %w", err)
	}
	dst, err := os.CreateTemp(s.tmpDir, "ingest-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tempPath := dst.Name()

	hasher := sha256.New()
	if err := download(io.MultiWriter(dst, hasher)); err != nil {
		dst.Close()
		_ = os.Remove(tempPath)
		return fmt.Errorf("download: %w", err)
	}
	dst.Close()
	checksum := hex.EncodeToString(hasher.Sum(nil))

	job, err := s.ingest.Accept(ctx, ownerID, sourceType, tempPath, checksum)
	if err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("accept: %w", err)
	}
	if job.Status == "pending" {
		if err := s.ingest.Process(ctx, job.ID); err != nil {
			return fmt.Errorf("process: %w", err)
		}
	}
	return nil
}

func fromRow(row sqlc.IngestSourceConnection) *Connection {
	email := ""
	if row.AccountEmail != nil {
		email = *row.AccountEmail
	}
	var lastSyncedAt *time.Time
	if row.LastSyncedAt.Valid {
		lastSyncedAt = &row.LastSyncedAt.Time
	}
	return &Connection{
		ID:           fromPgUUID(row.ID),
		Provider:     row.Provider,
		Label:        row.Label,
		AccountEmail: email,
		IsActive:     row.IsActive,
		LastSyncedAt: lastSyncedAt,
	}
}
