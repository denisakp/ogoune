package service

import (
	"context"
	"fmt"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/dto"
	"github.com/denisakp/ogoune/internal/port"
)

// StatusPageService provides data aggregation for public status pages.
type StatusPageService struct {
	resourceRepo           port.ResourceRepository
	incidentRepo           port.IncidentRepository
	monitoringActivityRepo port.MonitoringActivityRepository
	maintenanceRepo        port.MaintenanceRepository
	settingsRepo           port.StatusPageSettingsRepository
	componentRepo          port.ComponentRepository
}

// NewStatusPageService creates a new StatusPageService instance.
func NewStatusPageService(
	resourceRepo port.ResourceRepository,
	incidentRepo port.IncidentRepository,
	monitoringActivityRepo port.MonitoringActivityRepository,
	maintenanceRepo port.MaintenanceRepository,
	settingsRepo port.StatusPageSettingsRepository,
	componentRepo port.ComponentRepository,
) *StatusPageService {
	return &StatusPageService{
		resourceRepo:           resourceRepo,
		incidentRepo:           incidentRepo,
		monitoringActivityRepo: monitoringActivityRepo,
		maintenanceRepo:        maintenanceRepo,
		settingsRepo:           settingsRepo,
		componentRepo:          componentRepo,
	}
}

// GetData fetches and aggregates all data needed for the status page.
func (s *StatusPageService) GetData(ctx context.Context) (*dto.StatusPageData, error) {
	// Get settings to check if paused monitors should be hidden
	var settings *domain.StatusPageSettings
	if s.settingsRepo != nil {
		var err error
		settings, err = s.settingsRepo.Get(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch settings: %w", err)
		}
	}

	// Fetch all active resources
	resources, err := s.resourceRepo.FindActive(ctx, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch resources: %w", err)
	}

	components := []*domain.Component{}
	if s.componentRepo != nil {
		components, _ = s.componentRepo.List(ctx, 1000, 0)
	}

	// Filter paused resources if setting is enabled (default: hide paused)
	hidePaused := settings == nil || settings.HidePausedMonitors
	if hidePaused {
		filteredResources := make([]*domain.Resource, 0, len(resources))
		for _, r := range resources {
			if r.Status != domain.StatusPaused {
				filteredResources = append(filteredResources, r)
			}
		}
		resources = filteredResources
	}

	// Build resource status info with 90-day data
	resourceInfos := make([]dto.ResourceStatusInfo, 0, len(resources))
	resourcesByComponent := make(map[string][]dto.ResourceStatusInfo)
	hasNonOperationalResource := false

	for _, resource := range resources {
		info, err := s.buildResourceStatusInfo(ctx, resource)
		if err != nil {
			// Log error but continue processing other resources
			continue
		}

		// If resource belongs to a component, add it to the component's resource list
		// but do NOT add it to the standalone resourceInfos list
		if resource.ComponentID != nil {
			resourcesByComponent[*resource.ComponentID] = append(resourcesByComponent[*resource.ComponentID], info)
		} else {
			// Only standalone resources (no component) go to the main resources list
			resourceInfos = append(resourceInfos, info)
		}

		// Check if any resource is not "up", excluding pending/unknown resources
		// Newly added or unchecked resources should not count as an outage
		if info.CurrentStatus != "up" && info.CurrentStatus != "degraded" {
			hasNonOperationalResource = true
		} else if info.CurrentStatus == "degraded" {
			// Only count degraded if it's not from pending/unknown status
			if resource.Status != domain.StatusPending && resource.Status != domain.StatusUnknown {
				hasNonOperationalResource = true
			}
		}
	}

	// Calculate global status
	globalStatus := "all_systems_operational"
	if hasNonOperationalResource {
		globalStatus = "some_systems_down"
	}

	// Build settings DTO (only include public-facing fields)
	var settingsDTO *dto.StatusPageSettings
	if settings != nil {
		settingsDTO = &dto.StatusPageSettings{
			Name:                 settings.Name,
			HomepageURL:          settings.HomepageURL,
			GoogleAnalyticsID:    settings.GoogleAnalyticsID,
			EnableDetailsPage:    settings.EnableDetailsPage,
			ShowUptimePercentage: settings.ShowUptimePercentage,
		}
	}

	componentInfos := make([]dto.ComponentStatusInfo, 0, len(components))
	for _, component := range components {
		infos := resourcesByComponent[component.ID]
		componentStatus := s.mapComponentStatus(infos)
		componentInfos = append(componentInfos, dto.ComponentStatusInfo{
			ID:            component.ID,
			Name:          component.Name,
			Description:   component.Description,
			CurrentStatus: componentStatus,
			Resources:     infos,
		})
	}

	return &dto.StatusPageData{
		GlobalStatus: globalStatus,
		GeneratedAt:  time.Now(),
		Resources:    resourceInfos,
		Components:   componentInfos,
		Settings:     settingsDTO,
	}, nil
}

