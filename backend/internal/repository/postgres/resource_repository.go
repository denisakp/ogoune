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

// Create persists a new resource record to the database.
func (r *ResourceRepositoryImpl) Create(ctx context.Context, resource *domain.Resource) (*domain.Resource, error) {
	if err := r.db.WithContext(ctx).Create(resource).Error; err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}
	return resource, nil
}

// FindByID retrieves a resource by its ID.
func (r *ResourceRepositoryImpl) FindByID(ctx context.Context, id string) (*domain.Resource, error) {
	var resource domain.Resource
	err := r.db.WithContext(ctx).
		Preload("Tags").
		Preload("Component").
		First(&resource, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find resource by ID: %w", err)
	}
	return &resource, nil
}

// List retrieves all resources with pagination, ordered by creation time descending.
func (r *ResourceRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*domain.Resource, error) {
	var resources []*domain.Resource
	err := r.db.WithContext(ctx).
		Preload("Tags").
		Preload("Component").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&resources).Error

	if err != nil {
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}
	return resources, nil
}

// Update modifies an existing resource record in the database.
// It properly handles the many-to-many relationship with tags by replacing them.
func (r *ResourceRepositoryImpl) Update(ctx context.Context, resource *domain.Resource) error {
	// Use a transaction to ensure atomicity
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// First, update the resource fields (excluding associations)
		if err := tx.Model(resource).Updates(resource).Error; err != nil {
			return fmt.Errorf("failed to update resource fields: %w", err)
		}

		// Replace tags using Association API to properly handle many-to-many relationship
		if err := tx.Model(resource).Association("Tags").Replace(resource.Tags); err != nil {
			return fmt.Errorf("failed to update resource tags: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// Delete performs a soft delete by setting IsActive to false.
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

// FindActive retrieves all active resources with pagination, ordered by creation time descending.
func (r *ResourceRepositoryImpl) FindActive(ctx context.Context, limit, offset int) ([]*domain.Resource, error) {
	var resources []*domain.Resource
	err := r.db.WithContext(ctx).
		Preload("Tags").
		Preload("Component").
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

// FindScheduledResources retrieves all active resources (for scheduler startup loading).
// All active resources are assumed to be schedulable.
func (r *ResourceRepositoryImpl) FindScheduledResources(ctx context.Context) ([]*domain.Resource, error) {
	var resources []*domain.Resource
	err := r.db.WithContext(ctx).
		Preload("Tags").
		Preload("Component").
		Where("is_active = ?", true).
		Order("id ASC").
		Find(&resources).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find scheduled resources: %w", err)
	}
	return resources, nil
}

// FindByTag retrieves all resources associated with a specific tag name with pagination.
func (r *ResourceRepositoryImpl) FindByTag(ctx context.Context, tagName string, limit, offset int) ([]*domain.Resource, error) {
	var resources []*domain.Resource
	err := r.db.WithContext(ctx).
		Preload("Tags").
		Preload("Component").
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

// FindByComponentID returns resources assigned to a component.
func (r *ResourceRepositoryImpl) FindByComponentID(ctx context.Context, componentID string) ([]*domain.Resource, error) {
	var resources []*domain.Resource
	err := r.db.WithContext(ctx).
		Preload("Tags").
		Preload("Component").
		Where("component_id = ?", componentID).
		Order("created_at DESC").
		Find(&resources).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find resources by component: %w", err)
	}
	return resources, nil
}

// CountByComponentID returns how many resources are assigned to a component.
func (r *ResourceRepositoryImpl) CountByComponentID(ctx context.Context, componentID string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&domain.Resource{}).
		Where("component_id = ?", componentID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count resources for component: %w", err)
	}
	return count, nil
}

// UpdateMetadata updates only the metadata fields of a resource to avoid touching associations.
func (r *ResourceRepositoryImpl) UpdateMetadata(ctx context.Context, id string, metadata *domain.ResourceMetaData) error {
	if metadata == nil {
		return fmt.Errorf("metadata cannot be nil")
	}

	updates := map[string]interface{}{
		"ssl_expiration_date":    metadata.SSLExpirationDate,
		"ssl_issuer":             metadata.SSLIssuer,
		"domain_expiration_date": metadata.DomainExpirationDate,
		"domain_registrar":       metadata.DomainRegistrar,
	}

	result := r.db.WithContext(ctx).
		Model(&domain.Resource{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update resource metadata: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}
