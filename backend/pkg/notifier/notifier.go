package notifier

import (
	"context"

	"github.com/denisakp/pulseguard/internal/domain"
)

// Notifier defines the interface for sending notifications.
// Both SMTP and Webhook notifiers implement this interface.
type Notifier interface {
	Send(ctx context.Context, incident domain.Incident) error
}
