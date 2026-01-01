package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/denisakp/pulseguard/internal/api/response"
	"github.com/denisakp/pulseguard/internal/dto"
	"github.com/denisakp/pulseguard/internal/service"
	"github.com/go-chi/chi/v5"
)

// ComponentServiceInterface defines the methods required by ComponentHandler.
type ComponentServiceInterface interface {
	CreateComponent(ctx context.Context, payload *dto.CreateComponentPayload) (*dto.ComponentResponse, error)
	UpdateComponent(ctx context.Context, id string, payload *dto.UpdateComponentPayload) (*dto.ComponentResponse, error)
	DeleteComponent(ctx context.Context, id string) error
	GetComponent(ctx context.Context, id string) (*dto.ComponentResponse, error)
	ListComponents(ctx context.Context, limit, offset int) ([]*dto.ComponentResponse, error)
	BulkAssignToComponent(ctx context.Context, componentID string, payload *dto.BulkAssignPayload) error
	BulkRemoveFromComponent(ctx context.Context, payload *dto.BulkRemovePayload) error
}

// ComponentHandler handles CRUD endpoints for components.
type ComponentHandler struct {
	service ComponentServiceInterface
}

// NewComponentHandler constructs a ComponentHandler.
func NewComponentHandler(service ComponentServiceInterface) *ComponentHandler {
	return &ComponentHandler{service: service}
}

func (h *ComponentHandler) ListComponents(w http.ResponseWriter, r *http.Request) {
	components, err := h.service.ListComponents(r.Context(), 1000, 0)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, components)
}

func (h *ComponentHandler) GetComponent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	component, err := h.service.GetComponent(r.Context(), id)
	if err != nil {
		if err == service.ErrResourceNotFound {
			response.Error(w, http.StatusNotFound, "Component not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, component)
}

func (h *ComponentHandler) CreateComponent(w http.ResponseWriter, r *http.Request) {
	var payload dto.CreateComponentPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid payload")
		return
	}

	component, err := h.service.CreateComponent(r.Context(), &payload)
	if err != nil {
		if err == service.ErrValidationFailed {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, component)
}

func (h *ComponentHandler) UpdateComponent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var payload dto.UpdateComponentPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid payload")
		return
	}

	component, err := h.service.UpdateComponent(r.Context(), id, &payload)
	if err != nil {
		if err == service.ErrValidationFailed {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		if err == service.ErrResourceNotFound {
			response.Error(w, http.StatusNotFound, "Component not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, component)
}

func (h *ComponentHandler) DeleteComponent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.service.DeleteComponent(r.Context(), id); err != nil {
		if err == service.ErrValidationFailed {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		if err == service.ErrResourceNotFound {
			response.Error(w, http.StatusNotFound, "Component not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// BulkAssignToComponent assigns multiple resources to a component.
func (h *ComponentHandler) BulkAssignToComponent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var payload dto.BulkAssignPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid payload")
		return
	}

	if err := h.service.BulkAssignToComponent(r.Context(), id, &payload); err != nil {
		if err == service.ErrValidationFailed {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		if err == service.ErrResourceNotFound {
			response.Error(w, http.StatusNotFound, "Component not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "Resources assigned successfully"})
}

// BulkRemoveFromComponent removes resources from their components.
func (h *ComponentHandler) BulkRemoveFromComponent(w http.ResponseWriter, r *http.Request) {
	var payload dto.BulkRemovePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid payload")
		return
	}

	if err := h.service.BulkRemoveFromComponent(r.Context(), &payload); err != nil {
		if err == service.ErrValidationFailed {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		if err == service.ErrResourceNotFound {
			response.Error(w, http.StatusNotFound, "Resource not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "Resources removed successfully"})
}
