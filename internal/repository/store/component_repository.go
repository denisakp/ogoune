package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/repository"
	"gorm.io/gorm"
)

// ComponentRepositoryImpl provides GORM-based implementation of ComponentRepository.
type ComponentRepositoryImpl struct {
	db *gorm.DB
}

// NewComponentRepository creates a new component repository backed by GORM.
func NewComponentRepository(db *gorm.DB) repository.ComponentRepository {
	return &ComponentRepositoryImpl{db: db}
}

// Create persists a component record.
func (r *ComponentRepositoryImpl) Create(ctx context.Context, c *domain.Component) (*domain.Component, error) {
	if err := r.db.WithContext(ctx).Create(c).Error; err != nil {
		return nil, fmt.Errorf("failed to create component: %w", err)
	}
	return c, nil
}

// FindByID fetches a component with its resources.
func (r *ComponentRepositoryImpl) FindByID(ctx context.Context, id string) (*domain.Component, error) {
	var component domain.Component
	if err := r.db.WithContext(ctx).
		Preload("Resources").
		First(&component, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find component: %w", err)
	}
	return &component, nil
}

// List returns paginated components with their resources.
func (r *ComponentRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*domain.Component, error) {
	var components []*domain.Component
	if err := r.db.WithContext(ctx).
		Preload("Resources").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&components).Error; err != nil {
		return nil, fmt.Errorf("failed to list components: %w", err)
	}
	return components, nil
}

// Update modifies component fields (name/description) and associations if provided.
func (r *ComponentRepositoryImpl) Update(ctx context.Context, c *domain.Component) error {
	if err := r.db.WithContext(ctx).Model(c).Updates(map[string]any{
		"name":        c.Name,
		"description": c.Description,
	}).Error; err != nil {
		return fmt.Errorf("failed to update component: %w", err)
	}
	return nil
}

// Delete removes a component record.
func (r *ComponentRepositoryImpl) Delete(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Delete(&domain.Component{}, "id = ?", id)
	if res.Error != nil {
		return fmt.Errorf("failed to delete component: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

// UpdateLastNotificationStatus persists the latest notified status for a component.
func (r *ComponentRepositoryImpl) UpdateLastNotificationStatus(ctx context.Context, id string, status domain.ComponentStatus) error {
	res := r.db.WithContext(ctx).
		Model(&domain.Component{}).
		Where("id = ?", id).
		Update("last_notification_status", status)
	if res.Error != nil {
		return fmt.Errorf("failed to update component notification status: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}
