-- 0020: Daily uptime aggregates — spec 060 FR-004 + FR-008 + FR-026.
-- Per-resource, per-UTC-day counters populated by the aggregator cron every
-- 5 minutes. Reads serve the 90-day ribbon (`/`), the calendar (`/uptime`),
-- and the per-resource windows endpoint without scanning monitoring_activity.

CREATE TABLE IF NOT EXISTS uptime_daily_agg (
    resource_id   TEXT          NOT NULL,
    day           DATE          NOT NULL,
    samples       INTEGER       NOT NULL DEFAULT 0,
    up            INTEGER       NOT NULL DEFAULT 0,
    degraded      INTEGER       NOT NULL DEFAULT 0,
    down          INTEGER       NOT NULL DEFAULT 0,
    uptime_ratio  NUMERIC(5,4)  NOT NULL DEFAULT 1.0000,
    computed_at   TIMESTAMPTZ   NOT NULL,
    PRIMARY KEY (resource_id, day)
);
CREATE INDEX IF NOT EXISTS idx_uptime_daily_agg_day ON uptime_daily_agg(day);
