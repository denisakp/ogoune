package handler

import (
	"net/http"
	"strconv"

	"github.com/denisakp/pulseguard/internal/api/response"
	"github.com/denisakp/pulseguard/internal/service"
	"github.com/go-chi/chi/v5"
)

// MonitoringActivityHandler handles HTTP requests for monitoring activities.
type MonitoringActivityHandler struct {
	service *service.MonitoringActivityService
}

// NewMonitoringActivityHandler creates a new monitoring activity handler.
func NewMonitoringActivityHandler(service *service.MonitoringActivityService) *MonitoringActivityHandler {
	return &MonitoringActivityHandler{
		service: service,
	}
}

// ListActivities handles GET /monitoring-activities requests.
// Query parameters:
//   - limit: Maximum number of activities to return (default: 50)
//   - offset: Number of activities to skip (default: 0)
//   - resource_id: Filter activities by resource ID (optional)
func (h *MonitoringActivityHandler) ListActivities(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	resourceID := r.URL.Query().Get("resource_id")

	// Set defaults
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	// Fetch activities based on filters
	var activities interface{}
	var err error

	if resourceID != "" {
		activities, err = h.service.ListByResourceID(ctx, resourceID, limit, offset)
	} else {
		activities, err = h.service.ListAll(ctx, limit, offset)
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to fetch monitoring activities: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"activities": activities,
		"limit":      limit,
		"offset":     offset,
	})
}

// GetUptimeStats handles GET /resources/{resourceId}/uptime-stats requests.
// Returns hourly uptime percentage for the last 24 hours.
func (h *MonitoringActivityHandler) GetUptimeStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	resourceID := chi.URLParam(r, "resourceId")

	if resourceID == "" {
		response.Error(w, http.StatusBadRequest, "Resource ID is required")
		return
	}

	stats, err := h.service.GetUptimeStats(ctx, resourceID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to fetch uptime stats: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"resource_id": resourceID,
		"stats":       stats,
	})
}
