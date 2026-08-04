-- name: GetAlbumByArtistAndTitle :one
SELECT * FROM albums WHERE artist_id = $1 AND title = $2;

-- name: GetAlbumByID :one
SELECT * FROM albums WHERE id = $1;

-- name: CreateAlbum :one
INSERT INTO albums (id, artist_id, title, cover_url, released_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;
