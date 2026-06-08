-- name: CreateMonitoringActivity :exec
INSERT INTO monitoring_activities (
    id, created_at, updated_at, resource_id, message, success,
    response_time, response_data, is_maintenance
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: ListMonitoringActivities :many
SELECT * FROM monitoring_activities
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: FindMonitoringActivitiesByResourceID :many
SELECT * FROM monitoring_activities
WHERE resource_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: SelectMonitoringActivitySuccessInWindow :many
SELECT success FROM monitoring_activities
WHERE resource_id = ? AND created_at >= ?
ORDER BY created_at ASC;

-- name: SelectMonitoringActivityHourlyAggregateInputs :many
SELECT created_at, success FROM monitoring_activities
WHERE resource_id = ? AND created_at >= ?
ORDER BY created_at ASC;

-- name: CountMonitoringActivitySinceTotal :one
SELECT COUNT(*) FROM monitoring_activities WHERE created_at >= ?;

-- name: CountMonitoringActivitySinceSuccess :one
SELECT COUNT(*) FROM monitoring_activities WHERE created_at >= ? AND success = 1;

-- name: CountMonitoringActivityByResourceTotal :one
SELECT COUNT(*) FROM monitoring_activities WHERE resource_id = ? AND created_at >= ?;

-- name: CountMonitoringActivityByResourceSuccess :one
SELECT COUNT(*) FROM monitoring_activities
WHERE resource_id = ? AND created_at >= ? AND success = 1;

-- name: AvgResponseTimeByResourceInWindow :one
SELECT AVG(response_time) FROM monitoring_activities
WHERE resource_id = ? AND created_at >= ? AND success = 1;

-- name: AvgResponseTimeByResourcesSince :many
-- One round-trip bulk avg grouped by resource. Used by the list path to
-- enrich each resource with its avg response time over a sliding window (30d).
SELECT resource_id, AVG(response_time) AS avg_ms
FROM monitoring_activities
WHERE created_at >= sqlc.arg(since)
  AND success = 1
  AND resource_id IN (sqlc.slice('resource_ids'))
GROUP BY resource_id;

-- name: GetRecentResponseTimes :many
SELECT created_at, response_time FROM monitoring_activities
WHERE resource_id = ? AND success = 1
ORDER BY created_at DESC
LIMIT ?;
