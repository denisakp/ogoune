-- 0033: Kernel events reported by the host agent (spec 090). See postgres/0033.
CREATE TABLE IF NOT EXISTS host_events (
    id          TEXT PRIMARY KEY,
    host_id     TEXT NOT NULL REFERENCES hosts(id) ON DELETE CASCADE,
    occurred_at DATETIME NOT NULL,
    kind        TEXT NOT NULL,
    source      TEXT NOT NULL,
    occurrences INTEGER NOT NULL,
    detail      TEXT
);
CREATE INDEX IF NOT EXISTS idx_host_events_host_occurred ON host_events(host_id, occurred_at DESC);
