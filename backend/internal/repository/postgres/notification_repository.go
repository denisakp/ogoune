package postgres

import (
	"context"
	"errors"
	"fmt"

	domain "github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/repository"
	"gorm.io/gorm"
)

type NotificationRepositoryImpl struct {
	db *gorm.DB
}

// NewNotificationRepository creates a new NotificationRepository using GORM
func NewNotificationRepository(db *gorm.DB) repository.NotificationRepository {
	return &NotificationRepositoryImpl{db: db}
}

// Create persists a new notification event record to the database.
func (r *NotificationRepositoryImpl) Create(ctx context.Context, n *domain.NotificationEvent) error {
	if n == nil {
		return repository.ErrInvalidInput
	}
	return r.db.WithContext(ctx).Create(n).Error
}

// FindByID retrieves a notification event by its ID.
func (r *NotificationRepositoryImpl) FindByID(ctx context.Context, id string) (*domain.NotificationEvent, error) {
	if id == "" {
		return nil, repository.ErrInvalidInput
	}

	var notification domain.NotificationEvent
	err := r.db.WithContext(ctx).First(&notification, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find notification: %w", err)
	}
	return &notification, nil
}

// List retrieves all notification events with pagination.
func (r *NotificationRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*domain.NotificationEvent, error) {
	var notifications []*domain.NotificationEvent
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&notifications).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list notifications: %w", err)
	}
	return notifications, nil
}

// Update modifies an existing notification event record in the database.
func (r *NotificationRepositoryImpl) Update(ctx context.Context, n *domain.NotificationEvent) error {
	if n == nil || n.ID == "" {
		return repository.ErrInvalidInput
	}
	return r.db.WithContext(ctx).Save(n).Error
}

// Delete removes a notification event record from the database by its ID.
func (r *NotificationRepositoryImpl) Delete(ctx context.Context, id string) error {
	if id == "" {
		return repository.ErrInvalidInput
	}
	return r.db.WithContext(ctx).Delete(&domain.NotificationEvent{}, "id = ?", id).Error
}

// FindPending retrieves all pending notification events with pagination.
func (r *NotificationRepositoryImpl) FindPending(ctx context.Context, limit, offset int) ([]*domain.NotificationEvent, error) {
	var notifications []*domain.NotificationEvent
	// Note: This assumes there's a status field on NotificationEvent
	// For now, just return all notifications
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&notifications).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find pending notifications: %w", err)
	}
	return notifications, nil
}

// MarkAsSent updates the status of a notification event to sent.
func (r *NotificationRepositoryImpl) MarkAsSent(ctx context.Context, id string) error {
	if id == "" {
		return repository.ErrInvalidInput
	}
	// This is a placeholder implementation
	// In a real implementation, you'd update a status field
	return nil
}
