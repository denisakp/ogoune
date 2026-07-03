-- name: CreateReportHistory :one
INSERT INTO report_history (id, period, sent_at, status, uptime_pct, incident_count, downtime_seconds, recipient_email, resource_breakdown, created_at)
VALUES (sqlc.arg('id'), sqlc.arg('period'), sqlc.arg('sent_at'), sqlc.arg('status'), sqlc.arg('uptime_pct'), sqlc.arg('incident_count'), sqlc.arg('downtime_seconds'), sqlc.arg('recipient_email'), sqlc.arg('resource_breakdown'), sqlc.arg('created_at'))
RETURNING *;

-- name: ListRecentReportHistory :many
SELECT id, period, sent_at, status, uptime_pct, incident_count, downtime_seconds, recipient_email, resource_breakdown, created_at
FROM report_history
ORDER BY sent_at DESC
LIMIT sqlc.arg('lim');

-- name: FindReportHistoryByPeriod :one
SELECT id, period, sent_at, status, uptime_pct, incident_count, downtime_seconds, recipient_email, resource_breakdown, created_at
FROM report_history
WHERE period = sqlc.arg('period');
