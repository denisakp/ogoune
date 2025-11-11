package fake

import (
	"context"
	"sort"
	"sync"

	domain "github.com/denisakp/pulseguard/internal/domain"
)

// ResourceFake provides an in-memory implementation of ResourceRepository for testing.
type ResourceFake struct {
	mu        sync.RWMutex
	resources map[string]*domain.Resource
	tags      map[string]*domain.Tags // Simple tag store for FindByTag queries
}

// NewResourceFake creates a new in-memory ResourceRepository fake.
func NewResourceFake() *ResourceFake {
	return &ResourceFake{
		resources: make(map[string]*domain.Resource),
		tags:      make(map[string]*domain.Tags),
	}
}

func (r *ResourceFake) Create(ctx context.Context, resource *domain.Resource) (*domain.Resource, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Call BeforeCreate hook like GORM does - generates ID if not set
	if err := resource.BeforeCreate(nil); err != nil {
		return nil, ErrInvalidInput
	}

	if _, exists := r.resources[resource.ID]; exists {
		return nil, ErrDuplicate
	}

	// Store a copy to avoid external mutations
	copy := *resource
	r.resources[resource.ID] = &copy

	return &copy, nil
}

func (r *ResourceFake) FindByID(ctx context.Context, id string) (*domain.Resource, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	resource, exists := r.resources[id]
	if !exists {
		return nil, ErrNotFound
	}

	// Return a copy to avoid external mutations
	copy := *resource
	return &copy, nil
}

func (r *ResourceFake) List(ctx context.Context, limit, offset int) ([]*domain.Resource, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Convert to slice and sort by created_at DESC
	var resources []*domain.Resource
	for _, res := range r.resources {
		copy := *res
		resources = append(resources, &copy)
	}

	sort.Slice(resources, func(i, j int) bool {
		return resources[i].CreatedAt.After(resources[j].CreatedAt)
	})

	// Apply pagination
	if offset >= len(resources) {
		return []*domain.Resource{}, nil
	}

	end := offset + limit
	if end > len(resources) {
		end = len(resources)
	}

	return resources[offset:end], nil
}

func (r *ResourceFake) Update(ctx context.Context, resource *domain.Resource) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.resources[resource.ID]; !exists {
		return ErrNotFound
	}

	// Store a copy
	copy := *resource
	r.resources[resource.ID] = &copy

	return nil
}

func (r *ResourceFake) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	resource, exists := r.resources[id]
	if !exists {
		return ErrNotFound
	}

	// Soft delete - set IsActive to false
	resource.IsActive = false

	return nil
}

func (r *ResourceFake) FindActive(ctx context.Context, limit, offset int) ([]*domain.Resource, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Filter active resources
	var activeResources []*domain.Resource
	for _, res := range r.resources {
		if res.IsActive {
			copy := *res
			activeResources = append(activeResources, &copy)
		}
	}

	// Sort by created_at DESC
	sort.Slice(activeResources, func(i, j int) bool {
		return activeResources[i].CreatedAt.After(activeResources[j].CreatedAt)
	})

	// Apply pagination
	if offset >= len(activeResources) {
		return []*domain.Resource{}, nil
	}

	end := offset + limit
	if end > len(activeResources) {
		end = len(activeResources)
	}

	return activeResources[offset:end], nil
}

func (r *ResourceFake) FindByTag(ctx context.Context, tagName string, limit, offset int) ([]*domain.Resource, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Simple implementation: check if any resource has the tag name
	// In a real implementation, this would be a proper JOIN
	var tagged []*domain.Resource
	for _, res := range r.resources {
		if res.IsActive { // Only active resources
			for _, tag := range res.Tags {
				if tag.Name == tagName {
					copy := *res
					tagged = append(tagged, &copy)
					break
				}
			}
		}
	}

	// Sort by created_at DESC
	sort.Slice(tagged, func(i, j int) bool {
		return tagged[i].CreatedAt.After(tagged[j].CreatedAt)
	})

	// Apply pagination
	if offset >= len(tagged) {
		return []*domain.Resource{}, nil
	}

	end := offset + limit
	if end > len(tagged) {
		end = len(tagged)
	}

	return tagged[offset:end], nil
}

// AddTag is a helper for tests to associate tags with resources
func (r *ResourceFake) AddTag(resourceID string, tag *domain.Tags) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	resource, exists := r.resources[resourceID]
	if !exists {
		return ErrNotFound
	}

	// Add tag to resource's tag slice
	resource.Tags = append(resource.Tags, tag)

	// Store tag for reference
	r.tags[tag.ID] = tag

	return nil
}
