package service

import (
	"context"
	"fmt"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/dto"
	"github.com/denisakp/ogoune/internal/port"
)

// MonitoringActivityService handles business logic for monitoring activities.
type MonitoringActivityService struct {
	repo    port.MonitoringActivityRepository
	aggRepo port.UptimeDailyAggRepository
}

// NewMonitoringActivityService creates a new monitoring activity service.
// aggRepo may be nil; when nil, the 30d Uptime falls back to the per-hour
// monitoring_activities path (legacy behavior, kept for tests).
func NewMonitoringActivityService(repo port.MonitoringActivityRepository, aggRepo port.UptimeDailyAggRepository) *MonitoringActivityService {
	return &MonitoringActivityService{repo: repo, aggRepo: aggRepo}
}

// ListAll retrieves all monitoring activities with pagination.
func (s *MonitoringActivityService) ListAll(ctx context.Context, limit, offset int) ([]*domain.MonitoringActivity, error) {
	// Set default limit if not provided
	if limit <= 0 {
		limit = 50
	}

	// Validate offset
	if offset < 0 {
		return nil, fmt.Errorf("offset must be non-negative")
	}

	activities, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list monitoring activities: %w", err)
	}

	return activities, nil
}

// ListByResourceID retrieves monitoring activities for a specific resource with pagination.
func (s *MonitoringActivityService) ListByResourceID(ctx context.Context, resourceID string, limit, offset int) ([]*domain.MonitoringActivity, error) {
	// Set default limit if not provided
	if limit <= 0 {
		limit = 50
	}

	// Validate offset
	if offset < 0 {
		return nil, fmt.Errorf("offset must be non-negative")
	}

	// Validate resource ID
	if resourceID == "" {
		return nil, fmt.Errorf("resource_id is required")
	}

	activities, err := s.repo.FindByResourceID(ctx, resourceID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list monitoring activities for resource %s: %w", resourceID, err)
	}

	return activities, nil
}

// GetUptimeStats retrieves hourly uptime statistics for a resource over the last 24 hours.
func (s *MonitoringActivityService) GetUptimeStats(ctx context.Context, resourceID string) ([]dto.UptimeStatResponse, error) {
	// Validate resource ID
	if resourceID == "" {
		return nil, fmt.Errorf("resource_id is required")
	}

	// Get uptime stats from repository
	stats, err := s.repo.GetUptimeStats(ctx, resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get uptime stats for resource %s: %w", resourceID, err)
	}

	// Map domain objects to DTOs
	response := make([]dto.UptimeStatResponse, len(stats))
	for i, stat := range stats {
		response[i] = dto.UptimeStatResponse{
			Hour:            stat.Hour,
			UptimePercent:   stat.UptimePercent,
			SuccessfulCount: stat.SuccessfulCount,
			TotalCount:      stat.TotalCount,
		}
	}

	return response, nil
}

// GetRecentResponseTimes retrieves the most recent response times for a resource.
func (s *MonitoringActivityService) GetRecentResponseTimes(ctx context.Context, resourceID string, limit int) ([]dto.ResponseTimePoint, error) {
	// Validate resource ID
	if resourceID == "" {
		return nil, fmt.Errorf("resource_id is required")
	}

	// Set default limit if not provided
	if limit <= 0 {
		limit = 50
	}

	// Get response times from repository
	points, err := s.repo.GetRecentResponseTimes(ctx, resourceID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent response times for resource %s: %w", resourceID, err)
	}

	// Map domain objects to DTOs
	response := make([]dto.ResponseTimePoint, len(points))
	for i, point := range points {
		response[i] = dto.ResponseTimePoint{
			Timestamp:    point.Timestamp,
			ResponseTime: point.ResponseTime,
		}
	}

	return response, nil
}

// GetResourceUptimeStats computes uptime percentages and average response time for the four
// standard time windows: 2h, 24h, 7d (168h), 30d (720h). Fields are nil when no data exists.
func (s *MonitoringActivityService) GetResourceUptimeStats(ctx context.Context, resourceID string) (*dto.LiveStats, error) {
	if resourceID == "" {
		return nil, fmt.Errorf("resource_id is required")
	}

	stats := &dto.LiveStats{}

	if v, err := s.repo.GetUptimeByWindow(ctx, resourceID, 2); err == nil {
		stats.Uptime2h = v
	}
	if v, err := s.repo.GetUptimeByWindow(ctx, resourceID, 24); err == nil {
		stats.Uptime24h = v
	}
	if v, err := s.repo.GetUptimeByWindow(ctx, resourceID, 168); err == nil {
		stats.Uptime7d = v
	}
	if v, err := s.uptime30dFromAgg(ctx, resourceID); err == nil && v != nil {
		stats.Uptime30d = v
	} else if v, err := s.repo.GetUptimeByWindow(ctx, resourceID, 720); err == nil {
		stats.Uptime30d = v
	}
	if v, err := s.repo.GetAvgResponseTimeByWindow(ctx, resourceID, 24); err == nil {
		stats.AvgResponseTime24h = v
	}

	return stats, nil
}

// uptime30dFromAgg computes the 30-day uptime ratio from uptime_daily_agg,
// matching the list view's enrichment so both surfaces show the same number.
// Returns (nil, nil) when no aggregate rows exist (resource has not yet been
// rolled up — caller can fall back to the activity-based path).
func (s *MonitoringActivityService) uptime30dFromAgg(ctx context.Context, resourceID string) (*float64, error) {
	if s.aggRepo == nil {
		return nil, nil
	}
	now := time.Now().UTC()
	from := now.AddDate(0, 0, -29)
	rows, err := s.aggRepo.FindForResource(ctx, resourceID, from, now)
	if err != nil {
		return nil, err
	}
	var up, samples int
	for _, r := range rows {
		if r == nil {
			continue
		}
		up += r.Up
		samples += r.Samples
	}
	if samples == 0 {
		return nil, nil
	}
	ratio := float64(up) / float64(samples)
	return &ratio, nil
}