// buildResourceStatusInfo creates a ResourceStatusInfo from a Resource entity.
func (s *StatusPageService) buildResourceStatusInfo(ctx context.Context, resource *domain.Resource) (dto.ResourceStatusInfo, error) {
	// Calculate current status (map domain status to simplified enum)
	currentStatus := s.mapResourceStatus(resource.Status)

	// Calculate 90-day uptime and daily status
	uptimePercentage, dailyStatus, err := s.calculate90DayData(ctx, resource)
	if err != nil {
		return dto.ResourceStatusInfo{}, fmt.Errorf("failed to calculate 90-day data for resource %s: %w", resource.ID, err)
	}

	return dto.ResourceStatusInfo{
		ID:                         resource.ID,
		Name:                       resource.Name,
		CurrentStatus:              currentStatus,
		UptimePercentageLast90Days: uptimePercentage,
		DailyStatusLast90Days:      dailyStatus,
	}, nil
}

// mapResourceStatus maps domain.ResourceStatus to simplified status enum
func (s *StatusPageService) mapResourceStatus(status domain.ResourceStatus) string {
	switch status {
	case domain.StatusUp:
		return "up"
	case domain.StatusDown, domain.StatusError:
		return "down"
	case domain.StatusWarn, domain.StatusPending, domain.StatusUnknown:
		return "degraded"
	case domain.StatusPaused:
		return "up" // Paused resources are considered operational
	default:
		return "degraded"
	}
}

func (s *StatusPageService) mapComponentStatus(resources []dto.ResourceStatusInfo) string {
	hasDown := false
	hasDegraded := false

	for _, r := range resources {
		switch r.CurrentStatus {
		case "down":
			hasDown = true
		case "degraded":
			hasDegraded = true
		}
	}

	switch {
	case hasDown:
		return "down"
	case hasDegraded:
		return "degraded"
	default:
		return "up"
	}
}

// calculate90DayData calculates the 90-day uptime percentage and daily status array
func (s *StatusPageService) calculate90DayData(ctx context.Context, resource *domain.Resource) (float64, []string, error) {
	now := time.Now()
	startDate := now.AddDate(0, 0, -90)

	// Fetch all monitoring activities for the last 90 days
	activities, err := s.monitoringActivityRepo.FindByResourceID(ctx, resource.ID, 100000, 0)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to fetch monitoring activities: %w", err)
	}

	// Filter activities within the last 90 days
	filteredActivities := make([]*domain.MonitoringActivity, 0)
	for _, activity := range activities {
		if activity.CreatedAt.After(startDate) || activity.CreatedAt.Equal(startDate) {
			filteredActivities = append(filteredActivities, activity)
		}
	}

	// Fetch all incidents for the resource
	incidents, err := s.incidentRepo.FindByResource(ctx, resource.ID, 10000, 0)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to fetch incidents: %w", err)
	}

	// Filter incidents within the last 90 days
	filteredIncidents := make([]*domain.Incident, 0)
	for _, incident := range incidents {
		// Include incident if it started within the last 90 days or overlaps with the period
		if incident.StartedAt.After(startDate) ||
			(incident.ResolvedAt != nil && incident.ResolvedAt.After(startDate)) ||
			(incident.ResolvedAt == nil) {
			filteredIncidents = append(filteredIncidents, incident)
		}
	}

	// Calculate overall uptime percentage
	totalChecks := len(filteredActivities)
	successfulChecks := 0
	for _, activity := range filteredActivities {
		if activity.Success {
			successfulChecks++
		}
	}

	uptimePercentage := 100.0
	if totalChecks > 0 {
		uptimePercentage = (float64(successfulChecks) / float64(totalChecks)) * 100
	}

	// Calculate daily status for each of the last 90 days
	dailyStatus := make([]string, 90)
	for i := range 90 {
		dayStart := startDate.AddDate(0, 0, i)
		dayEnd := dayStart.AddDate(0, 0, 1)

		// Check if this day is before the resource was created
		if dayStart.Before(resource.CreatedAt) {
			dailyStatus[i] = "no_data"
		} else {
			status := s.calculateDayStatus(dayStart, dayEnd, filteredActivities, filteredIncidents)
			dailyStatus[i] = status
		}
	}

	return uptimePercentage, dailyStatus, nil
}

