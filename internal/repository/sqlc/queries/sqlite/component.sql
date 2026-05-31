-- name: CreateComponent :exec
INSERT INTO components (
    id, created_at, updated_at, name, description,
    last_notification_status, grouping_window_seconds
)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: FindComponentByID :one
SELECT * FROM components WHERE id = ?;

-- name: ListComponents :many
SELECT * FROM components
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: UpdateComponent :exec
UPDATE components
SET name = ?,
    description = ?,
    updated_at = ?
WHERE id = ?;

-- name: DeleteComponent :execrows
DELETE FROM components WHERE id = ?;

-- name: UpdateComponentLastNotificationStatus :execrows
UPDATE components
SET last_notification_status = ?
WHERE id = ?;

-- name: ListResourcesByComponentID :many
SELECT * FROM resources WHERE component_id = ? ORDER BY created_at;
