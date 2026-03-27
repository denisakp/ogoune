package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/denisakp/pulseguard/internal/api/response"
	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/service"
	"github.com/go-chi/chi/v5"
)

// TagServiceInterface defines the methods required by TagHandler.
type TagServiceInterface interface {
	CreateTag(ctx context.Context, tag *domain.Tags) error
	ListTags(ctx context.Context, limit, offset int) ([]*domain.Tags, error)
	GetTagByID(ctx context.Context, id string) (*domain.Tags, error)
	UpdateTag(ctx context.Context, id string, name string, color *string, description *string) (*domain.Tags, error)
	DeleteTag(ctx context.Context, id string) error
}

// TagHandler handles HTTP requests for tag management.
type TagHandler struct {
	tagService TagServiceInterface
}

// NewTagHandler creates a new TagHandler with injected dependencies.
func NewTagHandler(tagService TagServiceInterface) *TagHandler {
	return &TagHandler{
		tagService: tagService,
	}
}

// CreateTag handles POST /tags - creates a new tag.
func (h *TagHandler) CreateTag(w http.ResponseWriter, r *http.Request) {
	var tag domain.Tags

	if err := json.NewDecoder(r.Body).Decode(&tag); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	if tag.Name == "" {
		respondError(w, http.StatusBadRequest, "Tag name is required")
		return
	}

	if err := h.tagService.CreateTag(r.Context(), &tag); err != nil {
		if errors.Is(err, service.ErrValidationFailed) {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to create tag: "+err.Error())
		return
	}

	response.Created(w, tag)
}

// ListTags handles GET /tags - retrieves all tags.
func (h *TagHandler) ListTags(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50 // default
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	tags, err := h.tagService.ListTags(r.Context(), limit, offset)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to retrieve tags: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, tags)
}

// UpdateTag handles PATCH /tags/{id} - updates an existing tag.
func (h *TagHandler) UpdateTag(w http.ResponseWriter, r *http.Request) {
	tagID := chi.URLParam(r, "id")
	if tagID == "" {
		respondError(w, http.StatusBadRequest, "Tag ID is required")
		return
	}

	var payload struct {
		Name        *string `json:"name,omitempty"`
		Color       *string `json:"color,omitempty"`
		Description *string `json:"description,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	// Get existing tag to merge with updates
	existingTag, err := h.tagService.GetTagByID(r.Context(), tagID)
	if err != nil {
		if errors.Is(err, service.ErrResourceNotFound) {
			response.Error(w, http.StatusNotFound, "Tag not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to fetch tag: "+err.Error())
		return
	}

	// Merge updates with existing values
	name := existingTag.Name
	if payload.Name != nil && *payload.Name != "" {
		name = *payload.Name
	}

	color := existingTag.Color
	if payload.Color != nil {
		color = payload.Color
	}

	description := existingTag.Description
	if payload.Description != nil {
		description = payload.Description
	}

	updatedTag, err := h.tagService.UpdateTag(r.Context(), tagID, name, color, description)
	if err != nil {
		if errors.Is(err, service.ErrValidationFailed) {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, service.ErrResourceNotFound) {
			response.Error(w, http.StatusNotFound, "Tag not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to update tag: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, updatedTag)
}

// DeleteTag handles DELETE /tags/{id} - deletes a tag.
func (h *TagHandler) DeleteTag(w http.ResponseWriter, r *http.Request) {
	tagID := chi.URLParam(r, "id")
	if tagID == "" {
		response.Error(w, http.StatusBadRequest, "Tag ID is required")
		return
	}

	if err := h.tagService.DeleteTag(r.Context(), tagID); err != nil {
		if errors.Is(err, service.ErrResourceNotFound) {
			response.Error(w, http.StatusNotFound, "Tag not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to delete tag: "+err.Error())
		return
	}

	response.NoContent(w)
}
