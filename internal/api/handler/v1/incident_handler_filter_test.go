package v1_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/sqlc/dynquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type incidentFilterCapturing struct {
	*mockIncidentService
	gotFilter dynquery.IncidentFilter
	called    bool
}

func (s *incidentFilterCapturing) ListByFilter(ctx context.Context, f dynquery.IncidentFilter, page, perPage int) ([]*domain.Incident, int, error) {
	s.gotFilter = f
	s.called = true
	return s.mockIncidentService.ListByFilter(ctx, f, page, perPage)
}

func TestIncidentHandler_Filter_NoParams_PassesEmptyFilter(t *testing.T) {
	svc := &incidentFilterCapturing{mockIncidentService: &mockIncidentService{incidents: makeIncidents(2, false)}}
	router := newIncidentRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/incidents", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.True(t, svc.called)
	assert.Nil(t, svc.gotFilter.Status)
	assert.Nil(t, svc.gotFilter.MonitorID)
	assert.Nil(t, svc.gotFilter.From)
	assert.Nil(t, svc.gotFilter.To)
}

func TestIncidentHandler_Filter_AllParams_ParsedAndForwarded(t *testing.T) {
	svc := &incidentFilterCapturing{mockIncidentService: &mockIncidentService{incidents: nil}}
	router := newIncidentRouter(svc)

	req := httptest.NewRequest(http.MethodGet,
		"/api/v1/incidents?status=open&monitor_id=mon-123&from=2026-05-01T00:00:00Z&to=2026-05-31T23:59:59Z",
		nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.True(t, svc.called)
	require.NotNil(t, svc.gotFilter.Status)
	assert.Equal(t, "open", *svc.gotFilter.Status)
	require.NotNil(t, svc.gotFilter.MonitorID)
	assert.Equal(t, "mon-123", *svc.gotFilter.MonitorID)
	require.NotNil(t, svc.gotFilter.From)
	require.NotNil(t, svc.gotFilter.To)
	assert.Equal(t, time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC), svc.gotFilter.From.UTC())
	assert.Equal(t, time.Date(2026, 5, 31, 23, 59, 59, 0, time.UTC), svc.gotFilter.To.UTC())
}

func TestIncidentHandler_Filter_BadFromFormat_Returns400(t *testing.T) {
	svc := &mockIncidentService{}
	router := newIncidentRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/incidents?from=not-a-date", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
	var out problemDetailResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&out))
	require.NotEmpty(t, out.Errors)
	assert.Equal(t, "from", out.Errors[0].Field)
}

func TestIncidentHandler_Filter_FromAfterTo_Returns400(t *testing.T) {
	svc := &mockIncidentService{}
	router := newIncidentRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/incidents?from=2026-06-01T00:00:00Z&to=2026-05-01T00:00:00Z", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestIncidentHandler_Filter_BadPagination_Returns422(t *testing.T) {
	svc := &mockIncidentService{}
	router := newIncidentRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/incidents?page=abc", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusUnprocessableEntity, rr.Code)
}

func TestIncidentHandler_Filter_EmptyParams_TreatedAsNil(t *testing.T) {
	svc := &incidentFilterCapturing{mockIncidentService: &mockIncidentService{}}
	router := newIncidentRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/incidents?status=&monitor_id=&from=&to=", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.Nil(t, svc.gotFilter.Status)
	assert.Nil(t, svc.gotFilter.MonitorID)
	assert.Nil(t, svc.gotFilter.From)
	assert.Nil(t, svc.gotFilter.To)
}
