package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/denisakp/pulseguard/internal/api/response"
	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/service"
	"github.com/go-chi/chi/v5"
)

// IntegrationServiceInterface defines the methods required by IntegrationHandler.
type IntegrationServiceInterface interface {
	CreateIntegration(ctx context.Context, integration *domain.Integration) error
	ListIntegrations(ctx context.Context, limit, offset int) ([]*domain.Integration, error)
	GetIntegrationByID(ctx context.Context, id string) (*domain.Integration, error)
	UpdateIntegration(ctx context.Context, id string, name string, config []byte, isActive *bool) (*domain.Integration, error)
	ListActiveIntegrations(ctx context.Context) ([]*domain.Integration, error)
}

// IntegrationHandler handles HTTP requests for integration management.
type IntegrationHandler struct {
	integrationService IntegrationServiceInterface
}

// NewIntegrationHandler creates a new IntegrationHandler with injected dependencies.
func NewIntegrationHandler(integrationService IntegrationServiceInterface) *IntegrationHandler {
	return &IntegrationHandler{
		integrationService: integrationService,
	}
}

// CreateIntegration handles POST /integrations - creates a new integration.
func (h *IntegrationHandler) CreateIntegration(w http.ResponseWriter, r *http.Request) {
	var integration domain.Integration

	if err := json.NewDecoder(r.Body).Decode(&integration); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	// Validation
	if integration.Name == "" {
		response.Error(w, http.StatusBadRequest, "Integration name is required")
		return
	}

	// Validate Config contains required fields
	config, err := integration.GetConfig()
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid config format: "+err.Error())
		return
	}

	// Check if type is present in config
	if _, ok := config["type"]; !ok {
		response.Error(w, http.StatusBadRequest, "Integration config must include 'type' field")
		return
	}

	if err := h.integrationService.CreateIntegration(r.Context(), &integration); err != nil {
		if errors.Is(err, service.ErrValidationFailed) {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to create integration: "+err.Error())
		return
	}

	response.Created(w, integration)
}

// ListIntegrations handles GET /integrations - retrieves all integrations.
func (h *IntegrationHandler) ListIntegrations(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	activeStr := r.URL.Query().Get("active")

	limit := 50 // default
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	var integrations []*domain.Integration
	var err error

	// If active filter is specified, use the active integrations method
	if activeStr == "true" {
		integrations, err = h.integrationService.ListActiveIntegrations(r.Context())
	} else {
		integrations, err = h.integrationService.ListIntegrations(r.Context(), limit, offset)
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to retrieve integrations: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, integrations)
}

// UpdateIntegration handles PATCH /integrations/{id} - updates an existing integration.
func (h *IntegrationHandler) UpdateIntegration(w http.ResponseWriter, r *http.Request) {
	integrationID := chi.URLParam(r, "id")
	if integrationID == "" {
		response.Error(w, http.StatusBadRequest, "Integration ID is required")
		return
	}

	var payload struct {
		Name     string          `json:"name,omitempty"`
		Config   json.RawMessage `json:"config,omitempty"`
		IsActive *bool           `json:"is_active,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	updatedIntegration, err := h.integrationService.UpdateIntegration(
		r.Context(),
		integrationID,
		payload.Name,
		payload.Config,
		payload.IsActive,
	)

	if err != nil {
		if errors.Is(err, service.ErrValidationFailed) {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, service.ErrResourceNotFound) {
			response.Error(w, http.StatusNotFound, "Integration not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to update integration: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, updatedIntegration)
}
