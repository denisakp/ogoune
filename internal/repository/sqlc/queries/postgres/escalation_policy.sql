-- name: CreateEscalationPolicy :exec
INSERT INTO escalation_policies (id, name, scope_kind, scope_value, is_active, priority, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: FindEscalationPolicyByID :one
SELECT * FROM escalation_policies WHERE id = $1;

-- name: ListEscalationPolicies :many
SELECT * FROM escalation_policies ORDER BY priority ASC, created_at ASC;

-- name: UpdateEscalationPolicy :execrows
UPDATE escalation_policies
SET name = $2, scope_kind = $3, scope_value = $4, is_active = $5, priority = $6, updated_at = $7
WHERE id = $1;

-- name: SetEscalationPolicyPriority :execrows
UPDATE escalation_policies
SET priority = $2, updated_at = $3
WHERE id = $1;

-- name: DeleteEscalationPolicy :execrows
DELETE FROM escalation_policies WHERE id = $1;

-- name: NextEscalationPriority :one
SELECT COALESCE(MAX(priority), 0)::int + 1 AS next FROM escalation_policies WHERE is_active = TRUE;

-- name: CreateEscalationStep :exec
INSERT INTO escalation_steps (id, policy_id, step_order, delay_minutes, channel_ids)
VALUES ($1, $2, $3, $4, $5);

-- name: ListEscalationStepsByPolicy :many
SELECT * FROM escalation_steps WHERE policy_id = $1 ORDER BY step_order ASC;

-- name: DeleteEscalationStepsByPolicy :exec
DELETE FROM escalation_steps WHERE policy_id = $1;
