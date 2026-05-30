package fake

import (
	"context"
	"sort"
	"sync"
	"time"

	domain "github.com/denisakp/ogoune/internal/domain"
)

// ResourceFake provides an in-memory implementation of ResourceRepository for testing.
type ResourceFake struct {
	mu        sync.RWMutex
	resources map[string]*domain.Resource
	tags      map[string]*domain.Tags // Simple tag store for FindByTag queries
	updateErr error
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

	resource.EnsureID()

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

func (r *ResourceFake) FindByHeartbeatSlug(ctx context.Context, slug string) (*domain.Resource, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, resource := range r.resources {
		if resource.HeartbeatSlug != nil && *resource.HeartbeatSlug == slug && resource.IsActive && resource.Type == domain.ResourceHeartbeat {
			copy := *resource
			return &copy, nil
		}
	}

	return nil, ErrNotFound
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

	if r.updateErr != nil {
		err := r.updateErr
		r.updateErr = nil
		return err
	}

	if _, exists := r.resources[resource.ID]; !exists {
		return ErrNotFound
	}

	// Store a copy
	copy := *resource
	r.resources[resource.ID] = &copy

	return nil
}

// FailNextUpdate configures the fake to fail exactly one upcoming Update call.
func (r *ResourceFake) FailNextUpdate(err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.updateErr = err
}

func (r *ResourceFake) UpdateMetadata(ctx context.Context, id string, metadata *domain.ResourceMetaData) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	res, exists := r.resources[id]
	if !exists {
		return ErrNotFound
	}
	// Set metadata pointer to a copy to avoid external mutation
	if metadata != nil {
		mdCopy := *metadata
		res.Metadata = &mdCopy
	} else {
		res.Metadata = nil
	}
	return nil
}

// FindByComponentID returns resources associated with a component.
func (r *ResourceFake) FindByComponentID(ctx context.Context, componentID string) ([]*domain.Resource, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var resources []*domain.Resource
	for _, res := range r.resources {
		if res.ComponentID != nil && *res.ComponentID == componentID {
			copy := *res
			resources = append(resources, &copy)
		}
	}

	return resources, nil
}

// CountByComponentID returns number of resources for a component.
func (r *ResourceFake) CountByComponentID(ctx context.Context, componentID string) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var count int64
	for _, res := range r.resources {
		if res.ComponentID != nil && *res.ComponentID == componentID {
			count++
		}
	}

	return count, nil
}

func (r *ResourceFake) FindMissedHeartbeats(ctx context.Context, now time.Time, limit int) ([]*domain.Resource, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if limit <= 0 {
		limit = 1000
	}

	missed := make([]*domain.Resource, 0)
	for _, res := range r.resources {
		if !res.IsActive || res.Type != domain.ResourceHeartbeat || res.Status != domain.StatusUp || res.LastPingAt == nil {
			continue
		}
		if res.HeartbeatInterval == nil || res.HeartbeatGrace == nil {
			continue
		}
		deadline := res.LastPingAt.Add(time.Duration(*res.HeartbeatInterval+*res.HeartbeatGrace) * time.Second)
		if now.After(deadline) {
			copy := *res
			missed = append(missed, &copy)
		}
	}

	sort.Slice(missed, func(i, j int) bool {
		return missed[i].LastPingAt.Before(*missed[j].LastPingAt)
	})

	if len(missed) > limit {
		missed = missed[:limit]
	}
	return missed, nil
}

func (r *ResourceFake) UpdateLastPingAt(ctx context.Context, id string, at time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	res, exists := r.resources[id]
	if !exists || !res.IsActive || res.Type != domain.ResourceHeartbeat {
		return ErrNotFound
	}
	res.LastPingAt = &at
	return nil
}

func (r *ResourceFake) UpdateMonitoringState(ctx context.Context, resource *domain.Resource) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.updateErr != nil {
		err := r.updateErr
		r.updateErr = nil
		return err
	}

	res, exists := r.resources[resource.ID]
	if !exists {
		return ErrNotFound
	}
	res.Status = resource.Status
	res.FailureCount = resource.FailureCount
	res.LastChecked = resource.LastChecked
	res.LastStatusTransition = resource.LastStatusTransition
	res.FlapStartedAt = resource.FlapStartedAt
	return nil
}

func (r *ResourceFake) UpdateStatus(ctx context.Context, id string, status domain.ResourceStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	res, exists := r.resources[id]
	if !exists {
		return ErrNotFound
	}
	res.Status = status
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

func (r *ResourceFake) FindScheduledResources(ctx context.Context) ([]*domain.Resource, error) {
	return r.FindActive(ctx, 1000, 0)
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
