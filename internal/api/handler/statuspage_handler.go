package handler

import (
	"net/http"

	"github.com/denisakp/pulseguard/internal/api/response"
	"github.com/denisakp/pulseguard/internal/repository"
	"github.com/denisakp/pulseguard/internal/service"
	"github.com/go-chi/chi/v5"
)

// StatusPageHandler handles requests for the public status page.
type StatusPageHandler struct {
	service *service.StatusPageService
}

// NewStatusPageHandler creates a new StatusPageHandler instance.
func NewStatusPageHandler(service *service.StatusPageService) *StatusPageHandler {
	return &StatusPageHandler{
		service: service,
	}
}

// HandleStatusPage returns the public status page data as JSON.
func (h *StatusPageHandler) HandleStatusPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Fetch status page data
	data, err := h.service.GetData(ctx)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to load status page data: "+err.Error())
		return
	}

	// Return JSON response
	response.JSON(w, http.StatusOK, data)
}

// HandleResourceDetailStatus returns detailed status information for a single resource
func (h *StatusPageHandler) HandleResourceDetailStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get resource ID from URL path
	resourceID := chi.URLParam(r, "resourceId")

	if resourceID == "" {
		response.Error(w, http.StatusBadRequest, "Resource ID is required")
		return
	}

	// Fetch resource detail status
	data, err := h.service.GetResourceDetailStatus(ctx, resourceID)
	if err != nil {
		// Check if it's a not found error
		if err.Error() == "failed to fetch resource: "+repository.ErrNotFound.Error() {
			response.Error(w, http.StatusNotFound, "Resource not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to load resource status: "+err.Error())
		return
	}

	// Return JSON response
	response.JSON(w, http.StatusOK, data)
}
