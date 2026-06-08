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

-- name: FindEarliestUptimeDailyAggDay :one
SELECT MIN(day) AS earliest FROM uptime_daily_agg;

-- name: SumUptimeAggByResourcesSince :many
-- One round-trip bulk aggregation grouped by resource. Used by the list path
-- to enrich each resource with its uptime ratio over a sliding window (30d).
SELECT
    resource_id,
    SUM(up)      AS up_sum,
    SUM(samples) AS samples_sum
FROM uptime_daily_agg
WHERE day >= sqlc.arg(from_day)
  AND resource_id IN (sqlc.slice('resource_ids'))
GROUP BY resource_id;
