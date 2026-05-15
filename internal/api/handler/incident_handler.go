package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/service"
	"github.com/go-chi/chi/v5"
)

// IncidentServiceInterface defines the methods required by IncidentHandler.
// This interface allows for better testing by enabling mock implementations.
type IncidentServiceInterface interface {
	ListAll(ctx context.Context, limit, offset int) ([]*domain.Incident, error)
	ListUnresolved(ctx context.Context, limit, offset int) ([]*domain.Incident, error)
	GetIncidentByID(ctx context.Context, id string) (*domain.Incident, error)
	GetIncidentsByResource(ctx context.Context, resourceID string, limit, offset int) ([]*domain.Incident, error)
	GetEventStepsForIncident(ctx context.Context, incidentID string) ([]domain.IncidentEventStep, error)
}

// IncidentHandler handles HTTP requests for incident management.
// It follows the Handler -> Service -> Repository pattern, keeping all business
// logic in the service layer while handling HTTP concerns here.
type IncidentHandler struct {
	incidentService IncidentServiceInterface
}

type incidentResponse struct {
	ID                  string                      `json:"id"`
	CreatedAt           time.Time                   `json:"created_at"`
	UpdatedAt           time.Time                   `json:"updated_at"`
	ResourceID          string                      `json:"resource_id"`
	Resource            domain.Resource             `json:"resource"`
	Cause               string                      `json:"cause"`
	ResolvedAt          *time.Time                  `json:"resolved_at"`
	StartedAt           time.Time                   `json:"started_at"`
	Details             string                      `json:"details"`
	EventStep           []domain.IncidentEventStep  `json:"event_steps"`
	IncidentDiagnostics *domain.IncidentDiagnostics `json:"diagnostics"`
}

// NewIncidentHandler creates a new IncidentHandler with injected dependencies.
func NewIncidentHandler(incidentService IncidentServiceInterface) *IncidentHandler {
	return &IncidentHandler{
		incidentService: incidentService,
	}
}

// ListIncidents handles GET /incidents - retrieves all incidents with pagination.
// Query parameters:
//   - limit: Number of incidents to return (default: 25, max: 100)
//   - offset: Number of incidents to skip (default: 0)
//   - unresolved: If "true", only return unresolved incidents (optional)
//
// Response: 200 OK with array of incident objects
func (h *IncidentHandler) ListIncidents(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limit := 25
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	var incidents []*domain.Incident
	var err error

	// Check if requesting only unresolved incidents
	if r.URL.Query().Get("unresolved") == "true" {
		incidents, err = h.incidentService.ListUnresolved(r.Context(), limit, offset)
	} else {
		incidents, err = h.incidentService.ListAll(r.Context(), limit, offset)
	}

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve incidents: "+err.Error())
		return
	}

	// Return empty array if no incidents found
	if incidents == nil {
		incidents = []*domain.Incident{}
	}

	respondJSON(w, http.StatusOK, toIncidentResponses(incidents))
}

// GetIncidentDetail handles GET /incidents/{id} - retrieves a single incident with its event steps.
// Path parameters:
//   - id: The incident ID
//
// Response: 200 OK with incident object including event steps
// Response: 404 Not Found if incident doesn't exist
func (h *IncidentHandler) GetIncidentDetail(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "Incident ID is required")
		return
	}

	// Fetch incident with event steps
	incident, err := h.incidentService.GetIncidentByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrResourceNotFound) {
			respondError(w, http.StatusNotFound, "Incident not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to retrieve incident: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, toIncidentResponse(incident))
}

// GetIncidentEventSteps handles GET /incidents/{id}/event-steps - retrieves event steps for an incident.
// Path parameters:
//   - id: The incident ID
//
// Response: 200 OK with array of event step objects
func (h *IncidentHandler) GetIncidentEventSteps(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "Incident ID is required")
		return
	}

	// Fetch event steps for incident
	steps, err := h.incidentService.GetEventStepsForIncident(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve event steps: "+err.Error())
		return
	}

	// Return empty array if no steps found
	if steps == nil {
		steps = []domain.IncidentEventStep{}
	}

	respondJSON(w, http.StatusOK, steps)
}

// GetResourceIncidents handles GET /incidents?resource_id={resourceID} - retrieves incidents for a resource.
// Query parameters:
//   - resource_id: The resource ID (required)
//   - limit: Number of incidents to return (default: 25, max: 100)
//   - offset: Number of incidents to skip (default: 0)
//
// Response: 200 OK with array of incident objects for the resource
func (h *IncidentHandler) GetResourceIncidents(w http.ResponseWriter, r *http.Request) {
	resourceID := r.URL.Query().Get("resource_id")
	if resourceID == "" {
		respondError(w, http.StatusBadRequest, "resource_id query parameter is required")
		return
	}

	// Parse pagination parameters
	limit := 25
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	incidents, err := h.incidentService.GetIncidentsByResource(r.Context(), resourceID, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve incidents: "+err.Error())
		return
	}

	// Return empty array if no incidents found
	if incidents == nil {
		incidents = []*domain.Incident{}
	}

	respondJSON(w, http.StatusOK, toIncidentResponses(incidents))
}

func toIncidentResponses(incidents []*domain.Incident) []incidentResponse {
	out := make([]incidentResponse, 0, len(incidents))
	for _, incident := range incidents {
		if incident == nil {
			continue
		}
		out = append(out, toIncidentResponse(incident))
	}
	return out
}

func toIncidentResponse(incident *domain.Incident) incidentResponse {
	return incidentResponse{
		ID:                  incident.ID,
		CreatedAt:           incident.CreatedAt,
		UpdatedAt:           incident.UpdatedAt,
		ResourceID:          incident.ResourceID,
		Resource:            incident.Resource,
		Cause:               incident.Cause,
		ResolvedAt:          incident.ResolvedAt,
		StartedAt:           incident.StartedAt,
		Details:             string(incident.Details),
		EventStep:           incident.EventStep,
		IncidentDiagnostics: incident.IncidentDiagnostics,
	}
}
