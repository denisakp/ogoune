package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository"
	"gorm.io/gorm"
)

type NotificationRepositoryImpl struct {
	db *gorm.DB
}

// NewNotificationRepository creates a new NotificationRepository using GORM
func NewNotificationRepository(db *gorm.DB) port.NotificationRepository {
	return &NotificationRepositoryImpl{db: db}
}

// Create persists a new notification event record to the database.
func (r *NotificationRepositoryImpl) Create(ctx context.Context, n *domain.NotificationEvent) error {
	if n == nil {
		return repository.ErrInvalidInput
	}
	if n.Status == "" {
		n.Status = domain.NotificationEventStatusPending
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
	err := r.db.WithContext(ctx).
		Where("status = ?", domain.NotificationEventStatusPending).
		Where("type IN ?", []domain.NotificationEventType{domain.NotificationEventTypeDown, domain.NotificationEventTypeUp}).
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find pending notifications: %w", err)
	}
	return notifications, nil
}

// ClaimPending atomically acquires ownership of a pending notification event.
func (r *NotificationRepositoryImpl) ClaimPending(ctx context.Context, id, claimOwner string, claimedAt time.Time) (bool, error) {
	if id == "" || claimOwner == "" {
		return false, repository.ErrInvalidInput
	}

	result := r.db.WithContext(ctx).
		Model(&domain.NotificationEvent{}).
		Where("id = ?", id).
		Where("status = ?", domain.NotificationEventStatusPending).
		Where("claim_owner IS NULL OR claim_owner = ''").
		Updates(map[string]any{
			"claim_owner": claimOwner,
			"claimed_at":  claimedAt,
		})

	if result.Error != nil {
		return false, fmt.Errorf("failed to claim notification event %s: %w", id, result.Error)
	}

	return result.RowsAffected == 1, nil
}

// MarkAsSent updates the status of a notification event to sent.
func (r *NotificationRepositoryImpl) MarkAsSent(ctx context.Context, id string, processedAt time.Time) error {
	return r.markTerminal(ctx, id, domain.NotificationEventStatusSent, "", processedAt)
}

// MarkAsFailed updates the status of a notification event to failed.
func (r *NotificationRepositoryImpl) MarkAsFailed(ctx context.Context, id, lastError string, processedAt time.Time) error {
	return r.markTerminal(ctx, id, domain.NotificationEventStatusFailed, lastError, processedAt)
}

// MarkAsExpired updates the status of a notification event to expired.
func (r *NotificationRepositoryImpl) MarkAsExpired(ctx context.Context, id, lastError string, processedAt time.Time) error {
	return r.markTerminal(ctx, id, domain.NotificationEventStatusExpired, lastError, processedAt)
}

func (r *NotificationRepositoryImpl) markTerminal(ctx context.Context, id string, status domain.NotificationEventStatusType, lastError string, processedAt time.Time) error {
	if id == "" {
		return repository.ErrInvalidInput
	}

	updates := map[string]any{
		"status":       status,
		"processed_at": processedAt,
		"last_error":   lastError,
		"claim_owner":  nil,
		"claimed_at":   nil,
	}

	result := r.db.WithContext(ctx).Model(&domain.NotificationEvent{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update notification event %s terminal state: %w", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}

	return nil
}
