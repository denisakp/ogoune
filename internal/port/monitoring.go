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

// HostContextMetrics counts why an incident's host context came back empty.
//
// The enrichment is invisible when it produces nothing, so without a counter an
// operator cannot tell "this incident had no host data" from "the correlation has
// been broken for a week" (spec 089, FR-013a). Deliberately a one-method
// consumer-side contract rather than a method on domain.MetricsRecorder, which
// declares RecordCheck alone and exists for the check executor: an incident-read
// concern does not belong on it.
//
// no_host_attached is intentionally not a reason. It is the normal state of most
// monitors, not a signal.
type HostContextMetrics interface {
	// RecordHostContextAbsent increments the counter for one absence reason:
	// "no_samples", "out_of_retention" or "lookup_error".
	RecordHostContextAbsent(reason string)
}

// DatabaseHealthMetrics counts why a database check collected no health.
//
// All three reasons are invisible by design: the check passes and the field is
// simply absent, so without a counter an operator cannot tell "my credential
// lacks the grant" from "this has been broken since the last release"
// (spec 088, FR-016a).
//
// A one-method consumer-side contract, following the HostContextMetrics
// precedent, rather than a method on domain.MetricsRecorder -- that interface
// declares RecordCheck alone and exists for the check executor.
type DatabaseHealthMetrics interface {
	// RecordDatabaseHealthSkipped increments the counter for one skip reason:
	// "deadline", "privilege", "unsupported_version" or "no_fields".
	RecordDatabaseHealthSkipped(reason string)
}
