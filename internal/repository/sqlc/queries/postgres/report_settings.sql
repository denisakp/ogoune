-- name: GetReportSettings :one
SELECT id, enabled, recipient_email, schedule, scope, last_sent_at, created_at, updated_at
FROM report_settings
LIMIT 1;

-- name: UpsertReportSettings :one
INSERT INTO report_settings (id, enabled, recipient_email, schedule, scope, last_sent_at, created_at, updated_at)
VALUES (sqlc.arg('id'), sqlc.arg('enabled'), sqlc.arg('recipient_email'), sqlc.arg('schedule'), sqlc.arg('scope'), sqlc.arg('last_sent_at'), sqlc.arg('created_at'), sqlc.arg('updated_at'))
ON CONFLICT(id) DO UPDATE SET
    enabled = excluded.enabled,
    recipient_email = excluded.recipient_email,
    schedule = excluded.schedule,
    scope = excluded.scope,
    last_sent_at = excluded.last_sent_at,
    updated_at = excluded.updated_at
RETURNING *;
