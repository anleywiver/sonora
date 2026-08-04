-- name: CreateHistoryEntry :one
INSERT INTO play_history (id, user_id, song_id, progress_ms, played_at)
VALUES ($1, $2, $3, $4, now())
RETURNING *;

-- name: ListHistoryByUser :many
SELECT * FROM play_history
WHERE user_id = $1
  AND (
    sqlc.narg('cursor_played_at')::timestamptz IS NULL
    OR (played_at, id) < (sqlc.narg('cursor_played_at')::timestamptz, sqlc.narg('cursor_id')::uuid)
  )
ORDER BY played_at DESC, id DESC
LIMIT sqlc.arg('limit_count');

-- name: ListContinueListening :many
-- Most recent play per song, excluding songs barely started or basically
-- finished (both would be noise in a "continue listening" widget).
SELECT song_id, progress_ms, played_at, duration_ms FROM (
  SELECT DISTINCT ON (h.song_id) h.song_id, h.progress_ms, h.played_at, s.duration_ms
  FROM play_history h
  JOIN songs s ON s.id = h.song_id
  WHERE h.user_id = $1
    AND h.progress_ms > 5000
    AND h.progress_ms < (s.duration_ms * 0.95)
  ORDER BY h.song_id, h.played_at DESC
) AS recent
ORDER BY played_at DESC
LIMIT $2;
