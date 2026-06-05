// Package service — public status aggregator (spec 060 / US1).
package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/dto"
	"github.com/denisakp/ogoune/internal/port"
)

// PublicStatusService computes the snapshot exposed at GET /status.
type PublicStatusService struct {
	resources    port.ResourceRepository
	components   port.ComponentRepository
	incidents    port.IncidentRepository
	maintenances port.MaintenanceRepository
	uptimeAggs   port.UptimeDailyAggRepository
	clock        func() time.Time
}

func NewPublicStatusService(
	resources port.ResourceRepository,
	components port.ComponentRepository,
	incidents port.IncidentRepository,
	maintenances port.MaintenanceRepository,
	uptimeAggs port.UptimeDailyAggRepository,
) *PublicStatusService {
	return &PublicStatusService{
		resources:    resources,
		components:   components,
		incidents:    incidents,
		maintenances: maintenances,
		uptimeAggs:   uptimeAggs,
		clock:        time.Now,
	}
}

// SetClock overrides the wall clock — used by tests.
func (s *PublicStatusService) SetClock(c func() time.Time) { s.clock = c }

// GetCurrent returns the verdict + components + standalone resources +
// current-month incidents per spec 060 contract.
func (s *PublicStatusService) GetCurrent(ctx context.Context) (*dto.PublicStatus, error) {
	now := s.clock().UTC()

	// Load all active resources + components.
	components, err := s.components.List(ctx, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("public_status: list components: %w", err)
	}
	resources, err := s.resources.FindActive(ctx, 5000, 0)
	if err != nil {
		return nil, fmt.Errorf("public_status: list resources: %w", err)
	}

	// Index resources by component_id.
	byComponent := map[string][]*domain.Resource{}
	var standalone []*domain.Resource
	for _, r := range resources {
		if r.ComponentID != nil && *r.ComponentID != "" {
			byComponent[*r.ComponentID] = append(byComponent[*r.ComponentID], r)
		} else {
			standalone = append(standalone, r)
		}
	}

	// Build 90-day ribbon window once: [today-89 ... today].
	to := truncDayUTC(now)
	from := to.AddDate(0, 0, -89)

	// Bulk-fetch ribbons by resource id, grouped per resource.
	allIDs := make([]string, 0, len(resources))
	for _, r := range resources {
		allIDs = append(allIDs, r.ID)
	}
	aggs, err := s.uptimeAggs.FindRange(ctx, allIDs, from, to)
	if err != nil {
		return nil, fmt.Errorf("public_status: load uptime aggs: %w", err)
	}
	aggsByResource := map[string]map[string]float64{}
	for _, a := range aggs {
		key := a.Day.UTC().Format("2006-01-02")
		m := aggsByResource[a.ResourceID]
		if m == nil {
			m = map[string]float64{}
			aggsByResource[a.ResourceID] = m
		}
		m[key] = a.UptimeRatio
	}

	// Build per-resource summaries.
	resourceSummary := func(r *domain.Resource) dto.PublicResource {
		state := s.resourceState(ctx, r, now)
		ribbon := buildRibbon(from, to, aggsByResource[r.ID])
		return dto.PublicResource{
			ID:             r.ID,
			Name:           r.Name,
			Host:           r.Target,
			CurrentState:   state,
			Uptime90dRatio: averageRibbon(ribbon),
			UptimeRibbon:   ribbon,
		}
	}

	// Components in order (sorted by Name).
	sort.SliceStable(components, func(i, j int) bool { return components[i].Name < components[j].Name })
	out := &dto.PublicStatus{
		GeneratedAt:           now,
		Components:            make([]dto.PublicComponent, 0, len(components)),
		StandaloneResources:   make([]dto.PublicResource, 0, len(standalone)),
		CurrentMonthIncidents: []dto.PublicIncidentSummary{},
	}

	for _, c := range components {
		members := byComponent[c.ID]
		sort.SliceStable(members, func(i, j int) bool { return members[i].Name < members[j].Name })
		resSummaries := make([]dto.PublicResource, 0, len(members))
		for _, r := range members {
			resSummaries = append(resSummaries, resourceSummary(r))
		}
		out.Components = append(out.Components, dto.PublicComponent{
			ID:              c.ID,
			Name:            c.Name,
			AggregatedState: aggregateComponentState(resSummaries),
			Resources:       resSummaries,
		})
	}

	sort.SliceStable(standalone, func(i, j int) bool { return standalone[i].Name < standalone[j].Name })
	for _, r := range standalone {
		out.StandaloneResources = append(out.StandaloneResources, resourceSummary(r))
	}

	out.Verdict = computeVerdict(out.Components, out.StandaloneResources)

	// Current month incidents.
	monthIncidents, err := s.loadMonthIncidents(ctx, now)
	if err != nil {
		return nil, err
	}
	out.CurrentMonthIncidents = monthIncidents

	return out, nil
}

// resourceState maps the domain status into the public aggregated state,
// with maintenance overriding "down" when an active window covers the
// resource (FR-005).
func (s *PublicStatusService) resourceState(ctx context.Context, r *domain.Resource, now time.Time) dto.PublicAggregatedState {
	// Maintenance override first.
	if s.maintenances != nil {
		windows, err := s.maintenances.FindActiveForResource(ctx, r.ID, now)
		if err == nil && len(windows) > 0 {
			return dto.PublicStateMaintenance
		}
	}
	switch r.Status {
	case domain.StatusUp:
		return dto.PublicStateUp
	case domain.StatusDown, domain.StatusError:
		return dto.PublicStateDown
	case domain.StatusFlapping, domain.StatusWarn:
		return dto.PublicStateDegraded
	case domain.StatusPaused:
		return dto.PublicStateMaintenance
	default:
		return dto.PublicStateUnknown
	}
}

