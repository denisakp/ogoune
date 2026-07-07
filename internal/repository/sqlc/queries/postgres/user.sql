-- name: CreateUser :one
INSERT INTO users (
    id, email, name, hashed_password,
    password_initialized, force_password_change,
    two_factor_enabled, two_factor_secret, two_factor_backup_codes,
    last_login_at, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: FindUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: FindUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateUser :exec
UPDATE users
SET email = $2,
    name = $3,
    hashed_password = $4,
    password_initialized = $5,
    force_password_change = $6,
    two_factor_enabled = $7,
    two_factor_secret = $8,
    two_factor_backup_codes = $9,
    last_login_at = $10,
    updated_at = $11
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: UpdateUserPassword :exec
UPDATE users
SET hashed_password = $2,
    password_initialized = true,
    force_password_change = false
WHERE id = $1;

-- name: UpdateUserLastLogin :exec
UPDATE users
SET last_login_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: UpdateUserTwoFactorSecret :exec
UPDATE users
SET two_factor_secret = $2,
    two_factor_enabled = $3
WHERE id = $1;
