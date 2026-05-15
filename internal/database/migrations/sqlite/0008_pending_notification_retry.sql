-- 0008: notification event retry state for startup recovery
ALTER TABLE notification_events ADD COLUMN status TEXT NOT NULL DEFAULT 'pending';
ALTER TABLE notification_events ADD COLUMN claim_owner TEXT;
ALTER TABLE notification_events ADD COLUMN claimed_at DATETIME;
ALTER TABLE notification_events ADD COLUMN processed_at DATETIME;
ALTER TABLE notification_events ADD COLUMN last_error TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_notification_events_status_created_at
    ON notification_events(status, created_at);

CREATE INDEX IF NOT EXISTS idx_notification_events_claim_owner
    ON notification_events(claim_owner);
