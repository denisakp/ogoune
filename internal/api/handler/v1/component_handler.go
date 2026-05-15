package v1

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	dtoV1 "github.com/denisakp/ogoune/internal/dto/v1"
	"github.com/go-chi/chi/v5"
)

// ComponentV1RepositoryInterface defines the component repository methods used by the v1 component handler.
type ComponentV1RepositoryInterface interface {
	Create(ctx context.Context, c *domain.Component) (*domain.Component, error)
	List(ctx context.Context, limit, offset int) ([]*domain.Component, error)
	FindByID(ctx context.Context, id string) (*domain.Component, error)
	Update(ctx context.Context, c *domain.Component) error
	Delete(ctx context.Context, id string) error
}

// ComponentHandler handles v1 CRUD endpoints for components.
type ComponentHandler struct {
	repo ComponentV1RepositoryInterface
}

// NewComponentHandler creates a new ComponentHandler.
func NewComponentHandler(repo ComponentV1RepositoryInterface) *ComponentHandler {
	return &ComponentHandler{repo: repo}
}

// mapComponentResponse maps a domain.Component to a v1 ComponentResponse.
func mapComponentResponse(c *domain.Component) dtoV1.ComponentResponse {
	return dtoV1.ComponentResponse{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		CreatedAt:   c.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   c.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

// List handles GET /api/v1/components
//
// @Summary     List components
// @Tags        components
// @Security    BearerAuth
// @Produce     json
// @Param       page     query int false "Page number (default 1)"
// @Param       per_page query int false "Items per page (1-100, default 20)"
// @Success     200 {object} map[string]interface{}
// @Failure     401 {object} dtoV1.ErrorResponse
// @Router      /components [get]
func (h *ComponentHandler) List(w http.ResponseWriter, r *http.Request) {
	params, errs := parsePagination(r)
	if len(errs) > 0 {
		respondError(w, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "invalid pagination parameters", errs...)
		return
	}

	offset := (params.Page - 1) * params.PerPage
	items, err := h.repo.List(r.Context(), params.PerPage, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list components")
		return
	}

	// total: large limit fetch
	all, err := h.repo.List(r.Context(), 10000, 0)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to count components")
		return
	}

	data := make([]dtoV1.ComponentResponse, 0, len(items))
	for _, c := range items {
		data = append(data, mapComponentResponse(c))
	}

	respondPaginated(w, data, dtoV1.MetaResponse{
		Page:    params.Page,
		PerPage: params.PerPage,
		Total:   len(all),
	})
}

// Get handles GET /api/v1/components/{id}
//
// @Summary     Get a component by ID
// @Tags        components
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "Component ID"
// @Success     200 {object} dtoV1.SingleResponse[dtoV1.ComponentResponse]
// @Failure     404 {object} dtoV1.ErrorResponse
// @Router      /components/{id} [get]
func (h *ComponentHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	c, err := h.repo.FindByID(r.Context(), id)
	if err != nil || c == nil {
		respondError(w, http.StatusNotFound, "RESOURCE_NOT_FOUND", "component not found")
		return
	}
	respond(w, http.StatusOK, mapComponentResponse(c))
}

// Create handles POST /api/v1/components
//
// @Summary     Create a component
// @Tags        components
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       body body dtoV1.CreateComponentRequest true "Component payload"
// @Success     201 {object} dtoV1.SingleResponse[dtoV1.ComponentResponse]
// @Failure     422 {object} dtoV1.ErrorResponse
// @Failure     403 {object} dtoV1.ErrorResponse
// @Router      /components [post]
func (h *ComponentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dtoV1.CreateComponentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "invalid request body")
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		respondError(w, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "validation failed",
			dtoV1.FieldError{Field: "name", Message: "required"})
		return
	}

	c := &domain.Component{
		Name:        req.Name,
		Description: req.Description,
	}
	created, err := h.repo.Create(r.Context(), c)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create component")
		return
	}
	respond(w, http.StatusCreated, mapComponentResponse(created))
}

// Update handles PUT /api/v1/components/{id}
//
// @Summary     Update a component
// @Tags        components
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id   path string true "Component ID"
// @Param       body body dtoV1.UpdateComponentRequest true "Update payload"
// @Success     200 {object} dtoV1.SingleResponse[dtoV1.ComponentResponse]
// @Failure     404 {object} dtoV1.ErrorResponse
// @Failure     403 {object} dtoV1.ErrorResponse
// @Router      /components/{id} [put]
func (h *ComponentHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	c, err := h.repo.FindByID(r.Context(), id)
	if err != nil || c == nil {
		respondError(w, http.StatusNotFound, "RESOURCE_NOT_FOUND", "component not found")
		return
	}

	var req dtoV1.UpdateComponentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "invalid request body")
		return
	}

	if req.Name != nil {
		c.Name = *req.Name
	}
	if req.Description != nil {
		c.Description = req.Description
	}

	if err := h.repo.Update(r.Context(), c); err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update component")
		return
	}
	respond(w, http.StatusOK, mapComponentResponse(c))
}

// Delete handles DELETE /api/v1/components/{id}
//
// @Summary     Delete a component
// @Tags        components
// @Security    BearerAuth
// @Param       id path string true "Component ID"
// @Success     204
// @Failure     404 {object} dtoV1.ErrorResponse
// @Failure     403 {object} dtoV1.ErrorResponse
// @Router      /components/{id} [delete]
func (h *ComponentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.repo.Delete(r.Context(), id); err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to delete component")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
