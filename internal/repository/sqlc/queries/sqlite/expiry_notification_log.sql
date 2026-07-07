-- name: CountExpiryNotificationLogsByKey :one
SELECT COUNT(*) FROM expiry_notification_logs
WHERE resource_id = ? AND expiry_type = ? AND threshold = ?;

-- name: CreateExpiryNotificationLog :exec
INSERT INTO expiry_notification_logs (
    id, resource_id, expiry_type, threshold, sent_at, created_at, updated_at
)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: DeleteExpiryNotificationLogsByResourceIDAndType :exec
DELETE FROM expiry_notification_logs
WHERE resource_id = ? AND expiry_type = ?;

-- name: DeleteExpiryNotificationLogsOlderThan :exec
DELETE FROM expiry_notification_logs
WHERE sent_at < ?;
