-- name: CreateComponent :exec
INSERT INTO components (
    id, created_at, updated_at, name, description,
    last_notification_status, grouping_window_seconds
)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: FindComponentByID :one
SELECT * FROM components WHERE id = $1;

-- name: ListComponents :many
SELECT * FROM components
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateComponent :exec
UPDATE components
SET name = $2,
    description = $3,
    updated_at = $4
WHERE id = $1;

-- name: DeleteComponent :execrows
DELETE FROM components WHERE id = $1;

-- name: UpdateComponentLastNotificationStatus :execrows
UPDATE components
SET last_notification_status = $2
WHERE id = $1;

-- name: ListResourcesByComponentID :many
SELECT * FROM resources WHERE component_id = $1 ORDER BY created_at;
