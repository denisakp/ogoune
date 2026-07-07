-- name: CreateAPIKey :exec
INSERT INTO api_keys (
    id, created_at, updated_at, user_id, name, key_hash, key_prefix,
    scope, expires_at, last_used_at, last_used_ip, is_active
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);

-- name: FindAPIKeyByIDForUser :one
SELECT * FROM api_keys WHERE id = $1 AND user_id = $2;

-- name: FindAPIKeyByKeyHash :one
SELECT * FROM api_keys WHERE key_hash = $1;

-- name: ListAPIKeysByUserID :many
SELECT * FROM api_keys
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: UpdateAPIKeyLastUsed :execrows
UPDATE api_keys
SET last_used_at = $2, last_used_ip = $3
WHERE id = $1;

-- name: RevokeAPIKey :execrows
UPDATE api_keys
SET is_active = false
WHERE id = $1 AND user_id = $2;

-- name: CountAPIKeysByUserID :one
SELECT COUNT(*) FROM api_keys WHERE user_id = $1;
