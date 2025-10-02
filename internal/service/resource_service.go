package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/monitoring"
	"github.com/denisakp/pulseguard/internal/repository"
)

// UpdateResourcePayload contains the fields that can be updated for a resource
type UpdateResourcePayload struct {
	Name     *string              `json:"name,omitempty"`
	Type     *domain.ResourceType `json:"type,omitempty"`
	Target   *string              `json:"target,omitempty"`
	Interval *int                 `json:"interval,omitempty"`
	Timeout  *int                 `json:"timeout,omitempty"`
	IsActive *bool                `json:"is_active,omitempty"`
}

// ResourceService orchestrates resource-related operations using repository interfaces.
// This service demonstrates the dependency injection pattern and serves as an example
// of how to compose repository operations while maintaining clean boundaries.
type ResourceService struct {
	resources repository.ResourceRepository
	incidents repository.IncidentRepository
	scheduler monitoring.Scheduler
}

// NewResourceService creates a new ResourceService with the given repository dependencies.
func NewResourceService(
	resources repository.ResourceRepository,
	incidents repository.IncidentRepository,
	scheduler monitoring.Scheduler,
) *ResourceService {
	return &ResourceService{
		resources: resources,
		incidents: incidents,
		scheduler: scheduler,
	}
}

// CreateResource creates a new resource using domain validation and persistence.
// After successful creation, it schedules monitoring for the resource.
func (s *ResourceService) CreateResource(ctx context.Context, resource *domain.Resource) error {
	// Validate target format
	if err := domain.ValidateResourceTarget(resource.Target, resource.Type); err != nil {
		return fmt.Errorf("%w: %v", ErrValidationFailed, err)
	}

	// Create resource in database
	if err := s.resources.Create(ctx, resource); err != nil {
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

// UpdateResource updates an existing resource by ID with the provided payload.
// It fetches the resource, applies changes, validates, updates, and reschedules monitoring.
func (s *ResourceService) UpdateResource(ctx context.Context, id string, payload *UpdateResourcePayload) (*domain.Resource, error) {
	// Fetch existing resource
	resource, err := s.resources.FindByID(ctx, id)
	if err != nil {
		if errors.Is(repository.ErrNotFound, err) {
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
