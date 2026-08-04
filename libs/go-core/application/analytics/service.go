// Package analytics backs the admin Analytics page (Sprint 11,
// docs/screens-spec.md #21): top played songs and monthly storage growth.
package analytics

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"sonora.dev/go-core/infrastructure/postgres/sqlc"
)

type TopPlayedSong struct {
	SongID     string
	Title      string
	ArtistName string
	PlayCount  int64
}

type StorageGrowthPoint struct {
	Month      time.Time
	TotalBytes int64
}

type Service struct {
	q *sqlc.Queries
}

func NewService(q *sqlc.Queries) *Service {
	return &Service{q: q}
}

func (s *Service) TopPlayed(ctx context.Context, limit int32) ([]TopPlayedSong, error) {
	rows, err := s.q.GetTopPlayedSongs(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("analytics: top played: %w", err)
	}
	out := make([]TopPlayedSong, 0, len(rows))
	for _, row := range rows {
		out = append(out, TopPlayedSong{
			SongID:     uuid.UUID(row.ID.Bytes).String(),
			Title:      row.Title,
			ArtistName: row.ArtistName,
			PlayCount:  row.PlayCount,
		})
	}
	return out, nil
}

func (s *Service) StorageGrowth(ctx context.Context) ([]StorageGrowthPoint, error) {
	rows, err := s.q.GetStorageGrowth(ctx)
	if err != nil {
		return nil, fmt.Errorf("analytics: storage growth: %w", err)
	}
	out := make([]StorageGrowthPoint, 0, len(rows))
	for _, row := range rows {
		out = append(out, StorageGrowthPoint{Month: row.Month.Time, TotalBytes: row.TotalBytes})
	}
	return out, nil
}
