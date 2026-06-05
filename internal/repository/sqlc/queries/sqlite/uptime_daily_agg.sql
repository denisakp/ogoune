-- name: UpsertUptimeDailyAgg :exec
INSERT INTO uptime_daily_agg (
    resource_id, day, samples, up, degraded, down, uptime_ratio, computed_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?
)
ON CONFLICT(resource_id, day) DO UPDATE SET
    samples      = excluded.samples,
    up           = excluded.up,
    degraded     = excluded.degraded,
    down         = excluded.down,
    uptime_ratio = excluded.uptime_ratio,
    computed_at  = excluded.computed_at;

-- name: FindUptimeDailyAggForResource :many
SELECT resource_id, day, samples, up, degraded, down, uptime_ratio, computed_at
FROM uptime_daily_agg
WHERE resource_id = sqlc.arg(resource_id)
  AND day >= sqlc.arg(from_day)
  AND day <= sqlc.arg(to_day)
ORDER BY day ASC;
