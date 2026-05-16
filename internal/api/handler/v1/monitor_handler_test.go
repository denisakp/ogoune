package v1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	v1 "github.com/denisakp/ogoune/internal/api/handler/v1"
	"github.com/denisakp/ogoune/internal/api/middleware"
	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/dto"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// problemDetailResponse is a test helper struct matching the RFC 7807 ProblemDetail format.
type problemDetailResponse struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail"`
	Errors []struct {
		Field   string `json:"field"`
		Message string `json:"message"`
	} `json:"errors,omitempty"`
}

// --- mock service ---

type mockMonitorService struct {
	resources []*domain.Resource
	resource  *domain.Resource
	listErr   error
	getErr    error
	createErr error
	updateErr error
	deleteErr error
	pauseErr  error
	resumeErr error
}

func (m *mockMonitorService) ListActiveResources(_ context.Context, limit, offset int) ([]*domain.Resource, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	end := offset + limit
	if end > len(m.resources) {
		end = len(m.resources)
	}
	if offset > len(m.resources) {
		return []*domain.Resource{}, nil
	}
	return m.resources[offset:end], nil
}

func (m *mockMonitorService) ListAll(_ context.Context) ([]*domain.Resource, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.resources, nil
}

func (m *mockMonitorService) GetResourceByID(_ context.Context, id string) (*domain.Resource, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if m.resource != nil {
		return m.resource, nil
	}
	for _, r := range m.resources {
		if r.ID == id {
			return r, nil
		}
	}
	return nil, nil
}

func (m *mockMonitorService) CreateResource(_ context.Context, payload *dto.CreateResourcePayload) (*domain.Resource, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	return &domain.Resource{
		Base:   domain.Base{ID: "new-id", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		Name:   payload.Name,
		Type:   payload.Type,
		Target: payload.Target,
	}, nil
}

func (m *mockMonitorService) UpdateResource(_ context.Context, id string, _ *dto.UpdateResourcePayload) (*domain.Resource, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	return &domain.Resource{Base: domain.Base{ID: id, CreatedAt: time.Now(), UpdatedAt: time.Now()}}, nil
}

func (m *mockMonitorService) DeleteResource(_ context.Context, _ string) error {
	return m.deleteErr
}

func (m *mockMonitorService) PauseMonitoring(_ context.Context, _ string) error {
	return m.pauseErr
}

func (m *mockMonitorService) ResumeMonitoring(_ context.Context, _ string) error {
	return m.resumeErr
}

// --- helper: build a Chi router with v1 monitor routes ---

func newMonitorRouter(svc v1.MonitorV1ServiceInterface) *chi.Mux {
	r := chi.NewRouter()
	h := v1.NewMonitorHandler(svc)
	r.Get("/api/v1/monitors", h.List)
	r.With(middleware.RequireReadWrite).Post("/api/v1/monitors", h.Create)
	r.Get("/api/v1/monitors/{id}", h.Get)
	r.With(middleware.RequireReadWrite).Put("/api/v1/monitors/{id}", h.Update)
	r.With(middleware.RequireReadWrite).Delete("/api/v1/monitors/{id}", h.Delete)
	r.With(middleware.RequireReadWrite).Post("/api/v1/monitors/{id}/pause", h.Pause)
	r.With(middleware.RequireReadWrite).Post("/api/v1/monitors/{id}/resume", h.Resume)
	return r
}

// injectReadScope injects read-only API key context into a request.
func injectReadScope(r *http.Request) *http.Request {
	ctx := context.WithValue(r.Context(), "auth_method", "api_key")
	ctx = context.WithValue(ctx, "api_key_scope", domain.APIKeyScopeRead)
	return r.WithContext(ctx)
}

// injectReadWriteScope injects read-write API key context into a request.
func injectReadWriteScope(r *http.Request) *http.Request {
	ctx := context.WithValue(r.Context(), "auth_method", "api_key")
	ctx = context.WithValue(ctx, "api_key_scope", domain.APIKeyScopeReadWrite)
	return r.WithContext(ctx)
}

// ============================================================
// T012: Scope enforcement tests
// ============================================================

func TestMonitorHandler_ScopeEnforcement_ReadKeyOnWriteRoutes_Returns403(t *testing.T) {
	svc := &mockMonitorService{resources: []*domain.Resource{}}
	router := newMonitorRouter(svc)

	writeCases := []struct {
		method string
		path   string
		body   []byte
	}{
		{"POST", "/api/v1/monitors", []byte(`{"name":"x","type":"http","target":"http://x.com","interval":60,"timeout":10}`)},
		{"PUT", "/api/v1/monitors/abc", []byte(`{}`)},
		{"DELETE", "/api/v1/monitors/abc", nil},
		{"POST", "/api/v1/monitors/abc/pause", nil},
		{"POST", "/api/v1/monitors/abc/resume", nil},
	}

	for _, tc := range writeCases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			var bodyReader *bytes.Reader
			if tc.body != nil {
				bodyReader = bytes.NewReader(tc.body)
			} else {
				bodyReader = bytes.NewReader(nil)
			}
			req := httptest.NewRequest(tc.method, tc.path, bodyReader)
			req = injectReadScope(req)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
			assert.Equal(t, http.StatusForbidden, rr.Code, "%s %s with read scope should be 403", tc.method, tc.path)
		})
	}
}

