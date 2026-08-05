-- name: GetLyricsBySongID :one
SELECT * FROM lyrics WHERE song_id = $1;

-- name: CreateLyrics :one
INSERT INTO lyrics (id, song_id, provider_id, synced_content, plain_content)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetLyricsProviderByName :one
SELECT * FROM lyrics_providers WHERE name = $1;

-- name: CreateLyricsProvider :one
INSERT INTO lyrics_providers (id, name, base_url)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListLyricsProviders :many
SELECT * FROM lyrics_providers ORDER BY priority ASC, name ASC;

-- name: UpdateLyricsProviderPriority :exec
UPDATE lyrics_providers SET priority = $2, updated_at = now() WHERE id = $1;

-- name: SetLyricsProviderEnabled :exec
UPDATE lyrics_providers SET is_enabled = $2, updated_at = now() WHERE id = $1;

-- name: RecordLyricsLookup :exec
-- Sprint 14: every GetLyrics attempt for a provider increments
-- total_lookups; matched only increments successful_matches too — this is
-- what makes "Match Rate" on the admin Lyrics Source page real instead of
-- fabricated (a miss previously left no trace at all).
UPDATE lyrics_providers SET
  total_lookups = total_lookups + 1,
  successful_matches = successful_matches + (CASE WHEN sqlc.arg('matched')::bool THEN 1 ELSE 0 END),
  updated_at = now()
WHERE id = $1;

-- name: SetLyricsProviderHealth :exec
UPDATE lyrics_providers SET health_status = $2, updated_at = now() WHERE id = $1;
