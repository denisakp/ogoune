-- 0019: Status page branding fields — spec 060 FR-013..FR-018.
-- Adds 5 columns on status_page_settings to support logo (light + dark), favicon,
-- primary color, and a sanitized theme variable override map. All default to
-- empty / default brand so existing rows render with the built-in Ogoune brand.

ALTER TABLE status_page_settings
    ADD COLUMN IF NOT EXISTS logo_url_light  TEXT  NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS logo_url_dark   TEXT  NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS favicon_url     TEXT  NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS primary_color   TEXT  NOT NULL DEFAULT '#4f46e5',
    ADD COLUMN IF NOT EXISTS theme_overrides JSONB NOT NULL DEFAULT '{}'::jsonb;
