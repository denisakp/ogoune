-- 0031: Latest database health per monitor (spec 088). One row per PostgreSQL or
-- MySQL protocol monitor, replaced on every check, so the operator can watch
-- connection saturation climb before it becomes an outage.
--
-- Deliberately its own table rather than columns on `resources`: that table is
-- wide and read on every list endpoint under a benchmark gate, and these columns
-- would be null for every monitor that is not a database.
--
-- Every metric is nullable and independently so: a field is absent when the
-- credential lacks the privilege to read it correctly, when the value does not
-- apply (no replication), or when the server is below the supported version. The
-- two booleans are not nullable -- a consumer can always rely on them.
--
-- No retention job: one row per monitor, replaced in place, so size is constant
-- no matter how long a monitor runs.
CREATE TABLE IF NOT EXISTS resource_health (
    resource_id             TEXT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    collected_at            TIMESTAMPTZ NOT NULL,
    connections_active      BIGINT,
    connections_max         BIGINT,
    longest_query_seconds   DOUBLE PRECISION,
    replication_lag_seconds DOUBLE PRECISION,
    privilege_limited       BOOLEAN NOT NULL,
    unsupported_version     BOOLEAN NOT NULL
);
