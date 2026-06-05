-- 0017: Multi-domain custom domains (Q4 clarification, FR-028).
-- Single-tenant CE: no org_id. `status_page_id` nullable until PRD 008 ships status pages model.
-- Legacy `status_page_settings.custom_domain` (single field) remains for backward compat; new table is the source of truth going forward.
CREATE TABLE IF NOT EXISTS custom_domains (
    id              TEXT PRIMARY KEY,
    status_page_id  TEXT,
    domain          TEXT NOT NULL UNIQUE,
    status          TEXT NOT NULL DEFAULT 'pending',
    dns_records     JSONB NOT NULL DEFAULT '[]'::jsonb,
    ssl_status      TEXT NOT NULL DEFAULT 'none',
    created_at      TIMESTAMPTZ NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_custom_domains_status_page ON custom_domains(status_page_id);
