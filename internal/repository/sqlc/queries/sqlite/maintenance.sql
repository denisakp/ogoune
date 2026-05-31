-- name: CreateMaintenance :exec
INSERT INTO maintenances (
    id, created_at, updated_at, title, description, strategy, status,
    start_at, end_at, cron_expr, window_minutes, timezone,
    effective_from, effective_until, started_at, ended_at
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: FindMaintenanceByID :one
SELECT * FROM maintenances WHERE id = ?;

-- name: ListMaintenancesAll :many
SELECT * FROM maintenances ORDER BY created_at DESC LIMIT ? OFFSET ?;

-- name: ListMaintenancesByStatus :many
SELECT * FROM maintenances WHERE status = ? ORDER BY created_at DESC LIMIT ? OFFSET ?;

-- name: UpdateMaintenance :exec
UPDATE maintenances
SET title = ?, description = ?, strategy = ?, status = ?,
    start_at = ?, end_at = ?, cron_expr = ?, window_minutes = ?, timezone = ?,
    effective_from = ?, effective_until = ?,
    started_at = ?, ended_at = ?, updated_at = ?
WHERE id = ?;

-- name: DeleteMaintenance :execrows
DELETE FROM maintenances WHERE id = ?;

-- name: FindActiveMaintenancesForResource :many
SELECT DISTINCT m.* FROM maintenances m
JOIN maintenance_resources mr ON mr.maintenance_id = m.id
WHERE mr.resource_id = ?
  AND (
    m.status = 'active'
    OR (m.strategy = 'one_time' AND m.status = 'scheduled' AND m.start_at <= ? AND m.end_at >= ?)
  );

-- name: LinkMaintenanceResource :exec
INSERT INTO maintenance_resources (maintenance_id, resource_id) VALUES (?, ?) ON CONFLICT DO NOTHING;

-- name: UnlinkMaintenanceResource :exec
DELETE FROM maintenance_resources WHERE maintenance_id = ? AND resource_id = ?;

-- name: ListMaintenanceResourceIDsByMaintenanceID :many
SELECT resource_id FROM maintenance_resources WHERE maintenance_id = ?;
