-- name: CreateSession :one
INSERT INTO sessions (id, user_id, browser, os, ip, location, last_active_at, created_at, revoked_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: FindSessionByID :one
SELECT * FROM sessions WHERE id = $1;

-- name: ListActiveSessionsByUser :many
SELECT * FROM sessions
WHERE user_id = $1 AND revoked_at IS NULL
ORDER BY last_active_at DESC;

-- name: UpdateSessionLastActive :exec
UPDATE sessions
SET last_active_at = $2
WHERE id = $1;

-- name: RevokeSession :execrows
UPDATE sessions
SET revoked_at = $2
WHERE id = $1 AND revoked_at IS NULL;

-- name: RevokeAllSessionsExcept :execrows
UPDATE sessions
SET revoked_at = $3
WHERE user_id = $1 AND id <> $2 AND revoked_at IS NULL;
