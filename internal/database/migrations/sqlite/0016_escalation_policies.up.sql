-- 0016: Escalation policies + steps with priority-based first-match resolution (FR-026a).
-- Single-tenant CE: no org_id. Multi-tenant follow-up may add it later.
-- Lower `priority` = higher precedence. Partial unique index enforces no priority
-- collision among active policies (SQLite partial indexes supported since 3.8.0).
CREATE TABLE IF NOT EXISTS escalation_policies (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    scope_kind  TEXT NOT NULL,
    scope_value TEXT NOT NULL,
    is_active   INTEGER NOT NULL DEFAULT 1,
    priority    INTEGER NOT NULL,
    created_at  DATETIME NOT NULL,
    updated_at  DATETIME NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_escalation_policies_active ON escalation_policies(is_active, priority);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_escalation_priority_active
    ON escalation_policies(priority) WHERE is_active = 1;
CREATE TABLE IF NOT EXISTS escalation_steps (
    id            TEXT PRIMARY KEY,
    policy_id     TEXT NOT NULL,
    step_order    INTEGER NOT NULL,
    delay_minutes INTEGER NOT NULL,
    channel_ids   TEXT NOT NULL DEFAULT '[]',
    FOREIGN KEY (policy_id) REFERENCES escalation_policies(id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_escalation_steps_order
    ON escalation_steps(policy_id, step_order);
