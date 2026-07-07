-- SQLite cannot DROP a column inline; rebuilding the table is overkill for a
-- rollback path that's only used in dev. Leaving the columns in place is safe
-- (they're ignored by the previous code) — operators that truly need them gone
-- can recreate the schema from scratch.

CREATE TABLE IF NOT EXISTS custom_domains (
    id              TEXT PRIMARY KEY,
    status_page_id  TEXT,
    domain          TEXT NOT NULL UNIQUE,
    status          TEXT NOT NULL DEFAULT 'pending',
    dns_records     TEXT NOT NULL DEFAULT '[]',
    ssl_status      TEXT NOT NULL DEFAULT 'none',
    created_at      DATETIME NOT NULL,
    updated_at      DATETIME NOT NULL
);
