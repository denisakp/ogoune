package monitoring

import (
	"context"

	"github.com/denisakp/pulseguard/internal/domain"
)

// Scheduler defines the interface for scheduling monitoring tasks.
type Scheduler interface {
	Schedule(ctx context.Context, r *domain.Resource) error
	Unschedule(ctx context.Context, resourceID string) error
}

// Ensure SchedulerService implements the Scheduler interface
var _ Scheduler = (*SchedulerService)(nil)
