-- 0013: Optional auth credentials for protocol-aware monitors (Redis, MySQL, PostgreSQL).
-- Encryption: `password` and `options` are AES-256-GCM ciphertext (via pkg/crypto).
-- `username` is stored in plaintext (Redis ACL / MySQL / Postgres role name).
CREATE TABLE IF NOT EXISTS resource_credentials (
    id          TEXT PRIMARY KEY,
    resource_id TEXT NOT NULL UNIQUE,
    username    TEXT,
    password    BLOB NOT NULL,
    options     BLOB,
    created_at  DATETIME NOT NULL,
    updated_at  DATETIME NOT NULL,
    FOREIGN KEY (resource_id) REFERENCES resources(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_resource_credentials_resource_id ON resource_credentials(resource_id);
