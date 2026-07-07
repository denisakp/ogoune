-- Umami replaces Google Analytics. The legacy google_analytics_id column
-- is left in place to keep dialect parity simple; the app stops reading
-- or writing it.
ALTER TABLE status_page_settings
    ADD COLUMN IF NOT EXISTS umami_website_id TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS umami_script_url TEXT NOT NULL DEFAULT '';