func TestMonitorHandler_ScopeEnforcement_ReadWriteKeyOnWriteRoutes_NotForbidden(t *testing.T) {
	svc := &mockMonitorService{
		resources: []*domain.Resource{
			{Base: domain.Base{ID: "mon-1", CreatedAt: time.Now(), UpdatedAt: time.Now()}, Name: "Test", Type: domain.ResourceHTTP, Target: "http://test.com"},
		},
		resource: &domain.Resource{Base: domain.Base{ID: "mon-1", CreatedAt: time.Now(), UpdatedAt: time.Now()}, Name: "Test", Type: domain.ResourceHTTP, Target: "http://test.com"},
	}
	router := newMonitorRouter(svc)

	writeCases := []struct {
		method string
		path   string
		body   []byte
	}{
		{"POST", "/api/v1/monitors", []byte(`{"name":"x","type":"http","target":"http://x.com","interval":60,"timeout":10}`)},
		{"PUT", "/api/v1/monitors/mon-1", []byte(`{}`)},
		{"DELETE", "/api/v1/monitors/mon-1", nil},
		{"POST", "/api/v1/monitors/mon-1/pause", nil},
		{"POST", "/api/v1/monitors/mon-1/resume", nil},
	}

	for _, tc := range writeCases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			var bodyReader *bytes.Reader
			if tc.body != nil {
				bodyReader = bytes.NewReader(tc.body)
			} else {
				bodyReader = bytes.NewReader(nil)
			}
			req := httptest.NewRequest(tc.method, tc.path, bodyReader)
			req = injectReadWriteScope(req)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
			assert.NotEqual(t, http.StatusForbidden, rr.Code, "%s %s with read_write scope should not be 403", tc.method, tc.path)
		})
	}
}

// ============================================================
// T013: Pagination tests
// ============================================================

func TestMonitorHandler_Pagination_PerPage999_ClampedTo100(t *testing.T) {
	resources := make([]*domain.Resource, 5)
	for i := range resources {
		resources[i] = &domain.Resource{Base: domain.Base{ID: "r", CreatedAt: time.Now(), UpdatedAt: time.Now()}}
	}
	svc := &mockMonitorService{resources: resources}
	router := newMonitorRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/monitors?per_page=999", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var out map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &out))
	meta, ok := out["meta"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(100), meta["per_page"], "per_page should be clamped to 100")
}

