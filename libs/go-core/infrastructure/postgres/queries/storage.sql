-- name: GetStorageAccountByID :one
SELECT * FROM storage_accounts WHERE id = $1;

-- name: GetActiveStorageAccount :one
-- Sprint 9: quota-aware routing — pick the active, non-down account with
-- the most free space. An account with no quota info yet (NULL, before
-- its first health check) is treated as having plenty of room rather
-- than excluded, so a brand new account isn't starved until checked.
SELECT * FROM storage_accounts
WHERE is_active = true AND health_status <> 'down'
ORDER BY (COALESCE(quota_bytes, 999999999999) - used_bytes) DESC
LIMIT 1;

-- name: ListStorageAccounts :many
SELECT * FROM storage_accounts ORDER BY created_at ASC;

-- name: CreateStorageAccount :one
INSERT INTO storage_accounts (id, provider, label, account_email, credentials_encrypted, quota_bytes)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: IncrementStorageAccountUsedBytes :exec
UPDATE storage_accounts SET used_bytes = used_bytes + $2, updated_at = now() WHERE id = $1;

-- name: UpdateStorageAccountHealth :exec
UPDATE storage_accounts SET
  health_status = $2,
  quota_bytes = $3,
  used_bytes = $4,
  last_health_check_at = now(),
  updated_at = now()
WHERE id = $1;

-- name: DeleteStorageAccount :execrows
DELETE FROM storage_accounts WHERE id = $1;

-- name: GetStorageFileByChecksum :one
SELECT * FROM storage_files WHERE checksum = $1;

-- name: CreateStorageFile :one
INSERT INTO storage_files (id, storage_account_id, provider_file_id, checksum, size_bytes, mime_type)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;
