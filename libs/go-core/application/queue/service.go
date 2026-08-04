// Package queue implements the per-user play queue (queue_items) — plain
// CRUD, no Active Device / cross-device sync yet (that's Sprint 7/8's
// WebSocket + playback_states work).
package queue

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	appcatalog "sonora.dev/go-core/application/catalog"
	"sonora.dev/go-core/infrastructure/postgres/sqlc"
)

var ErrNotFound = errors.New("queue: not found")

const positionStep = 1000

type Item struct {
	ID       uuid.UUID
	SongID   uuid.UUID
	Position float64
	Song     *appcatalog.SongDetail
}

type Service struct {
	q       *sqlc.Queries
	catalog *appcatalog.Service
}

func NewService(q *sqlc.Queries, catalog *appcatalog.Service) *Service {
	return &Service{q: q, catalog: catalog}
}

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]*Item, error) {
	rows, err := s.q.ListQueueByUser(ctx, toPgUUID(userID))
	if err != nil {
		return nil, fmt.Errorf("queue: list: %w", err)
	}
	out := make([]*Item, 0, len(rows))
	for _, row := range rows {
		song, err := s.catalog.GetSong(ctx, fromPgUUID(row.SongID))
		if err != nil {
			continue // song was deleted from the catalog since being queued
		}
		out = append(out, &Item{ID: fromPgUUID(row.ID), SongID: fromPgUUID(row.SongID), Position: row.Position, Song: song})
	}
	return out, nil
}

func (s *Service) Add(ctx context.Context, userID, songID uuid.UUID) (*Item, error) {
	song, err := s.catalog.GetSong(ctx, songID)
	if err != nil {
		return nil, err
	}

	rows, err := s.q.ListQueueByUser(ctx, toPgUUID(userID))
	if err != nil {
		return nil, fmt.Errorf("queue: list for position: %w", err)
	}
	maxPos := 0.0
	for _, row := range rows {
		if row.Position > maxPos {
			maxPos = row.Position
		}
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("queue: generate id: %w", err)
	}
	row, err := s.q.AddQueueItem(ctx, sqlc.AddQueueItemParams{
		ID:       toPgUUID(id),
		UserID:   toPgUUID(userID),
		SongID:   toPgUUID(songID),
		Position: maxPos + positionStep,
	})
	if err != nil {
		return nil, fmt.Errorf("queue: add: %w", err)
	}
	return &Item{ID: fromPgUUID(row.ID), SongID: fromPgUUID(row.SongID), Position: row.Position, Song: song}, nil
}

func (s *Service) UpdatePosition(ctx context.Context, userID, itemID uuid.UUID, position float64) error {
	affected, err := s.q.UpdateQueueItemPosition(ctx, sqlc.UpdateQueueItemPositionParams{
		ID:       toPgUUID(itemID),
		UserID:   toPgUUID(userID),
		Position: position,
	})
	if err != nil {
		return fmt.Errorf("queue: update position: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Service) Remove(ctx context.Context, userID, itemID uuid.UUID) error {
	if err := s.q.RemoveQueueItem(ctx, sqlc.RemoveQueueItemParams{ID: toPgUUID(itemID), UserID: toPgUUID(userID)}); err != nil {
		return fmt.Errorf("queue: remove: %w", err)
	}
	return nil
}

func (s *Service) Clear(ctx context.Context, userID uuid.UUID) error {
	return s.q.ClearQueue(ctx, toPgUUID(userID))
}
