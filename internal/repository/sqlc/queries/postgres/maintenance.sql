-- name: CreateMaintenance :exec
INSERT INTO maintenances (
    id, created_at, updated_at, title, description, strategy, status,
    start_at, end_at, cron_expr, window_minutes, timezone,
    effective_from, effective_until, started_at, ended_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16);

-- name: FindMaintenanceByID :one
SELECT * FROM maintenances WHERE id = $1;

-- name: ListMaintenancesAll :many
SELECT * FROM maintenances ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: ListMaintenancesByStatus :many
SELECT * FROM maintenances WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: UpdateMaintenance :exec
UPDATE maintenances
SET title = $2, description = $3, strategy = $4, status = $5,
    start_at = $6, end_at = $7, cron_expr = $8, window_minutes = $9, timezone = $10,
    effective_from = $11, effective_until = $12,
    started_at = $13, ended_at = $14, updated_at = $15
WHERE id = $1;

-- name: DeleteMaintenance :execrows
DELETE FROM maintenances WHERE id = $1;

-- name: FindActiveMaintenancesForResource :many
SELECT DISTINCT m.* FROM maintenances m
JOIN maintenance_resources mr ON mr.maintenance_id = m.id
WHERE mr.resource_id = $1
  AND (
    m.status = 'active'
    OR (m.strategy = 'one_time' AND m.status = 'scheduled' AND m.start_at <= $2 AND m.end_at >= $3)
  );

-- name: LinkMaintenanceResource :exec
INSERT INTO maintenance_resources (maintenance_id, resource_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: UnlinkMaintenanceResource :exec
DELETE FROM maintenance_resources WHERE maintenance_id = $1 AND resource_id = $2;

-- name: ListMaintenanceResourceIDsByMaintenanceID :many
SELECT resource_id FROM maintenance_resources WHERE maintenance_id = $1;
