-- name: GetStatusPageSettings :one
SELECT * FROM status_page_settings LIMIT 1;

-- name: CreateStatusPageSettings :exec
INSERT INTO status_page_settings (
    id, name, homepage_url, custom_domain, google_analytics_id,
    enable_details_page, show_uptime_percentage, hide_paused_monitors,
    show_incident_history,
    custom_domain_status, custom_domain_ssl_status, custom_domain_dns_records,
    created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14);

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
    updated_at = $13
WHERE id = $1;
