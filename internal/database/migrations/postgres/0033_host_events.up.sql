-- 0033: Kernel events reported by the host agent (spec 090). One row per kind per
-- collection interval, carrying how many kernel reports it aggregates -- an OOM
-- storm is one row with occurrences=200, not two hundred rows.
--
-- detail holds classified fields only: process, pid, cgroup, and the bounded list
-- of distinct processes affected. Never the raw kernel line: /dev/kmsg carries
-- every subsystem's output in formats that change between kernel versions, and
-- storing it would mean keeping content nobody has examined.
--
-- Events get their own retention, longer than metrics and never thinned
-- (ADR 0011): metrics are dense and individually cheap, a kernel event is rare,
-- discrete, and is the point of the feature.
CREATE TABLE IF NOT EXISTS host_events (
    id          TEXT PRIMARY KEY,
    host_id     TEXT NOT NULL REFERENCES hosts(id) ON DELETE CASCADE,
    occurred_at TIMESTAMPTZ NOT NULL,
    kind        TEXT NOT NULL,
    source      TEXT NOT NULL,
    occurrences INTEGER NOT NULL,
    detail      JSONB
);
CREATE INDEX IF NOT EXISTS idx_host_events_host_occurred ON host_events(host_id, occurred_at DESC);
