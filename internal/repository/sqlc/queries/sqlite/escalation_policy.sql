-- name: CreateEscalationPolicy :exec
INSERT INTO escalation_policies (id, name, scope_kind, scope_value, is_active, priority, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: FindEscalationPolicyByID :one
SELECT * FROM escalation_policies WHERE id = ?;

-- name: ListEscalationPolicies :many
SELECT * FROM escalation_policies ORDER BY priority ASC, created_at ASC;

-- name: UpdateEscalationPolicy :execrows
UPDATE escalation_policies
SET name = ?, scope_kind = ?, scope_value = ?, is_active = ?, priority = ?, updated_at = ?
WHERE id = ?;

-- name: SetEscalationPolicyPriority :execrows
UPDATE escalation_policies
SET priority = ?, updated_at = ?
WHERE id = ?;

-- name: DeleteEscalationPolicy :execrows
DELETE FROM escalation_policies WHERE id = ?;

-- name: NextEscalationPriority :one
SELECT COALESCE(MAX(priority), 0) + 1 AS next FROM escalation_policies WHERE is_active = 1;

-- name: CreateEscalationStep :exec
INSERT INTO escalation_steps (id, policy_id, step_order, delay_minutes, channel_ids)
VALUES (?, ?, ?, ?, ?);

-- name: ListEscalationStepsByPolicy :many
SELECT * FROM escalation_steps WHERE policy_id = ? ORDER BY step_order ASC;

-- name: DeleteEscalationStepsByPolicy :exec
DELETE FROM escalation_steps WHERE policy_id = ?;
