-- name: GetArtistByName :one
SELECT * FROM artists WHERE name = $1;

-- name: GetArtistByID :one
SELECT * FROM artists WHERE id = $1;

-- name: CreateArtist :one
INSERT INTO artists (id, name, image_url) VALUES ($1, $2, $3) RETURNING *;

-- name: UpdateArtistMusicbrainzID :exec
-- Only fills a still-empty musicbrainz_id — never overwrites a value set
-- by an earlier, possibly more specific match.
UPDATE artists SET musicbrainz_id = $2, updated_at = now()
WHERE id = $1 AND musicbrainz_id IS NULL;
