package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/denisakp/ogoune/internal/repository"
	"github.com/hibiken/asynq"
)

// NotificationTaskHandler processes notification tasks from the Asynq queue.
// Note: Notifications are now handled directly in the IncidentService,
// so this handler is kept for backward compatibility but not actively used.
type NotificationTaskHandler struct {
	incidents     repository.IncidentRepository
	notifications repository.NotificationRepository
}

// NewNotificationTaskHandler creates a new notification task handler.
func NewNotificationTaskHandler(
	incidents repository.IncidentRepository,
	notifications repository.NotificationRepository,
) *NotificationTaskHandler {
	return &NotificationTaskHandler{
		incidents:     incidents,
		notifications: notifications,
	}
}

// ProcessTask processes a notification task from the queue.
// Notifications are now sent directly from IncidentService when incidents are created/resolved.
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

	// Notifications are now handled directly in IncidentService.CreateIncident and
	// IncidentService.ResolveIncident, so this handler is not actively used.
	// Keeping it for backward compatibility with existing task queue entries.
	fmt.Printf("Notification task received for incident %s (resource %s) - handled by IncidentService\n",
		payload.IncidentID, payload.ResourceID)

	return nil
}
