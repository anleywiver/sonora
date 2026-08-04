-- name: GetStorageAccountByID :one
SELECT * FROM storage_accounts WHERE id = $1;

-- name: GetActiveStorageAccount :one
-- Sprint 3: pick the first active account, oldest first. No quota-aware
-- routing yet — that's Sprint 9 (multi-drive pool).
SELECT * FROM storage_accounts WHERE is_active = true ORDER BY created_at ASC LIMIT 1;

-- name: ListStorageAccounts :many
SELECT * FROM storage_accounts ORDER BY created_at ASC;

-- name: CreateStorageAccount :one
INSERT INTO storage_accounts (id, provider, label, account_email, credentials_encrypted, quota_bytes)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: IncrementStorageAccountUsedBytes :exec
UPDATE storage_accounts SET used_bytes = used_bytes + $2, updated_at = now() WHERE id = $1;

-- name: GetStorageFileByChecksum :one
SELECT * FROM storage_files WHERE checksum = $1;

-- name: CreateStorageFile :one
INSERT INTO storage_files (id, storage_account_id, provider_file_id, checksum, size_bytes, mime_type)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;
