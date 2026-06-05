-- name: UpsertUptimeDailyAgg :exec
INSERT INTO uptime_daily_agg (
    resource_id, day, samples, up, degraded, down, uptime_ratio, computed_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
ON CONFLICT (resource_id, day) DO UPDATE SET
    samples      = EXCLUDED.samples,
    up           = EXCLUDED.up,
    degraded     = EXCLUDED.degraded,
    down         = EXCLUDED.down,
    uptime_ratio = EXCLUDED.uptime_ratio,
    computed_at  = EXCLUDED.computed_at;

-- name: FindUptimeDailyAggRange :many
SELECT resource_id, day, samples, up, degraded, down, uptime_ratio, computed_at
FROM uptime_daily_agg
WHERE resource_id = ANY(@resource_ids::text[])
  AND day BETWEEN @from_day AND @to_day
ORDER BY day ASC;

-- name: FindUptimeDailyAggForResource :many
SELECT resource_id, day, samples, up, degraded, down, uptime_ratio, computed_at
FROM uptime_daily_agg
WHERE resource_id = $1
  AND day BETWEEN $2 AND $3
ORDER BY day ASC;

-- name: FindEarliestUptimeDailyAggDay :one
SELECT MIN(day) AS earliest FROM uptime_daily_agg;
