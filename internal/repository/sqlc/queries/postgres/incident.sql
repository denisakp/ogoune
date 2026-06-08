-- US2 of spec 048: incident_repository sqlc migration. No M2M;
-- controlled-N+1 preload for Resource and IncidentDiagnostics
-- (1-to-1, FK uniqueIndex on incident_id).

-- name: CreateIncident :exec
INSERT INTO incidents (
    id, created_at, updated_at, resource_id, cause,
    resolved_at, started_at, details
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: FindIncidentByID :one
SELECT * FROM incidents WHERE id = $1;

-- name: ListIncidents :many
SELECT * FROM incidents
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateIncident :execrows
UPDATE incidents
SET resource_id = $2,
    cause       = $3,
    resolved_at = $4,
    started_at  = $5,
    details     = $6,
    updated_at  = $7
WHERE id = $1;

-- name: DeleteIncident :execrows
DELETE FROM incidents WHERE id = $1;

-- name: FindUnresolvedIncidents :many
SELECT * FROM incidents
WHERE resolved_at IS NULL
ORDER BY started_at DESC
LIMIT $1 OFFSET $2;

-- name: FindIncidentsByResource :many
SELECT * FROM incidents
WHERE resource_id = $1
ORDER BY started_at DESC
LIMIT $2 OFFSET $3;

-- name: FindActiveIncidentByResourceID :one
SELECT * FROM incidents
WHERE resource_id = $1 AND resolved_at IS NULL
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
SELECT COUNT(*) FROM incidents WHERE resource_id = $1;

-- name: GetIncidentStatsPG :one
-- Two aggregated counts in a single round-trip. CTE keeps the row scan
-- to one pass over the window.
WITH window_inc AS (
    SELECT resource_id FROM incidents WHERE started_at >= $1
)
SELECT
    COUNT(*)::bigint                       AS total_incidents,
    COUNT(DISTINCT resource_id)::bigint    AS affected_monitors
FROM window_inc;

-- name: CountIncidentsPerResourceSince :many
-- One round-trip count grouped by resource. Used by the list path to enrich
-- each resource with its incident count over a sliding window (e.g. 30d).
SELECT resource_id, COUNT(*)::bigint AS incident_count
FROM incidents
WHERE started_at >= $1 AND resource_id = ANY($2::text[])
GROUP BY resource_id;

-- name: ListIncidentDiagnosticsByIncidentIDs :many
SELECT * FROM incident_diagnostics
WHERE incident_id = ANY($1::text[]);

-- name: FindIncidentsByIDs :many
SELECT * FROM incidents WHERE id = ANY($1::text[]);
