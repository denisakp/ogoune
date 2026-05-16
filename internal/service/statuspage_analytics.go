package service

import (
	"context"
	"fmt"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/dto"
)

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
