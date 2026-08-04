package ingest

import (
	"time"

	"github.com/google/uuid"

	"sonora.dev/go-core/infrastructure/postgres/sqlc"
)

type Job struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	SourceType   string
	Status       string
	SongID       *uuid.UUID
	ErrorMessage *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func jobFromRow(row sqlc.IngestJob) *Job {
	return &Job{
		ID:           fromPgUUID(row.ID),
		UserID:       fromPgUUID(row.UserID),
		SourceType:   row.SourceType,
		Status:       row.Status,
		SongID:       fromPgUUIDPtr(row.SongID),
		ErrorMessage: row.ErrorMessage,
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
	}
}
