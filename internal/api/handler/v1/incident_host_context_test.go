package v1_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/domain"
)

// getIncidentBody issues GET /api/v1/incidents/{id} against a mock service and
// returns the decoded "data" object.
func getIncidentBody(t *testing.T, inc *domain.Incident) map[string]any {
	t.Helper()

	router := newIncidentRouter(&mockIncidentService{incident: inc})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/incidents/inc-1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var envelope struct {
		Data map[string]any `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
	return envelope.Data
}

func baseIncident() *domain.Incident {
	return &domain.Incident{
		Base:       domain.Base{ID: "inc-1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		ResourceID: "res-1",
		Cause:      "test failure",
		StartedAt:  time.Date(2026, 3, 10, 11, 58, 0, 0, time.UTC),
	}
}

// T017 — the field is present and fully populated when the service produced a
// context (FR-014, contract shape).
func TestIncidentHandler_Get_HostContextPresent(t *testing.T) {
	inc := baseIncident()
	inc.HostContext = &domain.HostContext{
		HostID:      "host-1",
		HostName:    "web-01",
		PeakCPUPct:  97.4,
		PeakMemPct:  99.1,
		WorstDisk:   &domain.DiskUsage{Mount: "/var", UsedPct: 96.2},
		SampleCount: 36,
		Resolution:  domain.HostContextFull,
		WindowFrom:  time.Date(2026, 3, 10, 11, 53, 0, 0, time.UTC),
		WindowTo:    time.Date(2026, 3, 10, 11, 59, 0, 0, time.UTC),
	}

	data := getIncidentBody(t, inc)
	raw, ok := data["host_context"]
	require.True(t, ok, "the field is always serialised, present or null")
	hc, ok := raw.(map[string]any)
	require.True(t, ok, "host_context is an object when a context exists")

	assert.Equal(t, "host-1", hc["host_id"])
	assert.Equal(t, "web-01", hc["host_name"])
	assert.InDelta(t, 97.4, hc["peak_cpu_pct"], 0.001)
	assert.InDelta(t, 99.1, hc["peak_mem_pct"], 0.001)
	assert.InDelta(t, 36, hc["sample_count"], 0.001)
	assert.Equal(t, "full", hc["resolution"])
	assert.Equal(t, "2026-03-10T11:53:00Z", hc["window_from"])
	assert.Equal(t, "2026-03-10T11:59:00Z", hc["window_to"])

	disk, ok := hc["worst_disk"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "/var", disk["mount"])
	assert.InDelta(t, 96.2, disk["used_pct"], 0.001)
}

// T017 — absence is null, never a zero-filled object (FR-010, SC-005).
func TestIncidentHandler_Get_HostContextAbsentIsNull(t *testing.T) {
	data := getIncidentBody(t, baseIncident())

	raw, ok := data["host_context"]
	require.True(t, ok, "the key is present so clients can rely on it")
	assert.Nil(t, raw, "absent context serialises as null, not as an empty or zeroed object")
}

// T017 — the worst disk is independently nullable: a host that reported no mounts
// still returns its CPU and memory peaks.
func TestIncidentHandler_Get_HostContextWorstDiskNullable(t *testing.T) {
	inc := baseIncident()
	inc.HostContext = &domain.HostContext{
		HostID:      "host-1",
		HostName:    "web-01",
		PeakCPUPct:  71.0,
		PeakMemPct:  62.0,
		SampleCount: 12,
		Resolution:  domain.HostContextReduced,
		WindowFrom:  time.Date(2026, 3, 10, 11, 53, 0, 0, time.UTC),
		WindowTo:    time.Date(2026, 3, 10, 11, 59, 0, 0, time.UTC),
	}

	data := getIncidentBody(t, inc)
	hc := data["host_context"].(map[string]any)
	assert.Nil(t, hc["worst_disk"])
	assert.InDelta(t, 71.0, hc["peak_cpu_pct"], 0.001)
	assert.Equal(t, "reduced", hc["resolution"])
}

// T017 — the addition is strictly additive: every pre-existing field keeps its
// name, type and nullability (FR-014, SC-006).
func TestIncidentHandler_Get_ExistingContractUnchanged(t *testing.T) {
	inc := baseIncident()
	resolved := time.Date(2026, 3, 10, 12, 30, 0, 0, time.UTC)
	inc.ResolvedAt = &resolved

	data := getIncidentBody(t, inc)

	for _, key := range []string{
		"id", "monitor_id", "resource_id", "resource", "cause", "status",
		"details", "event_steps", "diagnostics", "started_at", "resolved_at",
		"created_at", "updated_at",
	} {
		_, ok := data[key]
		assert.Truef(t, ok, "pre-existing field %q must still be serialised", key)
	}

	assert.IsType(t, "", data["id"])
	assert.IsType(t, "", data["started_at"])
	assert.IsType(t, "", data["resolved_at"])
	assert.Nil(t, data["diagnostics"], "diagnostics stays nullable and untouched")
}
