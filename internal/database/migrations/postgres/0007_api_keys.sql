-- 0007: API keys for programmatic authentication
CREATE TABLE IF NOT EXISTS api_keys (
    id           TEXT        NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_id      TEXT        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name         TEXT        NOT NULL,
    key_hash     TEXT        NOT NULL,
    key_prefix   TEXT        NOT NULL,
    scope        TEXT        NOT NULL DEFAULT 'read' CHECK (scope IN ('read', 'read_write')),
    expires_at   TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    last_used_ip TEXT        NOT NULL DEFAULT '',
    is_active    BOOLEAN     NOT NULL DEFAULT true,
    PRIMARY KEY (id),
    CONSTRAINT uq_api_keys_key_hash UNIQUE (key_hash)
);

CREATE INDEX IF NOT EXISTS idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX IF NOT EXISTS idx_api_keys_active ON api_keys(user_id, is_active);
