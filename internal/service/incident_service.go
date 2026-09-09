package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/dto"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository"
	"github.com/denisakp/ogoune/internal/repository/sqlc/dynquery"
)

// IncidentService provides business logic for incident management operations.
// Correlation window for incident host context (spec 089, FR-002). Deliberately
// constants, not configuration: the bounds are part of the response contract and
// changing them after the interface ships means re-testing every fixture.
const (
	hostContextWindowBefore = 5 * time.Minute
	hostContextWindowAfter  = 1 * time.Minute
)

type IncidentService struct {
	incidents   port.IncidentRepository
	eventSteps  port.IncidentEventStepRepository
	hostMetrics port.HostMetricsRepository
	hosts       port.HostRepository
	// rawWindow is the configured full-resolution retention window. It decides
	// the resolution marker and nothing else; the correlation window above is
	// independent of it.
	rawWindow time.Duration
	now       func() time.Time
	// hostMetricsObs counts why a context came back empty. Optional: a nil
	// recorder simply records nothing, so no call site is forced to supply one.
	hostCtxObs port.HostContextMetrics
}

// WithHostContextMetrics attaches the absence counter. Separate from the
// constructor so the observability wiring stays optional and every existing
// caller keeps working unchanged.
func (s *IncidentService) WithHostContextMetrics(m port.HostContextMetrics) *IncidentService {
	s.hostCtxObs = m
	return s
}

// recordAbsence counts one absence reason. no_host_attached is never passed here:
// it is the normal state of most monitors, not a signal (spec 089, FR-013a).
func (s *IncidentService) recordAbsence(reason string) {
	if s.hostCtxObs != nil {
		s.hostCtxObs.RecordHostContextAbsent(reason)
	}
}

// NewIncidentService creates a new IncidentService with the given repository dependencies.
func NewIncidentService(
	incidents port.IncidentRepository,
	eventSteps port.IncidentEventStepRepository,
	hostMetrics port.HostMetricsRepository,
	hosts port.HostRepository,
	rawWindow time.Duration,
) *IncidentService {
	return &IncidentService{
		incidents:   incidents,
		eventSteps:  eventSteps,
		hostMetrics: hostMetrics,
		hosts:       hosts,
		rawWindow:   rawWindow,
		now:         time.Now,
	}
}

// ListAll retrieves all incidents with pagination.
// Limit defaults to 25 if not provided or invalid.
// Offset defaults to 0 if negative.
func (s *IncidentService) ListAll(ctx context.Context, limit, offset int) ([]*domain.Incident, error) {
	if limit <= 0 {
		limit = 25
	}

	if offset < 0 {
		offset = 0
	}

	incidents, err := s.incidents.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list incidents: %w", err)
	}

	return incidents, nil
}

// ListByFilter passes the dynamic filter through to the repo.
func (s *IncidentService) ListByFilter(ctx context.Context, f dynquery.IncidentFilter, page, perPage int) ([]*domain.Incident, int, error) {
	return s.incidents.ListIncidentsByFilter(ctx, f, page, perPage)
}

// ListUnresolved retrieves all unresolved incidents with pagination.
func (s *IncidentService) ListUnresolved(ctx context.Context, limit, offset int) ([]*domain.Incident, error) {
	if limit <= 0 {
		limit = 25
	}

	if offset < 0 {
		offset = 0
	}

	incidents, err := s.incidents.FindUnresolved(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list unresolved incidents: %w", err)
	}

	return incidents, nil
}

// GetIncidentByID retrieves a single incident by its ID with all related event steps.
func (s *IncidentService) GetIncidentByID(ctx context.Context, id string) (*domain.Incident, error) {
	incident, err := s.incidents.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("%w: incident not found", ErrResourceNotFound)
		}
		return nil, fmt.Errorf("failed to get incident: %w", err)
	}

	// Fetch all event steps for this incident
	eventSteps, err := s.eventSteps.List(ctx, 1000, 0) // Get all steps (no pagination for now)
	if err != nil {
		// Don't fail the entire request if we can't get event steps
		// Just log and return incident without steps
		return incident, nil
	}

	// Filter event steps for this incident
	var incidentSteps []domain.IncidentEventStep
	for _, step := range eventSteps {
		if step.IncidentID == incident.ID {
			incidentSteps = append(incidentSteps, *step)
		}
	}

	incident.EventStep = incidentSteps

	incident.HostContext = s.buildHostContext(ctx, incident)

	return incident, nil
}

// GetIncidentsByResource retrieves all incidents for a specific resource with pagination.
func (s *IncidentService) GetIncidentsByResource(ctx context.Context, resourceID string, limit, offset int) ([]*domain.Incident, error) {
	if limit <= 0 {
		limit = 25
	}

	if offset < 0 {
		offset = 0
	}

	incidents, err := s.incidents.FindByResource(ctx, resourceID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get incidents for resource: %w", err)
	}

	return incidents, nil
}

