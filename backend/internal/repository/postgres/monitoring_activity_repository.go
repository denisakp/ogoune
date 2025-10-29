package postgres

import (
	"context"
	"fmt"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/repository"
	"gorm.io/gorm"
)

// MonitoringActivityRepositoryImpl implements the MonitoringActivityRepository interface using GORM.
type MonitoringActivityRepositoryImpl struct {
	db *gorm.DB
}

// NewMonitoringActivityRepository creates a new MonitoringActivityRepository using GORM.
func NewMonitoringActivityRepository(db *gorm.DB) repository.MonitoringActivityRepository {
	return &MonitoringActivityRepositoryImpl{db: db}
}

// Create persists a new monitoring activity record to the database.
func (r *MonitoringActivityRepositoryImpl) Create(ctx context.Context, activity *domain.MonitoringActivity) error {
	if activity == nil {
		return repository.ErrInvalidInput
	}
	return r.db.WithContext(ctx).Create(activity).Error
}

// List retrieves all monitoring activities with pagination, ordered by creation time descending.
func (r *MonitoringActivityRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*domain.MonitoringActivity, error) {
	var activities []*domain.MonitoringActivity
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Preload("Resource").
		Find(&activities).Error

	if err != nil {
		return nil, fmt.Errorf("failed to list monitoring activities: %w", err)
	}
	return activities, nil
}

// FindByResourceID retrieves all monitoring activities for a specific resource with pagination.
func (r *MonitoringActivityRepositoryImpl) FindByResourceID(ctx context.Context, resourceID string, limit, offset int) ([]*domain.MonitoringActivity, error) {
	if resourceID == "" {
		return nil, repository.ErrInvalidInput
	}

	var activities []*domain.MonitoringActivity
	err := r.db.WithContext(ctx).
		Where("resource_id = ?", resourceID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Preload("Resource").
		Find(&activities).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find monitoring activities for resource %s: %w", resourceID, err)
	}
	return activities, nil
}

// GetUptimeStats retrieves the hourly uptime percentage for a resource over the last 24 hours.
func (r *MonitoringActivityRepositoryImpl) GetUptimeStats(ctx context.Context, resourceID string) ([]domain.UptimeStat, error) {
	if resourceID == "" {
		return nil, repository.ErrInvalidInput
	}

	var stats []domain.UptimeStat

	// SQL query to calculate hourly uptime percentage for the last 24 hours
	query := `
		SELECT
			DATE_TRUNC('hour', created_at) as hour,
			ROUND((COUNT(CASE WHEN success = true THEN 1 END)::numeric / COUNT(*)::numeric * 100), 2) as uptime_percent,
			COUNT(CASE WHEN success = true THEN 1 END) as successful_count,
			COUNT(*) as total_count
		FROM monitoring_activities
		WHERE resource_id = ?
			AND created_at >= NOW() - INTERVAL '24 hours'
		GROUP BY DATE_TRUNC('hour', created_at)
		ORDER BY hour ASC
	`

	err := r.db.WithContext(ctx).Raw(query, resourceID).Scan(&stats).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get uptime stats for resource %s: %w", resourceID, err)
	}

	return stats, nil
}

// GetRecentResponseTimes retrieves the most recent response times for a resource.
func (r *MonitoringActivityRepositoryImpl) GetRecentResponseTimes(ctx context.Context, resourceID string, limit int) ([]domain.ResponseTimePoint, error) {
	if resourceID == "" {
		return nil, repository.ErrInvalidInput
	}

	if limit <= 0 {
		limit = 50 // default limit
	}

	var responsePoints []domain.ResponseTimePoint

	query := `
		SELECT
			created_at as timestamp,
			response_time
		FROM monitoring_activities
		WHERE resource_id = ?
			AND success = true
		ORDER BY created_at DESC
		LIMIT ?
	`

	err := r.db.WithContext(ctx).Raw(query, resourceID, limit).Scan(&responsePoints).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get recent response times for resource %s: %w", resourceID, err)
	}

	// Reverse the slice to have chronological order (oldest to newest)
	for i, j := 0, len(responsePoints)-1; i < j; i, j = i+1, j-1 {
		responsePoints[i], responsePoints[j] = responsePoints[j], responsePoints[i]
	}

	return responsePoints, nil
}
