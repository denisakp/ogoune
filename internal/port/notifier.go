package port

import (
	"github.com/denisakp/ogoune/pkg/notifier"
)

// Notifier is the port interface for sending notifications.
// It aliases the existing notifier.Notifier to avoid duplicating the
// NotificationPayload type hierarchy while centralizing the contract here.
type Notifier = notifier.Notifier
