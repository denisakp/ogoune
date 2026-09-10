package v1_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/domain"
)

func explainedIncident() *domain.Incident {
	inc := baseIncident()
	inc.Cause = "HTTP check failed: 502 Bad Gateway"
	event := domain.HostEvent{
		Base:        domain.Base{ID: "evt-1"},
		HostID:      "host-1",
		OccurredAt:  inc.StartedAt.Add(-13 * time.Second),
		Kind:        "oom_kill",
		Source:      "kmsg",
		Occurrences: 1,
		Detail:      &domain.HostEventDetail{Process: "postgres", PID: 4711},
	}
	inc.Explanation = &domain.IncidentExplanation{
		HostID:      "host-1",
		HostName:    "web-01",
		IncidentAt:  inc.StartedAt,
		Cause:       inc.Cause,
		Event:       event,
		Precedes:    true,
		OtherEvents: 2,
		WindowFrom:  inc.StartedAt.Add(-domain.HostContextWindowBefore),
		WindowTo:    inc.StartedAt.Add(domain.HostContextWindowAfter),
	}
	inc.HostEvents = []*domain.HostEvent{&event}
	return inc
}

func TestIncidentHandler_Get_ExplanationPresent(t *testing.T) {
	body := getIncidentBody(t, explainedIncident())

	raw, ok := body["explanation"]
	require.True(t, ok, "explanation must be present when the service produced one")
	e, ok := raw.(map[string]any)
	require.True(t, ok)

	assert.Equal(t, "host-1", e["host_id"])
	assert.Equal(t, "web-01", e["host_name"])
	assert.Equal(t, "HTTP check failed: 502 Bad Gateway", e["cause"])
	assert.Equal(t, "2026-03-10T11:58:00Z", e["incident_at"])
	assert.Equal(t, "2026-03-10T11:53:00Z", e["window_from"])
	assert.Equal(t, "2026-03-10T11:59:00Z", e["window_to"])
	assert.Equal(t, true, e["precedes"])
	assert.Equal(t, float64(2), e["other_events"])

	text, _ := e["text"].(string)
	assert.Contains(t, text, "OOM-killed postgres (pid 4711) on web-01")
	assert.Contains(t, text, "2026-03-10 11:58:00 UTC", "the failure time is always rendered")
	assert.Contains(t, text, "2026-03-10 11:57:47 UTC", "and so is the kernel's")
	assert.Contains(t, text, "2 other kernel events")

	event, ok := e["event"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "evt-1", event["id"])
	assert.Equal(t, "oom_kill", event["kind"])

	// The evidence travels with the claim, so the sentence is checkable in place
	// rather than merely trusted (FR-018b).
	events, ok := body["host_events"].([]any)
	require.True(t, ok)
	require.Len(t, events, 1)
}

// Absent means the key is not there at all -- asserted on the marshalled key
// set, because "explanation": null would already be a body change.
func TestIncidentHandler_Get_ExplanationAbsentOmitsTheKeys(t *testing.T) {
	for name, inc := range map[string]*domain.Incident{
		"nothing correlated": baseIncident(),
		"events but no sentence": func() *domain.Incident {
			i := baseIncident()
			i.HostEvents = nil
			return i
		}(),
	} {
		t.Run(name, func(t *testing.T) {
			body := getIncidentBody(t, inc)
			_, hasExplanation := body["explanation"]
			_, hasEvents := body["host_events"]
			assert.False(t, hasExplanation, "the key must be absent, not null")
			assert.False(t, hasEvents, "same for host_events")
		})
	}
}

// Events of a kind this version cannot phrase are served without a sentence:
// nothing is hidden, and nothing is malformed (FR-007, SC-009).
func TestIncidentHandler_Get_EventsWithoutExplanation(t *testing.T) {
	inc := baseIncident()
	inc.HostEvents = []*domain.HostEvent{{
		Base:        domain.Base{ID: "evt-unknown"},
		HostID:      "host-1",
		OccurredAt:  inc.StartedAt.Add(-time.Second),
		Kind:        "something_new",
		Source:      "kmsg",
		Occurrences: 1,
	}}

	body := getIncidentBody(t, inc)
	_, hasExplanation := body["explanation"]
	assert.False(t, hasExplanation)
	events, ok := body["host_events"].([]any)
	require.True(t, ok, "an unrecognised kind is still displayable")
	require.Len(t, events, 1)
}

// An explanation whose sentence came out empty is worse than none: the reader
// would get a block with a hole in it. Guarded in the mapper.
func TestIncidentHandler_Get_UnphrasableExplanationIsDropped(t *testing.T) {
	inc := explainedIncident()
	inc.Explanation.Event.Kind = "something_new"

	body := getIncidentBody(t, inc)
	_, hasExplanation := body["explanation"]
	assert.False(t, hasExplanation, "no sentence means no explanation object")
}

