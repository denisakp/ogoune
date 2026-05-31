-- name: CreateIncidentEventStep :exec
INSERT INTO incident_event_steps (id, created_at, updated_at, incident_id, step, message)
VALUES (?, ?, ?, ?, ?, ?);

-- name: FindIncidentEventStepByID :one
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
WHERE s.id = ?;

-- name: FindLastIncidentEventStepByIncidentAndStep :one
SELECT * FROM incident_event_steps
WHERE incident_id = ? AND step = ?
ORDER BY created_at DESC
LIMIT 1;

-- name: ListIncidentEventSteps :many
SELECT * FROM incident_event_steps
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: UpdateIncidentEventStep :execrows
UPDATE incident_event_steps
SET incident_id = ?, step = ?, message = ?, updated_at = ?
WHERE id = ?;

-- name: DeleteIncidentEventStep :execrows
DELETE FROM incident_event_steps WHERE id = ?;
