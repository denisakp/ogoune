package notifier

import (
	"fmt"

	"github.com/denisakp/pulseguard/internal/domain"
)

// NotifierFactory creates notifier instances based on integration types.
type NotifierFactory struct{}

// NewNotifierFactory creates a new NotifierFactory instance.
func NewNotifierFactory() *NotifierFactory {
	return &NotifierFactory{}
}

// GetNotifier creates a notifier instance based on the integration type.
func (f *NotifierFactory) GetNotifier(integrationType domain.IntegrationType) (Notifier, error) {
	switch integrationType {
	case domain.IntegrationSMTP:
		return NewSMTPNotifier(), nil
	case domain.IntegrationGoogleChat:
		return NewGoogleChatNotifier(), nil
	case domain.IntegrationSlack:
		return NewSlackNotifier(), nil
	case domain.IntegrationDiscord:
		return NewDiscordNotifier(), nil
	default:
		return nil, fmt.Errorf("unsupported integration type: %s", integrationType)
	}
}

// NewNotifier creates a notifier instance based on the integration type.
// Deprecated: Use NotifierFactory.GetNotifier instead.
func NewNotifier(integrationType domain.IntegrationType) (Notifier, error) {
	factory := NewNotifierFactory()
	return factory.GetNotifier(integrationType)
}
