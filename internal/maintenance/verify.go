package maintenance

import "github.com/denisakp/ogoune/internal/port"

// Compile-time interface satisfaction checks.
var _ port.MaintenanceScheduler = (*SchedulerService)(nil)
