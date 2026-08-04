// Package playback implements the per-user playback_states row — read and
// write only (Sprint 7). Active Device authority (only the active device
// may push state, remote control from other devices, Transfer Playback)
// is Sprint 8; this package doesn't enforce who is allowed to call it.
package playback

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"sonora.dev/go-core/infrastructure/postgres/sqlc"
)

type State struct {
	UserID         uuid.UUID
	ActiveDeviceID *uuid.UUID
	CurrentSongID  *uuid.UUID
	PositionMs     int
	IsPlaying      bool
}

type Service struct {
	q *sqlc.Queries
}

func NewService(q *sqlc.Queries) *Service {
	return &Service{q: q}
}

// GetState returns pgx.ErrNoRows (unwrapped) if the user has never had a
// playback state recorded — callers should treat that as "nothing playing"
// rather than an error.
func (s *Service) GetState(ctx context.Context, userID uuid.UUID) (*State, error) {
	row, err := s.q.GetPlaybackState(ctx, toPgUUID(userID))
	if err != nil {
		return nil, err
	}
	return &State{
		UserID:         userID,
		ActiveDeviceID: fromPgUUIDPtr(row.ActiveDeviceID),
		CurrentSongID:  fromPgUUIDPtr(row.CurrentSongID),
		PositionMs:     int(row.PositionMs),
		IsPlaying:      row.IsPlaying,
	}, nil
}

func (s *Service) UpsertState(ctx context.Context, userID uuid.UUID, activeDeviceID, currentSongID *uuid.UUID, positionMs int, isPlaying bool) (*State, error) {
	params := sqlc.UpsertPlaybackStateParams{
		UserID:     toPgUUID(userID),
		PositionMs: int32(positionMs),
		IsPlaying:  isPlaying,
	}
	if activeDeviceID != nil {
		params.ActiveDeviceID = toPgUUID(*activeDeviceID)
	}
	if currentSongID != nil {
		params.CurrentSongID = toPgUUID(*currentSongID)
	}

	row, err := s.q.UpsertPlaybackState(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("playback: upsert state: %w", err)
	}
	return &State{
		UserID:         userID,
		ActiveDeviceID: fromPgUUIDPtr(row.ActiveDeviceID),
		CurrentSongID:  fromPgUUIDPtr(row.CurrentSongID),
		PositionMs:     int(row.PositionMs),
		IsPlaying:      row.IsPlaying,
	}, nil
}
