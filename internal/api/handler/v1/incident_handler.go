package v1

import (
	"context"
	"net/http"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	dtoV1 "github.com/denisakp/ogoune/internal/dto/v1"
	"github.com/denisakp/ogoune/internal/narrative"
	"github.com/denisakp/ogoune/internal/repository/sqlc/dynquery"
	"github.com/go-chi/chi/v5"
)

// IncidentV1ServiceInterface defines the incident service methods used by the v1 incident handler.
type IncidentV1ServiceInterface interface {
	ListAll(ctx context.Context, limit, offset int) ([]*domain.Incident, error)
	ListByFilter(ctx context.Context, f dynquery.IncidentFilter, page, perPage int) ([]*domain.Incident, int, error)
	GetIncidentByID(ctx context.Context, id string) (*domain.Incident, error)
	GetEventStepsForIncident(ctx context.Context, incidentID string) ([]domain.IncidentEventStep, error)
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
// The rich fields (Resource, EventSteps, Diagnostics) are populated only when
// the service hydrated them (detail path); on the list path Resource is
// hydrated and the rest stay zero-valued — matching the legacy root behavior.
func mapIncidentResponse(inc *domain.Incident) dtoV1.IncidentResponse {
	resp := dtoV1.IncidentResponse{
		ID:          inc.ID,
		MonitorID:   inc.ResourceID,
		ResourceID:  inc.ResourceID,
		Resource:    inc.Resource,
		Cause:       inc.Cause,
		Status:      mapIncidentStatus(inc),
		Details:     string(inc.Details),
		EventSteps:  inc.EventStep,
		Diagnostics: inc.IncidentDiagnostics,
		StartedAt:   inc.StartedAt.UTC().Format(time.RFC3339),
		CreatedAt:   inc.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   inc.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if inc.ResolvedAt != nil {
		s := inc.ResolvedAt.UTC().Format(time.RFC3339)
		resp.ResolvedAt = &s
	}
	resp.HostContext = mapHostContext(inc.HostContext)
	resp.Explanation = mapIncidentExplanation(inc.Explanation)
	for _, e := range inc.HostEvents {
		resp.HostEvents = append(resp.HostEvents, mapHostEvent(e))
	}
	resp.HostLink = mapHostLink(inc.HostLink)
	return resp
}

// mapHostLink converts the read-side host link into its v1 shape (spec 092).
// Mapping only: the service resolved the machine and looked it up; nothing is
// decided or queried here. Nil maps to nil, which omitempty keeps out of the
// body -- absent, not null.
func mapHostLink(l *domain.HostLink) *dtoV1.HostLinkResponse {
	if l == nil {
		return nil
	}
	return &dtoV1.HostLinkResponse{
		HostID: l.HostID,
		Source: string(l.Source),
		Exists: l.Exists,
	}
}

// mapIncidentExplanation converts the read-side explanation into its v1 shape
// and renders the sentence (spec 091).
//
// A nil explanation maps to nil, which `omitempty` keeps out of the body
// entirely -- absent rather than null, so the list endpoint and a silent detail
// response are byte-identical to their pre-feature form.
//
// The sentence is rendered here rather than carried on the domain struct: prose
// belongs to a renderer, and the API is one. Serving it means the SPA does not
// re-implement the wording in TypeScript, where the two copies would drift.
func mapIncidentExplanation(e *domain.IncidentExplanation) *dtoV1.IncidentExplanationResponse {
	if e == nil {
		return nil
	}
	text := narrative.Sentence(e)
	if text == "" {
		// The correlator only names kinds it can phrase, so this is unreachable
		// today. Guarding anyway: an explanation whose sentence came out empty is
		// worse than no explanation, because the reader gets a block with a hole
		// in it (FR-007).
		return nil
	}
	return &dtoV1.IncidentExplanationResponse{
		Text:        text,
		HostID:      e.HostID,
		HostName:    e.HostName,
		IncidentAt:  e.IncidentAt.UTC().Format(time.RFC3339),
		Cause:       e.Cause,
		Event:       mapHostEvent(&e.Event),
		Precedes:    e.Precedes,
		OtherEvents: e.OtherEvents,
		WindowFrom:  e.WindowFrom.UTC().Format(time.RFC3339),
		WindowTo:    e.WindowTo.UTC().Format(time.RFC3339),
	}
}

// mapHostContext converts the read-side host context into its v1 shape, field by
// field. A nil context maps to a nil response, which serialises as
// "host_context": null -- the only absent representation the contract allows
// (spec 089, FR-010).
func mapHostContext(hc *domain.HostContext) *dtoV1.HostContextResponse {
	if hc == nil {
		return nil
	}
	out := &dtoV1.HostContextResponse{
		HostID:      hc.HostID,
		HostName:    hc.HostName,
		PeakCPUPct:  hc.PeakCPUPct,
		PeakMemPct:  hc.PeakMemPct,
		SampleCount: hc.SampleCount,
		Resolution:  string(hc.Resolution),
		WindowFrom:  hc.WindowFrom.UTC().Format(time.RFC3339),
		WindowTo:    hc.WindowTo.UTC().Format(time.RFC3339),
	}
	if hc.WorstDisk != nil {
		out.WorstDisk = &dtoV1.WorstDiskResponse{
			Mount:   hc.WorstDisk.Mount,
			UsedPct: hc.WorstDisk.UsedPct,
		}
	}
	return out
}

// List handles GET /api/v1/incidents
//
// @Summary     List incidents
// @Tags        incidents
// @Security    BearerAuth
// @Produce     json
// @Param       page       query int    false "Page number"
// @Param       per_page   query int    false "Items per page (1-100)"
// @Param       monitor_id query string false "Filter by monitor ID"
// @Param       status     query string false "Filter by status (open|resolved)"
// @Param       from       query string false "Filter incidents started_at >= (RFC 3339)"
// @Param       to         query string false "Filter incidents started_at <= (RFC 3339)"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} dtoV1.ErrorResponse
// @Failure     401 {object} dtoV1.ErrorResponse
// @Failure     422 {object} dtoV1.ErrorResponse
// @Router      /incidents [get]
func (h *IncidentHandler) List(w http.ResponseWriter, r *http.Request) {
	params, errs := parsePagination(r)
	if len(errs) > 0 {
		respondError(w, r, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "invalid pagination parameters", errs...)
		return
	}

	filter, ferrs := parseIncidentFilter(r)
	if len(ferrs) > 0 {
		respondError(w, r, http.StatusBadRequest, "VALIDATION_FAILED", "invalid filter parameters", ferrs...)
		return
	}

	items, total, err := h.service.ListByFilter(r.Context(), filter, params.Page, params.PerPage)
	if err != nil {
		respondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list incidents")
		return
	}

	out := make([]dtoV1.IncidentResponse, 0, len(items))
	for _, inc := range items {
		out = append(out, mapIncidentResponse(inc))
	}

	respondPaginated(w, out, dtoV1.MetaResponse{
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

// GetEventSteps handles GET /api/v1/incidents/{id}/event-steps
//
// @Summary     List an incident's event steps
// @Tags        incidents
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "Incident ID"
// @Success     200 {object} map[string]interface{}
// @Failure     500 {object} dtoV1.ErrorResponse
// @Router      /incidents/{id}/event-steps [get]
func (h *IncidentHandler) GetEventSteps(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	steps, err := h.service.GetEventStepsForIncident(r.Context(), id)
	if err != nil {
		respondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to retrieve event steps")
		return
	}
	if steps == nil {
		steps = []domain.IncidentEventStep{}
	}
	respond(w, http.StatusOK, steps)
}
