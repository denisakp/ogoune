package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/denisakp/pulseguard/internal/api/response"
	"github.com/denisakp/pulseguard/internal/dto"
	"github.com/denisakp/pulseguard/internal/service"
	"github.com/go-chi/chi/v5"
)

// MaintenanceServiceInterface encapsulates service methods used by handler
type MaintenanceServiceInterface interface {
	Create(ctxContext interface{}, payload *dto.MaintenanceCreatePayload) (interface{}, error)
}

// MaintenanceHandler handles HTTP endpoints for maintenances
type MaintenanceHandler struct {
	service *service.MaintenanceService
}

func NewMaintenanceHandler(service *service.MaintenanceService) *MaintenanceHandler {
	return &MaintenanceHandler{service: service}
}

// CreateMaintenance handles POST /maintenances
func (h *MaintenanceHandler) CreateMaintenance(w http.ResponseWriter, r *http.Request) {
	var payload dto.MaintenanceCreatePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}
	created, err := h.service.Create(r.Context(), &payload)
	if err != nil {
		if err == service.ErrValidationFailed {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, created)
}

// UpdateMaintenance handles PATCH /maintenances/{id}
func (h *MaintenanceHandler) UpdateMaintenance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var payload dto.MaintenanceUpdatePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}
	updated, err := h.service.Update(r.Context(), id, &payload)
	if err != nil {
		switch err {
		case service.ErrValidationFailed:
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		case service.ErrMaintenanceNotFound:
			response.Error(w, http.StatusNotFound, err.Error())
			return
		default:
			response.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	response.JSON(w, http.StatusOK, updated)
}

// DeleteMaintenance handles DELETE /maintenances/{id}
func (h *MaintenanceHandler) DeleteMaintenance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.service.Delete(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// FinishMaintenance handles POST /maintenances/{id}/finish
func (h *MaintenanceHandler) FinishMaintenance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	finished, err := h.service.Finish(r.Context(), id)
	if err != nil {
		if err == service.ErrMaintenanceNotFound {
			response.Error(w, http.StatusNotFound, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, finished)
}

// ListMaintenances handles GET /maintenances?status=scheduled|active|finished&limit=50&offset=0
func (h *MaintenanceHandler) ListMaintenances(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	limit := 50
	offset := 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	list, err := h.service.List(r.Context(), status, limit, offset)
	if err != nil {
		if err == service.ErrValidationFailed {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, list)
}
