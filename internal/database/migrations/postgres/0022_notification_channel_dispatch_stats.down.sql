ALTER TABLE notification_channels
    DROP COLUMN IF EXISTS last_sent_at,
    DROP COLUMN IF EXISTS last_failure_at,
    DROP COLUMN IF EXISTS failures_24h;
