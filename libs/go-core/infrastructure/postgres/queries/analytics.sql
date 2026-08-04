-- name: GetTopPlayedSongs :many
SELECT s.id, s.title, ar.name AS artist_name, count(h.id) AS play_count
FROM play_history h
JOIN songs s ON s.id = h.song_id
JOIN artists ar ON ar.id = s.artist_id
GROUP BY s.id, s.title, ar.name
ORDER BY play_count DESC, s.title ASC
LIMIT $1;

-- name: GetStorageGrowth :many
-- One row per month for the last 6 months (including the current one),
-- zero-filled via generate_series so a quiet month still shows up as a
-- bar at 0 instead of vanishing from the chart.
SELECT month::timestamptz AS month, COALESCE(sum(sf.size_bytes), 0)::bigint AS total_bytes
FROM generate_series(
  date_trunc('month', now()) - interval '5 months',
  date_trunc('month', now()),
  interval '1 month'
) AS month
LEFT JOIN storage_files sf ON date_trunc('month', sf.created_at) = month
GROUP BY month
ORDER BY month;
