-- name: CreateNotificationEvent :exec
INSERT INTO notification_events (
    id, created_at, updated_at, incident_id, type,
    status, claim_owner, claimed_at, processed_at, last_error
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: FindNotificationEventByID :one
SELECT * FROM notification_events WHERE id = $1;

-- name: ListNotificationEvents :many
SELECT * FROM notification_events
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateNotificationEvent :execrows
UPDATE notification_events
SET incident_id = $2, type = $3, status = $4,
    claim_owner = $5, claimed_at = $6, processed_at = $7, last_error = $8,
    updated_at = $9
WHERE id = $1;

-- name: DeleteNotificationEvent :exec
DELETE FROM notification_events WHERE id = $1;

-- name: FindPendingNotificationEvents :many
SELECT * FROM notification_events
WHERE status = 'pending'
  AND type IN ('down', 'up')
ORDER BY created_at ASC
LIMIT $1 OFFSET $2;

-- name: ClaimNotificationEventForUpdate :one
SELECT id FROM notification_events
WHERE id = $1
  AND status = 'pending'
  AND (claim_owner IS NULL OR claim_owner = '')
FOR UPDATE SKIP LOCKED;

-- name: UpdateNotificationEventClaim :exec
UPDATE notification_events
SET claim_owner = $2, claimed_at = $3
WHERE id = $1;

-- name: MarkNotificationEventTerminal :execrows
UPDATE notification_events
SET status = $2, processed_at = $3, last_error = $4,
    claim_owner = NULL, claimed_at = NULL
WHERE id = $1;
