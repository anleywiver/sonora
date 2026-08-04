-- name: GetAlbumByArtistAndTitle :one
SELECT * FROM albums WHERE artist_id = $1 AND title = $2;

-- name: GetAlbumByID :one
SELECT * FROM albums WHERE id = $1;

-- name: GetAlbumDetail :one
SELECT al.id, al.title, al.cover_url, al.released_at, al.artist_id, ar.name AS artist_name
FROM albums al
JOIN artists ar ON ar.id = al.artist_id
WHERE al.id = $1;

-- name: ListAlbumsByArtist :many
SELECT * FROM albums WHERE artist_id = $1 ORDER BY released_at DESC NULLS LAST, title ASC;

-- name: CreateAlbum :one
INSERT INTO albums (id, artist_id, title, cover_url, released_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;
