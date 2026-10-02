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
SELECT * FROM notification_channels WHERE type = $1 AND disabled_at IS NULL;

-- name: FindDefaultNotificationChannels :many
SELECT * FROM notification_channels WHERE enabled_by_default = true AND disabled_at IS NULL;

-- name: FindNotificationChannelsByResourceID :many
SELECT nc.* FROM notification_channels nc
JOIN resource_notification_channels rnc
    ON rnc.notification_channel_id = nc.id
WHERE rnc.resource_id = $1 AND nc.disabled_at IS NULL;

-- name: FindNotificationChannelsByComponentID :many
SELECT nc.* FROM notification_channels nc
JOIN component_notification_channels cnc
    ON cnc.notification_channel_id = nc.id
WHERE cnc.component_id = $1 AND nc.disabled_at IS NULL;

-- name: MarkNotificationChannelSent :exec
UPDATE notification_channels
SET last_sent_at = @at,
    updated_at   = @at
WHERE id = @id;

-- name: MarkNotificationChannelFailure :exec
UPDATE notification_channels
SET failures_24h    = CASE
    WHEN last_failure_at IS NULL OR last_failure_at < @cutoff_at THEN 1
    ELSE failures_24h + 1
END,
    last_failure_at = @at,
    updated_at      = @at
WHERE id = @id;

-- Spec 095: the four finders above are the send paths; they skip disabled
-- channels. Enable clears the disabled state; the erasure rewrites a channel's
-- configuration inside its transaction after re-reading it.

-- name: EnableNotificationChannel :execrows
UPDATE notification_channels
SET disabled_at     = NULL,
    disabled_reason = NULL,
    updated_at      = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id);

-- name: RewriteNotificationChannelConfig :execrows
UPDATE notification_channels
SET config          = sqlc.arg(config),
    disabled_at     = sqlc.narg(disabled_at),
    disabled_reason = sqlc.narg(disabled_reason),
    updated_at      = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id);
