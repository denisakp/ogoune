package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/repository"
	"gorm.io/gorm"
)

// ResourceRepositoryImpl provides GORM-based implementation of ResourceRepository
type ResourceRepositoryImpl struct {
	db *gorm.DB
}

// NewResourceRepository creates a new ResourceRepository using GORM
func NewResourceRepository(db *gorm.DB) repository.ResourceRepository {
	return &ResourceRepositoryImpl{db: db}
}

func (r *ResourceRepositoryImpl) Create(ctx context.Context, resource *domain.Resource) error {
	if err := r.db.WithContext(ctx).Create(resource).Error; err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}
	return nil
}

func (r *ResourceRepositoryImpl) FindByID(ctx context.Context, id string) (*domain.Resource, error) {
	var resource domain.Resource
	err := r.db.WithContext(ctx).Preload("Tags").First(&resource, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find resource by ID: %w", err)
	}
	return &resource, nil
}

func (r *ResourceRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*domain.Resource, error) {
	var resources []*domain.Resource
	err := r.db.WithContext(ctx).
		Preload("Tags").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&resources).Error

	if err != nil {
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}
	return resources, nil
}

func (r *ResourceRepositoryImpl) Update(ctx context.Context, resource *domain.Resource) error {
	result := r.db.WithContext(ctx).Save(resource)
	if result.Error != nil {
		return fmt.Errorf("failed to update resource: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *ResourceRepositoryImpl) Delete(ctx context.Context, id string) error {
	// Soft delete: set IsActive to false
	result := r.db.WithContext(ctx).
		Model(&domain.Resource{}).
		Where("id = ?", id).
		Update("is_active", false)

	if result.Error != nil {
		return fmt.Errorf("failed to delete resource: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *ResourceRepositoryImpl) FindActive(ctx context.Context, limit, offset int) ([]*domain.Resource, error) {
	var resources []*domain.Resource
	err := r.db.WithContext(ctx).
		Preload("Tags").
		Where("is_active = ?", true).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&resources).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find active resources: %w", err)
	}
	return resources, nil
}

func (r *ResourceRepositoryImpl) FindByTag(ctx context.Context, tagName string, limit, offset int) ([]*domain.Resource, error) {
	var resources []*domain.Resource
	err := r.db.WithContext(ctx).
		Preload("Tags").
		Joins("JOIN resource_tags ON resources.id = resource_tags.resource_id").
		Joins("JOIN tags ON resource_tags.tag_id = tags.id").
		Where("tags.name = ? AND resources.is_active = ?", tagName, true).
		Order("resources.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&resources).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find resources by tag: %w", err)
	}
	return resources, nil
}