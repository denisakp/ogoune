package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/dto"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateResource_ConfirmationValidation400(t *testing.T) {
	h := NewResourceHandler(&mockResourceService{})

	checks := 0
	payload := dto.CreateResourcePayload{
		Name:               "invalid",
		Type:               domain.ResourceHTTP,
		Target:             "https://example.com",
		Interval:           60,
		Timeout:            5,
		ConfirmationChecks: &checks,
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/resources", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.CreateResource(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPatchResource_ConfirmationValidation400(t *testing.T) {
	h := NewResourceHandler(&mockResourceService{})
	interval := 30
	confirmInterval := 30
	payload := dto.UpdateResourcePayload{
		Interval:             &interval,
		ConfirmationInterval: &confirmInterval,
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPatch, "/resources/r-1", bytes.NewReader(body))
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("id", "r-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))

	rec := httptest.NewRecorder()
	h.UpdateResource(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetResource_ReadResponseIncludesConfirmationFields(t *testing.T) {
	mockSvc := &mockResourceService{
		getResourceByIDWithResponseTimesFunc: func(ctx context.Context, id string, limit int) (*dto.ResourceResponse, error) {
			res := domain.Resource{
				Base:                 domain.Base{ID: id},
				Name:                 "r",
				Type:                 domain.ResourceHTTP,
				Target:               "https://example.com",
				Interval:             60,
				Timeout:              5,
				ConfirmationChecks:   3,
				ConfirmationInterval: 15,
			}
			return &dto.ResourceResponse{Resource: res}, nil
		},
	}
	h := NewResourceHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/resources/r-2", nil)
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("id", "r-2")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))

	rec := httptest.NewRecorder()
	h.GetResourceByID(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var payload map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
	_, hasChecks := payload["confirmation_checks"]
	_, hasInterval := payload["confirmation_interval"]
	assert.True(t, hasChecks)
	assert.True(t, hasInterval)
}
