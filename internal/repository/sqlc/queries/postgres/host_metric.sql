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

-- name: AggregateHostMetricsInWindow :many
-- Reduce a bounded correlation window to its peaks and sample count, and carry
-- back the disk documents for the same window, in ONE round trip.
--
-- The peaks are computed by the database via window functions so no numeric
-- column is ever reduced in Go (spec 089, FR-021); they repeat identically on
-- every row, which costs a few bytes and saves a round trip. The disks column
-- rides along because it is a stored document and nothing in this codebase
-- reaches inside JSON from SQL on either dialect, so the worst mount is picked
-- in Go (FR-021a).
--
-- No rows means no samples: the caller turns that into an absent context, never
-- a zero-filled one. Rows returned are bounded by the window (FR-021b).
SELECT
    COALESCE(MAX(cpu_pct) OVER (), 0)::double precision AS peak_cpu_pct,
    COALESCE(MAX(mem_pct) OVER (), 0)::double precision AS peak_mem_pct,
    COUNT(*) OVER () AS sample_count,
    disks
FROM host_metrics
WHERE host_id = $1 AND sampled_at >= $2 AND sampled_at <= $3;
