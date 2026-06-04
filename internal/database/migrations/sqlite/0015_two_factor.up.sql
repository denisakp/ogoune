-- 0015: TOTP 2FA magic-link reset tokens (FR-012a).
-- TOTP secret + backup codes already live on the `users` table (0001_initial).
-- This migration only adds the single-use reset-token table for the magic-link recovery flow.
CREATE TABLE IF NOT EXISTS two_factor_reset_tokens (
    token_hash  TEXT PRIMARY KEY,
    user_id     TEXT NOT NULL,
    expires_at  DATETIME NOT NULL,
    used_at     DATETIME,
    created_at  DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_two_factor_reset_tokens_user ON two_factor_reset_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_two_factor_reset_tokens_expires ON two_factor_reset_tokens(expires_at);
