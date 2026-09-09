-- name: InsertHostMetric :exec
INSERT INTO host_metrics (
    id, host_id, sampled_at, cpu_pct, mem_pct, net_in, net_out, disks
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: ListHostMetricsInRange :many
SELECT * FROM host_metrics
WHERE host_id = $1 AND sampled_at >= $2 AND sampled_at <= $3
ORDER BY sampled_at ASC;

-- name: DeleteHostMetricsOlderThan :execrows
DELETE FROM host_metrics WHERE sampled_at < $1;

-- name: DeleteHostMetricsByHost :exec
DELETE FROM host_metrics WHERE host_id = $1;

-- name: DecimateHostMetrics :execrows
-- Keep at most one sample per (host, minute) for rows older than the cutoff,
-- deleting the rest. MIN(id) is the earliest ULID in each minute bucket.
DELETE FROM host_metrics
WHERE host_metrics.sampled_at < $1
  AND host_metrics.id NOT IN (
    SELECT MIN(hm.id) FROM host_metrics hm
    WHERE hm.sampled_at < $1
    GROUP BY hm.host_id, date_trunc('minute', hm.sampled_at)
  );

-- name: AggregateHostMetricsInWindow :one
-- Reduce a bounded correlation window to its peaks and sample count. The peaks
-- are computed here rather than in Go so no numeric column ever crosses the wire
-- (spec 089, FR-021). A window with no samples returns sample_count = 0; the
-- caller turns that into an absent context, never a zero-filled one.
SELECT
    COALESCE(MAX(cpu_pct), 0)::double precision AS peak_cpu_pct,
    COALESCE(MAX(mem_pct), 0)::double precision AS peak_mem_pct,
    COUNT(*) AS sample_count
FROM host_metrics
WHERE host_id = $1 AND sampled_at >= $2 AND sampled_at <= $3;

-- name: ListHostDisksInWindow :many
-- The disks column only, for the same bounded window. Deliberately a narrow
-- projection: disk usage is a stored document and nothing in this codebase
-- reaches inside JSON from SQL on either dialect, so the worst mount is picked
-- in Go (spec 089, FR-021a). Rows returned are bounded by the window.
SELECT disks FROM host_metrics
WHERE host_id = $1 AND sampled_at >= $2 AND sampled_at <= $3
  AND disks IS NOT NULL;
