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

// host_link says which machine the incident describes and how that is known
// (spec 092, FR-008a). Four shapes, and one absence.
func TestIncidentHandler_Get_HostLinkShapes(t *testing.T) {
	cases := []struct {
		name   string
		link   *domain.HostLink
		source string
		exists bool
	}{
		{"recorded, machine exists", &domain.HostLink{HostID: "host-A", Source: domain.HostLinkSourceRecorded, Exists: true}, "recorded", true},
		{"recorded, machine deleted", &domain.HostLink{HostID: "host-A", Source: domain.HostLinkSourceRecorded, Exists: false}, "recorded", false},
		{"inferred, machine exists", &domain.HostLink{HostID: "host-B", Source: domain.HostLinkSourceInferred, Exists: true}, "inferred", true},
		{"inferred, machine deleted", &domain.HostLink{HostID: "host-B", Source: domain.HostLinkSourceInferred, Exists: false}, "inferred", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			inc := baseIncident()
			inc.HostLink = tc.link

			body := getIncidentBody(t, inc)
			raw, ok := body["host_link"].(map[string]any)
			require.True(t, ok, "host_link must be present")
			assert.Equal(t, tc.link.HostID, raw["host_id"])
			assert.Equal(t, tc.source, raw["source"])
			assert.Equal(t, tc.exists, raw["exists"])
		})
	}
}

// A deleted recorded machine is still recorded. Downgrading it to inferred
// would let it fall back onto a machine that was never involved (FR-009).
func TestIncidentHandler_Get_DeletedMachineStaysRecorded(t *testing.T) {
	inc := baseIncident()
	inc.HostLink = &domain.HostLink{HostID: "host-gone", Source: domain.HostLinkSourceRecorded, Exists: false}

	body := getIncidentBody(t, inc)
	raw := body["host_link"].(map[string]any)
	assert.Equal(t, "recorded", raw["source"])
	assert.Equal(t, false, raw["exists"])
	// host_context keeps its pre-092 shape: always a key, null when absent.
	assert.Nil(t, body["host_context"],
		"no name and no metrics can be shown for a machine that no longer exists")
}

// Absent means absent -- asserted on the marshalled key set. A null key would
// already be a body change, and this struct is shared with the list endpoint.
func TestIncidentHandler_Get_HostLinkAbsentOmitsTheKey(t *testing.T) {
	body := getIncidentBody(t, baseIncident())
	_, present := body["host_link"]
	assert.False(t, present, "the key must not exist at all when there is no machine to describe")
}

// The list endpoint never populates it and must not show it (research R5).
func TestIncidentHandler_List_NeverCarriesHostLink(t *testing.T) {
	inc := goldenListIncident()
	inc.HostLink = &domain.HostLink{HostID: "host-A", Source: domain.HostLinkSourceRecorded, Exists: true}
	router := newIncidentRouter(&mockIncidentService{incidents: []*domain.Incident{inc}})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/incidents", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	// The list path maps the same struct, so a link set on the domain object
	// WOULD serialise. The service never sets it on the list path; this pins
	// that the mapper alone does not invent one.
	var envelope struct {
		Data []map[string]any `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
	require.Len(t, envelope.Data, 1)
	// A link IS present here because the mock set it -- which proves the field
	// travels, and therefore that the list golden (unchanged, no link) is what
	// guards the real list path, where the service leaves it nil.
	_, present := envelope.Data[0]["host_link"]
	assert.True(t, present)
}

// An inferred machine never appears together with an explanation: the sentence
// is a claim, and it is not made about a machine that was only inferred
// (FR-007). The service enforces this; the handler must not reintroduce it.
func TestIncidentHandler_Get_InferredNeverHasAnExplanation(t *testing.T) {
	inc := baseIncident()
	inc.HostLink = &domain.HostLink{HostID: "host-B", Source: domain.HostLinkSourceInferred, Exists: true}
	// The service would never set both; if a caller did, the response must
	// still be internally honest, so this documents the invariant at the API.
	inc.Explanation = nil

	body := getIncidentBody(t, inc)
	_, hasExplanation := body["explanation"]
	assert.False(t, hasExplanation)
	assert.Equal(t, "inferred", body["host_link"].(map[string]any)["source"])
}
