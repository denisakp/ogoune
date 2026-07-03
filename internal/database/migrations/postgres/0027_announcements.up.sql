-- 0027: Operator announcement banners (option 2). Single-tenant, instance-wide.
CREATE TABLE IF NOT EXISTS announcements (
    id          TEXT PRIMARY KEY,
    severity    TEXT NOT NULL,
    title       TEXT NOT NULL,
    description TEXT NOT NULL,
    dismissible BOOLEAN NOT NULL,
    active      BOOLEAN NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_announcements_active ON announcements(active, created_at DESC);
