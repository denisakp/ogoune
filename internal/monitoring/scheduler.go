package monitoring

import (
	"github.com/denisakp/pulseguard/internal/repository"
)

// Ensure SchedulerService implements the repository.Scheduler interface
var _ repository.Scheduler = (*SchedulerService)(nil)
