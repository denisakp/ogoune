package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
)

// RepositorySchedulerAdapter adapts the runtime Scheduler interface to the port.ResourceScheduler interface.
// This allows services (ResourceService, MaintenanceService) to call Schedule/Unschedule via the
// port interface while internally delegating to the runtime scheduler (TimingWheel or Asynq).
type RepositorySchedulerAdapter struct {
	scheduler Scheduler
}

// IntervalScheduler provides an optional interval override scheduling path.
type IntervalScheduler interface {
	ScheduleWithInterval(ctx context.Context, resource *domain.Resource, interval time.Duration) error
}

// NewRepositorySchedulerAdapter creates a new adapter wrapping a runtime scheduler.
func NewRepositorySchedulerAdapter(rtScheduler Scheduler) port.ResourceScheduler {
	return &RepositorySchedulerAdapter{
		scheduler: rtScheduler,
	}
}

// Schedule implements port.ResourceScheduler.Schedule()
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

	// Heartbeat monitors are passively detected via incoming pings; no active polling needed
	if resource.Type == domain.ResourceHeartbeat {
		return nil
	}

	// Get interval from resource (in seconds)
	intervalSeconds := resource.Interval
	if intervalSeconds <= 0 {
		intervalSeconds = 300 // Default 300 seconds (5 minutes)
	}

	interval := time.Duration(intervalSeconds) * time.Second

	return a.ScheduleWithInterval(ctx, resource, interval)
}

// ScheduleWithInterval schedules a resource with an explicit interval override.
// The override is temporary and not persisted to the resource row.
func (a *RepositorySchedulerAdapter) ScheduleWithInterval(ctx context.Context, resource *domain.Resource, interval time.Duration) error {
	if resource == nil {
		return fmt.Errorf("repository scheduler: resource is nil")
	}

	if resource.ID == "" {
		return fmt.Errorf("repository scheduler: resource ID is empty")
	}

	if interval <= 0 {
		return fmt.Errorf("repository scheduler: interval must be > 0")
	}

	// Schedule via runtime scheduler using explicit interval override
	return a.scheduler.Schedule(resource.ID, interval)
}

// Unschedule implements port.ResourceScheduler.Unschedule()
// Removes a resource from the runtime scheduler.
func (a *RepositorySchedulerAdapter) Unschedule(ctx context.Context, resourceID string) error {
	if resourceID == "" {
		return fmt.Errorf("repository scheduler: resource ID is empty")
	}

	return a.scheduler.Unschedule(resourceID)
}
