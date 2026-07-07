-- name: CreateNotificationEvent :exec
INSERT INTO notification_events (
    id, created_at, updated_at, incident_id, type,
    status, claim_owner, claimed_at, processed_at, last_error
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: FindNotificationEventByID :one
SELECT * FROM notification_events WHERE id = ?;

-- name: ListNotificationEvents :many
SELECT * FROM notification_events
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: UpdateNotificationEvent :execrows
UPDATE notification_events
SET incident_id = ?, type = ?, status = ?,
    claim_owner = ?, claimed_at = ?, processed_at = ?, last_error = ?,
    updated_at = ?
WHERE id = ?;

-- name: DeleteNotificationEvent :exec
DELETE FROM notification_events WHERE id = ?;

-- name: FindPendingNotificationEvents :many
SELECT * FROM notification_events
WHERE status = 'pending'
  AND type IN ('down', 'up')
ORDER BY created_at ASC
LIMIT ? OFFSET ?;

-- name: ClaimNotificationEvent :execrows
UPDATE notification_events
SET claim_owner = ?, claimed_at = ?
WHERE id = ?
  AND status = 'pending'
  AND (claim_owner IS NULL OR claim_owner = '');

-- name: MarkNotificationEventTerminal :execrows
UPDATE notification_events
SET status = ?, processed_at = ?, last_error = ?,
    claim_owner = NULL, claimed_at = NULL
WHERE id = ?;
