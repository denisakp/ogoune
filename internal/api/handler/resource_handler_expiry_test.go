package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/dto"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateResource_InvalidExpiryThresholds verifies that invalid
// expiry_alert_thresholds values return 422 before hitting the service.
func TestCreateResource_InvalidExpiryThresholds(t *testing.T) {
	cases := []struct {
		name       string
		thresholds string
	}{
		{"letters", "abc"},
		{"value zero", "0,7,30"},
		{"value too large", "7,400"},
		{"mixed valid invalid", "7,abc,30"},
		{"negative", "-1,7"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			mockService := &mockResourceService{}
			handler := NewResourceHandler(mockService)

			thresholds := tc.thresholds
			payload := dto.CreateResourcePayload{
				Name:                  "Test Monitor",
				Target:                "https://example.com",
				Type:                  domain.ResourceHTTP,
				Timeout:               30,
				Interval:              60,
				ExpiryAlertThresholds: &thresholds,
			}
			body, _ := json.Marshal(payload)

			req := httptest.NewRequest(http.MethodPost, "/resources", bytes.NewReader(body))
			rec := httptest.NewRecorder()

			handler.CreateResource(rec, req)

			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)

			var resp map[string]string
			err := json.NewDecoder(rec.Body).Decode(&resp)
			require.NoError(t, err)
			assert.Contains(t, resp["error"], "expiry_alert_thresholds")
		})
	}
}

// TestCreateResource_ValidExpiryThresholds verifies that well-formed thresholds
// pass validation and reach the service.
func TestCreateResource_ValidExpiryThresholds(t *testing.T) {
	called := false
	mockService := &mockResourceService{
		createResourceFunc: func(_ context.Context, payload *dto.CreateResourcePayload) (*domain.Resource, error) {
			called = true
			require.NotNil(t, payload.ExpiryAlertThresholds)
			assert.Equal(t, "30,14,7,1", *payload.ExpiryAlertThresholds)
			return &domain.Resource{Base: domain.Base{ID: "res-1"}, Name: payload.Name}, nil
		},
	}
	handler := NewResourceHandler(mockService)

	thresholds := "30,14,7,1"
	payload := dto.CreateResourcePayload{
		Name:                  "Test Monitor",
		Target:                "https://example.com",
		Type:                  domain.ResourceHTTP,
		Timeout:               30,
		Interval:              60,
		ExpiryAlertThresholds: &thresholds,
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/resources", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.CreateResource(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.True(t, called, "service.CreateResource should have been invoked")
}

// TestUpdateResource_InvalidExpiryThresholds verifies that an invalid threshold
// string on update returns 422 before hitting the service.
func TestUpdateResource_InvalidExpiryThresholds(t *testing.T) {
	mockService := &mockResourceService{}
	handler := NewResourceHandler(mockService)

	thresholds := "xyz,0"
	payload := dto.UpdateResourcePayload{
		ExpiryAlertThresholds: &thresholds,
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPatch, "/resources/test-id", bytes.NewReader(body))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "test-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()

	handler.UpdateResource(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)

	var resp map[string]string
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Contains(t, resp["error"], "expiry_alert_thresholds")
}
