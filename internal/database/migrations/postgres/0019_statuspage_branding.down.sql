ALTER TABLE status_page_settings
    DROP COLUMN IF EXISTS logo_url_light,
    DROP COLUMN IF EXISTS logo_url_dark,
    DROP COLUMN IF EXISTS favicon_url,
    DROP COLUMN IF EXISTS primary_color,
    DROP COLUMN IF EXISTS theme_overrides;
