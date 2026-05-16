package v1

import (
	"context"
	"net/http"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	dtoV1 "github.com/denisakp/ogoune/internal/dto/v1"
	"github.com/go-chi/chi/v5"
)

// IncidentV1ServiceInterface defines the incident service methods used by the v1 incident handler.
type IncidentV1ServiceInterface interface {
	ListAll(ctx context.Context, limit, offset int) ([]*domain.Incident, error)
	GetIncidentByID(ctx context.Context, id string) (*domain.Incident, error)
}

// IncidentHandler handles v1 read endpoints for incidents.
type IncidentHandler struct {
	service IncidentV1ServiceInterface
}

// NewIncidentHandler creates a new IncidentHandler.
func NewIncidentHandler(svc IncidentV1ServiceInterface) *IncidentHandler {
	return &IncidentHandler{service: svc}
}

// mapIncidentStatus derives "open" or "resolved" from ResolvedAt.
func mapIncidentStatus(inc *domain.Incident) string {
	if inc.ResolvedAt == nil {
		return "open"
	}
	return "resolved"
}

// mapIncidentResponse maps a domain.Incident to a v1 IncidentResponse.
func mapIncidentResponse(inc *domain.Incident) dtoV1.IncidentResponse {
	resp := dtoV1.IncidentResponse{
		ID:        inc.ID,
		MonitorID: inc.ResourceID,
		Cause:     inc.Cause,
		Status:    mapIncidentStatus(inc),
		StartedAt: inc.StartedAt.UTC().Format(time.RFC3339),
		CreatedAt: inc.CreatedAt.UTC().Format(time.RFC3339),
	}
	if inc.ResolvedAt != nil {
		s := inc.ResolvedAt.UTC().Format(time.RFC3339)
		resp.ResolvedAt = &s
	}
	return resp
}

// List handles GET /api/v1/incidents
//
// @Summary     List incidents
// @Tags        incidents
// @Security    BearerAuth
// @Produce     json
// @Param       page       query string false "Page number"
// @Param       per_page   query string false "Items per page (1-100)"
// @Param       monitor_id query string false "Filter by monitor ID"
// @Param       status     query string false "Filter by status (open|resolved)"
// @Success     200 {object} map[string]interface{}
// @Failure     401 {object} dtoV1.ErrorResponse
// @Failure     422 {object} dtoV1.ErrorResponse
// @Router      /incidents [get]
func (h *IncidentHandler) List(w http.ResponseWriter, r *http.Request) {
	params, errs := parsePagination(r)
	if len(errs) > 0 {
		respondError(w, r, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "invalid pagination parameters", errs...)
		return
	}

	// Validate status filter
	statusFilter := r.URL.Query().Get("status")
	if statusFilter != "" && statusFilter != "open" && statusFilter != "resolved" {
		respondError(w, r, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "invalid status filter",
			dtoV1.FieldError{Field: "status", Message: "must be 'open' or 'resolved'"})
		return
	}

	monitorIDFilter := r.URL.Query().Get("monitor_id")

	offset := (params.Page - 1) * params.PerPage
	items, err := h.service.ListAll(r.Context(), params.PerPage, offset)
	if err != nil {
		respondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list incidents")
		return
	}

	all, err := h.service.ListAll(r.Context(), 10000, 0)
	if err != nil {
		respondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to count incidents")
		return
	}
	total := len(all)

	// Apply in-memory filters
	filtered := make([]dtoV1.IncidentResponse, 0, len(items))
	for _, inc := range items {
		if monitorIDFilter != "" && inc.ResourceID != monitorIDFilter {
			continue
		}
		status := mapIncidentStatus(inc)
		if statusFilter != "" && status != statusFilter {
			continue
		}
		filtered = append(filtered, mapIncidentResponse(inc))
	}

	respondPaginated(w, filtered, dtoV1.MetaResponse{
		Page:    params.Page,
		PerPage: params.PerPage,
		Total:   total,
	})
}

// Get handles GET /api/v1/incidents/{id}
//
// @Summary     Get an incident by ID
// @Tags        incidents
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "Incident ID"
// @Success     200 {object} dtoV1.SingleResponse[dtoV1.IncidentResponse]
// @Failure     404 {object} dtoV1.ErrorResponse
// @Router      /incidents/{id} [get]
func (h *IncidentHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	inc, err := h.service.GetIncidentByID(r.Context(), id)
	if err != nil || inc == nil {
		respondError(w, r, http.StatusNotFound, "RESOURCE_NOT_FOUND", "incident not found")
		return
	}
	respond(w, http.StatusOK, mapIncidentResponse(inc))
}
