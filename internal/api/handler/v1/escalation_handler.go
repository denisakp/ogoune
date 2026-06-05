package v1

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository"
	"github.com/denisakp/ogoune/internal/service"
	"github.com/go-chi/chi/v5"
)

// EscalationHandler — spec 059 FR-023..FR-026a.
type EscalationHandler struct {
	svc *service.EscalationService
}

func NewEscalationHandler(svc *service.EscalationService) *EscalationHandler {
	return &EscalationHandler{svc: svc}
}

type escalationStepDTO struct {
	ID           string   `json:"id,omitempty"`
	DelayMinutes int      `json:"delay_minutes"`
	ChannelIDs   []string `json:"channel_ids"`
}

type escalationPolicyDTO struct {
	ID       string                 `json:"id,omitempty"`
	Name     string                 `json:"name"`
	Scope    domain.EscalationScope `json:"scope"`
	IsActive bool                   `json:"is_active"`
	Priority int                    `json:"priority,omitempty"`
	Steps    []escalationStepDTO    `json:"steps"`
}

func toDTO(p *domain.EscalationPolicy) escalationPolicyDTO {
	steps := make([]escalationStepDTO, 0, len(p.Steps))
	for _, s := range p.Steps {
		steps = append(steps, escalationStepDTO{
			ID:           s.ID,
			DelayMinutes: s.DelayMinutes,
			ChannelIDs:   s.ChannelIDs,
		})
	}
	return escalationPolicyDTO{
		ID:       p.ID,
		Name:     p.Name,
		Scope:    p.Scope,
		IsActive: p.IsActive,
		Priority: p.Priority,
		Steps:    steps,
	}
}

func fromDTO(d escalationPolicyDTO) *domain.EscalationPolicy {
	steps := make([]domain.EscalationStep, 0, len(d.Steps))
	for i, s := range d.Steps {
		steps = append(steps, domain.EscalationStep{
			ID:           s.ID,
			StepOrder:    i + 1,
			DelayMinutes: s.DelayMinutes,
			ChannelIDs:   s.ChannelIDs,
		})
	}
	return &domain.EscalationPolicy{
		Name:     d.Name,
		Scope:    d.Scope,
		IsActive: d.IsActive,
		Priority: d.Priority,
		Steps:    steps,
	}
}

// List — GET /escalation-policies.
func (h *EscalationHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.svc.List(r.Context())
	if err != nil {
		respondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list policies")
		return
	}
	out := make([]escalationPolicyDTO, 0, len(rows))
	for _, p := range rows {
		out = append(out, toDTO(p))
	}
	respond(w, http.StatusOK, out)
}

// Create — POST /escalation-policies.
func (h *EscalationHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req escalationPolicyDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, http.StatusBadRequest, "BAD_REQUEST", "invalid body")
		return
	}
	p, err := h.svc.Create(r.Context(), fromDTO(req))
	if err != nil {
		respondError(w, r, mapValidationStatus(err), mapValidationCode(err), err.Error())
		return
	}
	respond(w, http.StatusCreated, toDTO(p))
}

// Update — PATCH /escalation-policies/{id}.
func (h *EscalationHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req escalationPolicyDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, http.StatusBadRequest, "BAD_REQUEST", "invalid body")
		return
	}
	p := fromDTO(req)
	p.ID = id
	updated, err := h.svc.Update(r.Context(), p)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondError(w, r, http.StatusNotFound, "NOT_FOUND", "policy not found")
			return
		}
		respondError(w, r, mapValidationStatus(err), mapValidationCode(err), err.Error())
		return
	}
	respond(w, http.StatusOK, toDTO(updated))
}

// Delete — DELETE /escalation-policies/{id}.
func (h *EscalationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondError(w, r, http.StatusNotFound, "NOT_FOUND", "policy not found")
			return
		}
		respondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to delete policy")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Reorder — PATCH /escalation-policies/reorder.
func (h *EscalationHandler) Reorder(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Order []string `json:"order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, http.StatusBadRequest, "BAD_REQUEST", "invalid body")
		return
	}
	if err := h.svc.Reorder(r.Context(), req.Order); err != nil {
		respondError(w, r, http.StatusUnprocessableEntity, "REORDER_INVALID", err.Error())
		return
	}
	rows, err := h.svc.List(r.Context())
	if err != nil {
		respondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list policies")
		return
	}
	out := make([]escalationPolicyDTO, 0, len(rows))
	for _, p := range rows {
		out = append(out, toDTO(p))
	}
	respond(w, http.StatusOK, out)
}

func mapValidationStatus(err error) int {
	switch {
	case errors.Is(err, service.ErrEscalationStepsRange),
		errors.Is(err, service.ErrEscalationChannelsEmpty),
		errors.Is(err, service.ErrEscalationDelayRange),
		errors.Is(err, service.ErrEscalationScopeInvalid),
		errors.Is(err, service.ErrEscalationReorderMissing),
		errors.Is(err, service.ErrEscalationReorderUnknown):
		return http.StatusUnprocessableEntity
	}
	return http.StatusInternalServerError
}

func mapValidationCode(err error) string {
	switch {
	case errors.Is(err, service.ErrEscalationStepsRange):
		return "STEPS_RANGE"
	case errors.Is(err, service.ErrEscalationChannelsEmpty):
		return "CHANNELS_EMPTY"
	case errors.Is(err, service.ErrEscalationDelayRange):
		return "DELAY_RANGE"
	case errors.Is(err, service.ErrEscalationScopeInvalid):
		return "SCOPE_INVALID"
	case errors.Is(err, service.ErrEscalationReorderMissing),
		errors.Is(err, service.ErrEscalationReorderUnknown):
		return "REORDER_INVALID"
	}
	return "INTERNAL_ERROR"
}
