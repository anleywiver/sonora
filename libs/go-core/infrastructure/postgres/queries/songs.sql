-- name: GetSongByID :one
SELECT * FROM songs WHERE id = $1;

-- name: GetSongByChecksum :one
-- Used by the ingest pipeline for dedup before touching storage.
SELECT * FROM songs WHERE checksum = $1;

-- name: ListSongsByAlbum :many
SELECT * FROM songs WHERE album_id = $1 ORDER BY track_number ASC NULLS LAST, title ASC;

-- name: ListSongsByArtist :many
SELECT * FROM songs
WHERE artist_id = $1
ORDER BY created_at DESC, id DESC
LIMIT $2;

-- name: CreateSong :one
INSERT INTO songs (id, album_id, artist_id, storage_file_id, title, duration_ms, track_number, checksum)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: DeleteSong :exec
DELETE FROM songs WHERE id = $1;

-- name: ListRecentSongs :many
-- Stand-in for "trending" until play_history exists (Sprint 6) to compute
-- real play-count trends.
SELECT s.id, s.title, s.artist_id, ar.name AS artist_name, s.album_id, s.duration_ms
FROM songs s
JOIN artists ar ON ar.id = s.artist_id
ORDER BY s.created_at DESC
LIMIT $1;

-- name: GetSongDetail :one
-- Joined view used by the song detail endpoint and by the stream handler
-- (needs storage_file_id/provider_file_id/storage_account_id to fetch the
-- bytes from the storage provider).
SELECT
  s.id, s.title, s.duration_ms, s.track_number, s.checksum,
  s.album_id, al.title AS album_title, al.cover_url AS album_cover_url,
  s.artist_id, ar.name AS artist_name,
  s.storage_file_id, sf.provider_file_id, sf.mime_type, sf.size_bytes,
  sf.storage_account_id
FROM songs s
JOIN artists ar ON ar.id = s.artist_id
LEFT JOIN albums al ON al.id = s.album_id
JOIN storage_files sf ON sf.id = s.storage_file_id
WHERE s.id = $1;
