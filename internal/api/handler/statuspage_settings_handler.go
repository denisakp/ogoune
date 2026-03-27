package handler

import (
	"encoding/json"
	"net/http"

	"github.com/denisakp/pulseguard/internal/api/response"
	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/dto"
	"github.com/denisakp/pulseguard/internal/service"
	"github.com/go-chi/chi/v5"
)

// StatusPageSettingsHandler handles HTTP requests for status page settings
type StatusPageSettingsHandler struct {
	service *service.StatusPageSettingsService
}

// NewStatusPageSettingsHandler creates a new handler instance
func NewStatusPageSettingsHandler(service *service.StatusPageSettingsService) *StatusPageSettingsHandler {
	return &StatusPageSettingsHandler{service: service}
}

// RegisterRoutes registers the settings routes
func (h *StatusPageSettingsHandler) RegisterRoutes(r chi.Router) {
	r.Get("/settings/statuspage", h.GetSettings)
	r.Put("/settings/statuspage", h.UpdateSettings)
}

// GetSettings retrieves the current status page settings
func (h *StatusPageSettingsHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.service.GetSettings(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to retrieve settings")
		return
	}

	resp := dto.StatusPageSettingsResponse{
		ID:                   settings.ID,
		Name:                 settings.Name,
		HomepageURL:          settings.HomepageURL,
		CustomDomain:         settings.CustomDomain,
		GoogleAnalyticsID:    settings.GoogleAnalyticsID,
		EnableDetailsPage:    settings.EnableDetailsPage,
		ShowUptimePercentage: settings.ShowUptimePercentage,
		HidePausedMonitors:   settings.HidePausedMonitors,
		ShowIncidentHistory:  settings.ShowIncidentHistory,
		CreatedAt:            settings.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:            settings.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.JSON(w, http.StatusOK, resp)
}

// UpdateSettings updates the status page settings
func (h *StatusPageSettingsHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var req dto.StatusPageSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	settings := &domain.StatusPageSettings{
		Name:                 req.Name,
		HomepageURL:          req.HomepageURL,
		CustomDomain:         req.CustomDomain,
		GoogleAnalyticsID:    req.GoogleAnalyticsID,
		EnableDetailsPage:    req.EnableDetailsPage,
		ShowUptimePercentage: req.ShowUptimePercentage,
		HidePausedMonitors:   req.HidePausedMonitors,
		ShowIncidentHistory:  req.ShowIncidentHistory,
	}

	if err := h.service.UpdateSettings(r.Context(), settings); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to update settings")
		return
	}

	// Return updated settings
	updated, err := h.service.GetSettings(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to retrieve updated settings")
		return
	}

	resp := dto.StatusPageSettingsResponse{
		ID:                   updated.ID,
		Name:                 updated.Name,
		HomepageURL:          updated.HomepageURL,
		CustomDomain:         updated.CustomDomain,
		GoogleAnalyticsID:    updated.GoogleAnalyticsID,
		EnableDetailsPage:    updated.EnableDetailsPage,
		ShowUptimePercentage: updated.ShowUptimePercentage,
		HidePausedMonitors:   updated.HidePausedMonitors,
		ShowIncidentHistory:  updated.ShowIncidentHistory,
		CreatedAt:            updated.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:            updated.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.JSON(w, http.StatusOK, resp)
}
