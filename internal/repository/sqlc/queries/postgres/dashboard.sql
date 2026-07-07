-- name: CreateDashboard :one
INSERT INTO dashboards (id, owner_id, name, scope, widgets, default_time_range, refresh_interval, visibility, created_at, updated_at)
VALUES (sqlc.arg('id'), sqlc.arg('owner_id'), sqlc.arg('name'), sqlc.arg('scope'), sqlc.arg('widgets'), sqlc.arg('default_time_range'), sqlc.arg('refresh_interval'), sqlc.arg('visibility'), sqlc.arg('created_at'), sqlc.arg('updated_at'))
RETURNING *;

-- name: FindDashboardByID :one
SELECT d.id, d.owner_id, u.name AS owner_name, d.name AS dashboard_name, d.scope, d.widgets, d.default_time_range, d.refresh_interval, d.visibility, d.created_at, d.updated_at
FROM dashboards d
JOIN users u ON u.id = d.owner_id
WHERE d.id = sqlc.arg('id');

-- name: ListDashboards :many
SELECT d.id, d.owner_id, u.name AS owner_name, d.name AS dashboard_name, d.scope, d.widgets, d.default_time_range, d.refresh_interval, d.visibility, d.created_at, d.updated_at
FROM dashboards d
JOIN users u ON u.id = d.owner_id
ORDER BY d.updated_at DESC
LIMIT sqlc.arg('lim') OFFSET sqlc.arg('off');

-- name: UpdateDashboard :execrows
UPDATE dashboards
SET name = sqlc.arg('name'), scope = sqlc.arg('scope'), widgets = sqlc.arg('widgets'), default_time_range = sqlc.arg('default_time_range'), refresh_interval = sqlc.arg('refresh_interval'), visibility = sqlc.arg('visibility'), updated_at = sqlc.arg('updated_at')
WHERE id = sqlc.arg('id');

-- name: UpdateDashboardWidgets :execrows
UPDATE dashboards
SET widgets = sqlc.arg('widgets'), updated_at = sqlc.arg('updated_at')
WHERE id = sqlc.arg('id');

-- name: DeleteDashboard :execrows
DELETE FROM dashboards WHERE id = sqlc.arg('id');
