package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/api/middleware"
	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/dto"
	"github.com/denisakp/ogoune/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockResourceService is a test double for ResourceService
type mockResourceService struct {
	createResourceFunc                   func(ctx context.Context, payload *dto.CreateResourcePayload) (*domain.Resource, error)
	updateResourceFunc                   func(ctx context.Context, id string, payload *dto.UpdateResourcePayload) (*domain.Resource, error)
	listAllFunc                          func(ctx context.Context) ([]*domain.Resource, error)
	deleteResourceFunc                   func(ctx context.Context, id string) error
	pauseMonitoringFunc                  func(ctx context.Context, resourceID string) error
	resumeMonitoringFunc                 func(ctx context.Context, resourceID string) error
	addTagsToResourceFunc                func(ctx context.Context, resourceID string, tagIDs []string) error
	removeTagFromResourceFunc            func(ctx context.Context, resourceID, tagID string) error
	getResourceByIDFunc                  func(ctx context.Context, id string) (*domain.Resource, error)
	getResourceByIDWithResponseTimesFunc func(ctx context.Context, id string, limit int) (*dto.ResourceResponse, error)
}

type mockLiveSnapshotService struct {
	getLiveSnapshotFunc func(ctx context.Context, resourceID string) (*dto.LiveSnapshotResponse, error)
}

func (m *mockLiveSnapshotService) GetLiveSnapshot(ctx context.Context, resourceID string) (*dto.LiveSnapshotResponse, error) {
	if m.getLiveSnapshotFunc != nil {
		return m.getLiveSnapshotFunc(ctx, resourceID)
	}
	return nil, nil
}

func (m *mockResourceService) CreateResource(ctx context.Context, payload *dto.CreateResourcePayload) (*domain.Resource, error) {
	if m.createResourceFunc != nil {
		return m.createResourceFunc(ctx, payload)
	}
	return &domain.Resource{Base: domain.Base{ID: "res-123"}, Name: payload.Name, Type: payload.Type, Target: payload.Target}, nil
}

func (m *mockResourceService) GetResourceByID(ctx context.Context, id string) (*domain.Resource, error) {
	if m.getResourceByIDFunc != nil {
		return m.getResourceByIDFunc(ctx, id)
	}
	return &domain.Resource{Base: domain.Base{ID: id}}, nil
}

func (m *mockResourceService) GetResourceByIDWithResponseTimes(ctx context.Context, id string, limit int) (*dto.ResourceResponse, error) {
	if m.getResourceByIDWithResponseTimesFunc != nil {
		return m.getResourceByIDWithResponseTimesFunc(ctx, id, limit)
	}
	return &dto.ResourceResponse{}, nil
}

func (m *mockResourceService) ListAll(ctx context.Context) ([]*domain.Resource, error) {
	if m.listAllFunc != nil {
		return m.listAllFunc(ctx)
	}
	return []*domain.Resource{}, nil
}

func (m *mockResourceService) UpdateResource(ctx context.Context, id string, payload *dto.UpdateResourcePayload) (*domain.Resource, error) {
	if m.updateResourceFunc != nil {
		return m.updateResourceFunc(ctx, id, payload)
	}
	// Default: return a simple updated resource
	return &domain.Resource{Base: domain.Base{ID: id}, Name: "Updated Resource"}, nil
}

func (m *mockResourceService) DeleteResource(ctx context.Context, id string) error {
	if m.deleteResourceFunc != nil {
		return m.deleteResourceFunc(ctx, id)
	}
	return nil
}

func (m *mockResourceService) PauseMonitoring(ctx context.Context, resourceID string) error {
	if m.pauseMonitoringFunc != nil {
		return m.pauseMonitoringFunc(ctx, resourceID)
	}
	return nil
}

func (m *mockResourceService) ResumeMonitoring(ctx context.Context, resourceID string) error {
	if m.resumeMonitoringFunc != nil {
		return m.resumeMonitoringFunc(ctx, resourceID)
	}
	return nil
}

