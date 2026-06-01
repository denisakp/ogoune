package v1_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	v1 "github.com/denisakp/ogoune/internal/api/handler/v1"
	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/sqlc/dynquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// filterCapturingService records the dynquery.MonitorFilter the handler passed
// it. Used to assert parse → handler pipeline behaviour without touching a DB.
type filterCapturingService struct {
	*mockMonitorService
	gotFilter dynquery.MonitorFilter
	called    bool
}

func (s *filterCapturingService) ListByFilter(_ context.Context, f dynquery.MonitorFilter, page, perPage int) ([]*domain.Resource, int, error) {
	s.gotFilter = f
	s.called = true
	total := len(s.resources)
	offset := (page - 1) * perPage
	if offset >= total {
		return []*domain.Resource{}, total, nil
	}
	end := offset + perPage
	if end > total {
		end = total
	}
	return s.resources[offset:end], total, nil
}

func newCapturingMonitorRouter(svc v1.MonitorV1ServiceInterface) http.Handler {
	return newMonitorRouter(svc)
}

func TestMonitorHandler_List_NoFilters_PassesEmptyFilter(t *testing.T) {
	res := &domain.Resource{Base: domain.Base{ID: "r1", CreatedAt: time.Now(), UpdatedAt: time.Now()}, Name: "n", Type: "http"}
	svc := &filterCapturingService{mockMonitorService: &mockMonitorService{resources: []*domain.Resource{res}}}
	router := newCapturingMonitorRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/monitors", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.True(t, svc.called)
	require.Nil(t, svc.gotFilter.Tag)
	require.Nil(t, svc.gotFilter.Type)
	require.Nil(t, svc.gotFilter.IsActive)
	require.Nil(t, svc.gotFilter.Q)
}

func TestMonitorHandler_List_AllFilters_ParsedAndForwarded(t *testing.T) {
	svc := &filterCapturingService{mockMonitorService: &mockMonitorService{resources: nil}}
	router := newCapturingMonitorRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/monitors?tag=prod&type=http&is_active=false&q=api", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.True(t, svc.called)
	require.NotNil(t, svc.gotFilter.Tag)
	assert.Equal(t, "prod", *svc.gotFilter.Tag)
	require.NotNil(t, svc.gotFilter.Type)
	assert.Equal(t, "http", *svc.gotFilter.Type)
	require.NotNil(t, svc.gotFilter.IsActive)
	assert.Equal(t, false, *svc.gotFilter.IsActive)
	require.NotNil(t, svc.gotFilter.Q)
	assert.Equal(t, "api", *svc.gotFilter.Q)
}

func TestMonitorHandler_List_InvalidType_Returns400(t *testing.T) {
	svc := &mockMonitorService{}
	router := newMonitorRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/monitors?type=bogus", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
	var body problemDetailResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&body))
	require.NotEmpty(t, body.Errors)
	assert.Equal(t, "type", body.Errors[0].Field)
}

func TestMonitorHandler_List_InvalidIsActive_Returns400(t *testing.T) {
	svc := &mockMonitorService{}
	router := newMonitorRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/monitors?is_active=maybe", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestMonitorHandler_List_TooLongQ_Returns400(t *testing.T) {
	long := make([]byte, 201)
	for i := range long {
		long[i] = 'x'
	}
	svc := &mockMonitorService{}
	router := newMonitorRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/monitors?q="+string(long), nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestMonitorHandler_List_EmptyParam_TreatedAsNil(t *testing.T) {
	svc := &filterCapturingService{mockMonitorService: &mockMonitorService{}}
	router := newCapturingMonitorRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/monitors?type=&tag=&q=", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.Nil(t, svc.gotFilter.Type)
	assert.Nil(t, svc.gotFilter.Tag)
	assert.Nil(t, svc.gotFilter.Q)
}

func TestMonitorHandler_List_BadPagination_Returns422(t *testing.T) {
	svc := &mockMonitorService{}
	router := newMonitorRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/monitors?page=abc", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusUnprocessableEntity, rr.Code)
}
