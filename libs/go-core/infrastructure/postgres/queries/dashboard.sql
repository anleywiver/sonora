-- name: GetDashboardStats :one
SELECT
  (SELECT count(*) FROM songs) AS total_songs,
  (SELECT count(*) FROM users) AS total_users,
  (SELECT count(*) FROM storage_accounts) AS total_drives,
  (SELECT COALESCE(sum(used_bytes), 0)::bigint FROM storage_accounts) AS total_storage_bytes;

-- name: ListStorageDistribution :many
SELECT id, label, used_bytes, quota_bytes FROM storage_accounts ORDER BY label ASC;

-- name: GetBackgroundJobsSummary :many
SELECT status, count(*) AS job_count FROM ingest_jobs GROUP BY status;
