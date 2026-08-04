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
