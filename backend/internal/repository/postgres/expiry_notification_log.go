package postgres

import (
	"context"
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
	"gorm.io/gorm"
)

// ExpiryNotificationLogRepository implements repository.ExpiryNotificationLogRepository using GORM.
type ExpiryNotificationLogRepository struct {
	db *gorm.DB
}

// NewExpiryNotificationLogRepository creates a new repository for expiry notification logs.
func NewExpiryNotificationLogRepository(db *gorm.DB) *ExpiryNotificationLogRepository {
	return &ExpiryNotificationLogRepository{db: db}
}

// CountByKey returns how many log entries exist for the given resource/type/threshold combination.
func (r *ExpiryNotificationLogRepository) CountByKey(ctx context.Context, resourceID, expiryType string, threshold int) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.ExpiryNotificationLog{}).
		Where("resource_id = ? AND expiry_type = ? AND threshold = ?", resourceID, expiryType, threshold).
		Count(&count).Error
	return count, err
}

// Create persists a new expiry notification log entry.
func (r *ExpiryNotificationLogRepository) Create(ctx context.Context, log *domain.ExpiryNotificationLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// DeleteByResourceIDAndType removes all log entries for a given resource and expiry type.
// This is called when a certificate or domain renewal is detected (renewal reset).
func (r *ExpiryNotificationLogRepository) DeleteByResourceIDAndType(ctx context.Context, resourceID, expiryType string) error {
	return r.db.WithContext(ctx).
		Where("resource_id = ? AND expiry_type = ?", resourceID, expiryType).
		Delete(&domain.ExpiryNotificationLog{}).Error
}

// DeleteOlderThan removes log entries with sent_at older than the given cutoff time.
func (r *ExpiryNotificationLogRepository) DeleteOlderThan(ctx context.Context, cutoff time.Time) error {
	return r.db.WithContext(ctx).
		Where("sent_at < ?", cutoff).
		Delete(&domain.ExpiryNotificationLog{}).Error
}
