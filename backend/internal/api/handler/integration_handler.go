package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/denisakp/pulseguard/internal/api/response"
	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/dto"
	"github.com/denisakp/pulseguard/internal/service"
	"github.com/go-chi/chi/v5"
)

// IntegrationServiceInterface defines the methods required by IntegrationHandler.
type IntegrationServiceInterface interface {
	CreateIntegration(ctx context.Context, integration *dto.CreateIntegrationPayload) (*domain.Integration, error)
	ListIntegrations(ctx context.Context, limit, offset int) ([]*domain.Integration, error)
	GetIntegrationByID(ctx context.Context, id string) (*domain.Integration, error)
	UpdateIntegration(ctx context.Context, id string, payload *dto.UpdateIntegrationPayload) (*domain.Integration, error)
	ListActiveIntegrations(ctx context.Context) ([]*domain.Integration, error)
}

// IntegrationHandler handles HTTP requests for integration management.
type IntegrationHandler struct {
	integrationService IntegrationServiceInterface
	validTypes         map[string]bool
	validEventTypes    map[domain.EventType]bool
}

// NewIntegrationHandler creates a new IntegrationHandler with injected dependencies.
func NewIntegrationHandler(integrationService IntegrationServiceInterface) *IntegrationHandler {
	return &IntegrationHandler{
		integrationService: integrationService,
		validTypes: map[string]bool{
			string(domain.IntegrationSlack):      true,
			string(domain.IntegrationGoogleChat): true,
			string(domain.IntegrationDiscord):    true,
		},
		validEventTypes: map[domain.EventType]bool{
			domain.EventTypeDown:   true,
			domain.EventTypeUp:     true,
			domain.EventTypeExpiry: true,
		},
	}
}

// CreateIntegration handles POST /integrations - creates a new integration.
func (h *IntegrationHandler) CreateIntegration(w http.ResponseWriter, r *http.Request) {
	var payload dto.CreateIntegrationPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	// Validation
	if err := h.validateIntegrationPayload(payload.Name, payload.Config, payload.EventTypes); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	integration, err := h.integrationService.CreateIntegration(r.Context(), &payload)
	if err != nil {
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

	var payload dto.UpdateIntegrationPayload

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	// validation
	if err := h.validateIntegrationPayload(payload.Name, payload.Config, payload.EventTypes); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	updatedIntegration, err := h.integrationService.UpdateIntegration(r.Context(), integrationID, &payload)
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

func (h *IntegrationHandler) validateIntegrationPayload(name string, config map[string]interface{}, eventTypes []domain.EventType) error {
	// validate name
	if name == "" {
		return fmt.Errorf("integration name is required")
	}

	if config == nil || len(config) == 0 {
		return fmt.Errorf("integration config is required")
	}

	// validate config.type
	typeVal, ok := config["type"].(string)
	if !ok || typeVal == "" {
		return fmt.Errorf("integration config must include 'type' field")
	}

	// validate integration type
	if !h.validTypes[typeVal] {
		return fmt.Errorf("invalid integration type '%s'")
	}

	if len(eventTypes) == 0 {
		return fmt.Errorf("at least one event type is required")
	}

	// validate each event type
	for _, eventType := range eventTypes {
		if !h.validEventTypes[eventType] {
			return fmt.Errorf("invalid event type '%s'", string(eventType))
		}
	}

	return nil
}
