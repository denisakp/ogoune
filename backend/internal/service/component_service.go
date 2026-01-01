package service

import (
	"context"
	"fmt"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/dto"
	"github.com/denisakp/pulseguard/internal/repository"
	"github.com/denisakp/pulseguard/pkg/notifier"
)

// ComponentService manages logical components and their derived status/notifications.
type ComponentService struct {
	components repository.ComponentRepository
	resources  repository.ResourceRepository
	channels   repository.NotificationChannelRepository
}

// NewComponentService creates a new ComponentService.
func NewComponentService(
	components repository.ComponentRepository,
	resources repository.ResourceRepository,
	channels repository.NotificationChannelRepository,
) *ComponentService {
	return &ComponentService{
		components: components,
		resources:  resources,
		channels:   channels,
	}
}

// CreateComponent creates a component with at least one resource and returns its DTO representation.
func (s *ComponentService) CreateComponent(ctx context.Context, payload *dto.CreateComponentPayload) (*dto.ComponentResponse, error) {
	if payload == nil || payload.Name == "" {
		return nil, fmt.Errorf("%w: component name is required", ErrValidationFailed)
	}

	if len(payload.ResourceIDs) == 0 {
		return nil, fmt.Errorf("%w: component must have at least one resource", ErrValidationFailed)
	}

	// Validate all resources exist
	for _, resourceID := range payload.ResourceIDs {
		if _, err := s.resources.FindByID(ctx, resourceID); err != nil {
			if err == repository.ErrNotFound {
				return nil, fmt.Errorf("%w: resource %s not found", ErrValidationFailed, resourceID)
			}
			return nil, err
		}
	}

	component := &domain.Component{
		Name:        payload.Name,
		Description: payload.Description,
	}

	created, err := s.components.Create(ctx, component)
	if err != nil {
		return nil, err
	}

	// Assign resources to the new component
	for _, resourceID := range payload.ResourceIDs {
		resource, _ := s.resources.FindByID(ctx, resourceID)
		resource.ComponentID = &created.ID
		if err := s.resources.Update(ctx, resource); err != nil {
			// Rollback: delete the created component if resource assignment fails
			_ = s.components.Delete(ctx, created.ID)
			return nil, fmt.Errorf("failed to assign resource %s: %w", resourceID, err)
		}
	}

	return s.toComponentResponse(ctx, created)
}

// UpdateComponent updates component metadata.
func (s *ComponentService) UpdateComponent(ctx context.Context, id string, payload *dto.UpdateComponentPayload) (*dto.ComponentResponse, error) {
	component, err := s.components.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if payload.Name != nil {
		if *payload.Name == "" {
			return nil, fmt.Errorf("%w: component name is required", ErrValidationFailed)
		}
		component.Name = *payload.Name
	}
	if payload.Description != nil {
		component.Description = payload.Description
	}

	if err := s.components.Update(ctx, component); err != nil {
		return nil, err
	}

	return s.toComponentResponse(ctx, component)
}

// DeleteComponent removes a component if no resources are attached.
func (s *ComponentService) DeleteComponent(ctx context.Context, id string) error {
	count, err := s.resources.CountByComponentID(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("%w: component still has %d resources", ErrValidationFailed, count)
	}
	return s.components.Delete(ctx, id)
}

// ListComponents returns components with derived status.
func (s *ComponentService) ListComponents(ctx context.Context, limit, offset int) ([]*dto.ComponentResponse, error) {
	components, err := s.components.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.ComponentResponse, 0, len(components))
	for _, c := range components {
		resp, err := s.toComponentResponse(ctx, c)
		if err != nil {
			return nil, err
		}
		responses = append(responses, resp)
	}
	return responses, nil
}

// GetComponent returns one component.
func (s *ComponentService) GetComponent(ctx context.Context, id string) (*dto.ComponentResponse, error) {
	component, err := s.components.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.toComponentResponse(ctx, component)
}

