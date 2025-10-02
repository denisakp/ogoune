package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/service"
	"github.com/go-chi/chi/v5"
)

// IntegrationServiceInterface defines the methods required by IntegrationHandler.
type IntegrationServiceInterface interface {
	CreateIntegration(ctx context.Context, integration *domain.Integration) error
	ListIntegrations(ctx context.Context, limit, offset int) ([]*domain.Integration, error)
	GetIntegrationByID(ctx context.Context, id string) (*domain.Integration, error)
	UpdateIntegration(ctx context.Context, id string, name, target string, isActive *bool) (*domain.Integration, error)
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
		respondError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	// Validation
	if integration.Name == "" {
		respondError(w, http.StatusBadRequest, "Integration name is required")
		return
	}

	if integration.Target == "" {
		respondError(w, http.StatusBadRequest, "Integration target is required")
		return
	}

	if integration.Type == "" {
		respondError(w, http.StatusBadRequest, "Integration type is required")
		return
	}

	if err := h.integrationService.CreateIntegration(r.Context(), &integration); err != nil {
		if errors.Is(err, service.ErrValidationFailed) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to create integration: "+err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, integration)
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
		respondError(w, http.StatusInternalServerError, "Failed to retrieve integrations: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, integrations)
}

// UpdateIntegration handles PATCH /integrations/{id} - updates an existing integration.
func (h *IntegrationHandler) UpdateIntegration(w http.ResponseWriter, r *http.Request) {
	integrationID := chi.URLParam(r, "id")
	if integrationID == "" {
		respondError(w, http.StatusBadRequest, "Integration ID is required")
		return
	}

	var payload struct {
		Name     string `json:"name,omitempty"`
		Target   string `json:"target,omitempty"`
		IsActive *bool  `json:"is_active,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	updatedIntegration, err := h.integrationService.UpdateIntegration(
		r.Context(),
		integrationID,
		payload.Name,
		payload.Target,
		payload.IsActive,
	)

	if err != nil {
		if errors.Is(err, service.ErrValidationFailed) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, service.ErrResourceNotFound) {
			respondError(w, http.StatusNotFound, "Integration not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to update integration: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, updatedIntegration)
}
