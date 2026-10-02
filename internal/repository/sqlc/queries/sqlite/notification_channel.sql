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
SELECT * FROM notification_channels WHERE type = ? AND disabled_at IS NULL;

-- name: FindDefaultNotificationChannels :many
SELECT * FROM notification_channels WHERE enabled_by_default = 1 AND disabled_at IS NULL;

-- name: FindNotificationChannelsByResourceID :many
SELECT nc.* FROM notification_channels nc
JOIN resource_notification_channels rnc
    ON rnc.notification_channel_id = nc.id
WHERE rnc.resource_id = ? AND nc.disabled_at IS NULL;

-- name: FindNotificationChannelsByComponentID :many
SELECT nc.* FROM notification_channels nc
JOIN component_notification_channels cnc
    ON cnc.notification_channel_id = nc.id
WHERE cnc.component_id = ? AND nc.disabled_at IS NULL;

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