// loadMonthIncidents returns the public-shape incidents that started in the
// current month and resolved (or are still open).
func (s *PublicStatusService) loadMonthIncidents(ctx context.Context, now time.Time) ([]dto.PublicIncidentSummary, error) {
	// Cheap heuristic for the MVP: pull last 200 incidents, filter by start month.
	rows, err := s.incidents.List(ctx, 200, 0)
	if err != nil {
		return nil, fmt.Errorf("public_status: list incidents: %w", err)
	}
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0)
	out := make([]dto.PublicIncidentSummary, 0)
	for _, inc := range rows {
		if inc.StartedAt.Before(monthStart) || !inc.StartedAt.Before(monthEnd) {
			continue
		}
		sev := dto.PublicSeverityMinor
		if inc.ResolvedAt == nil {
			sev = dto.PublicSeverityMajor
		}
		out = append(out, dto.PublicIncidentSummary{
			ID:         inc.ID,
			Title:      incidentTitle(inc),
			StartedAt:  inc.StartedAt,
			ResolvedAt: inc.ResolvedAt,
			Severity:   sev,
			ResourceID: inc.ResourceID,
		})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].StartedAt.After(out[j].StartedAt) })
	return out, nil
}

func incidentTitle(inc *domain.Incident) string {
	if inc.Cause != "" {
		return inc.Cause
	}
	return "Incident on " + inc.Resource.Name
}

// computeVerdict applies FR-002 with the clarified promotion rule:
// Major if ≥ 1 component fully down OR ≥ 50% components in degraded/down,
// else partial if any non-up resource exists, else operational.
func computeVerdict(components []dto.PublicComponent, standalone []dto.PublicResource) dto.PublicVerdict {
	totalComponents := len(components)
	componentDown := 0
	componentDegradedOrDown := 0
	hasNonUp := false

	for _, c := range components {
		switch c.AggregatedState {
		case dto.PublicStateDown:
			componentDown++
			componentDegradedOrDown++
			hasNonUp = true
		case dto.PublicStateDegraded:
			componentDegradedOrDown++
			hasNonUp = true
		case dto.PublicStateMaintenance, dto.PublicStateUnknown:
			hasNonUp = true
		}
	}
	for _, r := range standalone {
		switch r.CurrentState {
		case dto.PublicStateDown:
			hasNonUp = true
			componentDown++
		case dto.PublicStateDegraded:
			hasNonUp = true
		case dto.PublicStateMaintenance, dto.PublicStateUnknown:
			hasNonUp = true
		}
	}

	// 50% rule requires at least 2 affected components — a single degraded
	// component on a 1-component page is "partial", not "major".
	majorTrigger := componentDown >= 1 ||
		(totalComponents >= 2 && componentDegradedOrDown >= 2 && componentDegradedOrDown*2 >= totalComponents)

	switch {
	case majorTrigger:
		return dto.PublicVerdict{
			Status: dto.VerdictMajorOutage,
			Label:  "Major Outage",
			Color:  "red",
		}
	case hasNonUp:
		return dto.PublicVerdict{
			Status: dto.VerdictPartialDegradation,
			Label:  "Partial Degradation",
			Color:  "orange",
		}
	default:
		return dto.PublicVerdict{
			Status: dto.VerdictOperational,
			Label:  "All Systems Operational",
			Color:  "green",
		}
	}
}

// aggregateComponentState — max(severity) with maintenance override per FR-003.
// down > degraded > maintenance > unknown > up.
func aggregateComponentState(resources []dto.PublicResource) dto.PublicAggregatedState {
	if len(resources) == 0 {
		return dto.PublicStateUnknown
	}
	rank := map[dto.PublicAggregatedState]int{
		dto.PublicStateUp:          0,
		dto.PublicStateUnknown:     1,
		dto.PublicStateMaintenance: 2,
		dto.PublicStateDegraded:    3,
		dto.PublicStateDown:        4,
	}
	worst := dto.PublicStateUp
	for _, r := range resources {
		if rank[r.CurrentState] > rank[worst] {
			worst = r.CurrentState
		}
	}
	return worst
}

func buildRibbon(from, to time.Time, byDay map[string]float64) []dto.PublicRibbonEntry {
	out := make([]dto.PublicRibbonEntry, 0, 90)
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		entry := dto.PublicRibbonEntry{Day: key}
		if v, ok := byDay[key]; ok {
			entry.Ratio = v
		} else {
			entry.Ratio = 0 // 0 distinguishes "unknown" only if combined with no row marker; we keep 0 for null-as-zero baseline.
		}
		out = append(out, entry)
	}
	return out
}

func averageRibbon(ribbon []dto.PublicRibbonEntry) float64 {
	if len(ribbon) == 0 {
		return 0
	}
	sum := 0.0
	for _, e := range ribbon {
		sum += e.Ratio
	}
	return sum / float64(len(ribbon))
}

func truncDayUTC(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
