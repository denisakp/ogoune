package notifier

import (
	"context"
	"fmt"
	"log"

	"github.com/denisakp/pulseguard/internal/domain"
)

// InAppNotifier implements in-app notifications (logging for now).
type InAppNotifier struct{}

// NewInAppNotifier creates a new in-app notifier instance.
func NewInAppNotifier() *InAppNotifier {
	return &InAppNotifier{}
}

// Send logs an in-app notification for the incident.
// In a real system, this might create a notification record in the database
// or push to a WebSocket connection.
func (n *InAppNotifier) Send(ctx context.Context, config domain.Integration, incident domain.Incident) error {
	status := "DOWN"
	if incident.IsResolved {
		status = "UP"
	}

	message := fmt.Sprintf(
		"[IN-APP NOTIFICATION] Resource %s is %s | Incident: %s | Reason: %s | Time: %s",
		incident.ResourceID,
		status,
		incident.ID,
		incident.Reason,
		incident.StartedAt.Format("2006-01-02 15:04:05"),
	)

	log.Println(message)

	// TODO: In production, this could:
	// 1. Create a notification_events record in the database
	// 2. Push to connected WebSocket clients
	// 3. Store in a notification queue for user retrieval
	// 4. Send push notifications to mobile apps

	return nil
}
