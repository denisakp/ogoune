-- name: GetStatusPageSettings :one
SELECT * FROM status_page_settings LIMIT 1;

-- name: CreateStatusPageSettings :exec
INSERT INTO status_page_settings (
    id, name, homepage_url, custom_domain, google_analytics_id,
    enable_details_page, show_uptime_percentage, hide_paused_monitors,
    show_incident_history, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);

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
    updated_at = $10
WHERE id = $1;
