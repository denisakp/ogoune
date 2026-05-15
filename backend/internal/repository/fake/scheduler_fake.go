package fake

import (
	"context"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/repository"
)

// SchedulerFake is a mock implementation of repository.Scheduler for testing.
type SchedulerFake struct {
	scheduledResources   map[string]*domain.Resource
	unscheduledResources map[string]bool
}

// NewSchedulerFake creates a new fake scheduler service.
func NewSchedulerFake() *SchedulerFake {
	return &SchedulerFake{
		scheduledResources:   make(map[string]*domain.Resource),
		unscheduledResources: make(map[string]bool),
	}
}

// Schedule records that a resource was scheduled.
func (s *SchedulerFake) Schedule(ctx context.Context, r *domain.Resource) error {
	if r != nil {
		s.scheduledResources[r.ID] = r
	}
	return nil
}

// Unschedule records that a resource was unscheduled.
func (s *SchedulerFake) Unschedule(ctx context.Context, resourceID string) error {
	s.unscheduledResources[resourceID] = true
	delete(s.scheduledResources, resourceID)
	return nil
}

// Ensure SchedulerFake implements the repository.Scheduler interface
var _ repository.Scheduler = (*SchedulerFake)(nil)

// IsScheduled checks if a resource is scheduled (for testing).
func (s *SchedulerFake) IsScheduled(resourceID string) bool {
	_, exists := s.scheduledResources[resourceID]
	return exists
}

// IsUnscheduled checks if a resource was unscheduled (for testing).
func (s *SchedulerFake) IsUnscheduled(resourceID string) bool {
	return s.unscheduledResources[resourceID]
}
