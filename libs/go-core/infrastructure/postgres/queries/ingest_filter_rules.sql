-- name: ListIngestFilterRules :many
SELECT * FROM ingest_filter_rules WHERE source_type = $1 ORDER BY rule_type ASC, created_at ASC;

-- name: CreateIngestFilterRule :one
INSERT INTO ingest_filter_rules (id, source_type, rule_type, value)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: DeleteIngestFilterRule :execrows
DELETE FROM ingest_filter_rules WHERE id = $1 AND source_type = $2;