// calculateDayStatus determines the status for a specific day based on activities and incidents
func (s *StatusPageService) calculateDayStatus(dayStart, dayEnd time.Time, activities []*domain.MonitoringActivity, incidents []*domain.Incident) string {
	// Check if there are any incidents on this day
	hasMajorIncident := false
	hasMinorIncident := false

	for _, incident := range incidents {
		// Check if incident overlaps with this day
		incidentStart := incident.StartedAt
		incidentEnd := time.Now()
		if incident.ResolvedAt != nil {
			incidentEnd = *incident.ResolvedAt
		}

		// If incident overlaps with the day
		if incidentStart.Before(dayEnd) && incidentEnd.After(dayStart) {
			// Calculate incident duration for this day
			overlapStart := incidentStart
			if overlapStart.Before(dayStart) {
				overlapStart = dayStart
			}
			overlapEnd := incidentEnd
			if overlapEnd.After(dayEnd) {
				overlapEnd = dayEnd
			}

			duration := overlapEnd.Sub(overlapStart)
			dayDuration := dayEnd.Sub(dayStart)

			// If incident covers more than 50% of the day, consider it major
			if duration > dayDuration/2 {
				hasMajorIncident = true
			} else {
				hasMinorIncident = true
			}
		}
	}

	if hasMajorIncident {
		return "down"
	}

	// Check monitoring activities for this day
	dayActivities := make([]*domain.MonitoringActivity, 0)
	for _, activity := range activities {
		if (activity.CreatedAt.After(dayStart) || activity.CreatedAt.Equal(dayStart)) &&
			activity.CreatedAt.Before(dayEnd) {
			dayActivities = append(dayActivities, activity)
		}
	}

	if len(dayActivities) == 0 {
		// No data for this day
		if hasMinorIncident {
			return "degraded"
		}
		return "up" // Assume up if no data
	}

	// Calculate success rate for the day
	totalChecks := len(dayActivities)
	successfulChecks := 0
	for _, activity := range dayActivities {
		if activity.Success {
			successfulChecks++
		}
	}

	successRate := (float64(successfulChecks) / float64(totalChecks)) * 100

	// Determine status based on success rate
	if successRate < 50.0 {
		return "down"
	} else if successRate < 95.0 || hasMinorIncident {
		return "degraded"
	}

	return "up"
}

// formatDuration formats a duration into a human-readable string.
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	return fmt.Sprintf("%dd %dh", days, hours)
}

