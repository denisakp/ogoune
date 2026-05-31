-- name: CreateNotificationChannel :exec
INSERT INTO notification_channels (
    id, created_at, updated_at, name, type, config, enabled_by_default
)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: FindNotificationChannelByID :one
SELECT * FROM notification_channels WHERE id = $1;

-- name: ListNotificationChannels :many
SELECT * FROM notification_channels
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateNotificationChannel :execrows
UPDATE notification_channels
SET name = $2,
    type = $3,
    config = $4,
    enabled_by_default = $5,
    updated_at = $6
WHERE id = $1;

-- name: DeleteNotificationChannel :execrows
DELETE FROM notification_channels WHERE id = $1;

-- name: FindNotificationChannelsByType :many
SELECT * FROM notification_channels WHERE type = $1;

-- name: FindDefaultNotificationChannels :many
SELECT * FROM notification_channels WHERE enabled_by_default = true;

-- name: FindNotificationChannelsByResourceID :many
SELECT nc.* FROM notification_channels nc
JOIN resource_notification_channels rnc
    ON rnc.notification_channel_id = nc.id
WHERE rnc.resource_id = $1;

-- name: FindNotificationChannelsByComponentID :many
SELECT nc.* FROM notification_channels nc
JOIN component_notification_channels cnc
    ON cnc.notification_channel_id = nc.id
WHERE cnc.component_id = $1;
