package service

import (
	"context"
	"fmt"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/dto"
	"github.com/denisakp/pulseguard/internal/repository"
)

// MonitoringActivityService handles business logic for monitoring activities.
type MonitoringActivityService struct {
	repo repository.MonitoringActivityRepository
}

// NewMonitoringActivityService creates a new monitoring activity service.
func NewMonitoringActivityService(repo repository.MonitoringActivityRepository) *MonitoringActivityService {
	return &MonitoringActivityService{repo: repo}
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
