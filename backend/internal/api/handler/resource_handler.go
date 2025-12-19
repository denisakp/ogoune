package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/denisakp/pulseguard/internal/api/response"
	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/dto"
	"github.com/denisakp/pulseguard/internal/service"
	"github.com/go-chi/chi/v5"
)

// ResourceServiceInterface defines the methods required by ResourceHandler.
// This interface allows for better testing by enabling mock implementations.
type ResourceServiceInterface interface {
	CreateResource(ctx context.Context, payload *dto.CreateResourcePayload) (*domain.Resource, error)
	GetResourceByID(ctx context.Context, id string) (*domain.Resource, error)
	GetResourceByIDWithResponseTimes(ctx context.Context, id string, limit int) (*dto.ResourceResponse, error)
	UpdateResource(ctx context.Context, id string, payload *dto.UpdateResourcePayload) (*domain.Resource, error)
	ListAll(ctx context.Context) ([]*domain.Resource, error)
	DeleteResource(ctx context.Context, resourceID string) error
	PauseMonitoring(ctx context.Context, resourceID string) error
	ResumeMonitoring(ctx context.Context, resourceID string) error
	AddTagsToResource(ctx context.Context, resourceID string, tagIDs []string) error
	RemoveTagFromResource(ctx context.Context, resourceID string, tagID string) error
}

// ResourceHandler handles HTTP requests for monitoring resource management.
// It follows the Handler -> Service -> Repository pattern, keeping all business
// logic in the service layer while handling HTTP concerns here.
type ResourceHandler struct {
	resourceService ResourceServiceInterface
}

// NewResourceHandler creates a new ResourceHandler with injected dependencies.
func NewResourceHandler(resourceService ResourceServiceInterface) *ResourceHandler {
	return &ResourceHandler{
		resourceService: resourceService,
	}
}

// CreateResource handles POST /resources - creates a new monitoring resource.
// Request body: JSON representation of domain.Resource
// Response: 201 Created with the created resource (including generated ID)
func (h *ResourceHandler) CreateResource(w http.ResponseWriter, r *http.Request) {
	var resource dto.CreateResourcePayload

	// Decode JSON request body
	if err := json.NewDecoder(r.Body).Decode(&resource); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	// Basic validation - ensure required fields are present
	if resource.Name == "" {
		respondError(w, http.StatusBadRequest, "Resource name is required")
		return
	}
	if resource.Target == "" {
		respondError(w, http.StatusBadRequest, "Resource target is required")
		return
	}
	if resource.Type == "" {
		respondError(w, http.StatusBadRequest, "Resource type is required")
		return
	}

	// Validate resource type is one of the allowed values
	if resource.Type != domain.ResourceHTTP && resource.Type != domain.ResourceTCP {
		respondError(w, http.StatusBadRequest, "Invalid resource type. Must be 'http' or 'tcp'")
		return
	}

	// Set default values if not provided
	if resource.Interval == 0 {
		resource.Interval = 300 // Default to 300 seconds
	}
	if resource.Timeout == 0 {
		resource.Timeout = 10 // Default to 10 seconds
	}

	// Call service layer to create resource (which will also schedule monitoring)
	created, err := h.resourceService.CreateResource(r.Context(), &resource)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create resource: "+err.Error())
		return
	}

	// Respond with created resource (includes generated ID and metadata_pending)
	respondJSON(w, http.StatusCreated, created)
}

