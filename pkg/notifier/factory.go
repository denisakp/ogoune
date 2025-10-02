package notifier

import (
	"fmt"

	"github.com/denisakp/pulseguard/internal/domain"
)

// NewNotifier creates a notifier instance based on the integration type.
func NewNotifier(integrationType domain.IntegrationType) (Notifier, error) {
	switch integrationType {
	case domain.IntegrationSMTP:
		return NewSMTPNotifier(), nil
	case domain.IntegrationSlack:
		return NewInAppNotifier(), nil // Placeholder for Slack
	case domain.IntegrationGoogleChat:
		return NewInAppNotifier(), nil // Placeholder for Google Chat
	case "inapp": // Special case for in-app notifications
		return NewInAppNotifier(), nil
	default:
		return nil, fmt.Errorf("unsupported integration type: %s", integrationType)
	}
}
