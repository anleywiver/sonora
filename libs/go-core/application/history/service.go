// Package history implements play_history recording, listing, and the
// Continue Listening query — songs recently played but not finished.
package history

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"sonora.dev/go-core/infrastructure/postgres/sqlc"
)

type Entry struct {
	ID         uuid.UUID
	SongID     uuid.UUID
	ProgressMs int
	PlayedAt   time.Time
}

type ContinueListeningEntry struct {
	SongID     uuid.UUID
	ProgressMs int
	DurationMs int
	PlayedAt   time.Time
}

type Service struct {
	q *sqlc.Queries
}

func NewService(q *sqlc.Queries) *Service {
	return &Service{q: q}
}

func (s *Service) RecordPlay(ctx context.Context, userID, songID uuid.UUID, progressMs int) (*Entry, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("history: generate id: %w", err)
	}
	row, err := s.q.CreateHistoryEntry(ctx, sqlc.CreateHistoryEntryParams{
		ID:         toPgUUID(id),
		UserID:     toPgUUID(userID),
		SongID:     toPgUUID(songID),
		ProgressMs: int32(progressMs),
	})
	if err != nil {
		return nil, fmt.Errorf("history: record: %w", err)
	}
	return &Entry{
		ID:         fromPgUUID(row.ID),
		SongID:     fromPgUUID(row.SongID),
		ProgressMs: int(row.ProgressMs),
		PlayedAt:   row.PlayedAt.Time,
	}, nil
}

func (s *Service) ListHistory(ctx context.Context, userID uuid.UUID, cursor string, limit int32) ([]*Entry, string, bool, error) {
	params := sqlc.ListHistoryByUserParams{UserID: toPgUUID(userID), LimitCount: limit + 1}
	if cursor != "" {
		playedAt, id, err := decodeCursor(cursor)
		if err != nil {
			return nil, "", false, err
		}
		params.CursorPlayedAt = pgtype.Timestamptz{Time: playedAt, Valid: true}
		params.CursorID = toPgUUID(id)
	}

	rows, err := s.q.ListHistoryByUser(ctx, params)
	if err != nil {
		return nil, "", false, fmt.Errorf("history: list: %w", err)
	}

	hasMore := len(rows) > int(limit)
	if hasMore {
		rows = rows[:limit]
	}
	entries := make([]*Entry, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, &Entry{
			ID:         fromPgUUID(row.ID),
			SongID:     fromPgUUID(row.SongID),
			ProgressMs: int(row.ProgressMs),
			PlayedAt:   row.PlayedAt.Time,
		})
	}

	nextCursor := ""
	if hasMore && len(entries) > 0 {
		last := entries[len(entries)-1]
		nextCursor = encodeCursor(last.PlayedAt, last.ID)
	}
	return entries, nextCursor, hasMore, nil
}

func (s *Service) ListContinueListening(ctx context.Context, userID uuid.UUID, limit int32) ([]*ContinueListeningEntry, error) {
	rows, err := s.q.ListContinueListening(ctx, sqlc.ListContinueListeningParams{UserID: toPgUUID(userID), Limit: limit})
	if err != nil {
		return nil, fmt.Errorf("history: continue listening: %w", err)
	}
	out := make([]*ContinueListeningEntry, 0, len(rows))
	for _, row := range rows {
		out = append(out, &ContinueListeningEntry{
			SongID:     fromPgUUID(row.SongID),
			ProgressMs: int(row.ProgressMs),
			DurationMs: int(row.DurationMs),
			PlayedAt:   row.PlayedAt.Time,
		})
	}
	return out, nil
}
