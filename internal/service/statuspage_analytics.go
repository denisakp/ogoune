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

	activities, err := s.monitoringActivityRepo.FindByResourceID(ctx, resource.ID, 100000, 0)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to fetch monitoring activities: %w", err)
	}

	filteredActivities := filterActivitiesSince(activities, startDate)

	incidents, err := s.incidentRepo.FindByResource(ctx, resource.ID, 10000, 0)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to fetch incidents: %w", err)
	}

	filteredIncidents := filterIncidentsOverlapping(incidents, startDate)

	uptimePercentage := calcUptimePercentage(filteredActivities)

	dailyStatus := make([]string, 90)
	for i := range 90 {
		dayStart := startDate.AddDate(0, 0, i)
		dayEnd := dayStart.AddDate(0, 0, 1)

		if dayStart.Before(resource.CreatedAt) {
			dailyStatus[i] = "no_data"
		} else {
			dailyStatus[i] = s.calculateDayStatus(dayStart, dayEnd, filteredActivities, filteredIncidents)
		}
	}

	return uptimePercentage, dailyStatus, nil
}

func filterActivitiesSince(activities []*domain.MonitoringActivity, since time.Time) []*domain.MonitoringActivity {
	filtered := make([]*domain.MonitoringActivity, 0, len(activities))
	for _, a := range activities {
		if !a.CreatedAt.Before(since) {
			filtered = append(filtered, a)
		}
	}
	return filtered
}

func filterIncidentsOverlapping(incidents []*domain.Incident, since time.Time) []*domain.Incident {
	filtered := make([]*domain.Incident, 0, len(incidents))
	for _, inc := range incidents {
		if inc.StartedAt.After(since) ||
			(inc.ResolvedAt != nil && inc.ResolvedAt.After(since)) ||
			inc.ResolvedAt == nil {
			filtered = append(filtered, inc)
		}
	}
	return filtered
}

func calcUptimePercentage(activities []*domain.MonitoringActivity) float64 {
	total := len(activities)
	if total == 0 {
		return 100.0
	}
	success := 0
	for _, a := range activities {
		if a.Success {
			success++
		}
	}
	return (float64(success) / float64(total)) * 100
}

// calculateDayStatus determines the status for a specific day based on activities and incidents
func (s *StatusPageService) calculateDayStatus(dayStart, dayEnd time.Time, activities []*domain.MonitoringActivity, incidents []*domain.Incident) string {
	hasMajor, hasMinor := classifyDayIncidents(dayStart, dayEnd, incidents)

	if hasMajor {
		return "down"
	}

	dayActivities := activitiesInWindow(activities, dayStart, dayEnd)

	if len(dayActivities) == 0 {
		if hasMinor {
			return "degraded"
		}
		return "up"
	}

	successRate := calcUptimePercentage(dayActivities)

	if successRate < 50.0 {
		return "down"
	}
	if successRate < 95.0 || hasMinor {
		return "degraded"
	}

	return "up"
}

func classifyDayIncidents(dayStart, dayEnd time.Time, incidents []*domain.Incident) (hasMajor, hasMinor bool) {
	dayDuration := dayEnd.Sub(dayStart)

	for _, incident := range incidents {
		incidentEnd := time.Now()
		if incident.ResolvedAt != nil {
			incidentEnd = *incident.ResolvedAt
		}

		if !incident.StartedAt.Before(dayEnd) || !incidentEnd.After(dayStart) {
			continue
		}

		overlapStart := incident.StartedAt
		if overlapStart.Before(dayStart) {
			overlapStart = dayStart
		}
		overlapEnd := incidentEnd
		if overlapEnd.After(dayEnd) {
			overlapEnd = dayEnd
		}

		if overlapEnd.Sub(overlapStart) > dayDuration/2 {
			hasMajor = true
		} else {
			hasMinor = true
		}
	}

	return
}

