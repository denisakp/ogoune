-- 0010: Heartbeat monitoring schema fields and indexes
ALTER TABLE resources ADD COLUMN heartbeat_slug TEXT;
ALTER TABLE resources ADD COLUMN heartbeat_interval INTEGER;
ALTER TABLE resources ADD COLUMN heartbeat_grace INTEGER;
ALTER TABLE resources ADD COLUMN last_ping_at DATETIME;

CREATE UNIQUE INDEX IF NOT EXISTS idx_resources_heartbeat_slug_unique
    ON resources (heartbeat_slug)
    WHERE heartbeat_slug IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_resources_heartbeat_missed
    ON resources (type, status, last_ping_at);
