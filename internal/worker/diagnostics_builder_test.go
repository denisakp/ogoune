package worker

import (
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// --- Database health snapshot (spec 088, US2) ---

func i64p(v int64) *int64     { return &v }
func f64p(v float64) *float64 { return &v }

// T033 -- health is flattened into the four columns exactly as the keyword block
// is, and every field survives the trip.
func TestBuildIncidentDiagnostics_FlattensDatabaseHealth(t *testing.T) {
	result := domain.CheckResult{
		DatabaseHealth: &domain.DatabaseHealth{
			ConnectionsActive:     i64p(200),
			ConnectionsMax:        i64p(200),
			LongestQuerySeconds:   f64p(890.2),
			ReplicationLagSeconds: f64p(4.5),
		},
	}

	diag := BuildIncidentDiagnostics("inc-db", result, &domain.Resource{Timeout: 10})

	require.NotNil(t, diag.DbConnectionsActive)
	assert.Equal(t, int64(200), *diag.DbConnectionsActive)
	require.NotNil(t, diag.DbConnectionsMax)
	assert.Equal(t, int64(200), *diag.DbConnectionsMax)
	require.NotNil(t, diag.DbLongestQuerySeconds)
	assert.InDelta(t, 890.2, *diag.DbLongestQuerySeconds, 0.001)
	require.NotNil(t, diag.DbReplicationLagSeconds)
	assert.InDelta(t, 4.5, *diag.DbReplicationLagSeconds, 0.001)
}

// T033 -- a partial collection carries only what it had. A withheld field stays
// null on the incident too: the snapshot must not invent what the check could not
// read.
func TestBuildIncidentDiagnostics_PartialDatabaseHealthStaysPartial(t *testing.T) {
	result := domain.CheckResult{
		DatabaseHealth: &domain.DatabaseHealth{
			ConnectionsActive: i64p(12),
			ConnectionsMax:    i64p(100),
			PrivilegeLimited:  true,
		},
	}

	diag := BuildIncidentDiagnostics("inc-partial", result, &domain.Resource{Timeout: 10})

	assert.NotNil(t, diag.DbConnectionsActive)
	assert.Nil(t, diag.DbLongestQuerySeconds, "a withheld field is null on the incident too")
	assert.Nil(t, diag.DbReplicationLagSeconds)
}

// T034 -- an incident on a monitor that is not a database leaves all four null,
// and nothing else about its diagnostics changes.
func TestBuildIncidentDiagnostics_NonDatabaseIncidentHasNoHealth(t *testing.T) {
	cause := domain.HTTPInvalidStatusCode
	result := domain.CheckResult{
		Cause:           &cause,
		HTTPStatusCode:  500,
		ResponseHeaders: map[string]string{"Content-Type": "text/html"},
	}

	diag := BuildIncidentDiagnostics("inc-http", result, &domain.Resource{Timeout: 10})

	assert.Nil(t, diag.DbConnectionsActive)
	assert.Nil(t, diag.DbConnectionsMax)
	assert.Nil(t, diag.DbLongestQuerySeconds)
	assert.Nil(t, diag.DbReplicationLagSeconds)
	// Everything that was there before is still there.
	assert.Equal(t, "text/html", diag.ResponseHeaders["Content-Type"])
	assert.Equal(t, 500, diag.HTTPStatusCode)
}

// T035 -- the snapshot is a snapshot. Building diagnostics from a later check
// with different figures produces a different record; it does not reach back and
// change the one already written.
//
// This is what makes the incident columns worth having next to resource_health:
// one answers "how was it when this broke" and must never move, the other
// answers "how is it right now" and is overwritten constantly.
func TestBuildIncidentDiagnostics_SnapshotIsImmutable(t *testing.T) {
	atFailure := domain.CheckResult{
		DatabaseHealth: &domain.DatabaseHealth{
			ConnectionsActive: i64p(200),
			ConnectionsMax:    i64p(200),
		},
	}
	first := BuildIncidentDiagnostics("inc-frozen", atFailure, &domain.Resource{Timeout: 10})

	// The database recovers; a later check sees a healthy figure.
	afterRecovery := domain.CheckResult{
		DatabaseHealth: &domain.DatabaseHealth{
			ConnectionsActive: i64p(3),
			ConnectionsMax:    i64p(200),
		},
	}
	second := BuildIncidentDiagnostics("inc-frozen", afterRecovery, &domain.Resource{Timeout: 10})

	require.NotNil(t, first.DbConnectionsActive)
	assert.Equal(t, int64(200), *first.DbConnectionsActive,
		"the record taken at failure still holds the failure's figures")
	require.NotNil(t, second.DbConnectionsActive)
	assert.Equal(t, int64(3), *second.DbConnectionsActive)
	assert.NotSame(t, first.DbConnectionsActive, second.DbConnectionsActive,
		"the two records share no state")
}
