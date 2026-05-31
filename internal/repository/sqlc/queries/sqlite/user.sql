-- name: CreateUser :one
INSERT INTO users (
    id, email, name, hashed_password,
    password_initialized, force_password_change,
    two_factor_enabled, two_factor_secret, two_factor_backup_codes,
    last_login_at, created_at, updated_at
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: FindUserByID :one
SELECT * FROM users WHERE id = ?;

-- name: FindUserByEmail :one
SELECT * FROM users WHERE email = ?;

-- name: UpdateUser :exec
UPDATE users
SET email = ?,
    name = ?,
    hashed_password = ?,
    password_initialized = ?,
    force_password_change = ?,
    two_factor_enabled = ?,
    two_factor_secret = ?,
    two_factor_backup_codes = ?,
    last_login_at = ?,
    updated_at = ?
WHERE id = ?;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;

-- name: UpdateUserPassword :exec
UPDATE users
SET hashed_password = ?,
    password_initialized = 1,
    force_password_change = 0
WHERE id = ?;

-- name: UpdateUserLastLogin :exec
UPDATE users
SET last_login_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: UpdateUserTwoFactorSecret :exec
UPDATE users
SET two_factor_secret = ?,
    two_factor_enabled = ?
WHERE id = ?;
