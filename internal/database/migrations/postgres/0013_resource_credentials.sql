-- 0013: Optional auth credentials for protocol-aware monitors (Redis, MySQL, PostgreSQL).
-- Encryption: `password` and `options` are AES-256-GCM ciphertext (via pkg/crypto).
-- `username` is stored in plaintext (Redis ACL / MySQL / Postgres role name).
CREATE TABLE IF NOT EXISTS resource_credentials (
    id          VARCHAR(26) PRIMARY KEY,
    resource_id VARCHAR(26) NOT NULL UNIQUE REFERENCES resources(id) ON DELETE CASCADE,
    username    VARCHAR(128),
    password    BYTEA NOT NULL,
    options     BYTEA,
    created_at  TIMESTAMPTZ NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_resource_credentials_resource_id ON resource_credentials(resource_id);
