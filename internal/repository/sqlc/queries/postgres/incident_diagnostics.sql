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
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
        $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28)
RETURNING *;

-- name: FindIncidentDiagnosticsByIncidentID :one
SELECT * FROM incident_diagnostics WHERE incident_id = $1;

-- name: UpdateIncidentDiagnostics :execrows
UPDATE incident_diagnostics
SET request_method = $2,
    request_url = $3,
    request_headers = $4,
    request_timeout = $5,
    http_status_code = $6,
    response_headers = $7,
    response_body = $8,
    response_size = $9,
    failure_type = $10,
    error_message = $11,
    error_summary = $12,
    total_duration = $13,
    dns_duration = $14,
    tls_duration = $15,
    first_byte_duration = $16,
    body_truncated = $17,
    body_encoded = $18,
    keyword = $19,
    keyword_mode = $20,
    keyword_found = $21,
    icmp_available = $22,
    icmp_reachable = $23,
    icmp_rtt_ms = $24,
    root_cause_hint = $25,
    updated_at = $26
WHERE id = $1;

-- name: DeleteIncidentDiagnostics :execrows
DELETE FROM incident_diagnostics WHERE id = $1;
