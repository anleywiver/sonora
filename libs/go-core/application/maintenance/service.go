// Package maintenance implements Sprint 10's two scheduled jobs: a daily
// garbage collector and a weekly storage optimizer. See
// docs/decisions/0004-sprint10-scheduled-jobs-and-ingest-sources.md.
package maintenance

import (
	"context"
	"log"
	"os"
	"time"

	"sonora.dev/go-core/application/storageaccount"
	"sonora.dev/go-core/domain/identity"
	"sonora.dev/go-core/infrastructure/postgres/sqlc"
)

const gcBatchSize = 500

type Service struct {
	q             *sqlc.Queries
	refreshTokens identity.RefreshTokenRepository
	storage       *storageaccount.Service
}

func NewService(q *sqlc.Queries, refreshTokens identity.RefreshTokenRepository, storage *storageaccount.Service) *Service {
	return &Service{q: q, refreshTokens: refreshTokens, storage: storage}
}

// GarbageCollect deletes on-disk temp files for completed ingest jobs
// (already uploaded to storage, never needed again) and purges expired
// refresh_tokens rows. Deliberately does NOT touch failed jobs' temp
// files — RetryJob needs those to still exist.
func (s *Service) GarbageCollect(ctx context.Context) {
	rows, err := s.q.ListCompletedIngestJobsWithTempPath(ctx, gcBatchSize)
	if err != nil {
		log.Printf("maintenance: gc: list completed jobs: %v", err)
	} else {
		removed := 0
		for _, row := range rows {
			if row.TempPath == nil {
				continue
			}
			if err := os.Remove(*row.TempPath); err != nil && !os.IsNotExist(err) {
				log.Printf("maintenance: gc: remove temp file %s: %v", *row.TempPath, err)
				continue
			}
			if err := s.q.ClearIngestJobTempPath(ctx, row.ID); err != nil {
				log.Printf("maintenance: gc: clear temp_path for job %v: %v", row.ID, err)
				continue
			}
			removed++
		}
		log.Printf("maintenance: gc: removed %d completed ingest temp files", removed)
	}

	deleted, err := s.refreshTokens.DeleteExpired(ctx, time.Now())
	if err != nil {
		log.Printf("maintenance: gc: purge expired refresh tokens: %v", err)
		return
	}
	log.Printf("maintenance: gc: purged %d expired refresh tokens", deleted)
}

// StorageOptimize refreshes health/quota data for every active storage
// account, keeping application/ingest's quota-aware routing fed with
// numbers that aren't stale between manual admin health-checks.
func (s *Service) StorageOptimize(ctx context.Context) {
	checked, failed := s.storage.RunHealthChecks(ctx)
	log.Printf("maintenance: storage optimizer: checked %d accounts, %d failed", checked, failed)
}
