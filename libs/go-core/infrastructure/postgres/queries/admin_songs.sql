-- name: ListAdminSongs :many
-- Sprint 14 admin Manage Songs (ADR 0010) — search by title or artist
-- name, same cursor-pagination shape used everywhere else in this API.
SELECT s.id, s.title, s.duration_ms, s.created_at,
  ar.name AS artist_name, al.title AS album_title, sa.provider AS storage_provider
FROM songs s
JOIN artists ar ON ar.id = s.artist_id
LEFT JOIN albums al ON al.id = s.album_id
JOIN storage_files sf ON sf.id = s.storage_file_id
JOIN storage_accounts sa ON sa.id = sf.storage_account_id
WHERE s.deleted_at IS NULL
  AND (
    sqlc.narg('search')::text IS NULL
    OR s.title ILIKE '%' || sqlc.narg('search')::text || '%'
    OR ar.name ILIKE '%' || sqlc.narg('search')::text || '%'
  )
  AND (
    sqlc.narg('cursor_created_at')::timestamptz IS NULL
    OR (s.created_at, s.id) < (sqlc.narg('cursor_created_at')::timestamptz, sqlc.narg('cursor_id')::uuid)
  )
ORDER BY s.created_at DESC, s.id DESC
LIMIT sqlc.arg('limit_count');

-- name: SoftDeleteSong :execrows
UPDATE songs SET deleted_at = now(), updated_at = now() WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateSongTitle :exec
UPDATE songs SET title = $2, updated_at = now() WHERE id = $1;

-- name: UpdateSongArtistAlbum :exec
UPDATE songs SET artist_id = $2, album_id = $3, updated_at = now() WHERE id = $1;

-- name: ClearSongGenres :exec
DELETE FROM song_genres WHERE song_id = $1;

-- name: AddSongGenre :exec
INSERT INTO song_genres (song_id, genre_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: GetGenreByName :one
SELECT * FROM genres WHERE name = $1;

-- name: CreateGenre :one
INSERT INTO genres (id, name) VALUES ($1, $2) RETURNING *;
