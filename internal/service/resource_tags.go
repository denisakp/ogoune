package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository"
)

// findOrCreateTags finds existing tags by name or creates new ones if they don't exist.
// It accepts tag names as strings and returns tag entities.
func (s *ResourceService) findOrCreateTags(ctx context.Context, tagNames []string) ([]*domain.Tags, error) {
	var tags []*domain.Tags
	for _, rawTag := range tagNames {
		tagName := strings.TrimSpace(rawTag)
		if tagName == "" {
			continue
		}
		tag, err := s.resolveOrCreateTag(ctx, tagName)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

// resolveOrCreateTag returns the tag matching nameOrID, falling back to lookup-by-name
// and finally creating a new tag. Used by findOrCreateTags so the per-name decision
// tree stays under the cognitive-complexity budget of its caller.
func (s *ResourceService) resolveOrCreateTag(ctx context.Context, nameOrID string) (*domain.Tags, error) {
	// Backward compatibility: if client sends an existing tag ID, resolve it first.
	if tag, err := s.tags.FindByID(ctx, nameOrID); err == nil {
		return tag, nil
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, fmt.Errorf("failed to find tag by id '%s': %w", nameOrID, err)
	}

	tag, err := s.tags.FindByName(ctx, nameOrID)
	if err == nil {
		return tag, nil
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return nil, fmt.Errorf("failed to find tag '%s': %w", nameOrID, err)
	}

	newTag := &domain.Tags{Name: nameOrID}
	if err := s.tags.Create(ctx, newTag); err != nil {
		return nil, fmt.Errorf("failed to create tag '%s': %w", nameOrID, err)
	}
	return newTag, nil
}

// resolveChannelsByName resolves notification channels by exact name.
// Channels are never created here (they hold secrets); an unknown name is a
// validation error. The full channel list is loaded once and matched in memory,
// since there is no find-by-name repository method and channel counts are small.
func (s *ResourceService) resolveChannelsByName(ctx context.Context, names []string) ([]*domain.NotificationChannel, error) {
	if s.channels == nil {
		return nil, fmt.Errorf("%w: notification channel support is not configured", ErrValidationFailed)
	}
	all, err := s.channels.List(ctx, 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list notification channels: %w", err)
	}
	byName := make(map[string]*domain.NotificationChannel, len(all))
	for _, ch := range all {
		if ch != nil {
			byName[ch.Name] = ch
		}
	}
	var resolved []*domain.NotificationChannel
	for _, raw := range names {
		name := strings.TrimSpace(raw)
		if name == "" {
			continue
		}
		ch, ok := byName[name]
		if !ok {
			return nil, fmt.Errorf("%w: notification channel '%s' not found", ErrValidationFailed, name)
		}
		resolved = append(resolved, ch)
	}
	return resolved, nil
}

// AddTagsToResource adds multiple tags to a resource using GORM's Association mode.
func (s *ResourceService) AddTagsToResource(ctx context.Context, resourceID string, tagIDs []string) error {
	// Fetch the resource
	resource, err := s.resources.FindByID(ctx, resourceID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("%w: resource not found", ErrResourceNotFound)
		}
		return err
	}

	// Fetch all tags
	var tags []*domain.Tags
	for _, tagID := range tagIDs {
		tag, err := s.tags.FindByID(ctx, tagID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return fmt.Errorf("%w: tag with ID '%s' not found", ErrValidationFailed, tagID)
			}
			return err
		}
		tags = append(tags, tag)
	}

	// Use GORM association to append tags
	// This requires database access, so we need to get the DB instance
	// For now, we'll append tags to the resource and update
	resource.Tags = append(resource.Tags, tags...)

	if err := s.resources.Update(ctx, resource); err != nil {
		return fmt.Errorf("failed to add tags to resource: %w", err)
	}

	return nil
}

// RemoveTagFromResource removes a specific tag from a resource.
func (s *ResourceService) RemoveTagFromResource(ctx context.Context, resourceID string, tagID string) error {
	// Fetch the resource with its tags
	resource, err := s.resources.FindByID(ctx, resourceID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("%w: resource not found", ErrResourceNotFound)
		}
		return err
	}

	// Verify tag exists
	_, err = s.tags.FindByID(ctx, tagID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("%w: tag not found", ErrResourceNotFound)
		}
		return err
	}

	// Filter out the tag to be removed
	var filteredTags []*domain.Tags
	tagFound := false
	for _, tag := range resource.Tags {
		if tag.ID != tagID {
			filteredTags = append(filteredTags, tag)
		} else {
			tagFound = true
		}
	}

	if !tagFound {
		return fmt.Errorf("%w: tag not associated with resource", ErrValidationFailed)
	}

	resource.Tags = filteredTags

	if err := s.resources.Update(ctx, resource); err != nil {
		return fmt.Errorf("failed to remove tag from resource: %w", err)
	}

	return nil
}
