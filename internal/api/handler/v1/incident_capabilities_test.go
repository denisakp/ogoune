package v1_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/domain"
)

func recordedLink() *domain.HostLink {
	return &domain.HostLink{HostID: "host-A", Source: domain.HostLinkSourceRecorded, Exists: true}
}

func statePtr(s domain.HostCapabilitiesState) *domain.HostCapabilitiesState { return &s }

// host_capabilities is the declaration frozen when the incident opened (spec
// 093). Three shapes carry a state; "declared" carries the copy too.
func TestIncidentHandler_Get_FrozenCapabilityShapes(t *testing.T) {
	declared := &domain.HostCapabilities{
		Kmsg:      domain.Capability{Available: false, Reason: domain.CapabilityReasonUnreadable},
		CgroupOOM: domain.Capability{Available: true},
		Segfault:  domain.Capability{Available: false, Reason: domain.CapabilityReasonUnreadable},
	}
	cases := []struct {
		name  string
		state *domain.HostCapabilitiesState
		caps  *domain.HostCapabilities
		want  map[string]any
	}{
		{"declared: state, three capabilities, oom_detail, no declared_at",
			statePtr(domain.HostCapabilitiesDeclared), declared,
			map[string]any{
				"state":      "declared",
				"kmsg":       map[string]any{"available": false, "reason": "unreadable"},
				"cgroup_oom": map[string]any{"available": true},
				"segfault":   map[string]any{"available": false, "reason": "unreadable"},
				"oom_detail": "without_process",
			}},
		{"not_reported: state only", statePtr(domain.HostCapabilitiesNotReported), nil, map[string]any{"state": "not_reported"}},
		{"not_known: state only", statePtr(domain.HostCapabilitiesNotKnown), nil, map[string]any{"state": "not_known"}},
		{"pre-feature row with a host link: not_known, never the host's current declaration", nil, nil, map[string]any{"state": "not_known"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			inc := baseIncident()
			inc.HostLink = recordedLink()
			inc.HostCapabilitiesState = tc.state
			inc.HostCapabilities = tc.caps

			body := getIncidentBody(t, inc)
			raw, ok := body["host_capabilities"].(map[string]any)
			require.True(t, ok, "host_capabilities must be present alongside a host link")
			assert.Equal(t, tc.want, raw)
		})
	}
}

// Nothing to say, nothing said: a stored "no_machine", or a pre-feature row
// without a link, omit the key -- not null, absent.
func TestIncidentHandler_Get_FrozenCapabilitiesOmittedWhenNoMachine(t *testing.T) {
	t.Run("stored no_machine", func(t *testing.T) {
		inc := baseIncident()
		inc.HostCapabilitiesState = statePtr(domain.HostCapabilitiesNoMachine)
		body := getIncidentBody(t, inc)
		_, present := body["host_capabilities"]
		assert.False(t, present)
	})
	t.Run("pre-feature row, no link", func(t *testing.T) {
		body := getIncidentBody(t, baseIncident())
		_, present := body["host_capabilities"]
		assert.False(t, present)
	})
	t.Run("declared but the link is gone from the read model: omitted, detail-only by construction", func(t *testing.T) {
		inc := baseIncident()
		inc.HostCapabilitiesState = statePtr(domain.HostCapabilitiesDeclared)
		inc.HostCapabilities = &domain.HostCapabilities{Kmsg: domain.Capability{Available: true}}
		body := getIncidentBody(t, inc)
		_, present := body["host_capabilities"]
		assert.False(t, present, "the frozen columns are on every row; only the detail path sets the link that unlocks them")
	})
}

// The list endpoint reads the same rows, and the frozen columns ARE on them.
// The mapper must not surface them there: the list body stays byte-identical
// to its pre-feature form (guarded by the list golden as well).
func TestIncidentHandler_List_NeverCarriesFrozenCapabilities(t *testing.T) {
	inc := goldenListIncident()
	inc.HostCapabilitiesState = statePtr(domain.HostCapabilitiesDeclared)
	inc.HostCapabilities = &domain.HostCapabilities{Kmsg: domain.Capability{Available: true}, CgroupOOM: domain.Capability{Available: true}, Segfault: domain.Capability{Available: true}}
	router := newIncidentRouter(&mockIncidentService{incidents: []*domain.Incident{inc}})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/incidents", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var envelope struct {
		Data []map[string]any `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
	require.Len(t, envelope.Data, 1)
	_, present := envelope.Data[0]["host_capabilities"]
	assert.False(t, present, "stored on the row, never on the list")
}
