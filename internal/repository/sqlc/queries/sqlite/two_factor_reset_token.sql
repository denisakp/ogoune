-- name: CreateTwoFactorResetToken :exec
INSERT INTO two_factor_reset_tokens (token_hash, user_id, expires_at, created_at)
VALUES (?, ?, ?, ?);

-- name: FindActiveTwoFactorResetToken :one
SELECT * FROM two_factor_reset_tokens
WHERE token_hash = ? AND used_at IS NULL AND expires_at > ?;

-- name: MarkTwoFactorResetTokenUsed :execrows
UPDATE two_factor_reset_tokens
SET used_at = ?
WHERE token_hash = ? AND used_at IS NULL;

-- name: CountRecentTwoFactorResetTokensByUser :one
SELECT COUNT(*) FROM two_factor_reset_tokens
WHERE user_id = ? AND created_at > ?;

-- name: DeleteExpiredTwoFactorResetTokens :exec
DELETE FROM two_factor_reset_tokens WHERE expires_at < ?;
