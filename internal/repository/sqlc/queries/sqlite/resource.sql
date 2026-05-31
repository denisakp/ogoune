-- PR1 of US1: CRUD without M2M and without 1-to-1 preloads.
-- Mirror of postgres/resource.sql with SQLite-specific date math for
-- FindMissedHeartbeats (strftime instead of EXTRACT EPOCH).

-- name: CreateResource :exec
INSERT INTO resources (
    id, created_at, updated_at, name, type, interval, timeout, target,
    last_checked, status, is_active, failure_count,
    ssl_expiration_date, ssl_issuer, domain_expiration_date, domain_registrar,
    component_id, confirmation_checks, confirmation_interval, expiry_alert_thresholds,
    flap_detection_enabled, flap_threshold, flap_window_seconds, flap_max_duration_minutes,
    last_status_transition, flap_started_at, reminder_interval_minutes,
    heartbeat_slug, heartbeat_interval, heartbeat_grace, last_ping_at,
    keyword, keyword_mode, protocol_type, protocol_port
)
VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?,
    ?, ?, ?, ?,
    ?, ?, ?, ?,
    ?, ?, ?, ?,
    ?, ?, ?, ?,
    ?, ?, ?,
    ?, ?, ?, ?,
    ?, ?, ?, ?
);

-- name: FindResourceByID :one
SELECT * FROM resources WHERE id = ? AND is_active = 1;

-- name: FindResourceByHeartbeatSlug :one
SELECT * FROM resources
WHERE heartbeat_slug = ?
  AND type = 'heartbeat'
  AND is_active = 1;

-- name: ListResources :many
SELECT * FROM resources
WHERE is_active = 1
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: ListActiveResources :many
SELECT * FROM resources
WHERE is_active = 1
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: ListScheduledResources :many
SELECT * FROM resources
WHERE is_active = 1
ORDER BY id ASC;

-- name: FindResourcesByComponentID :many
SELECT * FROM resources
WHERE component_id = ? AND is_active = 1
ORDER BY created_at DESC;

-- name: CountResourcesByComponentID :one
SELECT COUNT(*) FROM resources
WHERE component_id = ? AND is_active = 1;

-- name: UpdateResourceMain :execrows
UPDATE resources
SET name                      = ?2,
    type                      = ?3,
    target                    = ?4,
    interval                  = ?5,
    timeout                   = ?6,
    is_active                 = ?7,
    confirmation_checks       = ?8,
    confirmation_interval     = ?9,
    component_id              = ?10,
    expiry_alert_thresholds   = ?11,
    flap_detection_enabled    = ?12,
    flap_threshold            = ?13,
    flap_window_seconds       = ?14,
    flap_max_duration_minutes = ?15,
    reminder_interval_minutes = ?16,
    heartbeat_interval        = ?17,
    heartbeat_grace           = ?18,
    updated_at                = ?19
WHERE id = ?1;

-- name: SoftDeleteResource :execrows
UPDATE resources SET is_active = 0
WHERE id = ? AND is_active = 1;

-- name: UpdateResourceStatus :execrows
UPDATE resources SET status = ?2 WHERE id = ?1;

-- name: UpdateResourceLastPingAt :execrows
UPDATE resources SET last_ping_at = ?2
WHERE id = ?1 AND type = 'heartbeat' AND is_active = 1;

-- name: FindMissedHeartbeatsSQLite :many
SELECT * FROM resources
WHERE type = 'heartbeat'
  AND status = 'up'
  AND is_active = 1
  AND last_ping_at IS NOT NULL
  AND (CAST(strftime('%s', last_ping_at) AS INTEGER) + heartbeat_interval + heartbeat_grace) < CAST(sqlc.arg('now_unix') AS INTEGER)
ORDER BY last_ping_at ASC
LIMIT CAST(sqlc.arg('row_limit') AS INTEGER);
