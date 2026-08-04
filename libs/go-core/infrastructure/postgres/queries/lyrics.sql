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
