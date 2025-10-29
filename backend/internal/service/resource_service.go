package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/dto"
	"github.com/denisakp/pulseguard/internal/repository"
)

// ResourceService orchestrates resource-related operations using repository interfaces.
// This service demonstrates the dependency injection pattern and serves as an example
// of how to compose repository operations while maintaining clean boundaries.
type ResourceService struct {
	resources repository.ResourceRepository
	incidents repository.IncidentRepository
	tags      repository.TagsRepository
	scheduler repository.Scheduler
}

// NewResourceService creates a new ResourceService with the given repository dependencies.
func NewResourceService(
	resources repository.ResourceRepository,
	incidents repository.IncidentRepository,
	tags repository.TagsRepository,
	scheduler repository.Scheduler,
) *ResourceService {
	return &ResourceService{
		resources: resources,
		incidents: incidents,
		tags:      tags,
		scheduler: scheduler,
	}
}

// findOrCreateTags finds existing tags by name or creates new ones if they don't exist.
// It accepts tag names as strings and returns tag entities.
func (s *ResourceService) findOrCreateTags(ctx context.Context, tagNames []string) ([]*domain.Tags, error) {
	var tags []*domain.Tags

	for _, tagName := range tagNames {
		// Try to find the tag by name
		tag, err := s.tags.FindByName(ctx, tagName)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				// Tag doesn't exist, create it
				newTag := &domain.Tags{
					Name: tagName,
				}
				if err := s.tags.Create(ctx, newTag); err != nil {
					return nil, fmt.Errorf("failed to create tag '%s': %w", tagName, err)
				}
				tags = append(tags, newTag)
			} else {
				return nil, fmt.Errorf("failed to find tag '%s': %w", tagName, err)
			}
		} else {
			// Tag exists, use it
			tags = append(tags, tag)
		}
	}

	return tags, nil
}

// CreateResource creates a new resource using domain validation and persistence.
// After successful creation, it schedules monitoring for the resource.
func (s *ResourceService) CreateResource(ctx context.Context, payload *dto.CreateResourcePayload) error {
	// Validate target format
	if err := domain.ValidateResourceTarget(payload.Target, payload.Type); err != nil {
		return fmt.Errorf("%w: %v", ErrValidationFailed, err)
	}

	resource := &domain.Resource{
		Name:     payload.Name,
		Type:     payload.Type,
		Interval: payload.Interval,
		Timeout:  payload.Timeout,
		Target:   payload.Target,
		IsActive: true,
		Status:   domain.StatusPending,
	}

	// Find or create tags by name if provided
	if len(payload.Tags) > 0 {
		tags, err := s.findOrCreateTags(ctx, payload.Tags)
		if err != nil {
			return fmt.Errorf("failed to process tags: %w", err)
		}
		resource.Tags = tags
	}

	// Create resource in database
	if _, err := s.resources.Create(ctx, resource); err != nil {
		return err
	}

	// Schedule monitoring for the newly created resource
	if err := s.scheduler.Schedule(ctx, resource); err != nil {
		// Log the error but don't fail the entire operation
		// The resource was created successfully, monitoring scheduling failed
		return fmt.Errorf("%w: %v", ErrSchedulerSync, err)
	}

	return nil
}

// GetResourceByID retrieves a resource by its ID.
func (s *ResourceService) GetResourceByID(ctx context.Context, id string) (*domain.Resource, error) {
	resource, err := s.resources.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrResourceNotFound
		}
		return nil, err
	}

	// load incidents
	// load incidents and convert to the expected value slice type
	incidentPtrs, err := s.incidents.FindByResource(ctx, id, 100, 0) // default limit 100
	if err != nil {
		return nil, fmt.Errorf("failed to load incidents for resource: %w", err)
	}
	incidents := make([]domain.Incident, 0, len(incidentPtrs))
	for _, inc := range incidentPtrs {
		if inc == nil {
			continue
		}
		incidents = append(incidents, *inc)
	}
	resource.Incidents = incidents

	return resource, nil

}

// UpdateResource updates an existing resource by ID with the provided payload.
// It fetches the resource, applies changes, validates, updates, and reschedules monitoring.
func (s *ResourceService) UpdateResource(ctx context.Context, id string, payload *dto.UpdateResourcePayload) (*domain.Resource, error) {
	// Fetch existing resource
	resource, err := s.resources.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrResourceNotFound
		}
		return nil, err
	}

	// Apply updates from payload
	if payload.Name != nil {
		resource.Name = *payload.Name
	}
	if payload.Type != nil {
		resource.Type = *payload.Type
	}
	if payload.Target != nil {
		resource.Target = *payload.Target
	}
	if payload.Interval != nil {
		resource.Interval = *payload.Interval
	}
	if payload.Timeout != nil {
		resource.Timeout = *payload.Timeout
	}
	if payload.IsActive != nil {
		resource.IsActive = *payload.IsActive
	}

	if payload.Tags != nil {
		if len(*payload.Tags) > 0 {
			// Find or create tags by name
			tags, err := s.findOrCreateTags(ctx, *payload.Tags)
			if err != nil {
				return nil, fmt.Errorf("failed to process tags: %w", err)
			}
			resource.Tags = tags
		} else {
			// Clear all tags if empty slice provided
			resource.Tags = []*domain.Tags{}
		}
	}

	// Validate target format after updates
	if err := domain.ValidateResourceTarget(resource.Target, resource.Type); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrValidationFailed, err)
	}

	// Update resource in database
	if err := s.resources.Update(ctx, resource); err != nil {
		return nil, err
	}

	// Reschedule monitoring for the updated resource
	// This will unschedule the old task and create a new one with updated parameters
	if err := s.scheduler.Schedule(ctx, resource); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSchedulerSync, err)
	}

	return resource, nil
}

