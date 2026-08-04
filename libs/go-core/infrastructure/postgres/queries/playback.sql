-- name: GetPlaybackState :one
SELECT * FROM playback_states WHERE user_id = $1;

-- name: UpsertPlaybackState :one
INSERT INTO playback_states (user_id, active_device_id, current_song_id, position_ms, is_playing, updated_at)
VALUES ($1, $2, $3, $4, $5, now())
ON CONFLICT (user_id) DO UPDATE SET
    active_device_id = EXCLUDED.active_device_id,
    current_song_id = EXCLUDED.current_song_id,
    position_ms = EXCLUDED.position_ms,
    is_playing = EXCLUDED.is_playing,
    updated_at = now()
RETURNING *;

-- name: ListQueueByUser :many
SELECT * FROM queue_items WHERE user_id = $1 ORDER BY position ASC;

-- name: AddQueueItem :one
INSERT INTO queue_items (id, user_id, song_id, position)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: RemoveQueueItem :exec
DELETE FROM queue_items WHERE id = $1 AND user_id = $2;

-- name: UpdateQueueItemPosition :execrows
UPDATE queue_items SET position = $3 WHERE id = $1 AND user_id = $2;

-- name: ClearQueue :exec
DELETE FROM queue_items WHERE user_id = $1;
