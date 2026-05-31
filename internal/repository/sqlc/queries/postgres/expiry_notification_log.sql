-- name: CountExpiryNotificationLogsByKey :one
SELECT COUNT(*) FROM expiry_notification_logs
WHERE resource_id = $1 AND expiry_type = $2 AND threshold = $3;

-- name: CreateExpiryNotificationLog :exec
INSERT INTO expiry_notification_logs (
    id, resource_id, expiry_type, threshold, sent_at, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: DeleteExpiryNotificationLogsByResourceIDAndType :exec
DELETE FROM expiry_notification_logs
WHERE resource_id = $1 AND expiry_type = $2;

-- name: DeleteExpiryNotificationLogsOlderThan :exec
DELETE FROM expiry_notification_logs
WHERE sent_at < $1;
