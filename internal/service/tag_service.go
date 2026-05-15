package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository"
)

// TagService provides business logic for tag management operations.
type TagService struct {
	tags repository.TagsRepository
}

// NewTagService creates a new TagService with the given repository dependency.
func NewTagService(tags repository.TagsRepository) *TagService {
	return &TagService{
		tags: tags,
	}
}

// CreateTag creates a new tag in the system.
func (s *TagService) CreateTag(ctx context.Context, tag *domain.Tags) error {
	if tag == nil {
		return fmt.Errorf("%w: tag cannot be nil", ErrValidationFailed)
	}

	if tag.Name == "" {
		return fmt.Errorf("%w: tag name is required", ErrValidationFailed)
	}

	// Check if tag with same name already exists
	existing, err := s.tags.FindByName(ctx, tag.Name)
	if err == nil && existing != nil {
		return fmt.Errorf("%w: tag with name '%s' already exists", ErrValidationFailed, tag.Name)
	}

	if err := s.tags.Create(ctx, tag); err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}

	return nil
}

// ListTags retrieves all tags with pagination.
func (s *TagService) ListTags(ctx context.Context, limit, offset int) ([]*domain.Tags, error) {
	return s.tags.List(ctx, limit, offset)
}

// GetTagByID retrieves a tag by its ID.
func (s *TagService) GetTagByID(ctx context.Context, id string) (*domain.Tags, error) {
	tag, err := s.tags.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("%w: tag not found", ErrResourceNotFound)
		}
		return nil, err
	}
	return tag, nil
}

// UpdateTag updates an existing tag.
func (s *TagService) UpdateTag(ctx context.Context, id string, name string, color *string, description *string) (*domain.Tags, error) {
	if name == "" {
		return nil, fmt.Errorf("%w: tag name is required", ErrValidationFailed)
	}

	// Fetch existing tag
	tag, err := s.tags.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("%w: tag not found", ErrResourceNotFound)
		}
		return nil, err
	}

	// Check if new name conflicts with another tag
	if tag.Name != name {
		existing, err := s.tags.FindByName(ctx, name)
		if err == nil && existing != nil && existing.ID != id {
			return nil, fmt.Errorf("%w: tag with name '%s' already exists", ErrValidationFailed, name)
		}
	}

	// Update tag fields
	tag.Name = name
	tag.Color = color
	tag.Description = description

	if err := s.tags.Update(ctx, tag); err != nil {
		return nil, fmt.Errorf("failed to update tag: %w", err)
	}

	return tag, nil
}

// DeleteTag removes a tag from the system.
// GORM will automatically handle the many-to-many relationship cleanup.
func (s *TagService) DeleteTag(ctx context.Context, id string) error {
	// Verify tag exists
	_, err := s.tags.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("%w: tag not found", ErrResourceNotFound)
		}
		return err
	}

	if err := s.tags.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	return nil
}
