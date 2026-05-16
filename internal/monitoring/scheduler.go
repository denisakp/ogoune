package monitoring

import (
	"github.com/denisakp/ogoune/internal/port"
)

// Ensure SchedulerService implements the port.ResourceScheduler interface
var _ port.ResourceScheduler = (*SchedulerService)(nil)
