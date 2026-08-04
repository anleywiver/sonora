-- name: GetIngestSourceConnectionByID :one
SELECT * FROM ingest_source_connections WHERE id = $1;

-- name: ListIngestSourceConnections :many
SELECT * FROM ingest_source_connections ORDER BY created_at ASC;

-- name: ListActiveIngestSourceConnections :many
SELECT * FROM ingest_source_connections WHERE is_active = true ORDER BY created_at ASC;

-- name: CreateIngestSourceConnection :one
INSERT INTO ingest_source_connections (id, provider, label, account_email, credentials_encrypted)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateIngestSourceConnectionSync :exec
UPDATE ingest_source_connections SET last_synced_at = $2, updated_at = now() WHERE id = $1;

-- name: DeleteIngestSourceConnection :execrows
DELETE FROM ingest_source_connections WHERE id = $1;
