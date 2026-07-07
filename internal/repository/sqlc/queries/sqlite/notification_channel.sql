-- name: CreateNotificationChannel :exec
INSERT INTO notification_channels (
    id, created_at, updated_at, name, type, config, enabled_by_default
)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: FindNotificationChannelByID :one
SELECT * FROM notification_channels WHERE id = ?;

-- name: ListNotificationChannels :many
SELECT * FROM notification_channels
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: UpdateNotificationChannel :execrows
UPDATE notification_channels
SET name = ?,
    type = ?,
    config = ?,
    enabled_by_default = ?,
    updated_at = ?
WHERE id = ?;

-- name: DeleteNotificationChannel :execrows
DELETE FROM notification_channels WHERE id = ?;

-- name: FindNotificationChannelsByType :many
SELECT * FROM notification_channels WHERE type = ?;

-- name: FindDefaultNotificationChannels :many
SELECT * FROM notification_channels WHERE enabled_by_default = 1;

-- name: FindNotificationChannelsByResourceID :many
SELECT nc.* FROM notification_channels nc
JOIN resource_notification_channels rnc
    ON rnc.notification_channel_id = nc.id
WHERE rnc.resource_id = ?;

-- name: FindNotificationChannelsByComponentID :many
SELECT nc.* FROM notification_channels nc
JOIN component_notification_channels cnc
    ON cnc.notification_channel_id = nc.id
WHERE cnc.component_id = ?;

-- name: MarkNotificationChannelSent :exec
UPDATE notification_channels
SET last_sent_at = sqlc.arg(at),
    updated_at   = sqlc.arg(at)
WHERE id = sqlc.arg(id);

-- name: MarkNotificationChannelFailure :exec
UPDATE notification_channels
SET failures_24h    = CASE
    WHEN last_failure_at IS NULL OR last_failure_at < sqlc.arg(cutoff_at) THEN 1
    ELSE failures_24h + 1
END,
    last_failure_at = sqlc.arg(at),
    updated_at      = sqlc.arg(at)
WHERE id = sqlc.arg(id);
