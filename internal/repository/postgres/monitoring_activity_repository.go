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
