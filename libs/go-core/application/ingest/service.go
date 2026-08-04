// Package ingest implements the upload -> checksum-dedup -> storage ->
// metadata pipeline (Sprint 3). Accept runs synchronously in the HTTP
// request (fast: hash + one DB row) — Process runs in the worker (slow:
// Drive upload + ffprobe), decoupled via an Asynq task carrying only the
// job ID, since the temp file path is already persisted on the row.
package ingest

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"sonora.dev/go-core/infrastructure/crypto"
	"sonora.dev/go-core/infrastructure/mediainfo"
	"sonora.dev/go-core/infrastructure/meilisearch"
	"sonora.dev/go-core/infrastructure/postgres/sqlc"
	"sonora.dev/go-core/infrastructure/storage"
)

type Service struct {
	q                  *sqlc.Queries
	box                *crypto.Box
	search             *meilisearch.Client
	googleClientID     string
	googleClientSecret string
}

func NewService(q *sqlc.Queries, box *crypto.Box, search *meilisearch.Client, googleClientID, googleClientSecret string) *Service {
	return &Service{q: q, box: box, search: search, googleClientID: googleClientID, googleClientSecret: googleClientSecret}
}

// Accept records an uploaded file as an ingest job. If a song with this
// checksum already exists, the job completes immediately (dedup) and the
// temp file is discarded — no background processing needed. The caller
// (HTTP handler) should only enqueue the Asynq "ingest:process" task when
// the returned job's Status is "pending".
func (s *Service) Accept(ctx context.Context, userID uuid.UUID, tempPath, checksum string) (*Job, error) {
	if existing, err := s.q.GetSongByChecksum(ctx, checksum); err == nil {
		_ = os.Remove(tempPath)
		return s.createJobRow(ctx, userID, "completed", existing.ID, nil)
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("ingest: check checksum: %w", err)
	}

	return s.createJobRow(ctx, userID, "pending", pgtype.UUID{}, &tempPath)
}

func (s *Service) createJobRow(ctx context.Context, userID uuid.UUID, status string, songID pgtype.UUID, tempPath *string) (*Job, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("ingest: generate job id: %w", err)
	}
	row, err := s.q.CreateIngestJob(ctx, sqlc.CreateIngestJobParams{
		ID:         toPgUUID(id),
		UserID:     toPgUUID(userID),
		SourceType: "manual_upload",
		Status:     status,
		SongID:     songID,
		TempPath:   tempPath,
	})
	if err != nil {
		return nil, fmt.Errorf("ingest: create job: %w", err)
	}
	return jobFromRow(row), nil
}

