// Package handler — admin endpoints for incident lifecycle updates (US7).
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository"
)

type IncidentUpdateProvider interface {
	ListByIncident(ctx context.Context, incidentID string) ([]*domain.IncidentUpdate, error)
	Create(ctx context.Context, incidentID string, status domain.IncidentUpdateStatus, message, postedBy string) (*domain.IncidentUpdate, error)
	Update(ctx context.Context, id string, status domain.IncidentUpdateStatus, message string) (*domain.IncidentUpdate, error)
	Delete(ctx context.Context, id string) error
}

type IncidentUpdateHandler struct {
	svc IncidentUpdateProvider
}

func NewIncidentUpdateHandler(svc IncidentUpdateProvider) *IncidentUpdateHandler {
	return &IncidentUpdateHandler{svc: svc}
}

type incidentUpdatePayload struct {
	Status  domain.IncidentUpdateStatus `json:"status"`
	Message string                      `json:"message"`
}

// List — GET /api/v1/incidents/{id}/updates.
func (h *IncidentUpdateHandler) List(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	rows, err := h.svc.ListByIncident(r.Context(), id)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

// Create — POST /api/v1/incidents/{id}/updates.
func (h *IncidentUpdateHandler) Create(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var p incidentUpdatePayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeProblem(w, http.StatusBadRequest, "invalid_payload", err.Error())
		return
	}
	postedBy := postedByFromContext(r)
	out, err := h.svc.Create(r.Context(), id, p.Status, p.Message, postedBy)
	if err != nil {
		writeProblem(w, http.StatusUnprocessableEntity, "invalid_update", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

// Update — PATCH /api/v1/incidents/{id}/updates/{updateID}.
func (h *IncidentUpdateHandler) Update(w http.ResponseWriter, r *http.Request) {
	updateID := chi.URLParam(r, "updateID")
	var p incidentUpdatePayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeProblem(w, http.StatusBadRequest, "invalid_payload", err.Error())
		return
	}
	out, err := h.svc.Update(r.Context(), updateID, p.Status, p.Message)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeProblem(w, http.StatusNotFound, "not_found", "update not found")
			return
		}
		writeProblem(w, http.StatusUnprocessableEntity, "invalid_update", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// Delete — DELETE /api/v1/incidents/{id}/updates/{updateID}.
func (h *IncidentUpdateHandler) Delete(w http.ResponseWriter, r *http.Request) {
	updateID := chi.URLParam(r, "updateID")
	if err := h.svc.Delete(r.Context(), updateID); err != nil {
		writeProblem(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// postedByFromContext extracts the actor id from the auth context populated
// by the project's auth middleware. Returns empty when unauthenticated.
func postedByFromContext(r *http.Request) string {
	if v := r.Context().Value("user_id"); v != nil { //nolint:staticcheck // existing pattern
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
