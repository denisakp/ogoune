package port

import (
	"context"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
)

// MonitoringIncidentProcessor defines the contract for incident lifecycle management
// as consumed by the worker layer and heartbeat detector.
type MonitoringIncidentProcessor interface {
	CreateIncident(ctx context.Context, r *domain.Resource, result domain.CheckResult) error
	ResolveIncident(ctx context.Context, r *domain.Resource, result domain.CheckResult) error
	NotifyFlapping(ctx context.Context, r *domain.Resource, transitionCount, windowSeconds, maxDurationMinutes int) error
	NotifyStabilized(ctx context.Context, r *domain.Resource, finalStatus domain.ResourceStatus) error
	SendReminderIfDue(ctx context.Context, r *domain.Resource) error
	FindLatestIncidentForResource(ctx context.Context, resourceID string) (*domain.Incident, error)
	SetComponentRepository(repo ComponentRepository)
}

// MaintenanceScheduler defines the contract for scheduling maintenance windows.
type MaintenanceScheduler interface {
	EnsureScheduled(ctx context.Context) error
}

// ConfirmationRescheduler defines the contract for rescheduling a resource check
// with a temporary interval override (used during confirmation checks).
type ConfirmationRescheduler interface {
	ScheduleWithInterval(ctx context.Context, r *domain.Resource, interval time.Duration) error
}
