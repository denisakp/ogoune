-- name: CreateMonitoringActivity :exec
INSERT INTO monitoring_activities (
    id, created_at, updated_at, resource_id, message, success,
    response_time, response_data, is_maintenance
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: ListMonitoringActivities :many
SELECT * FROM monitoring_activities
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: FindMonitoringActivitiesByResourceID :many
SELECT * FROM monitoring_activities
WHERE resource_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: SelectMonitoringActivitySuccessInWindow :many
SELECT success FROM monitoring_activities
WHERE resource_id = $1 AND created_at >= $2
ORDER BY created_at ASC;

-- name: SelectMonitoringActivityHourlyAggregateInputs :many
SELECT created_at, success FROM monitoring_activities
WHERE resource_id = $1 AND created_at >= $2
ORDER BY created_at ASC;

-- name: CountMonitoringActivitySinceTotal :one
SELECT COUNT(*) FROM monitoring_activities WHERE created_at >= $1;

-- name: CountMonitoringActivitySinceSuccess :one
SELECT COUNT(*) FROM monitoring_activities WHERE created_at >= $1 AND success = true;

-- name: CountMonitoringActivityByResourceTotal :one
SELECT COUNT(*) FROM monitoring_activities WHERE resource_id = $1 AND created_at >= $2;

-- name: CountMonitoringActivityByResourceSuccess :one
SELECT COUNT(*) FROM monitoring_activities
WHERE resource_id = $1 AND created_at >= $2 AND success = true;

-- name: AvgResponseTimeByResourceInWindow :one
SELECT AVG(response_time)::float8 FROM monitoring_activities
WHERE resource_id = $1 AND created_at >= $2 AND success = true;

-- name: GetRecentResponseTimes :many
SELECT created_at, response_time FROM monitoring_activities
WHERE resource_id = $1 AND success = true
ORDER BY created_at DESC
LIMIT $2;
