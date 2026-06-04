-- 0014: User sessions with immediate revoke semantics (FR-009a).
-- AuthMiddleware reads `revoked_at` on every authenticated request — NULL = active, non-NULL = invalid.
CREATE TABLE IF NOT EXISTS sessions (
    id             TEXT PRIMARY KEY,
    user_id        TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    browser        TEXT NOT NULL DEFAULT '',
    os             TEXT NOT NULL DEFAULT '',
    ip             TEXT NOT NULL DEFAULT '',
    location       TEXT,
    last_active_at TIMESTAMPTZ NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL,
    revoked_at     TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_sessions_user_active ON sessions(user_id, revoked_at);