// GetResourceDetailStatus fetches detailed status information for a single resource
func (s *StatusPageService) GetResourceDetailStatus(ctx context.Context, resourceID string) (*dto.ResourceDetailStatusData, error) {
	// Fetch the resource
	resource, err := s.resourceRepo.FindByID(ctx, resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch resource: %w", err)
	}

	// Get current status
	currentStatus := s.mapResourceStatus(resource.Status)

	// Get 90-day uptime history (reuse existing logic)
	uptimePercentage90Days, dailyStatus90Days, err := s.calculate90DayData(ctx, resource)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate 90-day data: %w", err)
	}

	// Calculate uptime summary for different time windows
	uptimeSummary, err := s.calculateUptimeSummary(ctx, resource)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate uptime summary: %w", err)
	}
	uptimeSummary.Last90Days = uptimePercentage90Days

	// Calculate response time summary for last 7 days
	responseTimeSummary, err := s.calculateResponseTimeSummary7Days(ctx, resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate response time summary: %w", err)
	}

	// Get recent events (incidents and resolutions)
	recentEvents, err := s.buildRecentEvents(ctx, resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to build recent events: %w", err)
	}

	// Determine last updated time
	lastUpdated := time.Now()
	if resource.LastChecked != nil {
		lastUpdated = *resource.LastChecked
	}

	detail := &dto.ResourceDetailStatusData{
		ID:                    resource.ID,
		Name:                  resource.Name,
		CurrentStatus:         currentStatus,
		LastUpdated:           lastUpdated,
		UptimeHistory90Days:   dailyStatus90Days,
		UptimeSummary:         uptimeSummary,
		ResponseTimeSummary7D: responseTimeSummary,
		RecentEvents:          recentEvents,
	}

	// Attach maintenance banner information: active has priority, else upcoming scheduled
	// Active or currently within window
	// Maintenance repo may be nil in some test contexts
	now := time.Now()
	if s.maintenanceRepo != nil {
		activeMaintenances, err := s.maintenanceRepo.FindActiveForResource(ctx, resourceID, now)
		if err == nil && len(activeMaintenances) > 0 {
			m := activeMaintenances[0]
			detail.Maintenance = &dto.MaintenanceBanner{
				Status:   "active",
				Title:    m.Title,
				StartAt:  m.StartAt,
				EndAt:    m.EndAt,
				Timezone: m.Timezone,
			}
			return detail, nil
		}

		// Upcoming scheduled: pick the nearest future StartAt
		scheduled, err := s.maintenanceRepo.List(ctx, "scheduled", 100, 0)
		if err == nil && len(scheduled) > 0 {
			var candidate *domain.Maintenance
			for _, m := range scheduled {
				if m.StartAt == nil || m.StartAt.Before(now) {
					continue
				}
				// ensure the maintenance applies to this resource
				applies := false
				for _, r := range m.Resources {
					if r.ID == resourceID {
						applies = true
						break
					}
				}
				if !applies {
					continue
				}
				if candidate == nil || m.StartAt.Before(*candidate.StartAt) {
					candidate = m
				}
			}
			if candidate != nil {
				detail.Maintenance = &dto.MaintenanceBanner{
					Status:   "scheduled",
					Title:    candidate.Title,
					StartAt:  candidate.StartAt,
					EndAt:    candidate.EndAt,
					Timezone: candidate.Timezone,
				}
			}
		}
	}

	return detail, nil
}

// calculateUptimeSummary calculates uptime percentages for different time windows
func (s *StatusPageService) calculateUptimeSummary(ctx context.Context, resource *domain.Resource) (dto.UptimeSummary, error) {
	now := time.Now()
	resourceCreatedAt := resource.CreatedAt

	// Helper function to calculate uptime for a specific time window
	calculateWindowUptime := func(hours int) (float64, error) {
		windowStart := now.Add(-time.Duration(hours) * time.Hour)

		// If window starts before resource was created, adjust to creation time
		if windowStart.Before(resourceCreatedAt) {
			windowStart = resourceCreatedAt
		}

		// Fetch activities for this window
		activities, err := s.monitoringActivityRepo.FindByResourceID(ctx, resource.ID, 100000, 0)
		if err != nil {
			return 0, err
		}

		// Filter activities within the window
		var totalChecks, successfulChecks int
		for _, activity := range activities {
			if activity.CreatedAt.After(windowStart) && activity.CreatedAt.Before(now) {
				totalChecks++
				if activity.Success {
					successfulChecks++
				}
			}
		}

		if totalChecks == 0 {
			return 100.0, nil // Assume up if no data
		}

		return (float64(successfulChecks) / float64(totalChecks)) * 100, nil
	}

	uptime24h, err := calculateWindowUptime(24)
	if err != nil {
		return dto.UptimeSummary{}, err
	}

	uptime7d, err := calculateWindowUptime(24 * 7)
	if err != nil {
		return dto.UptimeSummary{}, err
	}

	uptime30d, err := calculateWindowUptime(24 * 30)
	if err != nil {
		return dto.UptimeSummary{}, err
	}

	return dto.UptimeSummary{
		Last24Hours: uptime24h,
		Last7Days:   uptime7d,
		Last30Days:  uptime30d,
		// Last90Days will be set by the caller
	}, nil
}

