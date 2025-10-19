package handler

import (
	"net/http"

	"github.com/denisakp/pulseguard/internal/api/response"
	"github.com/denisakp/pulseguard/internal/service"
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
