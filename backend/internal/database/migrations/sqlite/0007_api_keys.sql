-- 0007: API keys for programmatic authentication
CREATE TABLE IF NOT EXISTS api_keys (
    id           TEXT     NOT NULL PRIMARY KEY,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_id      TEXT     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name         TEXT     NOT NULL,
    key_hash     TEXT     NOT NULL UNIQUE,
    key_prefix   TEXT     NOT NULL,
    scope        TEXT     NOT NULL DEFAULT 'read',
    expires_at   DATETIME,
    last_used_at DATETIME,
    last_used_ip TEXT     NOT NULL DEFAULT '',
    is_active    INTEGER  NOT NULL DEFAULT 1
);

CREATE INDEX IF NOT EXISTS idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX IF NOT EXISTS idx_api_keys_active ON api_keys(user_id, is_active);
