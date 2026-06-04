ALTER TABLE status_page_settings
    DROP COLUMN IF EXISTS custom_domain_status,
    DROP COLUMN IF EXISTS custom_domain_ssl_status,
    DROP COLUMN IF EXISTS custom_domain_dns_records;

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
