package worker

import (
	"testing"
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestBuildIncidentDiagnostics_PersistsResponseHeaders(t *testing.T) {
	cause := domain.HTTPInvalidStatusCode
	result := domain.CheckResult{
		Cause:           &cause,
		ResponseHeaders: map[string]string{"Content-Type": "application/json"},
	}
	resource := &domain.Resource{Timeout: 10}

	diag := BuildIncidentDiagnostics("inc-1", result, resource)

	assert.Equal(t, "application/json", diag.ResponseHeaders["Content-Type"])
}

func TestBuildIncidentDiagnostics_NilResponseHeaders_DefaultsToEmptyMap(t *testing.T) {
	result := domain.CheckResult{}
	resource := &domain.Resource{Timeout: 10}

	diag := BuildIncidentDiagnostics("inc-3", result, resource)

	assert.NotNil(t, diag.ResponseHeaders)
	assert.Equal(t, 0, len(diag.ResponseHeaders))
}

func TestBuildIncidentDiagnostics_RemovesAuthorizationHeader(t *testing.T) {
	result := domain.CheckResult{
		RequestHeaders: map[string]string{
			"Authorization": "Bearer secret",
			"X-Trace-ID":    "abc-123",
		},
		ResponseTime: time.Second,
	}
	resource := &domain.Resource{Timeout: 10}

	diag := BuildIncidentDiagnostics("inc-2", result, resource)

	_, found := diag.RequestHeaders["Authorization"]
	assert.False(t, found)
	assert.Equal(t, "abc-123", diag.RequestHeaders["X-Trace-ID"])
}
