-- name: CreateSession :one
INSERT INTO sessions (id, user_id, browser, os, ip, location, last_active_at, created_at, revoked_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: FindSessionByID :one
SELECT * FROM sessions WHERE id = ?;

-- name: ListActiveSessionsByUser :many
SELECT * FROM sessions
WHERE user_id = ? AND revoked_at IS NULL
ORDER BY last_active_at DESC;

-- name: UpdateSessionLastActive :exec
UPDATE sessions
SET last_active_at = ?
WHERE id = ?;

-- name: RevokeSession :execrows
UPDATE sessions
SET revoked_at = ?
WHERE id = ? AND revoked_at IS NULL;

-- name: RevokeAllSessionsExcept :execrows
UPDATE sessions
SET revoked_at = ?
WHERE user_id = ? AND id <> ? AND revoked_at IS NULL;
