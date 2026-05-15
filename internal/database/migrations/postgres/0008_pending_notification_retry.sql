-- 0008: notification event retry state for startup recovery
ALTER TABLE notification_events
    ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'pending',
    ADD COLUMN IF NOT EXISTS claim_owner TEXT,
    ADD COLUMN IF NOT EXISTS claimed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS processed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS last_error TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_notification_events_status_created_at
    ON notification_events(status, created_at);

CREATE INDEX IF NOT EXISTS idx_notification_events_claim_owner
    ON notification_events(claim_owner);
