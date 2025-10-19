package service

import (
	"context"
	"fmt"

	"github.com/denisakp/pulseguard/internal/domain"
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
