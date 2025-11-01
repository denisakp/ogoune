package service

import (
	"context"
	"fmt"
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/dto"
	"github.com/denisakp/pulseguard/internal/repository"
)

// StatusPageService provides data aggregation for public status pages.
type StatusPageService struct {
	resourceRepo           repository.ResourceRepository
	incidentRepo           repository.IncidentRepository
	monitoringActivityRepo repository.MonitoringActivityRepository
}

// NewStatusPageService creates a new StatusPageService instance.
func NewStatusPageService(
	resourceRepo repository.ResourceRepository,
	incidentRepo repository.IncidentRepository,
	monitoringActivityRepo repository.MonitoringActivityRepository,
) *StatusPageService {
	return &StatusPageService{
		resourceRepo:           resourceRepo,
		incidentRepo:           incidentRepo,
		monitoringActivityRepo: monitoringActivityRepo,
	}
}

// GetData fetches and aggregates all data needed for the status page.
func (s *StatusPageService) GetData(ctx context.Context) (*dto.StatusPageData, error) {
	// Fetch all active resources
	resources, err := s.resourceRepo.FindActive(ctx, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch resources: %w", err)
	}

	// Build resource status info with 90-day data
	resourceInfos := make([]dto.ResourceStatusInfo, 0, len(resources))
	hasNonOperationalResource := false

	for _, resource := range resources {
		info, err := s.buildResourceStatusInfo(ctx, resource)
		if err != nil {
			// Log error but continue processing other resources
			continue
		}
		resourceInfos = append(resourceInfos, info)

		// Check if any resource is not "up"
		if info.CurrentStatus != "up" {
			hasNonOperationalResource = true
		}
	}

	// Calculate global status
	globalStatus := "all_systems_operational"
	if hasNonOperationalResource {
		globalStatus = "some_systems_down"
	}

	return &dto.StatusPageData{
		GlobalStatus: globalStatus,
		GeneratedAt:  time.Now(),
		Resources:    resourceInfos,
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
	for i := 0; i < 90; i++ {
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

	return &dto.ResourceDetailStatusData{
		ID:                    resource.ID,
		Name:                  resource.Name,
		CurrentStatus:         currentStatus,
		LastUpdated:           lastUpdated,
		UptimeHistory90Days:   dailyStatus90Days,
		UptimeSummary:         uptimeSummary,
		ResponseTimeSummary7D: responseTimeSummary,
		RecentEvents:          recentEvents,
	}, nil
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
