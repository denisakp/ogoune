-- name: CreateIncidentUpdate :exec
INSERT INTO incident_updates (
    id, incident_id, status, message, posted_by, posted_at, created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?
);

-- name: GetIncidentUpdate :one
SELECT id, incident_id, status, message, posted_by, posted_at, created_at, updated_at
FROM incident_updates
WHERE id = ?;

-- name: ListIncidentUpdates :many
SELECT id, incident_id, status, message, posted_by, posted_at, created_at, updated_at
FROM incident_updates
WHERE incident_id = ?
ORDER BY posted_at DESC;

-- name: UpdateIncidentUpdate :exec
UPDATE incident_updates
SET status = ?,
    message = ?,
    posted_at = ?,
    updated_at = ?
WHERE id = ?;

-- name: DeleteIncidentUpdate :exec
DELETE FROM incident_updates WHERE id = ?;
