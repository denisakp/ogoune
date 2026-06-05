-- name: GetStatusPageSettings :one
SELECT * FROM status_page_settings LIMIT 1;

-- name: CreateStatusPageSettings :exec
INSERT INTO status_page_settings (
    id, name, homepage_url, custom_domain, google_analytics_id,
    enable_details_page, show_uptime_percentage, hide_paused_monitors,
    show_incident_history,
    custom_domain_status, custom_domain_ssl_status, custom_domain_dns_records,
    logo_url_light, logo_url_dark, favicon_url, primary_color, theme_overrides,
    created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19);

-- name: UpdateStatusPageSettings :exec
UPDATE status_page_settings
SET name = $2,
    homepage_url = $3,
    custom_domain = $4,
    google_analytics_id = $5,
    enable_details_page = $6,
    show_uptime_percentage = $7,
    hide_paused_monitors = $8,
    show_incident_history = $9,
    custom_domain_status = $10,
    custom_domain_ssl_status = $11,
    custom_domain_dns_records = $12,
    logo_url_light = $13,
    logo_url_dark = $14,
    favicon_url = $15,
    primary_color = $16,
    theme_overrides = $17,
    updated_at = $18
WHERE id = $1;
