package notifier

import (
	"context"

	"github.com/denisakp/pulseguard/internal/domain"
)

// Notifier defines the interface for sending notifications across different channels.
type Notifier interface {
	Send(ctx context.Context, config domain.Integration, incident domain.Incident) error
}
