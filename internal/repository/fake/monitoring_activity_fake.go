package fake

import (
	"context"
	"sort"
	"sync"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/repository"
)

// MonitoringActivityFake is an in-memory implementation of MonitoringActivityRepository for testing.
type MonitoringActivityFake struct {
	activities map[string]*domain.MonitoringActivity
	mu         sync.RWMutex
}

// NewMonitoringActivityFake creates a new fake monitoring activity repository.
func NewMonitoringActivityFake() *MonitoringActivityFake {
	return &MonitoringActivityFake{
		activities: make(map[string]*domain.MonitoringActivity),
	}
}

// Create stores a new monitoring activity in memory.
func (f *MonitoringActivityFake) Create(ctx context.Context, activity *domain.MonitoringActivity) error {
	if activity == nil {
		return repository.ErrInvalidInput
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	// ID should be set by BeforeCreate hook in the domain model
	if activity.ID == "" {
		return repository.ErrInvalidInput
	}

	f.activities[activity.ID] = activity
	return nil
}

// List retrieves all monitoring activities with pagination.
func (f *MonitoringActivityFake) List(ctx context.Context, limit, offset int) ([]*domain.MonitoringActivity, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// Collect all activities
	activities := make([]*domain.MonitoringActivity, 0, len(f.activities))
	for _, activity := range f.activities {
		activities = append(activities, activity)
	}

	// Sort by created_at descending
	sort.Slice(activities, func(i, j int) bool {
		return activities[i].CreatedAt.After(activities[j].CreatedAt)
	})

	// Apply pagination
	start := offset
	if start > len(activities) {
		return []*domain.MonitoringActivity{}, nil
	}

	end := start + limit
	if end > len(activities) {
		end = len(activities)
	}

	return activities[start:end], nil
}

// FindByResourceID retrieves all monitoring activities for a specific resource.
func (f *MonitoringActivityFake) FindByResourceID(ctx context.Context, resourceID string, limit, offset int) ([]*domain.MonitoringActivity, error) {
	if resourceID == "" {
		return nil, repository.ErrInvalidInput
	}

	f.mu.RLock()
	defer f.mu.RUnlock()

	// Filter activities by resource ID
	filtered := make([]*domain.MonitoringActivity, 0)
	for _, activity := range f.activities {
		if activity.ResourceID == resourceID {
			filtered = append(filtered, activity)
		}
	}

	// Sort by created_at descending
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	// Apply pagination
	start := offset
	if start > len(filtered) {
		return []*domain.MonitoringActivity{}, nil
	}

	end := start + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[start:end], nil
}
