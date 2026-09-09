-- name: InsertHostMetric :exec
INSERT INTO host_metrics (
    id, host_id, sampled_at, cpu_pct, mem_pct, net_in, net_out, disks
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: ListHostMetricsInRange :many
SELECT * FROM host_metrics
WHERE host_id = ?1 AND sampled_at >= ?2 AND sampled_at <= ?3
ORDER BY sampled_at ASC;

-- name: DeleteHostMetricsOlderThan :execrows
DELETE FROM host_metrics WHERE sampled_at < ?1;

-- name: DeleteHostMetricsByHost :exec
DELETE FROM host_metrics WHERE host_id = ?;

-- name: DecimateHostMetrics :execrows
-- Keep at most one sample per (host, minute) for rows older than the cutoff,
-- deleting the rest. MIN(id) is the earliest ULID in each minute bucket;
-- substr(sampled_at,1,16) buckets by 'YYYY-MM-DD HH:MM'.
DELETE FROM host_metrics
WHERE host_metrics.sampled_at < ?1
  AND host_metrics.id NOT IN (
    SELECT MIN(hm.id) FROM host_metrics hm
    WHERE hm.sampled_at < ?1
    GROUP BY hm.host_id, substr(hm.sampled_at, 1, 16)
  );

-- name: AggregateHostMetricsInWindow :many
-- Reduce a bounded correlation window to its peaks and sample count, and carry
-- back the disk documents for the same window, in ONE round trip. Mirrors the
-- Postgres query exactly in name and result shape.
--
-- The peaks are computed by the database via window functions so no numeric
-- column is ever reduced in Go (spec 089, FR-021); they repeat identically on
-- every row, which costs a few bytes and saves a round trip. The disks column
-- rides along because it is a stored document and nothing in this codebase
-- reaches inside JSON from SQL on either dialect (FR-021a).
--
-- The range predicate mirrors ListHostMetricsInRange exactly -- no strftime is
-- needed for a plain comparison, only for extracting epoch seconds.
-- Keep this file pure ASCII: sqlc slices SQLite query text by byte offset.
SELECT
    CAST(COALESCE(MAX(cpu_pct) OVER (), 0.0) AS REAL) AS peak_cpu_pct,
    CAST(COALESCE(MAX(mem_pct) OVER (), 0.0) AS REAL) AS peak_mem_pct,
    COUNT(*) OVER () AS sample_count,
    disks
FROM host_metrics
WHERE host_id = ?1 AND sampled_at >= ?2 AND sampled_at <= ?3;
