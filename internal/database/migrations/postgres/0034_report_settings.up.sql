-- 0034: Monthly report settings. Renumbered from 0026, which it shared with
-- 0026_report_history -- applied state is keyed on the version alone, so a
-- shared prefix makes two migrations indistinguishable and the migrator now
-- refuses to start on one.
--
-- Safe to renumber: every statement here is IF NOT EXISTS, so on a database that
-- already ran it under 0026 the new version applies as a no-op and simply
-- records itself.
-- 0026: Single-tenant: one instance-wide config row.
CREATE TABLE IF NOT EXISTS report_settings (
    id              TEXT PRIMARY KEY,
    enabled         BOOLEAN NOT NULL,
    recipient_email TEXT NOT NULL,
    schedule        TEXT NOT NULL,
    scope           TEXT NOT NULL,
    last_sent_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL
);