// RecalculateAndNotify derives component status and emits a notification when it changes.
func (s *ComponentService) RecalculateAndNotify(ctx context.Context, componentID string) error {
	component, err := s.components.FindByID(ctx, componentID)
	if err != nil {
		return err
	}

	resources, err := s.resources.FindByComponentID(ctx, componentID)
	if err != nil {
		return err
	}

	status, impacted := deriveComponentStatus(resources)

	// Avoid duplicate notifications when status is unchanged
	if component.LastNotificationStatus == status {
		return nil
	}

	channels, err := s.collectChannels(ctx, resources)
	if err != nil {
		return err
	}

	payload := notifier.NotificationPayload{
		Component: &notifier.ComponentNotification{
			Component: *component,
			Status:    status,
			Previous:  &component.LastNotificationStatus,
			Impacted:  impacted,
		},
	}

	for _, channel := range channels {
		// Skip errors on individual channels to avoid blocking others
		_ = s.sendNotification(ctx, payload, channel)
	}

	// Update last notified status even if no channels
	if err := s.components.UpdateLastNotificationStatus(ctx, componentID, status); err != nil {
		return err
	}

	return nil
}

func (s *ComponentService) sendNotification(ctx context.Context, payload notifier.NotificationPayload, channel *domain.NotificationChannel) error {
	var n notifier.Notifier
	var err error

	switch channel.Type {
	case "smtp":
		n, err = notifier.NewSMTPNotifierFromConfig(string(channel.Config))
	case "webhook":
		n, err = notifier.NewWebhookNotifierFromConfig(string(channel.Config))
	default:
		err = fmt.Errorf("unknown notification channel type: %s", channel.Type)
	}
	if err != nil {
		return err
	}

	return n.Send(ctx, payload)
}

func (s *ComponentService) collectChannels(ctx context.Context, resources []*domain.Resource) ([]*domain.NotificationChannel, error) {
	seen := make(map[string]struct{})
	channels := make([]*domain.NotificationChannel, 0)

	for _, r := range resources {
		list, err := s.channels.FindByResourceID(ctx, r.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load channels for resource %s: %w", r.ID, err)
		}
		for _, ch := range list {
			if _, exists := seen[ch.ID]; exists {
				continue
			}
			seen[ch.ID] = struct{}{}
			channels = append(channels, ch)
		}
	}

	return channels, nil
}

func (s *ComponentService) toComponentResponse(ctx context.Context, component *domain.Component) (*dto.ComponentResponse, error) {
	if component == nil {
		return nil, fmt.Errorf("component cannot be nil")
	}

	resources := component.Resources
	// If not preloaded, fetch
	if resources == nil {
		var err error
		resources, err = s.resources.FindByComponentID(ctx, component.ID)
		if err != nil {
			return nil, err
		}
	}

	status, _ := deriveComponentStatus(resources)

	snaps := make([]dto.ComponentResourceSnapshot, 0, len(resources))
	impacted := make([]dto.ComponentResourceSnapshot, 0)
	for _, r := range resources {
		snap := dto.ComponentResourceSnapshot{ID: r.ID, Name: r.Name, Status: r.Status}
		snaps = append(snaps, snap)
		if r.Status != domain.StatusUp {
			impacted = append(impacted, snap)
		}
	}

	return &dto.ComponentResponse{
		ID:                component.ID,
		Name:              component.Name,
		Description:       component.Description,
		Status:            status,
		ImpactedResources: impacted,
		Resources:         snaps,
	}, nil
}

func deriveComponentStatus(resources []*domain.Resource) (domain.ComponentStatus, []notifier.ComponentResource) {
	hasDown := false
	hasDegraded := false
	impacted := make([]notifier.ComponentResource, 0)

	for _, r := range resources {
		switch r.Status {
		case domain.StatusDown, domain.StatusError:
			hasDown = true
			impacted = append(impacted, notifier.ComponentResource{ID: r.ID, Name: r.Name, Status: r.Status})
		case domain.StatusWarn, domain.StatusPending, domain.StatusUnknown:
			hasDegraded = true
			impacted = append(impacted, notifier.ComponentResource{ID: r.ID, Name: r.Name, Status: r.Status})
		default:
			// treat paused as up
		}
	}

	switch {
	case hasDown:
		return domain.ComponentStatusDown, impacted
	case hasDegraded:
		return domain.ComponentStatusDegraded, impacted
	default:
		return domain.ComponentStatusUp, impacted
	}
}

