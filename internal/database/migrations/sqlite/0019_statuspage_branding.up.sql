-- 0019: Status page branding fields — spec 060 FR-013..FR-018.
-- SQLite forbids multi-column ALTER and `IF NOT EXISTS` on columns; each
-- ADD lives on its own statement. theme_overrides stored as TEXT (JSON).

ALTER TABLE status_page_settings ADD COLUMN logo_url_light  TEXT NOT NULL DEFAULT '';
ALTER TABLE status_page_settings ADD COLUMN logo_url_dark   TEXT NOT NULL DEFAULT '';
ALTER TABLE status_page_settings ADD COLUMN favicon_url     TEXT NOT NULL DEFAULT '';
ALTER TABLE status_page_settings ADD COLUMN primary_color   TEXT NOT NULL DEFAULT '#4f46e5';
ALTER TABLE status_page_settings ADD COLUMN theme_overrides TEXT NOT NULL DEFAULT '{}';
