-- name: UpsertResourceHealth :exec
-- At most one row per monitor: replace rather than accumulate. This is what makes
-- storage constant per monitor and removes any need for a retention job.
INSERT INTO resource_health (
    resource_id, collected_at, connections_active, connections_max,
    longest_query_seconds, replication_lag_seconds, privilege_limited, unsupported_version
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (resource_id) DO UPDATE SET
    collected_at            = EXCLUDED.collected_at,
    connections_active      = EXCLUDED.connections_active,
    connections_max         = EXCLUDED.connections_max,
    longest_query_seconds   = EXCLUDED.longest_query_seconds,
    replication_lag_seconds = EXCLUDED.replication_lag_seconds,
    privilege_limited       = EXCLUDED.privilege_limited,
    unsupported_version     = EXCLUDED.unsupported_version;

-- name: FindResourceHealth :one
SELECT * FROM resource_health WHERE resource_id = $1;

-- name: DeleteResourceHealth :exec
-- Called when a check collects nothing at all, so the interface shows no figures
-- rather than yesterday's.
DELETE FROM resource_health WHERE resource_id = $1;
