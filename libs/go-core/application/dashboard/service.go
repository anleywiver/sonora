// Package dashboard backs the admin Dashboard page (Sprint 14,
// docs/screens-spec.md #16): the 4 stat cards, storage distribution, and
// background jobs summary. Top Played reuses application/analytics.
package dashboard

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"sonora.dev/go-core/infrastructure/postgres/sqlc"
)

type Stats struct {
	TotalSongs        int64
	TotalUsers        int64
	TotalDrives       int64
	TotalStorageBytes int64
}

type StorageAccountUsage struct {
	ID         uuid.UUID
	Label      string
	UsedBytes  int64
	QuotaBytes *int64
}

type Service struct {
	q *sqlc.Queries
}

func NewService(q *sqlc.Queries) *Service {
	return &Service{q: q}
}

func (s *Service) GetStats(ctx context.Context) (*Stats, error) {
	row, err := s.q.GetDashboardStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("dashboard: stats: %w", err)
	}
	return &Stats{
		TotalSongs:        row.TotalSongs,
		TotalUsers:        row.TotalUsers,
		TotalDrives:       row.TotalDrives,
		TotalStorageBytes: row.TotalStorageBytes,
	}, nil
}

func (s *Service) StorageDistribution(ctx context.Context) ([]StorageAccountUsage, error) {
	rows, err := s.q.ListStorageDistribution(ctx)
	if err != nil {
		return nil, fmt.Errorf("dashboard: storage distribution: %w", err)
	}
	out := make([]StorageAccountUsage, 0, len(rows))
	for _, row := range rows {
		out = append(out, StorageAccountUsage{
			ID:         uuid.UUID(row.ID.Bytes),
			Label:      row.Label,
			UsedBytes:  row.UsedBytes,
			QuotaBytes: row.QuotaBytes,
		})
	}
	return out, nil
}

// BackgroundJobsSummary maps ingest_jobs.status -> count.
func (s *Service) BackgroundJobsSummary(ctx context.Context) (map[string]int64, error) {
	rows, err := s.q.GetBackgroundJobsSummary(ctx)
	if err != nil {
		return nil, fmt.Errorf("dashboard: background jobs summary: %w", err)
	}
	out := make(map[string]int64, len(rows))
	for _, row := range rows {
		out[row.Status] = row.JobCount
	}
	return out, nil
}
