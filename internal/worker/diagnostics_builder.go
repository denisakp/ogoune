package worker

import (
	"encoding/base64"
	"strings"

	"github.com/denisakp/ogoune/internal/domain"
)

// BuildIncidentDiagnostics constructs an IncidentDiagnostics record from a CheckResult.
// This helper captures rich diagnostic information to help users understand what went wrong.
// It sanitizes sensitive headers before storage for security.
func BuildIncidentDiagnostics(incidentID string, result domain.CheckResult, resource *domain.Resource) *domain.IncidentDiagnostics {
	diag := &domain.IncidentDiagnostics{
		IncidentID:        incidentID,
		RequestMethod:     result.RequestMethod,
		RequestURL:        result.RequestURL,
		RequestHeaders:    sanitizeHeaders(result.RequestHeaders),
		RequestTimeout:    resource.Timeout,
		HTTPStatusCode:    result.HTTPStatusCode,
		ResponseHeaders:   normalizeHeaders(result.ResponseHeaders),
		TotalDuration:     int(result.ResponseTime.Milliseconds()),
		DNSDuration:       int(result.DNSDuration.Milliseconds()),
		TLSDuration:       int(result.TLSDuration.Milliseconds()),
		FirstByteDuration: int(result.FirstByteDuration.Milliseconds()),
	}

	// Set error context if available
	if result.Cause != nil {
		diag.FailureType = string(*result.Cause)
		diag.ErrorSummary = domain.GenerateErrorSummary(diag.FailureType, result.HTTPStatusCode)
	}

	if result.ErrorMessage != "" {
		diag.ErrorMessage = result.ErrorMessage
	}

	// Populate keyword-specific fields when available
	if result.KeywordContext != nil {
		kw := result.KeywordContext.Keyword
		mode := result.KeywordContext.KeywordMode
		found := result.KeywordContext.KeywordFound
		diag.Keyword = &kw
		diag.KeywordMode = &mode
		diag.KeywordFound = &found
	}

	// For keyword monitors, use ReadBodySize as the authoritative response size
	// and propagate the 512 KB truncation flag from the strategy.
	if result.ReadBodySize > 0 {
		diag.ResponseSize = int(result.ReadBodySize)
	}
	if result.BodyTruncated && result.KeywordContext != nil {
		diag.BodyTruncated = true
	}

	// Encode response body if present (could be binary)
	if result.ResponseBody != "" {
		// Check if it looks like binary data
		if strings.Contains(result.ResponseBody, "\x00") {
			diag.ResponseBody = base64.StdEncoding.EncodeToString([]byte(result.ResponseBody))
			diag.BodyEncoded = true
		} else {
			diag.ResponseBody = result.ResponseBody
			diag.BodyEncoded = false
		}

		// For non-keyword monitors: apply the 5 KB cap and set ResponseSize from excerpt length
		if result.KeywordContext == nil {
			const maxBodySize = 5 * 1024
			if len(diag.ResponseBody) > maxBodySize {
				diag.ResponseBody = diag.ResponseBody[:maxBodySize]
				diag.BodyTruncated = true
			}
			diag.ResponseSize = len(result.ResponseBody)
		}
	}

	return diag
}

func normalizeHeaders(headers map[string]string) map[string]string {
	if headers == nil {
		return make(map[string]string)
	}

	copyHeaders := make(map[string]string, len(headers))
	for k, v := range headers {
		copyHeaders[k] = v
	}
	return copyHeaders
}

// sanitizeHeaders removes sensitive headers before storage.
// Headers like Authorization, Cookie, X-API-Key are filtered out to prevent credential leaks.
func sanitizeHeaders(headers map[string]string) map[string]string {
	if headers == nil {
		return make(map[string]string)
	}

	// List of header names that should not be stored (case-insensitive)
	blocked := map[string]bool{
		"authorization":     true,
		"cookie":            true,
		"x-api-key":         true,
		"x-auth-token":      true,
		"authorization-key": true,
		"x-access-token":    true,
		"password":          true,
		"secret":            true,
		"token":             true,
		"apikey":            true,
	}

	sanitized := make(map[string]string)
	for k, v := range headers {
		lowerKey := strings.ToLower(k)
		if !blocked[lowerKey] {
			sanitized[k] = v
		}
	}
	return sanitized
}
