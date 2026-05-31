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
