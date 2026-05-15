package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type incidentServiceStub struct {
	incidents []*domain.Incident
	incident  *domain.Incident
}

func (s *incidentServiceStub) ListAll(ctx context.Context, limit, offset int) ([]*domain.Incident, error) {
	return s.incidents, nil
}

func (s *incidentServiceStub) ListUnresolved(ctx context.Context, limit, offset int) ([]*domain.Incident, error) {
	return s.incidents, nil
}

func (s *incidentServiceStub) GetIncidentByID(ctx context.Context, id string) (*domain.Incident, error) {
	return s.incident, nil
}

func (s *incidentServiceStub) GetIncidentsByResource(ctx context.Context, resourceID string, limit, offset int) ([]*domain.Incident, error) {
	return s.incidents, nil
}

func (s *incidentServiceStub) GetEventStepsForIncident(ctx context.Context, incidentID string) ([]domain.IncidentEventStep, error) {
	return nil, nil
}

func TestIncidentHandler_ListIncidents_DetailsAreReadableText(t *testing.T) {
	h := NewIncidentHandler(&incidentServiceStub{
		incidents: []*domain.Incident{{
			Base:       domain.Base{ID: "inc-1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			ResourceID: "res-1",
			Cause:      "dns_resolution_failure",
			StartedAt:  time.Now(),
			Details:    []byte("dial tcp: lookup api.example.com: no such host"),
		}},
	})

	req := httptest.NewRequest(http.MethodGet, "/incidents", nil)
	rr := httptest.NewRecorder()

	h.ListIncidents(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var out []map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &out))
	require.Len(t, out, 1)
	assert.Equal(t, "dial tcp: lookup api.example.com: no such host", out[0]["details"])
}

func TestIncidentHandler_GetIncidentDetail_DetailsAreReadableText(t *testing.T) {
	h := NewIncidentHandler(&incidentServiceStub{
		incident: &domain.Incident{
			Base:       domain.Base{ID: "inc-2", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			ResourceID: "res-2",
			Cause:      "health_check_failed",
			StartedAt:  time.Now(),
			Details:    []byte("connection refused by remote host"),
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/incidents/inc-2", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "inc-2")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()

	h.GetIncidentDetail(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var out map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &out))
	assert.Equal(t, "connection refused by remote host", out["details"])
}
