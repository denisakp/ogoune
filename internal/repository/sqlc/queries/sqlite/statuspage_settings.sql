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
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateStatusPageSettings :exec
UPDATE status_page_settings
SET name = ?,
    homepage_url = ?,
    custom_domain = ?,
    google_analytics_id = ?,
    enable_details_page = ?,
    show_uptime_percentage = ?,
    hide_paused_monitors = ?,
    show_incident_history = ?,
    custom_domain_status = ?,
    custom_domain_ssl_status = ?,
    custom_domain_dns_records = ?,
    updated_at = ?
WHERE id = ?;
