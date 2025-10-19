package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/denisakp/pulseguard/internal/repository"
	"github.com/hibiken/asynq"
)

// NotificationTaskHandler processes notification tasks from the Asynq queue.
type NotificationTaskHandler struct {
	incidents     repository.IncidentRepository
	integrations  repository.IntegrationRepository
	notifications repository.NotificationRepository
}

// NewNotificationTaskHandler creates a new notification task handler.
func NewNotificationTaskHandler(
	incidents repository.IncidentRepository,
	integrations repository.IntegrationRepository,
	notifications repository.NotificationRepository,
) *NotificationTaskHandler {
	return &NotificationTaskHandler{
		incidents:     incidents,
		integrations:  integrations,
		notifications: notifications,
	}
}

// ProcessTask processes a notification task from the queue.
func (h *NotificationTaskHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	// Parse the task payload
	var payload struct {
		IncidentID string `json:"incident_id"`
		EventType  string `json:"event_type"`
		ResourceID string `json:"resource_id"`
		Timestamp  int64  `json:"timestamp"`
	}

	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal notification payload: %w", err)
	}

	// Get the incident details
	_, err := h.incidents.FindByID(ctx, payload.IncidentID)
	if err != nil {
		return fmt.Errorf("failed to find incident %s: %w", payload.IncidentID, err)
	}

	// TODO: Implement actual notification sending logic
	// This would involve:
	// 1. Getting active integrations from h.integrations
	// 2. Formatting the notification message based on event type
	// 3. Sending notifications via the pkg/notifier package
	// 4. Recording notification events in h.notifications

	// For now, just log that we would send a notification
	fmt.Printf("Would send %s notification for incident %s (resource %s)\n",
		payload.EventType, payload.IncidentID, payload.ResourceID)

	return nil
}
