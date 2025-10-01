package service

import (
	"context"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/monitoring"
	"github.com/denisakp/pulseguard/internal/repository"
)

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
	// Domain validation can be added here if needed
	if err := s.resources.Create(ctx, resource); err != nil {
		return err
	}

	// Schedule monitoring for the newly created resource
	if err := s.scheduler.Schedule(ctx, resource); err != nil {
		// Log the error but don't fail the entire operation
		// The resource was created successfully, monitoring scheduling failed
		return err
	}

	return nil
}

// UpdateResource updates an existing resource and reschedules monitoring if needed.
func (s *ResourceService) UpdateResource(ctx context.Context, resource *domain.Resource) error {
	if err := s.resources.Update(ctx, resource); err != nil {
		return err
	}

	// Reschedule monitoring for the updated resource
	// This will unschedule the old task and create a new one with updated parameters
	if err := s.scheduler.Schedule(ctx, resource); err != nil {
		// Log the error but don't fail the entire operation
		return err
	}

	return nil
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
