package notifier

import (
	"context"

	"github.com/denisakp/pulseguard/internal/domain"
)

// ComponentNotification carries derived component state for aggregated notifications.
type ComponentNotification struct {
	Component domain.Component
	Status    domain.ComponentStatus
	Previous  *domain.ComponentStatus
	Impacted  []ComponentResource
}

// ComponentResource captures a simplified view of a resource for notifications.
type ComponentResource struct {
	ID     string
	Name   string
	Status domain.ResourceStatus
}

// NotificationPayload is a union of either a resource incident or a component update.
// Only one of Incident or Component should be non-nil.
type NotificationPayload struct {
	Incident  *domain.Incident
	Component *ComponentNotification
}

// Notifier defines the interface for sending notifications.
// Both SMTP and Webhook notifiers implement this interface.
type Notifier interface {
	Send(ctx context.Context, payload NotificationPayload) error
}
