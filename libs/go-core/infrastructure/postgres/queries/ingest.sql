-- name: CreateIngestJob :one
INSERT INTO ingest_jobs (id, user_id, source_type, status, song_id, temp_path)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetIngestJobByID :one
SELECT * FROM ingest_jobs WHERE id = $1;

-- name: ListIngestJobsByUser :many
SELECT * FROM ingest_jobs
WHERE user_id = $1
  AND (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status'))
  AND (
    sqlc.narg('cursor_created_at')::timestamptz IS NULL
    OR (created_at, id) < (sqlc.narg('cursor_created_at')::timestamptz, sqlc.narg('cursor_id')::uuid)
  )
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg('limit_count');

-- name: CompleteIngestJob :exec
UPDATE ingest_jobs SET status = 'completed', song_id = $2, temp_path = NULL, updated_at = now() WHERE id = $1;

-- name: FailIngestJob :exec
UPDATE ingest_jobs SET status = 'failed', error_message = $2, updated_at = now() WHERE id = $1;

-- name: SkipIngestJobByFilter :exec
-- Sprint 14 sisipan (ADR 0008) — terminal state distinct from 'failed':
-- the item was never actually broken, it just didn't match a configured
-- filter rule for its source. temp_path is cleared same as a completed
-- job (ADR 0008: skipped items shouldn't linger as temp files either).
UPDATE ingest_jobs SET status = 'skipped_by_filter', error_message = $2, temp_path = NULL, updated_at = now() WHERE id = $1;

-- name: MarkIngestJobProcessing :exec
UPDATE ingest_jobs SET status = 'processing', updated_at = now() WHERE id = $1;

-- name: ResetIngestJobToPending :exec
UPDATE ingest_jobs SET status = 'pending', error_message = NULL, updated_at = now() WHERE id = $1;

-- name: DeleteIngestJob :execrows
DELETE FROM ingest_jobs WHERE id = $1 AND user_id = $2;

-- name: ListAllIngestJobs :many
-- Sprint 14 admin Job Queue (docs/screens-spec.md #20) — same cursor
-- pattern as ListIngestJobsByUser, minus the per-user scoping.
SELECT * FROM ingest_jobs
WHERE (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status'))
  AND (
    sqlc.narg('cursor_created_at')::timestamptz IS NULL
    OR (created_at, id) < (sqlc.narg('cursor_created_at')::timestamptz, sqlc.narg('cursor_id')::uuid)
  )
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg('limit_count');

-- name: ListCompletedIngestJobsWithTempPath :many
-- Sprint 10 garbage collector target: completed jobs still holding a
-- temp_path. Deliberately excludes 'failed' jobs — RetryJob needs that
-- file to still be on disk.
SELECT id, temp_path FROM ingest_jobs WHERE status = 'completed' AND temp_path IS NOT NULL LIMIT $1;

-- name: ClearIngestJobTempPath :exec
UPDATE ingest_jobs SET temp_path = NULL, updated_at = now() WHERE id = $1;
