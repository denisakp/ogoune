-- 0031: Latest database health per monitor (spec 088). See postgres/0031 for rationale.
CREATE TABLE IF NOT EXISTS resource_health (
    resource_id             TEXT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    collected_at            DATETIME NOT NULL,
    connections_active      INTEGER,
    connections_max         INTEGER,
    longest_query_seconds   REAL,
    replication_lag_seconds REAL,
    privilege_limited       INTEGER NOT NULL,
    unsupported_version     INTEGER NOT NULL
);
