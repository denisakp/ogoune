package postgres

import (
	"context"
	"fmt"

	domain "github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/repository"
	"gorm.io/gorm"
)

type TagsRepositoryImpl struct {
	db *gorm.DB
}

// NewTagsRepository creates a new TagsRepository using GORM
func NewTagsRepository(db *gorm.DB) repository.TagsRepository {
	return &TagsRepositoryImpl{db: db}
}

func (r *TagsRepositoryImpl) Create(ctx context.Context, t *domain.Tags) error {
	if err := r.db.WithContext(ctx).Create(t).Error; err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}
	return nil
}

func (r *TagsRepositoryImpl) FindByID(ctx context.Context, id string) (*domain.Tags, error) {
	var tag domain.Tags
	err := r.db.WithContext(ctx).First(&tag, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find tag by ID: %w", err)
	}
	return &tag, nil
}

func (r *TagsRepositoryImpl) FindByName(ctx context.Context, name string) (*domain.Tags, error) {
	var tag domain.Tags
	err := r.db.WithContext(ctx).First(&tag, "name = ?", name).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find tag by name: %w", err)
	}
	return &tag, nil
}

func (r *TagsRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*domain.Tags, error) {
	var tags []*domain.Tags
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&tags).Error

	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}
	return tags, nil
}

func (r *TagsRepositoryImpl) Update(ctx context.Context, t *domain.Tags) error {
	result := r.db.WithContext(ctx).Save(t)
	if result.Error != nil {
		return fmt.Errorf("failed to update tag: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *TagsRepositoryImpl) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&domain.Tags{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete tag: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}
