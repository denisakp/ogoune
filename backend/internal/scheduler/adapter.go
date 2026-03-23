package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/repository"
)

// RepositorySchedulerAdapter adapts the runtime Scheduler interface to the repository.Scheduler interface.
// This allows services (ResourceService, MaintenanceService) to call Schedule/Unschedule via the
// repository interface while internally delegating to the runtime scheduler (TimingWheel or Asynq).
type RepositorySchedulerAdapter struct {
	scheduler Scheduler
}

// NewRepositorySchedulerAdapter creates a new adapter wrapping a runtime scheduler.
func NewRepositorySchedulerAdapter(rtScheduler Scheduler) repository.Scheduler {
	return &RepositorySchedulerAdapter{
		scheduler: rtScheduler,
	}
}

// Schedule implements repository.Scheduler.Schedule()
// Converts a domain.Resource to a schedule call on the runtime scheduler.
func (a *RepositorySchedulerAdapter) Schedule(ctx context.Context, resource *domain.Resource) error {
	if resource == nil {
		return fmt.Errorf("repository scheduler: resource is nil")
	}

	if resource.ID == "" {
		return fmt.Errorf("repository scheduler: resource ID is empty")
	}

	// Only schedule if resource is active
	if !resource.IsActive {
		// If inactive, ensure it's unscheduled
		_ = a.scheduler.Unschedule(resource.ID)
		return nil
	}

	// Get interval from resource (in seconds)
	intervalSeconds := resource.Interval
	if intervalSeconds <= 0 {
		intervalSeconds = 300 // Default 300 seconds (5 minutes)
	}

	interval := time.Duration(intervalSeconds) * time.Second

	// Schedule via runtime scheduler
	return a.scheduler.Schedule(resource.ID, interval)
}

// Unschedule implements repository.Scheduler.Unschedule()
// Removes a resource from the runtime scheduler.
func (a *RepositorySchedulerAdapter) Unschedule(ctx context.Context, resourceID string) error {
	if resourceID == "" {
		return fmt.Errorf("repository scheduler: resource ID is empty")
	}

	return a.scheduler.Unschedule(resourceID)
}
