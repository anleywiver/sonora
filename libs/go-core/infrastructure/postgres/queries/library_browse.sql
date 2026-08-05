-- Sprint 14 sisipan — Browse Library (docs/screens-spec.md, ADR 0011):
-- ALL of the catalog (not just favorites), with search + a simple sort
-- toggle. No cursor pagination here (unlike most other list endpoints in
-- this API) — a personal-scale library is realistically hundreds of
-- items, not enough to need keyset pagination per sort mode; a flat
-- LIMIT is simpler and correct at this scale.

-- name: ListLibrarySongs :many
SELECT s.id, s.title, s.duration_ms, ar.name AS artist_name, al.title AS album_title
FROM songs s
JOIN artists ar ON ar.id = s.artist_id
LEFT JOIN albums al ON al.id = s.album_id
WHERE s.deleted_at IS NULL
  AND (sqlc.narg('search')::text IS NULL OR s.title ILIKE '%' || sqlc.narg('search')::text || '%' OR ar.name ILIKE '%' || sqlc.narg('search')::text || '%')
ORDER BY
  CASE WHEN sqlc.arg('sort_alpha')::bool THEN s.title END ASC,
  CASE WHEN NOT sqlc.arg('sort_alpha')::bool THEN s.created_at END DESC
LIMIT sqlc.arg('limit_count');

-- name: ListLibraryAlbums :many
SELECT al.id, al.title, al.cover_url, ar.name AS artist_name
FROM albums al
JOIN artists ar ON ar.id = al.artist_id
WHERE (sqlc.narg('search')::text IS NULL OR al.title ILIKE '%' || sqlc.narg('search')::text || '%' OR ar.name ILIKE '%' || sqlc.narg('search')::text || '%')
ORDER BY
  CASE WHEN sqlc.arg('sort_alpha')::bool THEN al.title END ASC,
  CASE WHEN NOT sqlc.arg('sort_alpha')::bool THEN al.created_at END DESC
LIMIT sqlc.arg('limit_count');

-- name: ListLibraryArtists :many
SELECT id, name, image_url
FROM artists
WHERE (sqlc.narg('search')::text IS NULL OR name ILIKE '%' || sqlc.narg('search')::text || '%')
ORDER BY
  CASE WHEN sqlc.arg('sort_alpha')::bool THEN name END ASC,
  CASE WHEN NOT sqlc.arg('sort_alpha')::bool THEN created_at END DESC
LIMIT sqlc.arg('limit_count');
