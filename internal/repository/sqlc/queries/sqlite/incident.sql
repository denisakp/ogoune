-- US2 of spec 048: incident_repository sqlc migration.
-- Mirror of postgres/incident.sql with SQLite-specific GetIncidentStats
-- (two correlated sub-queries; SQLite lacks COUNT(*) FILTER and the
-- one-pass CTE form).

-- name: CreateIncident :exec
INSERT INTO incidents (
    id, created_at, updated_at, resource_id, cause,
    resolved_at, started_at, details
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: FindIncidentByID :one
SELECT * FROM incidents WHERE id = ?;

-- name: ListIncidents :many
SELECT * FROM incidents
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: UpdateIncident :execrows
UPDATE incidents
SET resource_id = ?2,
    cause       = ?3,
    resolved_at = ?4,
    started_at  = ?5,
    details     = ?6,
    updated_at  = ?7
WHERE id = ?1;

-- name: DeleteIncident :execrows
DELETE FROM incidents WHERE id = ?;

-- name: FindUnresolvedIncidents :many
SELECT * FROM incidents
WHERE resolved_at IS NULL
ORDER BY started_at DESC
LIMIT ? OFFSET ?;

-- name: FindIncidentsByResource :many
SELECT * FROM incidents
WHERE resource_id = ?
ORDER BY started_at DESC
LIMIT ? OFFSET ?;

-- name: FindActiveIncidentByResourceID :one
SELECT * FROM incidents
WHERE resource_id = ? AND resolved_at IS NULL
ORDER BY started_at DESC
LIMIT 1;

-- name: HasActiveIncident :one
SELECT EXISTS(SELECT 1 FROM incidents WHERE resolved_at IS NULL) AS has_active;

-- name: FindLastResolvedIncident :one
SELECT * FROM incidents
WHERE resolved_at IS NOT NULL
ORDER BY resolved_at DESC
LIMIT 1;

-- name: CountIncidentsByResourceID :one
SELECT COUNT(*) FROM incidents WHERE resource_id = ?;

-- name: GetIncidentStatsSQLite :one
SELECT
    (SELECT COUNT(*)                       FROM incidents i1 WHERE i1.started_at >= sqlc.arg('since')) AS total_incidents,
    (SELECT COUNT(DISTINCT i2.resource_id) FROM incidents i2 WHERE i2.started_at >= sqlc.arg('since')) AS affected_monitors;

-- name: ListIncidentDiagnosticsByIncidentIDs :many
SELECT * FROM incident_diagnostics
WHERE incident_id IN (sqlc.slice('incident_ids'));

-- name: FindIncidentsByIDs :many
SELECT * FROM incidents WHERE id IN (sqlc.slice('ids'));