// SC-004: nothing served records a causal claim, a score, or a confidence.
// A key-name assertion, so a future field cannot be added quietly.
func TestIncidentHandler_Get_NoCausalFieldIsEverServed(t *testing.T) {
	forbidden := []string{
		"cause_id", "caused_by", "causedby", "score", "confidence",
		"probability", "severity", "certainty", "weight", "rank",
	}
	for name, inc := range map[string]*domain.Incident{
		"correlated": explainedIncident(),
		"silent":     baseIncident(),
	} {
		t.Run(name, func(t *testing.T) {
			router := newIncidentRouter(&mockIncidentService{incident: inc})
			req := httptest.NewRequest(http.MethodGet, "/api/v1/incidents/inc-1", nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			require.Equal(t, http.StatusOK, rec.Code)

			for _, key := range forbidden {
				assert.NotContainsf(t, strings.ToLower(rec.Body.String()), `"`+key+`"`,
					"the wording may interpret; the data may not assert %q", key)
			}
		})
	}
}

// --- T021a: the endpoints this feature must not move --------------------------
//
// The list and detail responses share one mapper (mapIncidentResponse), which is
// exactly how an untouched endpoint changes by accident. These goldens are the
// API-side equivalent of the notifier ones: comparison, not inspection.

func assertAPIGolden(t *testing.T, name string, got []byte) {
	t.Helper()
	path := filepath.Join("testdata", name)
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		require.NoError(t, os.WriteFile(path, got, 0o644))
		t.Logf("golden updated: %s", path)
		return
	}
	want, err := os.ReadFile(path)
	require.NoErrorf(t, err, "missing golden %s", path)
	require.Equalf(t, string(want), string(got),
		"%s changed. An incident with nothing correlated must serialise exactly as it did before spec 091 (FR-017, SC-008)", name)
}

// goldenListIncident is fixed in every respect that reaches the body.
func goldenListIncident() *domain.Incident {
	fixed := time.Date(2026, 3, 10, 11, 55, 0, 0, time.UTC)
	return &domain.Incident{
		Base:       domain.Base{ID: "inc-golden", CreatedAt: fixed, UpdatedAt: fixed},
		ResourceID: "res-golden",
		Resource: domain.Resource{
			Base: domain.Base{ID: "res-golden"},
			Name: "storefront",
		},
		Cause:     "HTTP check failed: 502 Bad Gateway",
		StartedAt: time.Date(2026, 3, 10, 11, 58, 0, 0, time.UTC),
	}
}

// preFeatureIncidentKeys is the exact top-level key set an IncidentResponse
// carried before spec 091, taken from `git show HEAD:internal/dto/v1/incident.go`
// while the goldens were captured. The goldens themselves were written after the
// DTO gained its two fields, so this list is what proves omitempty actually kept
// them out rather than merely being expected to.
var preFeatureIncidentKeys = []string{
	"cause", "created_at", "details", "diagnostics", "event_steps",
	"host_context", "id", "monitor_id", "resolved_at", "resource",
	"resource_id", "started_at", "status", "updated_at",
}

func assertPreFeatureKeySet(t *testing.T, obj map[string]any) {
	t.Helper()
	var keys []string
	for k := range obj {
		keys = append(keys, k)
	}
	assert.ElementsMatch(t, preFeatureIncidentKeys, keys,
		"an incident with nothing correlated must serialise exactly the keys it did before spec 091")
}

func TestIncidentHandler_ListGolden_Unchanged(t *testing.T) {
	router := newIncidentRouter(&mockIncidentService{incidents: []*domain.Incident{goldenListIncident()}})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/incidents", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	assertAPIGolden(t, "incident_list.golden.json", rec.Body.Bytes())
	assert.NotContains(t, rec.Body.String(), "explanation",
		"the list endpoint never populates it, and omitempty keeps the key out")
	assert.NotContains(t, rec.Body.String(), "host_events")

	var envelope struct {
		Data []map[string]any `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
	require.Len(t, envelope.Data, 1)
	assertPreFeatureKeySet(t, envelope.Data[0])
}

func TestIncidentHandler_GetGolden_NoHostUnchanged(t *testing.T) {
	router := newIncidentRouter(&mockIncidentService{incident: goldenListIncident()})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/incidents/inc-golden", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	assertAPIGolden(t, "incident_detail_silent.golden.json", rec.Body.Bytes())

	var envelope struct {
		Data map[string]any `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
	assertPreFeatureKeySet(t, envelope.Data)
}

// The host context is read, not altered (FR-017): a correlated incident still
// serialises its host_context exactly as WI-1 defined it.
func TestIncidentHandler_Get_HostContextUntouchedByCorrelation(t *testing.T) {
	inc := explainedIncident()
	inc.HostContext = &domain.HostContext{
		HostID:      "host-1",
		HostName:    "web-01",
		PeakCPUPct:  97.4,
		PeakMemPct:  62.5,
		SampleCount: 12,
		Resolution:  domain.HostContextFull,
		WindowFrom:  inc.StartedAt.Add(-domain.HostContextWindowBefore),
		WindowTo:    inc.StartedAt.Add(domain.HostContextWindowAfter),
	}

	body := getIncidentBody(t, inc)
	hc, ok := body["host_context"].(map[string]any)
	require.True(t, ok)

	var keys []string
	for k := range hc {
		keys = append(keys, k)
	}
	assert.ElementsMatch(t, []string{
		"host_id", "host_name", "peak_cpu_pct", "peak_mem_pct",
		"worst_disk", "sample_count", "resolution", "window_from", "window_to",
	}, keys, "spec 091 reads the host context; it adds nothing to it and takes nothing away")

	raw, err := json.Marshal(hc)
	require.NoError(t, err)
	assert.NotContains(t, string(raw), "explanation")
}