// GetResourceByID handles GET /resources/{id} - retrieves a monitoring resource by ID with response times.
// URL parameter: id (resource ID)
// Query parameter: limit (optional, default 50) - number of recent response times to include
// Response: 200 OK with dto.ResourceResponse object (includes response times)
func (h *ResourceHandler) GetResourceByID(w http.ResponseWriter, r *http.Request) {
	resourceID := chi.URLParam(r, "id")
	if resourceID == "" {
		respondError(w, http.StatusBadRequest, "Resource ID is required")
		return
	}

	// Parse optional limit query parameter
	limit := 50 // default
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Get resource with response times
	resource, err := h.resourceService.GetResourceByIDWithResponseTimes(r.Context(), resourceID, limit)
	if err != nil {
		if errors.Is(err, service.ErrResourceNotFound) {
			respondError(w, http.StatusNotFound, "Resource not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to retrieve resource: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, resource)
}

// ListResources handles GET /resources - retrieves all monitoring resources.
// Response: 200 OK with array of domain.Resource objects
func (h *ResourceHandler) ListResources(w http.ResponseWriter, r *http.Request) {
	// Call service layer to list all resources
	resources, err := h.resourceService.ListAll(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve resources: "+err.Error())
		return
	}

	// Respond with resource list (empty array if no resources exist)
	respondJSON(w, http.StatusOK, resources)
}

// UpdateResource handles PUT /resources/{id} - updates an existing monitoring resource.
// URL parameter: id (resource ID)
// Request body: JSON representation of domain.Resource with updated fields
// Response: 200 OK with the updated resource
// UpdateResource handles PATCH /resources/{id} - updates an existing monitoring resource.
// Request body: JSON representation of service.UpdateResourcePayload (partial update)
// Response: 200 OK with the updated resource
func (h *ResourceHandler) UpdateResource(w http.ResponseWriter, r *http.Request) {
	// Extract resource ID from URL path parameter (set by Chi router)
	resourceID := chi.URLParam(r, "id")
	if resourceID == "" {
		respondError(w, http.StatusBadRequest, "Resource ID is required")
		return
	}

	var payload dto.UpdateResourcePayload

	// Decode JSON request body
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	// Call service layer to update resource (which will also reschedule monitoring)
	updatedResource, err := h.resourceService.UpdateResource(r.Context(), resourceID, &payload)
	if err != nil {
		// Check for validation errors
		if errors.Is(err, service.ErrValidationFailed) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		// Check for not found errors
		if errors.Is(err, service.ErrResourceNotFound) {
			respondError(w, http.StatusNotFound, "Resource not found")
			return
		}
		// Generic error
		respondError(w, http.StatusInternalServerError, "Failed to update resource: "+err.Error())
		return
	}

	// Respond with updated resource
	respondJSON(w, http.StatusOK, updatedResource)
}

// DeleteResource handles DELETE /resources/{id} - soft deletes a monitoring resource.
// URL parameter: id (resource ID)
// Response: 204 No Content on success
func (h *ResourceHandler) DeleteResource(w http.ResponseWriter, r *http.Request) {
	// Extract resource ID from URL path parameter (set by Chi router)
	resourceID := chi.URLParam(r, "id")
	if resourceID == "" {
		respondError(w, http.StatusBadRequest, "Resource ID is required")
		return
	}

	// Call service layer to delete resource (which will also unschedule monitoring)
	if err := h.resourceService.DeleteResource(r.Context(), resourceID); err != nil {
		// Check for not found errors
		if errors.Is(err, service.ErrResourceNotFound) {
			respondError(w, http.StatusNotFound, "Resource not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to delete resource: "+err.Error())
		return
	}

	// Respond with 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

// PauseResourceMonitoring handles POST /resources/{id}/pause - pauses monitoring for a resource.
// URL parameter: id (resource ID)
// Response: 200 OK with success message
func (h *ResourceHandler) PauseResourceMonitoring(w http.ResponseWriter, r *http.Request) {
	// Extract resource ID from URL path parameter (set by Chi router)
	resourceID := chi.URLParam(r, "id")
	if resourceID == "" {
		respondError(w, http.StatusBadRequest, "Resource ID is required")
		return
	}

	// Call service layer to pause monitoring
	if err := h.resourceService.PauseMonitoring(r.Context(), resourceID); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to pause monitoring: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Monitoring paused successfully",
	})
}

// ResumeResourceMonitoring handles POST /resources/{id}/resume - resumes monitoring for a resource.
// URL parameter: id (resource ID)
// Response: 200 OK with success message
func (h *ResourceHandler) ResumeResourceMonitoring(w http.ResponseWriter, r *http.Request) {
	// Extract resource ID from URL path parameter (set by Chi router)
	resourceID := chi.URLParam(r, "id")
	if resourceID == "" {
		respondError(w, http.StatusBadRequest, "Resource ID is required")
		return
	}

	// Call service layer to resume monitoring
	if err := h.resourceService.ResumeMonitoring(r.Context(), resourceID); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to resume monitoring: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Monitoring resumed successfully",
	})
}

// AddTagsToResource handles POST /resources/{resourceID}/tags - adds tags to a resource.
// Request body: JSON array of tag IDs
// Response: 200 OK with success message
func (h *ResourceHandler) AddTagsToResource(w http.ResponseWriter, r *http.Request) {
	resourceID := chi.URLParam(r, "resourceID")
	if resourceID == "" {
		respondError(w, http.StatusBadRequest, "Resource ID is required")
		return
	}

	var payload struct {
		TagIDs []string `json:"tag_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	if len(payload.TagIDs) == 0 {
		respondError(w, http.StatusBadRequest, "At least one tag ID is required")
		return
	}

	if err := h.resourceService.AddTagsToResource(r.Context(), resourceID, payload.TagIDs); err != nil {
		if errors.Is(err, service.ErrResourceNotFound) {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, service.ErrValidationFailed) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to add tags to resource: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Tags added successfully",
	})
}

// RemoveTagFromResource handles DELETE /resources/{resourceID}/tags/{tagID} - removes a tag from a resource.
// URL parameters: resourceID, tagID
// Response: 204 No Content on success
func (h *ResourceHandler) RemoveTagFromResource(w http.ResponseWriter, r *http.Request) {
	resourceID := chi.URLParam(r, "resourceID")
	tagID := chi.URLParam(r, "tagID")

	if resourceID == "" {
		respondError(w, http.StatusBadRequest, "Resource ID is required")
		return
	}

	if tagID == "" {
		respondError(w, http.StatusBadRequest, "Tag ID is required")
		return
	}

	if err := h.resourceService.RemoveTagFromResource(r.Context(), resourceID, tagID); err != nil {
		if errors.Is(err, service.ErrResourceNotFound) {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, service.ErrValidationFailed) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to remove tag from resource: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// respondJSON writes a JSON response with the given status code and payload.
// This is a helper function to ensure consistent JSON response formatting.
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	response.JSON(w, status, payload)
}

// respondError writes a JSON error response with the given status code and message.
// Error responses follow a consistent format: {"error": "message"}
func respondError(w http.ResponseWriter, status int, message string) {
	response.Error(w, status, message)
}
