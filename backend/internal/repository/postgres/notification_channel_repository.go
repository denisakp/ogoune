package postgres

import (
	"context"
	"errors"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/repository"
	"gorm.io/gorm"
)

// NotificationChannelRepository implements repository.NotificationChannelRepository using Postgres
type NotificationChannelRepository struct {
	db *gorm.DB
}

// NewNotificationChannelRepository creates a new Postgres-backed notification channel repository
func NewNotificationChannelRepository(db *gorm.DB) *NotificationChannelRepository {
	return &NotificationChannelRepository{db: db}
}

// Create inserts a new notification channel
func (r *NotificationChannelRepository) Create(ctx context.Context, channel *domain.NotificationChannel) error {
	if err := r.db.WithContext(ctx).Create(channel).Error; err != nil {
		return err
	}
	return nil
}

// FindByID retrieves a notification channel by ID
func (r *NotificationChannelRepository) FindByID(ctx context.Context, id string) (*domain.NotificationChannel, error) {
	var channel domain.NotificationChannel
	if err := r.db.WithContext(ctx).First(&channel, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &channel, nil
}

// List retrieves all notification channels with pagination
func (r *NotificationChannelRepository) List(ctx context.Context, limit, offset int) ([]*domain.NotificationChannel, error) {
	var channels []*domain.NotificationChannel
	if err := r.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&channels).Error; err != nil {
		return nil, err
	}
	return channels, nil
}

// Update updates an existing notification channel
func (r *NotificationChannelRepository) Update(ctx context.Context, channel *domain.NotificationChannel) error {
	result := r.db.WithContext(ctx).Save(channel)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

// Delete removes a notification channel by ID
func (r *NotificationChannelRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&domain.NotificationChannel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

// FindByType retrieves all notification channels of a specific type
func (r *NotificationChannelRepository) FindByType(ctx context.Context, channelType domain.NotificationChannelType) ([]*domain.NotificationChannel, error) {
	var channels []*domain.NotificationChannel
	if err := r.db.WithContext(ctx).
		Where("type = ?", channelType).
		Find(&channels).Error; err != nil {
		return nil, err
	}
	return channels, nil
}

// FindDefaultChannels retrieves all channels marked as enabled by default
func (r *NotificationChannelRepository) FindDefaultChannels(ctx context.Context) ([]*domain.NotificationChannel, error) {
	var channels []*domain.NotificationChannel
	if err := r.db.WithContext(ctx).
		Where("enabled_by_default = ?", true).
		Find(&channels).Error; err != nil {
		return nil, err
	}
	return channels, nil
}
