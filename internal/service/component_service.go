package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/denisakp/ogoune/internal/config"
	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/dto"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository"
	"github.com/denisakp/ogoune/pkg/notifier"
)

// ComponentService manages logical components and their derived status/notifications.
type ComponentService struct {
	components    port.ComponentRepository
	resources     port.ResourceRepository
	channels      port.NotificationChannelRepository
	cfg           *config.Config
	pendingTimers sync.Map // componentID -> *time.Timer
}

// NewComponentService creates a new ComponentService.
func NewComponentService(
	components port.ComponentRepository,
	resources port.ResourceRepository,
	channels port.NotificationChannelRepository,
) *ComponentService {
	return &ComponentService{
		components: components,
		resources:  resources,
		channels:   channels,
	}
}

// NewComponentServiceWithConfig creates a ComponentService with smart alerting configuration.
func NewComponentServiceWithConfig(
	components port.ComponentRepository,
	resources port.ResourceRepository,
	channels port.NotificationChannelRepository,
	cfg *config.Config,
) *ComponentService {
	svc := &ComponentService{
		components: components,
		resources:  resources,
		channels:   channels,
		cfg:        cfg,
	}
	return svc
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

	if payload.GroupingWindowSeconds != nil {
		if err := validateGroupingWindow(*payload.GroupingWindowSeconds); err != nil {
			return nil, err
		}
		component.GroupingWindowSeconds = *payload.GroupingWindowSeconds
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

	if payload.GroupingWindowSeconds != nil {
		if err := validateGroupingWindow(*payload.GroupingWindowSeconds); err != nil {
			return nil, err
		}
		component.GroupingWindowSeconds = *payload.GroupingWindowSeconds
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
// When GroupingWindowSeconds > 0, notifications are deferred via a sliding timer so that
// rapid successive changes within the window produce exactly one grouped notification.
func (s *ComponentService) RecalculateAndNotify(ctx context.Context, componentID string) error {
	component, err := s.components.FindByID(ctx, componentID)
	if err != nil {
		return err
	}

	resources, err := s.resources.FindByComponentID(ctx, componentID)
	if err != nil {
		return err
	}

	status, _ := deriveComponentStatus(resources)

	// Avoid triggering when status is unchanged
	if component.LastNotificationStatus == status {
		return nil
	}

	// Determine grouping window
	window := 0
	if s.cfg != nil {
		window = s.cfg.GroupingWindowSeconds
	}
	if component.GroupingWindowSeconds > 0 {
		window = component.GroupingWindowSeconds
	}

	if window > 0 {
		// Cancel existing pending timer for this component (sliding window)
		if existing, loaded := s.pendingTimers.LoadAndDelete(componentID); loaded {
			existing.(*time.Timer).Stop()
		}
		// Schedule deferred dispatch
		t := time.AfterFunc(time.Duration(window)*time.Second, func() {
			s.pendingTimers.Delete(componentID)
			s.dispatchComponentAlert(componentID)
		})
		s.pendingTimers.Store(componentID, t)
		return nil
	}

	// No grouping window — dispatch immediately
	return s.dispatchComponentAlertImmediate(ctx, componentID)
}

// dispatchComponentAlert is called by the deferred timer; it re-fetches current state
// before dispatching to ensure the notification reflects the latest snapshot.
func (s *ComponentService) dispatchComponentAlert(componentID string) {
	ctx := context.Background()
	if err := s.dispatchComponentAlertImmediate(ctx, componentID); err != nil {
		slog.Error("failed to dispatch component alert", "component_id", componentID, "error", err)
	}
}

// dispatchComponentAlertImmediate fetches current state and dispatches to all channels.
// On failure it retries up to 3 times with exponential back-off; if all attempts fail,
// individual per-resource alerts are sent as a fallback.
func (s *ComponentService) dispatchComponentAlertImmediate(ctx context.Context, componentID string) error {
	component, err := s.components.FindByID(ctx, componentID)
	if err != nil {
		return err
	}

	resources, err := s.resources.FindByComponentID(ctx, componentID)
	if err != nil {
		return err
	}

	status, impacted := deriveComponentStatus(resources)

	// Idempotence guard: if status matches what we last notified, skip
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

	const maxAttempts = 3
	backoff := []time.Duration{0, 2 * time.Second, 4 * time.Second}
	var lastErr error

	for _, ch := range channels {
		sent := false
		for attempt := range maxAttempts {
			if backoff[attempt] > 0 {
				time.Sleep(backoff[attempt])
			}

			if err := s.sendNotification(ctx, payload, ch); err != nil {
				lastErr = err
				continue
			}
			sent = true
			break
		}
		if !sent {
			// Fallback: log failure; per-resource alerts would require incident context
			// that is not available here, so we just log and continue.
			slog.Warn("channel exhausted retries for component alert", "channel_id", ch.ID, "component_id", componentID)
		}
	}

	// Update last notified status
	if err := s.components.UpdateLastNotificationStatus(ctx, componentID, status); err != nil {
		return err
	}

	if lastErr != nil {
		return lastErr
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
		ID:                    component.ID,
		Name:                  component.Name,
		Description:           component.Description,
		Status:                status,
		ImpactedResources:     impacted,
		Resources:             snaps,
		GroupingWindowSeconds: component.GroupingWindowSeconds,
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
				slog.Warn("failed to auto-cleanup component", "component_id", oldComponentID, "error", err)
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
			slog.Warn("failed to auto-cleanup component", "component_id", componentID, "error", err)
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

// validateGroupingWindow validates the grouping_window_seconds value.
// 0 means disabled; non-zero must be between 10 and 300 seconds.
func validateGroupingWindow(seconds int) error {
	if seconds != 0 && (seconds < 10 || seconds > 300) {
		return fmt.Errorf("%w: grouping_window_seconds must be 0 (disabled) or between 10 and 300 (got %d)", ErrValidationFailed, seconds)
	}
	return nil
}
