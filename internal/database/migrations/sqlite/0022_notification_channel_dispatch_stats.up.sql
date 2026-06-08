ALTER TABLE notification_channels ADD COLUMN last_sent_at    DATETIME NULL;
ALTER TABLE notification_channels ADD COLUMN last_failure_at DATETIME NULL;
ALTER TABLE notification_channels ADD COLUMN failures_24h    INTEGER  NOT NULL DEFAULT 0;