// BulkAssignToComponent assigns multiple resources to a component.
func (s *ComponentService) BulkAssignToComponent(ctx context.Context, componentID string, payload *dto.BulkAssignPayload) error {
	if payload == nil || len(payload.ResourceIDs) == 0 {
		return fmt.Errorf("%w: at least one resource ID is required", ErrValidationFailed)
	}

	// Validate component exists
	if _, err := s.components.FindByID(ctx, componentID); err != nil {
		return err
	}

	// Assign each resource to the component
	for _, resourceID := range payload.ResourceIDs {
		resource, err := s.resources.FindByID(ctx, resourceID)
		if err != nil {
			if err == repository.ErrNotFound {
				return fmt.Errorf("%w: resource %s not found", ErrValidationFailed, resourceID)
			}
			return err
		}

		// Unschedule from old component if exists
		if resource.ComponentID != nil && *resource.ComponentID != componentID {
			oldComponentID := *resource.ComponentID
			resource.ComponentID = &componentID
			if err := s.resources.Update(ctx, resource); err != nil {
				return fmt.Errorf("failed to assign resource %s: %w", resourceID, err)
			}
			// Auto-cleanup old component if now empty
			if err := s.autoCleanupComponent(ctx, oldComponentID); err != nil {
				// Log but don't fail the operation
				fmt.Printf("failed to auto-cleanup component %s: %v\n", oldComponentID, err)
			}
		} else {
			resource.ComponentID = &componentID
			if err := s.resources.Update(ctx, resource); err != nil {
				return fmt.Errorf("failed to assign resource %s: %w", resourceID, err)
			}
		}
	}

	// Recalculate component status
	return s.RecalculateAndNotify(ctx, componentID)
}

// BulkRemoveFromComponent removes resources from their components.
func (s *ComponentService) BulkRemoveFromComponent(ctx context.Context, payload *dto.BulkRemovePayload) error {
	if payload == nil || len(payload.ResourceIDs) == 0 {
		return fmt.Errorf("%w: at least one resource ID is required", ErrValidationFailed)
	}

	affectedComponentIDs := make(map[string]struct{})

	for _, resourceID := range payload.ResourceIDs {
		resource, err := s.resources.FindByID(ctx, resourceID)
		if err != nil {
			if err == repository.ErrNotFound {
				return fmt.Errorf("%w: resource %s not found", ErrValidationFailed, resourceID)
			}
			return err
		}

		if resource.ComponentID != nil {
			affectedComponentIDs[*resource.ComponentID] = struct{}{}
			resource.ComponentID = nil
			if err := s.resources.Update(ctx, resource); err != nil {
				return fmt.Errorf("failed to remove resource %s from component: %w", resourceID, err)
			}
		}
	}

	// Auto-cleanup empty components and recalculate status for non-empty ones
	for componentID := range affectedComponentIDs {
		if err := s.autoCleanupComponent(ctx, componentID); err != nil {
			// Log but continue with other components
			fmt.Printf("failed to auto-cleanup component %s: %v\n", componentID, err)
		}
	}

	return nil
}

// autoCleanupComponent deletes a component if it has no resources.
func (s *ComponentService) autoCleanupComponent(ctx context.Context, componentID string) error {
	count, err := s.resources.CountByComponentID(ctx, componentID)
	if err != nil {
		return err
	}
	if count == 0 {
		return s.components.Delete(ctx, componentID)
	}
	// Recalculate status if component still has resources
	return s.RecalculateAndNotify(ctx, componentID)
}
