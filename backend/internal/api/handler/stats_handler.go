package handler

import (
	"net/http"

	"github.com/denisakp/pulseguard/internal/api/response"
	"github.com/denisakp/pulseguard/internal/service"
)

// StatsHandler handles HTTP requests for aggregated statistics.
type StatsHandler struct {
	service *service.StatsService
}

// NewStatsHandler creates a new stats handler.
func NewStatsHandler(service *service.StatsService) *StatsHandler {
	return &StatsHandler{
		service: service,
	}
}

// GetSummary handles GET /stats/summary?range=24h requests.
// Returns aggregated uptime and incident statistics for all monitored resources.
// Query parameters:
//   - range: Time range for statistics (2h, 24h, 7d, 30d). Default: 24h
func (h *StatsHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameter
	timeRange := r.URL.Query().Get("range")
	if timeRange == "" {
		timeRange = "24h" // Default to 24 hours
	}

	// Validate time range
	validRanges := map[string]bool{
		"2h":  true,
		"24h": true,
		"7d":  true,
		"30d": true,
	}

	if !validRanges[timeRange] {
		response.Error(w, http.StatusBadRequest, "Invalid time range. Must be one of: 2h, 24h, 7d, 30d")
		return
	}

	// Get aggregated statistics
	summary, err := h.service.GetSummary(ctx, timeRange)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to fetch statistics summary: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, summary)
}