func (m *mockResourceService) ListActiveResources(ctx context.Context, limit, offset int) ([]*domain.Resource, error) {
	return []*domain.Resource{}, nil
}

func (m *mockResourceService) ListResourcesByTag(ctx context.Context, tagName string, limit, offset int) ([]*domain.Resource, error) {
	return []*domain.Resource{}, nil
}

func (m *mockResourceService) ListUnresolvedIncidents(ctx context.Context, resourceID string) ([]*domain.Incident, error) {
	return []*domain.Incident{}, nil
}

func (m *mockResourceService) AddTagsToResource(ctx context.Context, resourceID string, tagIDs []string) error {
	if m.addTagsToResourceFunc != nil {
		return m.addTagsToResourceFunc(ctx, resourceID, tagIDs)
	}
	return nil
}

func (m *mockResourceService) RemoveTagFromResource(ctx context.Context, resourceID, tagID string) error {
	if m.removeTagFromResourceFunc != nil {
		return m.removeTagFromResourceFunc(ctx, resourceID, tagID)
	}
	return nil
}

func TestCreateResource_Success(t *testing.T) {
	mockService := &mockResourceService{}
	handler := NewResourceHandler(mockService)

	resource := dto.CreateResourcePayload{
		Name:     "Test Monitor",
		Target:   "https://example.com",
		Type:     domain.ResourceHTTP,
		Timeout:  30,
		Interval: 60,
	}
	body, _ := json.Marshal(resource)

	req := httptest.NewRequest(http.MethodPost, "/resources", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.CreateResource(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var created domain.Resource
	err := json.NewDecoder(rec.Body).Decode(&created)
	require.NoError(t, err)

	assert.Equal(t, "Test Monitor", created.Name)
	assert.Equal(t, "https://example.com", created.Target)
}

func TestCreateResource_ValidationErrors(t *testing.T) {
	mockService := &mockResourceService{}
	handler := NewResourceHandler(mockService)

	tests := []struct {
		name     string
		resource domain.Resource
		expected string
	}{
		{
			name:     "missing name",
			resource: domain.Resource{Target: "https://example.com", Type: domain.ResourceHTTP},
			expected: "Resource name is required",
		},
		{
			name:     "missing target",
			resource: domain.Resource{Name: "Test", Type: domain.ResourceHTTP},
			expected: "Resource target is required",
		},
		{
			name:     "missing type",
			resource: domain.Resource{Name: "Test", Target: "https://example.com"},
			expected: "Resource type is required",
		},
		{
			name:     "invalid type",
			resource: domain.Resource{Name: "Test", Target: "https://example.com", Type: "invalid"},
			expected: "Invalid resource type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.resource)
			req := httptest.NewRequest(http.MethodPost, "/resources", bytes.NewReader(body))
			rec := httptest.NewRecorder()

			handler.CreateResource(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)

			var errorResp map[string]string
			err := json.NewDecoder(rec.Body).Decode(&errorResp)
			require.NoError(t, err)

			assert.Contains(t, errorResp["error"], tt.expected)
		})
	}
}

func TestCreateResource_ServiceError(t *testing.T) {
	mockService := &mockResourceService{
		createResourceFunc: func(ctx context.Context, payload *dto.CreateResourcePayload) (*domain.Resource, error) {
			return nil, errors.New("database connection failed")
		},
	}
	handler := NewResourceHandler(mockService)

	resource := dto.CreateResourcePayload{
		Name:     "Test Monitor",
		Target:   "https://example.com",
		Type:     domain.ResourceHTTP,
		Interval: 60,
		Timeout:  5,
	}
	body, _ := json.Marshal(resource)

	req := httptest.NewRequest(http.MethodPost, "/resources", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.CreateResource(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var errorResp map[string]string
	err := json.NewDecoder(rec.Body).Decode(&errorResp)
	require.NoError(t, err)

	assert.Contains(t, errorResp["error"], "Failed to create resource")
}

func TestCreateResource_InvalidJSON(t *testing.T) {
	mockService := &mockResourceService{}
	handler := NewResourceHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/resources", bytes.NewReader([]byte("invalid json")))
	rec := httptest.NewRecorder()

	handler.CreateResource(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestListResources_Success(t *testing.T) {
	expectedResources := []*domain.Resource{
		{
			Base:   domain.Base{ID: "1"},
			Name:   "Monitor 1",
			Target: "https://example1.com",
			Type:   domain.ResourceHTTP,
		},
		{
			Base:   domain.Base{ID: "2"},
			Name:   "Monitor 2",
			Target: "https://example2.com",
			Type:   domain.ResourceTCP,
		},
	}

	mockService := &mockResourceService{
		listAllFunc: func(ctx context.Context) ([]*domain.Resource, error) {
			return expectedResources, nil
		},
	}
	handler := NewResourceHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/resources", nil)
	rec := httptest.NewRecorder()

	handler.ListResources(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resources []*domain.Resource
	err := json.NewDecoder(rec.Body).Decode(&resources)
	require.NoError(t, err)

	assert.Len(t, resources, len(expectedResources))
	assert.Equal(t, "Monitor 1", resources[0].Name)
	assert.Equal(t, "Monitor 2", resources[1].Name)
}

func TestListResources_EmptyList(t *testing.T) {
	mockService := &mockResourceService{
		listAllFunc: func(ctx context.Context) ([]*domain.Resource, error) {
			return []*domain.Resource{}, nil
		},
	}
	handler := NewResourceHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/resources", nil)
	rec := httptest.NewRecorder()

	handler.ListResources(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resources []*domain.Resource
	err := json.NewDecoder(rec.Body).Decode(&resources)
	require.NoError(t, err)

	assert.Empty(t, resources)
}

func TestListResources_ServiceError(t *testing.T) {
	mockService := &mockResourceService{
		listAllFunc: func(ctx context.Context) ([]*domain.Resource, error) {
			return nil, errors.New("database query failed")
		},
	}
	handler := NewResourceHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/resources", nil)
	rec := httptest.NewRecorder()

	handler.ListResources(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var errorResp map[string]string
	err := json.NewDecoder(rec.Body).Decode(&errorResp)
	require.NoError(t, err)

	assert.Contains(t, errorResp["error"], "Failed to retrieve resources")
}

func TestPauseResourceMonitoring_Success(t *testing.T) {
	mockService := &mockResourceService{
		pauseMonitoringFunc: func(ctx context.Context, resourceID string) error {
			assert.Equal(t, "test-resource-id", resourceID)
			return nil
		},
	}
	handler := NewResourceHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/resources/test-resource-id/pause", nil)
	// Simulate Chi URL parameter
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("id", "test-resource-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))

	rec := httptest.NewRecorder()

	handler.PauseResourceMonitoring(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Monitoring paused successfully", response["message"])
}

func TestPauseResourceMonitoring_MissingID(t *testing.T) {
	mockService := &mockResourceService{}
	handler := NewResourceHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/resources//pause", nil)
	// No URL parameter provided
	ctx := chi.NewRouteContext()
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))

	rec := httptest.NewRecorder()

	handler.PauseResourceMonitoring(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var errorResp map[string]string
	err := json.NewDecoder(rec.Body).Decode(&errorResp)
	require.NoError(t, err)

	assert.Contains(t, errorResp["error"], "Resource ID is required")
}

func TestPauseResourceMonitoring_ServiceError(t *testing.T) {
	mockService := &mockResourceService{
		pauseMonitoringFunc: func(ctx context.Context, resourceID string) error {
			return errors.New("resource not found")
		},
	}
	handler := NewResourceHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/resources/test-resource-id/pause", nil)
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("id", "test-resource-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))

	rec := httptest.NewRecorder()

	handler.PauseResourceMonitoring(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var errorResp map[string]string
	err := json.NewDecoder(rec.Body).Decode(&errorResp)
	require.NoError(t, err)

	assert.Contains(t, errorResp["error"], "Failed to pause monitoring")
}

func TestResumeResourceMonitoring_Success(t *testing.T) {
	mockService := &mockResourceService{
		resumeMonitoringFunc: func(ctx context.Context, resourceID string) error {
			assert.Equal(t, "test-resource-id", resourceID)
			return nil
		},
	}
	handler := NewResourceHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/resources/test-resource-id/resume", nil)
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("id", "test-resource-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))

	rec := httptest.NewRecorder()

	handler.ResumeResourceMonitoring(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Monitoring resumed successfully", response["message"])
}

func TestResumeResourceMonitoring_MissingID(t *testing.T) {
	mockService := &mockResourceService{}
	handler := NewResourceHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/resources//resume", nil)
	ctx := chi.NewRouteContext()
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))

	rec := httptest.NewRecorder()

	handler.ResumeResourceMonitoring(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var errorResp map[string]string
	err := json.NewDecoder(rec.Body).Decode(&errorResp)
	require.NoError(t, err)

	assert.Contains(t, errorResp["error"], "Resource ID is required")
}

func TestResumeResourceMonitoring_ServiceError(t *testing.T) {
	mockService := &mockResourceService{
		resumeMonitoringFunc: func(ctx context.Context, resourceID string) error {
			return errors.New("resource not found")
		},
	}
	handler := NewResourceHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/resources/test-resource-id/resume", nil)
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("id", "test-resource-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))

	rec := httptest.NewRecorder()

	handler.ResumeResourceMonitoring(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var errorResp map[string]string
	err := json.NewDecoder(rec.Body).Decode(&errorResp)
	require.NoError(t, err)

	assert.Contains(t, errorResp["error"], "Failed to resume monitoring")
}

func TestUpdateResource_Success(t *testing.T) {
	name := "Updated Monitor"
	interval := 120

	mockService := &mockResourceService{
		updateResourceFunc: func(ctx context.Context, id string, payload *dto.UpdateResourcePayload) (*domain.Resource, error) {
			assert.Equal(t, "test-resource-id", id)
			assert.NotNil(t, payload.Name)
			assert.Equal(t, name, *payload.Name)
			assert.NotNil(t, payload.Interval)
			assert.Equal(t, interval, *payload.Interval)

			return &domain.Resource{
				Base:     domain.Base{ID: id},
				Name:     name,
				Type:     domain.ResourceHTTP,
				Target:   "https://example.com",
				Interval: interval,
				Timeout:  30,
				IsActive: true,
			}, nil
		},
	}
	handler := NewResourceHandler(mockService)

	payload := map[string]interface{}{
		"name":     name,
		"interval": interval,
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPatch, "/resources/test-resource-id", bytes.NewReader(body))
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("id", "test-resource-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))

	rec := httptest.NewRecorder()

	handler.UpdateResource(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var result domain.Resource
	err := json.NewDecoder(rec.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, "test-resource-id", result.ID)
	assert.Equal(t, name, result.Name)
	assert.Equal(t, interval, result.Interval)
}

func TestUpdateResource_PartialUpdate(t *testing.T) {
	newInterval := 240
	mockService := &mockResourceService{
		updateResourceFunc: func(ctx context.Context, id string, payload *dto.UpdateResourcePayload) (*domain.Resource, error) {
			assert.Equal(t, "test-resource-id", id)
			assert.NotNil(t, payload.Interval)
			assert.Equal(t, newInterval, *payload.Interval)
			assert.Nil(t, payload.Name) // Name should not be provided

			return &domain.Resource{
				Base:     domain.Base{ID: id},
				Name:     "Original Name",
				Type:     domain.ResourceHTTP,
				Target:   "https://example.com",
				Interval: newInterval,
				IsActive: true,
			}, nil
		},
	}
	handler := NewResourceHandler(mockService)

	payload := map[string]interface{}{
		"interval": newInterval,
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPut, "/resources/test-resource-id", bytes.NewReader(body))
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("id", "test-resource-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
	rec := httptest.NewRecorder()

	handler.UpdateResource(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var updated domain.Resource
	err := json.NewDecoder(rec.Body).Decode(&updated)
	require.NoError(t, err)

	assert.Equal(t, "Original Name", updated.Name)
	assert.Equal(t, newInterval, updated.Interval)
}

func TestUpdateResource_ValidationFailed(t *testing.T) {
	mockService := &mockResourceService{
		updateResourceFunc: func(ctx context.Context, id string, payload *dto.UpdateResourcePayload) (*domain.Resource, error) {
			return nil, service.ErrValidationFailed
		},
	}
	handler := NewResourceHandler(mockService)

	payload := map[string]interface{}{
		"target": "invalid-url",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPatch, "/resources/test-resource-id", bytes.NewReader(body))
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("id", "test-resource-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))

	rec := httptest.NewRecorder()

	handler.UpdateResource(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var errorResp map[string]string
	err := json.NewDecoder(rec.Body).Decode(&errorResp)
	require.NoError(t, err)

	assert.Contains(t, errorResp["error"], "validation failed")
}

func TestUpdateResource_NotFound(t *testing.T) {
	mockService := &mockResourceService{
		updateResourceFunc: func(ctx context.Context, id string, payload *dto.UpdateResourcePayload) (*domain.Resource, error) {
			return nil, service.ErrResourceNotFound
		},
	}
	handler := NewResourceHandler(mockService)

	payload := map[string]interface{}{
		"name": "Updated Name",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPatch, "/resources/non-existent-id", bytes.NewReader(body))
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("id", "non-existent-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))

	rec := httptest.NewRecorder()

	handler.UpdateResource(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var errorResp map[string]string
	err := json.NewDecoder(rec.Body).Decode(&errorResp)
	require.NoError(t, err)

	assert.Contains(t, errorResp["error"], "Resource not found")
}

func TestUpdateResource_MissingID(t *testing.T) {
	mockService := &mockResourceService{}
	handler := NewResourceHandler(mockService)

	payload := map[string]interface{}{
		"name": "Updated Name",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPatch, "/resources/", bytes.NewReader(body))
	ctx := chi.NewRouteContext()
	// Don't add ID to URL params
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))

	rec := httptest.NewRecorder()

	handler.UpdateResource(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var errorResp map[string]string
	err := json.NewDecoder(rec.Body).Decode(&errorResp)
	require.NoError(t, err)

	assert.Contains(t, errorResp["error"], "Resource ID is required")
}

func TestUpdateResource_InvalidJSON(t *testing.T) {
	mockService := &mockResourceService{}
	handler := NewResourceHandler(mockService)

	req := httptest.NewRequest(http.MethodPatch, "/resources/test-resource-id", bytes.NewReader([]byte("invalid json")))
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("id", "test-resource-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))

	rec := httptest.NewRecorder()

	handler.UpdateResource(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var errorResp map[string]string
	err := json.NewDecoder(rec.Body).Decode(&errorResp)
	require.NoError(t, err)

	assert.Contains(t, errorResp["error"], "Invalid request payload")
}

func TestUpdateResource_ServiceError(t *testing.T) {
	mockService := &mockResourceService{
		updateResourceFunc: func(ctx context.Context, id string, payload *dto.UpdateResourcePayload) (*domain.Resource, error) {
			return nil, errors.New("database connection failed")
		},
	}
	handler := NewResourceHandler(mockService)

	payload := map[string]interface{}{
		"name": "Updated Name",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPatch, "/resources/test-resource-id", bytes.NewReader(body))
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("id", "test-resource-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))

	rec := httptest.NewRecorder()

	handler.UpdateResource(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var errorResp map[string]string
	err := json.NewDecoder(rec.Body).Decode(&errorResp)
	require.NoError(t, err)

	assert.Contains(t, errorResp["error"], "Failed to update resource")
}

func TestUpdateResource_UpdateTargetWithValidation(t *testing.T) {
	newTarget := "https://api.newexample.com"
	mockService := &mockResourceService{
		updateResourceFunc: func(ctx context.Context, id string, payload *dto.UpdateResourcePayload) (*domain.Resource, error) {
			assert.Equal(t, "test-resource-id", id)
			assert.NotNil(t, payload.Target)
			assert.Equal(t, newTarget, *payload.Target)

			return &domain.Resource{
				Base:     domain.Base{ID: id},
				Name:     "API Monitor",
				Type:     domain.ResourceHTTP,
				Target:   newTarget,
				IsActive: true,
			}, nil
		},
	}
	handler := NewResourceHandler(mockService)

	payload := map[string]interface{}{
		"target": newTarget,
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPatch, "/resources/test-resource-id", bytes.NewReader(body))
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("id", "test-resource-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))

	rec := httptest.NewRecorder()

	handler.UpdateResource(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var result domain.Resource
	err := json.NewDecoder(rec.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, newTarget, result.Target)
}

func TestUpdateResource_PauseViaIsActive(t *testing.T) {
	isActive := false
	mockService := &mockResourceService{
		updateResourceFunc: func(ctx context.Context, id string, payload *dto.UpdateResourcePayload) (*domain.Resource, error) {
			assert.Equal(t, "test-resource-id", id)
			assert.NotNil(t, payload.IsActive)
			assert.Equal(t, isActive, *payload.IsActive)

			return &domain.Resource{
				Base:     domain.Base{ID: id},
				Name:     "API Monitor",
				Type:     domain.ResourceHTTP,
				Target:   "https://example.com",
				IsActive: isActive,
			}, nil
		},
	}
	handler := NewResourceHandler(mockService)

	payload := map[string]interface{}{
		"is_active": false,
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPatch, "/resources/test-resource-id", bytes.NewReader(body))
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("id", "test-resource-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))

	rec := httptest.NewRecorder()

	handler.UpdateResource(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var result domain.Resource
	err := json.NewDecoder(rec.Body).Decode(&result)
	require.NoError(t, err)

	assert.False(t, result.IsActive)
}

func TestDeleteResource_Success(t *testing.T) {
	mockService := &mockResourceService{
		deleteResourceFunc: func(ctx context.Context, id string) error {
			assert.Equal(t, "test-resource-id", id)
			return nil
		},
	}
	handler := NewResourceHandler(mockService)

	req := httptest.NewRequest(http.MethodDelete, "/resources/test-resource-id", nil)
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("id", "test-resource-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))

	rec := httptest.NewRecorder()

	handler.DeleteResource(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Empty(t, rec.Body.String(), "Expected empty response body for 204")
}

func TestDeleteResource_NotFound(t *testing.T) {
	mockService := &mockResourceService{
		deleteResourceFunc: func(ctx context.Context, id string) error {
			return service.ErrResourceNotFound
		},
	}
	handler := NewResourceHandler(mockService)

	req := httptest.NewRequest(http.MethodDelete, "/resources/non-existent-id", nil)
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("id", "non-existent-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))

	rec := httptest.NewRecorder()

	handler.DeleteResource(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var errorResp map[string]string
	err := json.NewDecoder(rec.Body).Decode(&errorResp)
	require.NoError(t, err)

	assert.Contains(t, errorResp["error"], "Resource not found")
}

func TestDeleteResource_MissingID(t *testing.T) {
	mockService := &mockResourceService{}
	handler := NewResourceHandler(mockService)

	req := httptest.NewRequest(http.MethodDelete, "/resources/", nil)
	ctx := chi.NewRouteContext()
	// Don't add ID to URL params
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))

	rec := httptest.NewRecorder()

	handler.DeleteResource(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var errorResp map[string]string
	err := json.NewDecoder(rec.Body).Decode(&errorResp)
	require.NoError(t, err)

	assert.Contains(t, errorResp["error"], "Resource ID is required")
}

func TestDeleteResource_ServiceError(t *testing.T) {
	mockService := &mockResourceService{
		deleteResourceFunc: func(ctx context.Context, id string) error {
			return errors.New("database connection failed")
		},
	}
	handler := NewResourceHandler(mockService)

	req := httptest.NewRequest(http.MethodDelete, "/resources/test-resource-id", nil)
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("id", "test-resource-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))

	rec := httptest.NewRecorder()

	handler.DeleteResource(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var errorResp map[string]string
	err := json.NewDecoder(rec.Body).Decode(&errorResp)
	require.NoError(t, err)

	assert.Contains(t, errorResp["error"], "Failed to delete resource")
}

func TestGetLive_200_AllFields(t *testing.T) {
	fetchedAt := time.Now().UTC()
	avgResponse := 142
	lastResponse := 121
	resourceID := "res-live-123"
	liveService := &mockLiveSnapshotService{
		getLiveSnapshotFunc: func(ctx context.Context, id string) (*dto.LiveSnapshotResponse, error) {
			assert.Equal(t, resourceID, id)
			return &dto.LiveSnapshotResponse{
				Resource: &domain.Resource{Base: domain.Base{ID: id}, Name: "API Monitor", Type: domain.ResourceHTTP, Target: "https://example.com"},
				Stats: dto.LiveStats{
					Uptime2h:           ptrFloat64(100),
					Uptime24h:          ptrFloat64(99.9),
					Uptime7d:           ptrFloat64(99.7),
					Uptime30d:          ptrFloat64(99.5),
					AvgResponseTime24h: &avgResponse,
					LastResponseTime:   &lastResponse,
				},
				ActiveIncident:   &dto.LiveActiveIncident{ID: "inc-1", StartedAt: fetchedAt.Add(-10 * time.Minute), Cause: "timeout"},
				RecentActivities: []*domain.MonitoringActivity{{ResponseTime: 120}, {ResponseTime: 125}},
				FetchedAt:        fetchedAt,
			}, nil
		},
	}
	h := NewResourceHandler(&mockResourceService{}, liveService)

	req := httptest.NewRequest(http.MethodGet, "/resources/"+resourceID+"/live", nil)
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("id", resourceID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
	rec := httptest.NewRecorder()

	h.GetLive(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&body)
	require.NoError(t, err)
	require.NotNil(t, body["resource"])
	require.NotNil(t, body["stats"])
	require.NotNil(t, body["active_incident"])
	require.NotNil(t, body["recent_activities"])
	require.NotNil(t, body["fetched_at"])

	stats, ok := body["stats"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, stats, "uptime_2h")
	assert.Contains(t, stats, "uptime_24h")
	assert.Contains(t, stats, "uptime_7d")
	assert.Contains(t, stats, "uptime_30d")
	assert.Contains(t, stats, "avg_response_time_24h")
	assert.Contains(t, stats, "last_response_time")

	if fetchedAtString, ok := body["fetched_at"].(string); ok {
		parsed, parseErr := time.Parse(time.RFC3339Nano, fetchedAtString)
		require.NoError(t, parseErr)
		assert.Equal(t, time.UTC, parsed.Location())
	}
}

func TestGetLive_404_UnknownID(t *testing.T) {
	liveService := &mockLiveSnapshotService{
		getLiveSnapshotFunc: func(ctx context.Context, resourceID string) (*dto.LiveSnapshotResponse, error) {
			return nil, service.ErrResourceNotFound
		},
	}
	h := NewResourceHandler(&mockResourceService{}, liveService)

	req := httptest.NewRequest(http.MethodGet, "/resources/unknown/live", nil)
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("id", "unknown")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
	rec := httptest.NewRecorder()

	h.GetLive(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var errorResp map[string]string
	err := json.NewDecoder(rec.Body).Decode(&errorResp)
	require.NoError(t, err)
	assert.Contains(t, errorResp["error"], "Resource not found")
}

func TestGetLive_NullActiveIncident(t *testing.T) {
	liveService := &mockLiveSnapshotService{
		getLiveSnapshotFunc: func(ctx context.Context, resourceID string) (*dto.LiveSnapshotResponse, error) {
			return &dto.LiveSnapshotResponse{
				Resource:         &domain.Resource{Base: domain.Base{ID: resourceID}},
				Stats:            dto.LiveStats{},
				ActiveIncident:   nil,
				RecentActivities: []*domain.MonitoringActivity{},
				FetchedAt:        time.Now().UTC(),
			}, nil
		},
	}
	h := NewResourceHandler(&mockResourceService{}, liveService)

	req := httptest.NewRequest(http.MethodGet, "/resources/r1/live", nil)
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("id", "r1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
	rec := httptest.NewRecorder()

	h.GetLive(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&body)
	require.NoError(t, err)
	assert.Nil(t, body["active_incident"])
}

func TestGetLive_Unauthorized(t *testing.T) {
	liveService := &mockLiveSnapshotService{
		getLiveSnapshotFunc: func(ctx context.Context, resourceID string) (*dto.LiveSnapshotResponse, error) {
			t.Fatalf("live service should not be called without auth")
			return nil, nil
		},
	}
	h := NewResourceHandler(&mockResourceService{}, liveService)

	r := chi.NewRouter()
	r.Use(middleware.AuthMiddleware(nil, nil))
	r.Get("/resources/{id}/live", h.GetLive)

	req := httptest.NewRequest(http.MethodGet, "/resources/r1/live", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestGetLive_RecentActivitiesMax20(t *testing.T) {
	recent := make([]*domain.MonitoringActivity, 20)
	for i := range recent {
		recent[i] = &domain.MonitoringActivity{ResponseTime: i + 100}
	}

	liveService := &mockLiveSnapshotService{
		getLiveSnapshotFunc: func(ctx context.Context, resourceID string) (*dto.LiveSnapshotResponse, error) {
			return &dto.LiveSnapshotResponse{
				Resource:         &domain.Resource{Base: domain.Base{ID: resourceID}},
				Stats:            dto.LiveStats{},
				ActiveIncident:   nil,
				RecentActivities: recent,
				FetchedAt:        time.Now().UTC(),
			}, nil
		},
	}
	h := NewResourceHandler(&mockResourceService{}, liveService)

	req := httptest.NewRequest(http.MethodGet, "/resources/r1/live", nil)
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("id", "r1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
	rec := httptest.NewRecorder()

	h.GetLive(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&body)
	require.NoError(t, err)
	recentActivities, ok := body["recent_activities"].([]interface{})
	require.True(t, ok)
	assert.LessOrEqual(t, len(recentActivities), 20)
}

func TestGetLive_FetchedAtIsUTC(t *testing.T) {
	liveService := &mockLiveSnapshotService{
		getLiveSnapshotFunc: func(ctx context.Context, resourceID string) (*dto.LiveSnapshotResponse, error) {
			return &dto.LiveSnapshotResponse{
				Resource:         &domain.Resource{Base: domain.Base{ID: resourceID}},
				Stats:            dto.LiveStats{},
				ActiveIncident:   nil,
				RecentActivities: []*domain.MonitoringActivity{},
				FetchedAt:        time.Now().UTC(),
			}, nil
		},
	}
	h := NewResourceHandler(&mockResourceService{}, liveService)

	req := httptest.NewRequest(http.MethodGet, "/resources/r1/live", nil)
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("id", "r1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
	rec := httptest.NewRecorder()

	h.GetLive(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&body)
	require.NoError(t, err)

	fetchedAtStr, ok := body["fetched_at"].(string)
	require.True(t, ok)
	parsed, parseErr := time.Parse(time.RFC3339Nano, fetchedAtStr)
	require.NoError(t, parseErr)
	assert.Equal(t, time.UTC, parsed.Location())
}

func ptrFloat64(v float64) *float64 {
	return &v
}
