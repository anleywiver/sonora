-- name: GetArtistByName :one
SELECT * FROM artists WHERE name = $1;

-- name: GetArtistByID :one
SELECT * FROM artists WHERE id = $1;

-- name: CreateArtist :one
INSERT INTO artists (id, name, image_url) VALUES ($1, $2, $3) RETURNING *;