func (s *Service) GetJob(ctx context.Context, jobID, userID uuid.UUID) (*Job, error) {
	row, err := s.q.GetIngestJobByID(ctx, toPgUUID(jobID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("ingest: get job: %w", err)
	}
	if fromPgUUID(row.UserID) != userID {
		return nil, ErrNotFound
	}
	return jobFromRow(row), nil
}

// ListJobs returns up to limit jobs plus an opaque next-cursor and
// has_more flag, per the API's cursor-pagination convention.
func (s *Service) ListJobs(ctx context.Context, userID uuid.UUID, status, cursor string, limit int32) ([]*Job, string, bool, error) {
	params := sqlc.ListIngestJobsByUserParams{
		UserID:     toPgUUID(userID),
		LimitCount: limit + 1,
	}
	if status != "" {
		params.Status = &status
	}
	if cursor != "" {
		createdAt, id, err := decodeCursor(cursor)
		if err != nil {
			return nil, "", false, err
		}
		params.CursorCreatedAt = pgtype.Timestamptz{Time: createdAt, Valid: true}
		params.CursorID = toPgUUID(id)
	}

	rows, err := s.q.ListIngestJobsByUser(ctx, params)
	if err != nil {
		return nil, "", false, fmt.Errorf("ingest: list jobs: %w", err)
	}

	hasMore := len(rows) > int(limit)
	if hasMore {
		rows = rows[:limit]
	}

	jobs := make([]*Job, 0, len(rows))
	for _, row := range rows {
		jobs = append(jobs, jobFromRow(row))
	}

	nextCursor := ""
	if hasMore && len(jobs) > 0 {
		last := jobs[len(jobs)-1]
		nextCursor = encodeCursor(last.CreatedAt, last.ID)
	}
	return jobs, nextCursor, hasMore, nil
}

// RetryJob resets a failed job back to pending so the worker can re-run it
// against the same temp file. The caller (HTTP handler) must re-enqueue
// the Asynq task after this succeeds.
func (s *Service) RetryJob(ctx context.Context, jobID, userID uuid.UUID) (*Job, error) {
	row, err := s.q.GetIngestJobByID(ctx, toPgUUID(jobID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("ingest: get job: %w", err)
	}
	if fromPgUUID(row.UserID) != userID {
		return nil, ErrNotFound
	}
	if row.Status != "failed" || row.TempPath == nil {
		return nil, ErrNotRetryable
	}
	if err := s.q.ResetIngestJobToPending(ctx, toPgUUID(jobID)); err != nil {
		return nil, fmt.Errorf("ingest: reset job: %w", err)
	}
	row.Status = "pending"
	return jobFromRow(row), nil
}

func (s *Service) DeleteJob(ctx context.Context, jobID, userID uuid.UUID) error {
	row, err := s.q.GetIngestJobByID(ctx, toPgUUID(jobID))
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("ingest: get job: %w", err)
	}
	if fromPgUUID(row.UserID) != userID {
		return ErrNotFound
	}

	affected, err := s.q.DeleteIngestJob(ctx, sqlc.DeleteIngestJobParams{ID: toPgUUID(jobID), UserID: toPgUUID(userID)})
	if err != nil {
		return fmt.Errorf("ingest: delete job: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	if row.TempPath != nil {
		_ = os.Remove(*row.TempPath)
	}
	return nil
}

// Process runs the slow part of the pipeline: upload to storage, extract
// metadata, and create the catalog rows. Called by the worker's Asynq task
// handler only — never from an HTTP request.
func (s *Service) Process(ctx context.Context, jobID uuid.UUID) error {
	row, err := s.q.GetIngestJobByID(ctx, toPgUUID(jobID))
	if err != nil {
		return fmt.Errorf("ingest: load job: %w", err)
	}
	if row.TempPath == nil {
		return fmt.Errorf("ingest: job %s has no temp file to process", jobID)
	}
	tempPath := *row.TempPath

	fail := func(cause error) error {
		msg := cause.Error()
		if updErr := s.q.FailIngestJob(ctx, sqlc.FailIngestJobParams{ID: toPgUUID(jobID), ErrorMessage: &msg}); updErr != nil {
			return fmt.Errorf("ingest: mark job failed (%v) after: %w", updErr, cause)
		}
		return cause
	}

	if err := s.q.MarkIngestJobProcessing(ctx, toPgUUID(jobID)); err != nil {
		return fmt.Errorf("ingest: mark processing: %w", err)
	}

	account, err := s.q.GetActiveStorageAccount(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fail(ErrNoActiveStorageAccount)
		}
		return fail(fmt.Errorf("ingest: load storage account: %w", err))
	}

	refreshToken, err := s.box.Decrypt(account.CredentialsEncrypted)
	if err != nil {
		return fail(fmt.Errorf("ingest: decrypt storage credentials: %w", err))
	}

	info, err := mediainfo.Probe(ctx, tempPath)
	if err != nil {
		return fail(fmt.Errorf("ingest: probe media: %w", err))
	}

	stat, err := os.Stat(tempPath)
	if err != nil {
		return fail(fmt.Errorf("ingest: stat temp file: %w", err))
	}

	// Re-checksum from disk rather than trusting a value from earlier in
	// the pipeline — this is the "checksum validation right before storage
	// relay" step from ADR 0001, and it's robust to a worker restart
	// between accept and process too.
	checksum, err := sha256File(tempPath)
	if err != nil {
		return fail(fmt.Errorf("ingest: checksum temp file: %w", err))
	}

	storageFile, err := s.q.GetStorageFileByChecksum(ctx, checksum)
	if errors.Is(err, pgx.ErrNoRows) {
		storageFile, err = s.uploadToStorage(ctx, account, refreshToken, tempPath, checksum, stat.Size())
		if err != nil {
			return fail(err)
		}
	} else if err != nil {
		return fail(fmt.Errorf("ingest: check storage file: %w", err))
	}

	artistName := info.Artist
	if artistName == "" {
		artistName = "Unknown Artist"
	}
	artist, err := s.findOrCreateArtist(ctx, artistName)
	if err != nil {
		return fail(err)
	}

	var albumID pgtype.UUID
	if info.Album != "" {
		album, err := s.findOrCreateAlbum(ctx, artist.ID, info.Album)
		if err != nil {
			return fail(err)
		}
		albumID = album.ID
	}

	title := info.Title
	if title == "" {
		title = originalFilenameFrom(tempPath)
	}

	var trackNumber *int32
	if info.TrackNumber != nil {
		n := int32(*info.TrackNumber)
		trackNumber = &n
	}

	songID, err := uuid.NewV7()
	if err != nil {
		return fail(fmt.Errorf("ingest: generate song id: %w", err))
	}
	song, err := s.q.CreateSong(ctx, sqlc.CreateSongParams{
		ID:            toPgUUID(songID),
		AlbumID:       albumID,
		ArtistID:      artist.ID,
		StorageFileID: storageFile.ID,
		Title:         title,
		DurationMs:    int32(info.DurationMs),
		TrackNumber:   trackNumber,
		Checksum:      checksum,
	})
	if err != nil {
		return fail(fmt.Errorf("ingest: create song: %w", err))
	}

	if err := s.q.CompleteIngestJob(ctx, sqlc.CompleteIngestJobParams{ID: toPgUUID(jobID), SongID: song.ID}); err != nil {
		return fmt.Errorf("ingest: complete job: %w", err)
	}

	// The song row is the source of truth and is already durable — a search
	// index hiccup shouldn't fail a completed ingest job.
	if indexErr := s.indexSong(ctx, song, artist, info.Album); indexErr != nil {
		log.Printf("ingest: index song %s in search: %v", jobID, indexErr)
	}

	_ = os.Remove(tempPath)
	return nil
}

func (s *Service) indexSong(ctx context.Context, song sqlc.Song, artist sqlc.Artist, albumTitle string) error {
	doc := meilisearch.SongDocument{
		ID:         fromPgUUID(song.ID).String(),
		Title:      song.Title,
		ArtistID:   fromPgUUID(song.ArtistID).String(),
		ArtistName: artist.Name,
		AlbumTitle: albumTitle,
		DurationMs: int(song.DurationMs),
	}
	if song.AlbumID.Valid {
		doc.AlbumID = fromPgUUID(song.AlbumID).String()
	}
	return s.search.IndexSong(ctx, doc)
}

func (s *Service) uploadToStorage(ctx context.Context, account sqlc.StorageAccount, refreshToken, tempPath, checksum string, sizeBytes int64) (sqlc.StorageFile, error) {
	mimeType, err := detectMimeType(tempPath)
	if err != nil {
		return sqlc.StorageFile{}, fmt.Errorf("ingest: detect mime type: %w", err)
	}

	f, err := os.Open(tempPath)
	if err != nil {
		return sqlc.StorageFile{}, fmt.Errorf("ingest: open temp file: %w", err)
	}
	defer f.Close()

	provider := storage.NewGoogleDriveProvider(ctx, s.googleClientID, s.googleClientSecret, refreshToken)
	providerFileID, err := provider.Upload(ctx, filepath.Base(tempPath), mimeType, f)
	if err != nil {
		return sqlc.StorageFile{}, fmt.Errorf("ingest: upload to drive: %w", err)
	}

	fileID, err := uuid.NewV7()
	if err != nil {
		return sqlc.StorageFile{}, fmt.Errorf("ingest: generate storage file id: %w", err)
	}
	storageFile, err := s.q.CreateStorageFile(ctx, sqlc.CreateStorageFileParams{
		ID:               toPgUUID(fileID),
		StorageAccountID: account.ID,
		ProviderFileID:   providerFileID,
		Checksum:         checksum,
		SizeBytes:        sizeBytes,
		MimeType:         mimeType,
	})
	if err != nil {
		return sqlc.StorageFile{}, fmt.Errorf("ingest: record storage file: %w", err)
	}

	if err := s.q.IncrementStorageAccountUsedBytes(ctx, sqlc.IncrementStorageAccountUsedBytesParams{
		ID:        account.ID,
		UsedBytes: sizeBytes,
	}); err != nil {
		return sqlc.StorageFile{}, fmt.Errorf("ingest: update storage usage: %w", err)
	}
	return storageFile, nil
}
