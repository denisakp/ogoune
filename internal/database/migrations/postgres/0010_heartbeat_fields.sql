-- 0010: Heartbeat monitoring schema fields and indexes
ALTER TABLE resources
    ADD COLUMN IF NOT EXISTS heartbeat_slug TEXT,
    ADD COLUMN IF NOT EXISTS heartbeat_interval INTEGER,
    ADD COLUMN IF NOT EXISTS heartbeat_grace INTEGER,
    ADD COLUMN IF NOT EXISTS last_ping_at TIMESTAMP NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_resources_heartbeat_slug_unique
    ON resources (heartbeat_slug)
    WHERE heartbeat_slug IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_resources_heartbeat_missed
    ON resources (type, status, last_ping_at)
    WHERE type = 'heartbeat' AND is_active = true;