// GetEventStepsForIncident retrieves all event steps for a specific incident.
func (s *IncidentService) GetEventStepsForIncident(ctx context.Context, incidentID string) ([]domain.IncidentEventStep, error) {
	// Get all event steps (considering a small limit for now, can be optimized)
	steps, err := s.eventSteps.List(ctx, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get event steps: %w", err)
	}

	// Filter steps for this incident
	var result []domain.IncidentEventStep
	for _, step := range steps {
		if step.IncidentID == incidentID {
			result = append(result, *step)
		}
	}

	return result, nil
}

// GetActiveIncident returns the most recent unresolved incident for a resource.
// When no active incident exists, it returns (nil, nil).
func (s *IncidentService) GetActiveIncident(ctx context.Context, resourceID string) (*dto.LiveActiveIncident, error) {
	if resourceID == "" {
		return nil, fmt.Errorf("resource_id is required")
	}

	incident, err := s.incidents.FindActiveByResourceID(ctx, resourceID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get active incident for resource %s: %w", resourceID, err)
	}

	return &dto.LiveActiveIncident{
		ID:        incident.ID,
		StartedAt: incident.StartedAt,
		Cause:     incident.Cause,
	}, nil
}

// buildHostContext aggregates what the monitor's host was doing around the
// incident start. It is best-effort by contract: every failure path returns nil
// so the incident detail response is never lost to its own enrichment
// (spec 089, FR-012).
//
// The monitor -> host link is resolved here, at read time, not frozen when the
// incident opened. Re-attaching a monitor therefore changes what its past
// incidents display; that is a documented limitation (FR-017), not an oversight.
func (s *IncidentService) buildHostContext(ctx context.Context, incident *domain.Incident) *domain.HostContext {
	if s.hostMetrics == nil || s.hosts == nil {
		return nil
	}
	if incident.Resource.HostID == nil || *incident.Resource.HostID == "" {
		// The normal state of most monitors, not a signal. Not logged, not counted.
		return nil
	}
	hostID := *incident.Resource.HostID

	from := incident.StartedAt.Add(-hostContextWindowBefore)
	to := incident.StartedAt.Add(hostContextWindowAfter)
	// The window may still be elapsing when the operator opens a fresh incident;
	// aggregate over what exists rather than waiting for it (FR-004).
	if now := s.now(); to.After(now) {
		to = now
	}
	if !to.After(from) {
		return nil
	}

	agg, err := s.hostMetrics.AggregateWindow(ctx, hostID, from, to)
	if err != nil {
		slog.Debug("incident host context: aggregate failed",
			"incident_id", incident.ID, "host_id", hostID, "error", err)
		s.recordAbsence("lookup_error")
		return nil
	}
	if agg == nil || agg.SampleCount == 0 {
		// Either the host never reported in this window, or retention has since
		// purged it. The window's age tells them apart.
		if s.rawWindow > 0 && from.Before(s.now().Add(-s.rawWindow)) {
			s.recordAbsence("out_of_retention")
		} else {
			s.recordAbsence("no_samples")
		}
		return nil
	}

	host, err := s.hosts.FindByID(ctx, hostID)
	if err != nil || host == nil {
		slog.Debug("incident host context: host lookup failed",
			"incident_id", incident.ID, "host_id", hostID, "error", err)
		s.recordAbsence("lookup_error")
		return nil
	}

	return &domain.HostContext{
		HostID:      hostID,
		HostName:    host.Name,
		PeakCPUPct:  agg.PeakCPUPct,
		PeakMemPct:  agg.PeakMemPct,
		WorstDisk:   worstDisk(agg.Disks),
		SampleCount: agg.SampleCount,
		Resolution:  s.resolutionFor(from),
		WindowFrom:  from,
		WindowTo:    to,
	}
}

// resolutionFor reports whether the window is entirely inside the configured
// full-resolution retention window. It takes the window's *start*, not its end:
// a window that straddles the threshold has thinned samples in it, so judging it
// by its most recent edge would call a partly-degraded aggregate exact.
//
// It errs toward reduced -- the thinning job runs periodically, so samples just
// past the threshold may still be at full resolution, and understating confidence
// is harmless where overstating it is not (spec 089, FR-007b). It never inspects
// the samples themselves: the agent's reporting interval is configurable, so a
// host natively reporting once a minute would otherwise be mislabelled as
// degraded while its data is intact (FR-007a).
func (s *IncidentService) resolutionFor(windowStart time.Time) domain.HostContextResolution {
	if s.rawWindow <= 0 {
		return domain.HostContextReduced
	}
	if windowStart.After(s.now().Add(-s.rawWindow)) {
		return domain.HostContextFull
	}
	return domain.HostContextReduced
}

// worstDisk returns the highest-utilisation mount seen anywhere in the window,
// or nil when no sample reported one. A host with no disks still gets its CPU and
// memory peaks: one missing signal must not suppress the others.
func worstDisk(docs [][]domain.DiskUsage) *domain.DiskUsage {
	var worst *domain.DiskUsage
	for _, doc := range docs {
		for i := range doc {
			if worst == nil || doc[i].UsedPct > worst.UsedPct {
				d := doc[i]
				worst = &d
			}
		}
	}
	return worst
}
