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

-- name: MarkIngestJobProcessing :exec
UPDATE ingest_jobs SET status = 'processing', updated_at = now() WHERE id = $1;

-- name: ResetIngestJobToPending :exec
UPDATE ingest_jobs SET status = 'pending', error_message = NULL, updated_at = now() WHERE id = $1;

-- name: DeleteIngestJob :execrows
DELETE FROM ingest_jobs WHERE id = $1 AND user_id = $2;
