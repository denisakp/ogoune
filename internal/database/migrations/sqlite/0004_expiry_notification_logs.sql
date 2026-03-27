CREATE TABLE IF NOT EXISTS expiry_notification_logs (
    id          TEXT     PRIMARY KEY,
    resource_id TEXT     NOT NULL REFERENCES resources(id) ON DELETE CASCADE,
    expiry_type TEXT     NOT NULL CHECK (expiry_type IN ('ssl', 'domain')),
    threshold   INTEGER  NOT NULL,
    sent_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (resource_id, expiry_type, threshold)
);

CREATE INDEX IF NOT EXISTS idx_expiry_notification_logs_resource_id ON expiry_notification_logs(resource_id);
CREATE INDEX IF NOT EXISTS idx_expiry_notification_logs_expiry_type ON expiry_notification_logs(expiry_type);
CREATE INDEX IF NOT EXISTS idx_expiry_notification_logs_sent_at ON expiry_notification_logs(sent_at);

ALTER TABLE resources ADD COLUMN expiry_alert_thresholds TEXT DEFAULT NULL;
