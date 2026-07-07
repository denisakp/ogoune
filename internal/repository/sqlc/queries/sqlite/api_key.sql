-- name: CreateAPIKey :exec
INSERT INTO api_keys (
    id, created_at, updated_at, user_id, name, key_hash, key_prefix,
    scope, expires_at, last_used_at, last_used_ip, is_active
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: FindAPIKeyByIDForUser :one
SELECT * FROM api_keys WHERE id = ? AND user_id = ?;

-- name: FindAPIKeyByKeyHash :one
SELECT * FROM api_keys WHERE key_hash = ?;

-- name: ListAPIKeysByUserID :many
SELECT * FROM api_keys
WHERE user_id = ?
ORDER BY created_at DESC;

-- name: UpdateAPIKeyLastUsed :execrows
UPDATE api_keys
SET last_used_at = ?, last_used_ip = ?
WHERE id = ?;

-- name: RevokeAPIKey :execrows
UPDATE api_keys
SET is_active = 0
WHERE id = ? AND user_id = ?;

-- name: CountAPIKeysByUserID :one
SELECT COUNT(*) FROM api_keys WHERE user_id = ?;
