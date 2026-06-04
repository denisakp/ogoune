-- 0018: Fold custom-domain DNS state into status_page_settings.
-- The standalone `custom_domains` table (0017) was a mis-abstraction: a
-- self-hosted instance has exactly one status page, so its public domain is
-- an attribute of that page — not of an abstract org. SaaS multi-tenant will
-- carry these same fields per (org-scoped) status page row.

ALTER TABLE status_page_settings
    ADD COLUMN IF NOT EXISTS custom_domain_status      TEXT  NOT NULL DEFAULT 'pending',
    ADD COLUMN IF NOT EXISTS custom_domain_ssl_status  TEXT  NOT NULL DEFAULT 'none',
    ADD COLUMN IF NOT EXISTS custom_domain_dns_records JSONB NOT NULL DEFAULT '[]'::jsonb;

DROP TABLE IF EXISTS custom_domains;