// calculateResponseTimeSummary7Days calculates response time statistics for the last 7 days
func (s *StatusPageService) calculateResponseTimeSummary7Days(ctx context.Context, resourceID string) (dto.ResponseTimeSummary, error) {
	now := time.Now()
	sevenDaysAgo := now.AddDate(0, 0, -7)

	// Fetch activities for the last 7 days
	activities, err := s.monitoringActivityRepo.FindByResourceID(ctx, resourceID, 100000, 0)
	if err != nil {
		return dto.ResponseTimeSummary{}, err
	}

	// Filter successful activities from last 7 days and calculate stats
	var responseTimes []int
	for _, activity := range activities {
		if activity.Success && activity.CreatedAt.After(sevenDaysAgo) && activity.CreatedAt.Before(now) {
			responseTimes = append(responseTimes, activity.ResponseTime)
		}
	}

	// If no data, try to get the latest successful response time
	if len(responseTimes) == 0 {
		for _, activity := range activities {
			if activity.Success && activity.ResponseTime > 0 {
				return dto.ResponseTimeSummary{
					AvgMs: activity.ResponseTime,
					MinMs: activity.ResponseTime,
					MaxMs: activity.ResponseTime,
				}, nil
			}
		}
		// No successful checks at all
		return dto.ResponseTimeSummary{
			AvgMs: 0,
			MinMs: 0,
			MaxMs: 0,
		}, nil
	}

	// Calculate min, max, and average
	minMs := responseTimes[0]
	maxMs := responseTimes[0]
	sum := 0

	for _, rt := range responseTimes {
		if rt < minMs {
			minMs = rt
		}
		if rt > maxMs {
			maxMs = rt
		}
		sum += rt
	}

	avgMs := sum / len(responseTimes)

	return dto.ResponseTimeSummary{
		AvgMs: avgMs,
		MinMs: minMs,
		MaxMs: maxMs,
	}, nil
}

// buildRecentEvents builds a list of recent up/down events from incidents
func (s *StatusPageService) buildRecentEvents(ctx context.Context, resourceID string) ([]dto.ResourceEvent, error) {
	// Fetch recent incidents for the resource (last 20)
	incidents, err := s.incidentRepo.FindByResource(ctx, resourceID, 20, 0)
	if err != nil {
		return nil, err
	}

	events := make([]dto.ResourceEvent, 0)

	for _, incident := range incidents {
		// Add "down" event
		downEvent := dto.ResourceEvent{
			Type:      "down",
			Timestamp: incident.StartedAt,
			Reason:    incident.Cause,
		}

		// Calculate duration if resolved
		if incident.ResolvedAt != nil {
			duration := formatDuration(incident.ResolvedAt.Sub(incident.StartedAt))
			downEvent.Duration = &duration
		} else {
			// Ongoing incident
			duration := formatDuration(time.Since(incident.StartedAt))
			downEvent.Duration = &duration
		}

		// Add details if available
		if len(incident.Details) > 0 {
			detailsStr := string(incident.Details)
			downEvent.Details = &detailsStr
		}

		events = append(events, downEvent)

		// Add "up" event if resolved
		if incident.ResolvedAt != nil {
			upEvent := dto.ResourceEvent{
				Type:      "up",
				Timestamp: *incident.ResolvedAt,
				Reason:    "Running again",
				Duration:  nil,
				Details:   nil,
			}
			events = append(events, upEvent)
		}
	}

	// Sort events by timestamp descending (most recent first)
	// Events are already roughly sorted since incidents are sorted by started_at DESC
	// But we need to interleave down and up events properly
	for i := 0; i < len(events)-1; i++ {
		for j := i + 1; j < len(events); j++ {
			if events[i].Timestamp.Before(events[j].Timestamp) {
				events[i], events[j] = events[j], events[i]
			}
		}
	}

	// Limit to 20 most recent events
	if len(events) > 20 {
		events = events[:20]
	}

	return events, nil
}