// DeleteResource soft deletes a resource and unschedules its monitoring.
func (s *ResourceService) DeleteResource(ctx context.Context, resourceID string) error {
	if err := s.resources.Delete(ctx, resourceID); err != nil {
		return err
	}

	// Unschedule monitoring for the deleted resource
	if err := s.scheduler.Unschedule(ctx, resourceID); err != nil {
		// Log the error but don't fail the entire operation
		return err
	}

	return nil
}

// ListActiveResources returns all active resources with pagination.
func (s *ResourceService) ListActiveResources(ctx context.Context, limit, offset int) ([]*domain.Resource, error) {
	return s.resources.FindActive(ctx, limit, offset)
}

// ListResourcesByTag returns resources filtered by a specific tag.
func (s *ResourceService) ListResourcesByTag(ctx context.Context, tagName string, limit, offset int) ([]*domain.Resource, error) {
	return s.resources.FindByTag(ctx, tagName, limit, offset)
}

// ListUnresolvedIncidents returns unresolved incidents for a specific resource.
func (s *ResourceService) ListUnresolvedIncidents(ctx context.Context, resourceID string) ([]*domain.Incident, error) {
	// First verify resource exists
	_, err := s.resources.FindByID(ctx, resourceID)
	if err != nil {
		return nil, err
	}

	// Get unresolved incidents for this resource
	return s.incidents.FindByResource(ctx, resourceID, 50, 0) // Default limit of 50
}

// ListAll retrieves all monitoring resources from the repository.
// This method supports listing all resources without pagination for simple use cases.
func (s *ResourceService) ListAll(ctx context.Context) ([]*domain.Resource, error) {
	// Use a large limit to get all resources (can be optimized with proper pagination later)
	return s.resources.List(ctx, 1000, 0)
}

// PauseMonitoring pauses monitoring for a specific resource by setting IsActive to false
// and unscheduling its monitoring tasks.
func (s *ResourceService) PauseMonitoring(ctx context.Context, resourceID string) error {
	// Retrieve the resource
	resource, err := s.resources.FindByID(ctx, resourceID)
	if err != nil {
		return err
	}

	// Check if already paused
	if !resource.IsActive {
		return nil // Already paused, nothing to do
	}

	// Set IsActive to false
	resource.IsActive = false

	// Update the resource in the database
	if err := s.resources.Update(ctx, resource); err != nil {
		return err
	}

	// Unschedule monitoring tasks for this resource
	if err := s.scheduler.Unschedule(ctx, resourceID); err != nil {
		// Log the error but consider the pause operation successful
		// since the database state has been updated
		return err
	}

	return nil
}

// ResumeMonitoring resumes monitoring for a specific resource by setting IsActive to true
// and rescheduling its monitoring tasks.
func (s *ResourceService) ResumeMonitoring(ctx context.Context, resourceID string) error {
	// Retrieve the resource
	resource, err := s.resources.FindByID(ctx, resourceID)
	if err != nil {
		return err
	}

	// Check if already active
	if resource.IsActive {
		return nil // Already active, nothing to do
	}

	// Set IsActive to true
	resource.IsActive = true

	// Update the resource in the database
	if err := s.resources.Update(ctx, resource); err != nil {
		return err
	}

	// Schedule monitoring tasks for this resource
	if err := s.scheduler.Schedule(ctx, resource); err != nil {
		// Log the error but consider the resume operation successful
		// since the database state has been updated
		return err
	}

	return nil
}

// AddTagsToResource adds multiple tags to a resource using GORM's Association mode.
func (s *ResourceService) AddTagsToResource(ctx context.Context, resourceID string, tagIDs []string) error {
	// Fetch the resource
	resource, err := s.resources.FindByID(ctx, resourceID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("%w: resource not found", ErrResourceNotFound)
		}
		return err
	}

	// Fetch all tags
	var tags []*domain.Tags
	for _, tagID := range tagIDs {
		tag, err := s.tags.FindByID(ctx, tagID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return fmt.Errorf("%w: tag with ID '%s' not found", ErrValidationFailed, tagID)
			}
			return err
		}
		tags = append(tags, tag)
	}

	// Use GORM association to append tags
	// This requires database access, so we need to get the DB instance
	// For now, we'll append tags to the resource and update
	resource.Tags = append(resource.Tags, tags...)

	if err := s.resources.Update(ctx, resource); err != nil {
		return fmt.Errorf("failed to add tags to resource: %w", err)
	}

	return nil
}

// RemoveTagFromResource removes a specific tag from a resource.
func (s *ResourceService) RemoveTagFromResource(ctx context.Context, resourceID string, tagID string) error {
	// Fetch the resource with its tags
	resource, err := s.resources.FindByID(ctx, resourceID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("%w: resource not found", ErrResourceNotFound)
		}
		return err
	}

	// Verify tag exists
	_, err = s.tags.FindByID(ctx, tagID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("%w: tag not found", ErrResourceNotFound)
		}
		return err
	}

	// Filter out the tag to be removed
	var filteredTags []*domain.Tags
	tagFound := false
	for _, tag := range resource.Tags {
		if tag.ID != tagID {
			filteredTags = append(filteredTags, tag)
		} else {
			tagFound = true
		}
	}

	if !tagFound {
		return fmt.Errorf("%w: tag not associated with resource", ErrValidationFailed)
	}

	resource.Tags = filteredTags

	if err := s.resources.Update(ctx, resource); err != nil {
		return fmt.Errorf("failed to remove tag from resource: %w", err)
	}

	return nil
}
