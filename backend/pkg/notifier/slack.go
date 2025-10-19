package notifier

import (
	"context"
	"fmt"

	"github.com/denisakp/pulseguard/internal/domain"
)

// SlackNotifier is a placeholder for Slack notification implementation.
type SlackNotifier struct{}

// NewSlackNotifier creates a new Slack notifier instance.
func NewSlackNotifier() Notifier {
	return &SlackNotifier{}
}

// Send is a placeholder method that returns an error indicating the notifier is not yet implemented.
func (n *SlackNotifier) Send(ctx context.Context, integration domain.Integration, incident domain.Incident) error {
	return fmt.Errorf("slack notifier not yet implemented")
}
