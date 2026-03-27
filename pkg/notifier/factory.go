package notifier

// NotifierFactory is kept for compatibility but is no longer needed
// for integration-based notifier creation. SMTP and Webhook notifiers
// are now instantiated directly with configuration.
type NotifierFactory struct{}

// NewNotifierFactory creates a new NotifierFactory instance.
func NewNotifierFactory() *NotifierFactory {
	return &NotifierFactory{}
}
