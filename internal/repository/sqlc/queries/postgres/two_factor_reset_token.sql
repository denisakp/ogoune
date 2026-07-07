-- name: CreateTwoFactorResetToken :exec
INSERT INTO two_factor_reset_tokens (token_hash, user_id, expires_at, created_at)
VALUES ($1, $2, $3, $4);

-- name: FindActiveTwoFactorResetToken :one
SELECT * FROM two_factor_reset_tokens
WHERE token_hash = $1 AND used_at IS NULL AND expires_at > $2;

-- name: MarkTwoFactorResetTokenUsed :execrows
UPDATE two_factor_reset_tokens
SET used_at = $2
WHERE token_hash = $1 AND used_at IS NULL;

-- name: CountRecentTwoFactorResetTokensByUser :one
SELECT COUNT(*)::bigint FROM two_factor_reset_tokens
WHERE user_id = $1 AND created_at > $2;

-- name: DeleteExpiredTwoFactorResetTokens :exec
DELETE FROM two_factor_reset_tokens WHERE expires_at < $1;
