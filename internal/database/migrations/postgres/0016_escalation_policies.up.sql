-- 0016: Escalation policies + steps with priority-based first-match resolution (FR-026a).
-- Single-tenant CE: no org_id. Multi-tenant follow-up may add it later.
-- Lower `priority` = higher precedence. Partial unique index ensures no priority collision among active policies.
CREATE TABLE IF NOT EXISTS escalation_policies (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    scope_kind  TEXT NOT NULL,
    scope_value TEXT NOT NULL,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    priority    INTEGER NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_escalation_priority_active
    ON escalation_policies(priority) WHERE is_active = TRUE;
CREATE INDEX IF NOT EXISTS idx_escalation_policies_active ON escalation_policies(is_active, priority);

CREATE TABLE IF NOT EXISTS escalation_steps (
    id            TEXT PRIMARY KEY,
    policy_id     TEXT NOT NULL REFERENCES escalation_policies(id) ON DELETE CASCADE,
    step_order    INTEGER NOT NULL,
    delay_minutes INTEGER NOT NULL,
    channel_ids   JSONB NOT NULL DEFAULT '[]'::jsonb
);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_escalation_steps_order
    ON escalation_steps(policy_id, step_order);
