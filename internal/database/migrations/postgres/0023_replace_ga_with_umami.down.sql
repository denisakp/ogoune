ALTER TABLE status_page_settings
    DROP COLUMN IF EXISTS umami_website_id,
    DROP COLUMN IF EXISTS umami_script_url;
