-- name: CreateIncidentEventStep :exec
INSERT INTO incident_event_steps (id, created_at, updated_at, incident_id, step, message)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: FindIncidentEventStepByID :one
-- Single JOIN preload (Clarification Q1) — incident embedded into event step.
SELECT
    s.id, s.created_at, s.updated_at, s.incident_id, s.step, s.message,
    i.created_at AS i_created_at,
    i.updated_at AS i_updated_at,
    i.resource_id,
    i.cause,
    i.resolved_at,
    i.started_at,
    i.details
FROM incident_event_steps s
JOIN incidents i ON i.id = s.incident_id
WHERE s.id = $1;

-- name: FindLastIncidentEventStepByIncidentAndStep :one
SELECT * FROM incident_event_steps
WHERE incident_id = $1 AND step = $2
ORDER BY created_at DESC
LIMIT 1;

-- name: ListIncidentEventSteps :many
SELECT * FROM incident_event_steps
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateIncidentEventStep :execrows
UPDATE incident_event_steps
SET incident_id = $2, step = $3, message = $4, updated_at = $5
WHERE id = $1;

-- name: DeleteIncidentEventStep :execrows
DELETE FROM incident_event_steps WHERE id = $1;
