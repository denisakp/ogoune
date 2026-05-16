package notifier

// Compile-time interface satisfaction checks.
var (
	_ Notifier = (*SMTPNotifier)(nil)
	_ Notifier = (*WebHookNotifier)(nil)
)
