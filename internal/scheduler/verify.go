package scheduler

import "github.com/denisakp/ogoune/internal/port"

// Compile-time interface satisfaction checks.
var (
	_ port.Scheduler                      = (*TimingWheelScheduler)(nil)
	_ port.Scheduler                      = (*Asynq)(nil)
	_ port.ResourceScheduler              = (*RepositorySchedulerAdapter)(nil)
	_ port.AsynqSchedulerAdapterWithInterval = (*RepositorySchedulerAdapter)(nil)
	_ port.ConfirmationRescheduler        = (*RepositorySchedulerAdapter)(nil)
)