func activitiesInWindow(activities []*domain.MonitoringActivity, start, end time.Time) []*domain.MonitoringActivity {
	result := make([]*domain.MonitoringActivity, 0)
	for _, a := range activities {
		if !a.CreatedAt.Before(start) && a.CreatedAt.Before(end) {
			result = append(result, a)
		}
	}
	return result
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

	activities, err := s.monitoringActivityRepo.FindByResourceID(ctx, resource.ID, 100000, 0)
	if err != nil {
		return dto.UptimeSummary{}, err
	}

	windowUptime := func(hours int) float64 {
		windowStart := now.Add(-time.Duration(hours) * time.Hour)
		if windowStart.Before(resource.CreatedAt) {
			windowStart = resource.CreatedAt
		}
		windowActivities := activitiesInWindow(activities, windowStart, now)
		return calcUptimePercentage(windowActivities)
	}

	return dto.UptimeSummary{
		Last24Hours: windowUptime(24),
		Last7Days:   windowUptime(24 * 7),
		Last30Days:  windowUptime(24 * 30),
	}, nil
}

// calculateResponseTimeSummary7Days calculates response time statistics for the last 7 days
func (s *StatusPageService) calculateResponseTimeSummary7Days(ctx context.Context, resourceID string) (dto.ResponseTimeSummary, error) {
	now := time.Now()
	sevenDaysAgo := now.AddDate(0, 0, -7)

	activities, err := s.monitoringActivityRepo.FindByResourceID(ctx, resourceID, 100000, 0)
	if err != nil {
		return dto.ResponseTimeSummary{}, err
	}

	responseTimes := collectResponseTimes(activities, sevenDaysAgo, now)

	if len(responseTimes) == 0 {
		return fallbackResponseTime(activities), nil
	}

	return computeResponseTimeStats(responseTimes), nil
}

func collectResponseTimes(activities []*domain.MonitoringActivity, start, end time.Time) []int {
	var times []int
	for _, a := range activities {
		if a.Success && a.CreatedAt.After(start) && a.CreatedAt.Before(end) {
			times = append(times, a.ResponseTime)
		}
	}
	return times
}

func fallbackResponseTime(activities []*domain.MonitoringActivity) dto.ResponseTimeSummary {
	for _, a := range activities {
		if a.Success && a.ResponseTime > 0 {
			return dto.ResponseTimeSummary{AvgMs: a.ResponseTime, MinMs: a.ResponseTime, MaxMs: a.ResponseTime}
		}
	}
	return dto.ResponseTimeSummary{}
}

func computeResponseTimeStats(times []int) dto.ResponseTimeSummary {
	minMs, maxMs, sum := times[0], times[0], 0
	for _, rt := range times {
		if rt < minMs {
			minMs = rt
		}
		if rt > maxMs {
			maxMs = rt
		}
		sum += rt
	}
	return dto.ResponseTimeSummary{AvgMs: sum / len(times), MinMs: minMs, MaxMs: maxMs}
}

// buildRecentEvents builds a list of recent up/down events from incidents
func (s *StatusPageService) buildRecentEvents(ctx context.Context, resourceID string) ([]dto.ResourceEvent, error) {
	incidents, err := s.incidentRepo.FindByResource(ctx, resourceID, 20, 0)
	if err != nil {
		return nil, err
	}

	events := make([]dto.ResourceEvent, 0, len(incidents)*2)
	for _, incident := range incidents {
		events = append(events, incidentToEvents(incident)...)
	}

	sortEventsDescending(events)

	if len(events) > 20 {
		events = events[:20]
	}

	return events, nil
}

func incidentToEvents(incident *domain.Incident) []dto.ResourceEvent {
	downEvent := dto.ResourceEvent{
		Type:      "down",
		Timestamp: incident.StartedAt,
		Reason:    incident.Cause,
	}

	if incident.ResolvedAt != nil {
		d := formatDuration(incident.ResolvedAt.Sub(incident.StartedAt))
		downEvent.Duration = &d
	} else {
		d := formatDuration(time.Since(incident.StartedAt))
		downEvent.Duration = &d
	}

	if len(incident.Details) > 0 {
		s := string(incident.Details)
		downEvent.Details = &s
	}

	events := []dto.ResourceEvent{downEvent}

	if incident.ResolvedAt != nil {
		events = append(events, dto.ResourceEvent{
			Type:      "up",
			Timestamp: *incident.ResolvedAt,
			Reason:    "Running again",
		})
	}

	return events
}

func sortEventsDescending(events []dto.ResourceEvent) {
	for i := 0; i < len(events)-1; i++ {
		for j := i + 1; j < len(events); j++ {
			if events[i].Timestamp.Before(events[j].Timestamp) {
				events[i], events[j] = events[j], events[i]
			}
		}
	}
}
