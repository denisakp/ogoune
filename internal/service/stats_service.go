package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/denisakp/ogoune/internal/dto"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository"
)

// StatsService handles business logic for aggregated statistics.
type StatsService struct {
	monitoringActivity port.MonitoringActivityRepository
	incidents          port.IncidentRepository
}

// NewStatsService creates a new stats service.
func NewStatsService(
	monitoringActivity port.MonitoringActivityRepository,
	incidents port.IncidentRepository,
) *StatsService {
	return &StatsService{
		monitoringActivity: monitoringActivity,
		incidents:          incidents,
	}
}

// GetSummary retrieves aggregated uptime and incident statistics for all monitored resources
// within the specified time range.
func (s *StatsService) GetSummary(ctx context.Context, timeRange string) (*dto.StatsSummaryResponse, error) {
	// Parse time range to hours
	hours, err := parseTimeRange(timeRange)
	if err != nil {
		return nil, fmt.Errorf("invalid time range: %w", err)
	}

	// Get global uptime stats
	overallUptime, err := s.monitoringActivity.GetGlobalUptimeStats(ctx, hours)
	if err != nil {
		return nil, fmt.Errorf("failed to get global uptime stats: %w", err)
	}

	// Get incident statistics
	totalIncidents, affectedMonitors, err := s.incidents.GetIncidentStats(ctx, hours)
	if err != nil {
		return nil, fmt.Errorf("failed to get incident stats: %w", err)
	}

	// Calculate duration since last resolved incident
	hasActive, err := s.incidents.HasActiveIncident(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check for active incident: %w", err)
	}

	var withoutIncidentsDuration string
	if hasActive {
		withoutIncidentsDuration = "0m"
	} else {
		lastResolved, err := s.incidents.FindLastResolved(ctx)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				withoutIncidentsDuration = "∞"
			} else {
				return nil, fmt.Errorf("failed to find last resolved incident: %w", err)
			}
		} else {
			elapsed := int64(time.Since(*lastResolved.ResolvedAt).Seconds())
			withoutIncidentsDuration = formatDurationFromSeconds(elapsed)
		}
	}

	return &dto.StatsSummaryResponse{
		Range:                    timeRange,
		OverallUptime:            overallUptime,
		Incidents:                totalIncidents,
		WithoutIncidentsDuration: withoutIncidentsDuration,
		AffectedMonitors:         affectedMonitors,
	}, nil
}

// parseTimeRange converts time range string to hours.
// Supported formats: 2h, 24h, 7d, 30d
func parseTimeRange(timeRange string) (int, error) {
	switch timeRange {
	case "2h":
		return 2, nil
	case "24h":
		return 24, nil
	case "7d":
		return 24 * 7, nil // 168 hours
	case "30d":
		return 24 * 30, nil // 720 hours
	default:
		return 0, fmt.Errorf("unsupported time range '%s', must be one of: 2h, 24h, 7d, 30d", timeRange)
	}
}

// formatDurationFromSeconds formats seconds into a human-readable duration string.
func formatDurationFromSeconds(seconds int64) string {
	if seconds == 0 {
		return "0m"
	}

	duration := time.Duration(seconds) * time.Second
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60

	if hours > 0 {
		if minutes > 0 {
			return fmt.Sprintf("%dh %dm", hours, minutes)
		}
		return fmt.Sprintf("%dh", hours)
	}

	return fmt.Sprintf("%dm", minutes)
}
