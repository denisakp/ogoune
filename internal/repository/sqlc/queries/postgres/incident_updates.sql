-- name: CreateIncidentUpdate :exec
INSERT INTO incident_updates (
    id, incident_id, status, message, posted_by, posted_at, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
);

-- name: GetIncidentUpdate :one
SELECT id, incident_id, status, message, posted_by, posted_at, created_at, updated_at
FROM incident_updates
WHERE id = $1;

-- name: ListIncidentUpdates :many
SELECT id, incident_id, status, message, posted_by, posted_at, created_at, updated_at
FROM incident_updates
WHERE incident_id = $1
ORDER BY posted_at DESC;

-- name: UpdateIncidentUpdate :exec
UPDATE incident_updates
SET status = $2,
    message = $3,
    posted_at = $4,
    updated_at = $5
WHERE id = $1;

-- name: DeleteIncidentUpdate :exec
DELETE FROM incident_updates WHERE id = $1;
