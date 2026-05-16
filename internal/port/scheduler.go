package port

import (
	"context"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
)

// Scheduler defines the interface for the full scheduler lifecycle.
type Scheduler interface {
	Start(ctx context.Context, repo ActiveResourceRepository) error
	Schedule(resourceID string, interval time.Duration) error
	Unschedule(resourceID string) error
	Pause(resourceID string) error
	Resume(resourceID string) error
	Stop(ctx context.Context) error
}

// ActiveResourceRepository defines the interface for accessing active resources
// during scheduler startup.
type ActiveResourceRepository interface {
	FindScheduledResources(ctx context.Context) ([]ScheduleItem, error)
}

// ScheduleItem represents a schedulable resource.
type ScheduleItem struct {
	ResourceID string
	Interval   time.Duration
	Paused     bool
}

// AsynqSchedulerAdapter bridges the scheduler runtime to the existing
// Asynq-backed scheduler service.
type AsynqSchedulerAdapter interface {
	Schedule(ctx context.Context, r *domain.Resource) error
	Unschedule(ctx context.Context, resourceID string) error
}

// AsynqSchedulerAdapterWithInterval optionally supports scheduling with
// a temporary interval override.
type AsynqSchedulerAdapterWithInterval interface {
	AsynqSchedulerAdapter
	ScheduleWithInterval(ctx context.Context, r *domain.Resource, interval time.Duration) error
}
