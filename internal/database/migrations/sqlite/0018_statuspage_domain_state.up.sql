-- 0018: Fold custom-domain DNS state into status_page_settings.
-- See postgres mirror for rationale. SQLite forbids multi-column ALTER and
-- has no `IF NOT EXISTS` on columns; each ADD lives on its own statement.

ALTER TABLE status_page_settings ADD COLUMN custom_domain_status      TEXT NOT NULL DEFAULT 'pending';
ALTER TABLE status_page_settings ADD COLUMN custom_domain_ssl_status  TEXT NOT NULL DEFAULT 'none';
ALTER TABLE status_page_settings ADD COLUMN custom_domain_dns_records TEXT NOT NULL DEFAULT '[]';

DROP TABLE IF EXISTS custom_domains;
