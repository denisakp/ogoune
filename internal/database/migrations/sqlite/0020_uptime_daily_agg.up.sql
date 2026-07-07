-- 0020: Daily uptime aggregates — SQLite mirror.
-- `day` stored as TEXT (`YYYY-MM-DD` UTC) per the cross-dialect convention.
-- `uptime_ratio` stored as REAL (precision sufficient since recomputed, never accumulated).

CREATE TABLE IF NOT EXISTS uptime_daily_agg (
    resource_id   TEXT     NOT NULL,
    day           TEXT     NOT NULL,
    samples       INTEGER  NOT NULL DEFAULT 0,
    up            INTEGER  NOT NULL DEFAULT 0,
    degraded      INTEGER  NOT NULL DEFAULT 0,
    down          INTEGER  NOT NULL DEFAULT 0,
    uptime_ratio  REAL     NOT NULL DEFAULT 1.0,
    computed_at   DATETIME NOT NULL,
    PRIMARY KEY (resource_id, day)
);
CREATE INDEX IF NOT EXISTS idx_uptime_daily_agg_day ON uptime_daily_agg(day);
