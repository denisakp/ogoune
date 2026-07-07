-- name: CreateIncidentDiagnostics :one
INSERT INTO incident_diagnostics (
    id, created_at, updated_at, incident_id,
    request_method, request_url, request_headers, request_timeout,
    http_status_code, response_headers, response_body, response_size,
    failure_type, error_message, error_summary,
    total_duration, dns_duration, tls_duration, first_byte_duration,
    body_truncated, body_encoded,
    keyword, keyword_mode, keyword_found,
    icmp_available, icmp_reachable, icmp_rtt_ms, root_cause_hint
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: FindIncidentDiagnosticsByIncidentID :one
SELECT * FROM incident_diagnostics WHERE incident_id = ?;

-- name: UpdateIncidentDiagnostics :execrows
UPDATE incident_diagnostics
SET request_method = ?,
    request_url = ?,
    request_headers = ?,
    request_timeout = ?,
    http_status_code = ?,
    response_headers = ?,
    response_body = ?,
    response_size = ?,
    failure_type = ?,
    error_message = ?,
    error_summary = ?,
    total_duration = ?,
    dns_duration = ?,
    tls_duration = ?,
    first_byte_duration = ?,
    body_truncated = ?,
    body_encoded = ?,
    keyword = ?,
    keyword_mode = ?,
    keyword_found = ?,
    icmp_available = ?,
    icmp_reachable = ?,
    icmp_rtt_ms = ?,
    root_cause_hint = ?,
    updated_at = ?
WHERE id = ?;

-- name: DeleteIncidentDiagnostics :execrows
DELETE FROM incident_diagnostics WHERE id = ?;