func TestMonitorHandler_Pagination_PerPage0_Returns422(t *testing.T) {
	svc := &mockMonitorService{}
	router := newMonitorRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/monitors?per_page=0", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	var out problemDetailResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &out))
	assert.Equal(t, "VALIDATION_FAILED", out.Type)
}

func TestMonitorHandler_Pagination_PageNegative_Returns422(t *testing.T) {
	svc := &mockMonitorService{}
	router := newMonitorRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/monitors?page=-1", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
}

func TestMonitorHandler_Pagination_PageNonInteger_Returns422(t *testing.T) {
	svc := &mockMonitorService{}
	router := newMonitorRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/monitors?page=abc", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
}

// ============================================================
// T014: Envelope shape tests
// ============================================================

func TestMonitorHandler_List_EnvelopeHasDataArrayAndMeta(t *testing.T) {
	svc := &mockMonitorService{resources: []*domain.Resource{
		{Base: domain.Base{ID: "m1", CreatedAt: time.Now(), UpdatedAt: time.Now()}, Name: "M1", Type: domain.ResourceHTTP, Target: "http://m1.com"},
	}}
	router := newMonitorRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/monitors", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var out map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &out))
	_, hasData := out["data"]
	assert.True(t, hasData, "response should have 'data' key")
	_, hasMeta := out["meta"]
	assert.True(t, hasMeta, "response should have 'meta' key")
	data, ok := out["data"].([]interface{})
	assert.True(t, ok, "'data' should be an array")
	assert.Len(t, data, 1)
}

func TestMonitorHandler_Get_EnvelopeHasDataObjectAndNullMeta(t *testing.T) {
	svc := &mockMonitorService{
		resource: &domain.Resource{Base: domain.Base{ID: "m1", CreatedAt: time.Now(), UpdatedAt: time.Now()}, Name: "M1", Type: domain.ResourceHTTP, Target: "http://m1.com"},
	}
	router := newMonitorRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/monitors/m1", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var out map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &out))
	_, hasData := out["data"]
	assert.True(t, hasData, "single response should have 'data' key")
	assert.Nil(t, out["meta"], "single response 'meta' should be null")
}

func TestMonitorHandler_Get_NotFound_ErrorCodeResourceNotFound(t *testing.T) {
	svc := &mockMonitorService{getErr: assert.AnError}
	// override to return ErrResourceNotFound
	svcWithNotFound := &mockMonitorServiceNotFound{}
	router := newMonitorRouter(svcWithNotFound)
	_ = svc

	req := httptest.NewRequest(http.MethodGet, "/api/v1/monitors/unknown", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)
	var out problemDetailResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &out))
	assert.Equal(t, "RESOURCE_NOT_FOUND", out.Type)
}

func TestMonitorHandler_Create_MissingType_Returns422WithFields(t *testing.T) {
	svc := &mockMonitorService{}
	router := newMonitorRouter(svc)

	body := []byte(`{"name":"Test","target":"http://test.com","interval":60,"timeout":10}`) // missing type
	req := httptest.NewRequest(http.MethodPost, "/api/v1/monitors", bytes.NewReader(body))
	req = injectReadWriteScope(req)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	var out problemDetailResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &out))
	assert.Equal(t, "VALIDATION_FAILED", out.Type)
	assert.NotEmpty(t, out.Errors, "errors should be non-empty")
}

// mockMonitorServiceNotFound always returns service.ErrResourceNotFound on GetResourceByID.
type mockMonitorServiceNotFound struct {
	mockMonitorService
}

func (m *mockMonitorServiceNotFound) GetResourceByID(_ context.Context, _ string) (*domain.Resource, error) {
	return nil, errNotFound
}

var errNotFound = notFoundError{}

type notFoundError struct{}

func (notFoundError) Error() string { return "resource not found" }
func (notFoundError) Is(target error) bool {
	return target.Error() == "resource: not found" || target.Error() == "resource not found"
}
