package worker

import (
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
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

func boolPtr(b bool) *bool { return &b }

func TestBuildIncidentDiagnostics_KeywordContext_Populated(t *testing.T) {
	cause := domain.KeywordNotFound
	result := domain.CheckResult{
		Cause:        &cause,
		ReadBodySize: 48291,
		KeywordContext: &domain.KeywordCheckContext{
			Keyword:      "operational",
			KeywordMode:  "contains",
			KeywordFound: false,
		},
	}
	resource := &domain.Resource{Timeout: 10}

	diag := BuildIncidentDiagnostics("inc-kw-1", result, resource)

	assert.Equal(t, "operational", *diag.Keyword)
	assert.Equal(t, "contains", *diag.KeywordMode)
	assert.Equal(t, boolPtr(false), diag.KeywordFound)
	assert.Equal(t, 48291, diag.ResponseSize)
}

func TestBuildIncidentDiagnostics_KeywordContext_Nil_FieldsZero(t *testing.T) {
	cause := domain.HTTPInvalidStatusCode
	result := domain.CheckResult{
		Cause: &cause,
	}
	resource := &domain.Resource{Timeout: 10}

	diag := BuildIncidentDiagnostics("inc-kw-2", result, resource)

	assert.Nil(t, diag.Keyword)
	assert.Nil(t, diag.KeywordMode)
	assert.Nil(t, diag.KeywordFound)
}

func TestBuildIncidentDiagnostics_BodyTruncated_PreservedFromResult(t *testing.T) {
	result := domain.CheckResult{
		ResponseBody:  "short excerpt",
		BodyTruncated: true,
		ReadBodySize:  512 * 1024,
		KeywordContext: &domain.KeywordCheckContext{
			Keyword:      "x",
			KeywordMode:  "contains",
			KeywordFound: false,
		},
	}
	resource := &domain.Resource{Timeout: 10}

	diag := BuildIncidentDiagnostics("inc-kw-3", result, resource)

	assert.True(t, diag.BodyTruncated)
	assert.Equal(t, 512*1024, diag.ResponseSize)
}

func TestBuildIncidentDiagnostics_ReadBodySize_UsedAsResponseSize(t *testing.T) {
	result := domain.CheckResult{
		ResponseBody: "abc",
		ReadBodySize: 99999,
		KeywordContext: &domain.KeywordCheckContext{
			Keyword:      "abc",
			KeywordMode:  "contains",
			KeywordFound: true,
		},
	}
	resource := &domain.Resource{Timeout: 10}

	diag := BuildIncidentDiagnostics("inc-kw-4", result, resource)

	assert.Equal(t, 99999, diag.ResponseSize)
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
