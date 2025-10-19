package notifier

import (
	"context"
	"fmt"

	"github.com/denisakp/pulseguard/internal/domain"
)

// DiscordNotifier is a placeholder for Discord notification implementation.
type DiscordNotifier struct{}

// NewDiscordNotifier creates a new Discord notifier instance.
func NewDiscordNotifier() Notifier {
	return &DiscordNotifier{}
}

// Send is a placeholder method that returns an error indicating the notifier is not yet implemented.
func (n *DiscordNotifier) Send(ctx context.Context, integration domain.Integration, incident domain.Incident) error {
	return fmt.Errorf("discord notifier not yet implemented")
}
