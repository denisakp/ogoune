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
	default:
		return nil, fmt.Errorf("unsupported integration type: %s", integrationType)
	}
}
