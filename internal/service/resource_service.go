package service

import (
	"context"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/repository"
)

// ResourceService orchestrates resource-related operations using repository interfaces.
// This service demonstrates the dependency injection pattern and serves as an example
// of how to compose repository operations while maintaining clean boundaries.
type ResourceService struct {
	resources repository.ResourceRepository
	incidents repository.IncidentRepository
}

// NewResourceService creates a new ResourceService with the given repository dependencies.
func NewResourceService(resources repository.ResourceRepository, incidents repository.IncidentRepository) *ResourceService {
	return &ResourceService{
		resources: resources,
		incidents: incidents,
	}
}

// CreateResource creates a new resource using domain validation and persistence.
func (s *ResourceService) CreateResource(ctx context.Context, resource *domain.Resource) error {
	// Domain validation can be added here if needed
	return s.resources.Create(ctx, resource)
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
