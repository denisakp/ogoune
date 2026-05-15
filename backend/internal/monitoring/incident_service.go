package monitoring

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/repository"
	"github.com/denisakp/pulseguard/pkg/notifier"
	"github.com/hibiken/asynq"
)

// IncidentService manages incident creation and resolution with dynamic notification dispatch.
// It creates incidents only after 3 consecutive failures and tracks the full lifecycle.
// Notifications are sent via user-configured notification channels (SMTP, Webhook, etc.).
type IncidentService struct {
	incidents            repository.IncidentRepository
	eventSteps           repository.IncidentEventStepRepository
	notifications        repository.NotificationRepository
	notificationChannels repository.NotificationChannelRepository
	client               *asynq.Client
}

// NewIncidentService creates a new incident service with the given dependencies.
func NewIncidentService(
	incidents repository.IncidentRepository,
	eventSteps repository.IncidentEventStepRepository,
	notifications repository.NotificationRepository,
	notificationChannels repository.NotificationChannelRepository,
	client *asynq.Client,
) *IncidentService {
	return &IncidentService{
		incidents:            incidents,
		eventSteps:           eventSteps,
		notifications:        notifications,
		notificationChannels: notificationChannels,
		client:               client,
	}
}

// CreateIncident creates a new incident when a resource reaches 3 consecutive failures.
// It checks for existing active incidents, creates event steps, and dispatches notifications
// to all configured notification channels associated with the resource.
func (s *IncidentService) CreateIncident(ctx context.Context, r *domain.Resource, result domain.CheckResult) error {
	if r == nil {
		return fmt.Errorf("resource cannot be nil")
	}

	// Check if there's already an active incident for this resource (ResolvedAt is nil)
	incidents, err := s.incidents.FindByResource(ctx, r.ID, 1, 0)
	if err != nil {
		return fmt.Errorf("failed to check for existing incidents: %w", err)
	}

	// Look for unresolved incidents (where ResolvedAt is nil)
	for _, incident := range incidents {
		if incident.ResolvedAt == nil {
			log.Printf("[INCIDENT] Active incident %s already exists for resource %s (started: %s), skipping creation to avoid duplicates",
				incident.ID, r.ID, incident.StartedAt.Format(time.RFC3339))
			return nil // Active incident already exists, avoid duplicates
		}
	}

	log.Printf("[INCIDENT] Creating NEW incident for resource %s after %d consecutive failures", r.ID, r.FailureCount)

	// Extract cause from result - this is the structured failure reason
	cause := extractCause(result)

	// Create new incident
	incident := &domain.Incident{
		ResourceID: r.ID,
		Resource:   *r,
		Cause:      cause,
		ResolvedAt: nil, // nil means active
		StartedAt:  time.Now(),
		Details:    []byte(result.ResponseData),
	}

	if _, err := s.incidents.Create(ctx, incident); err != nil {
		return fmt.Errorf("failed to create incident: %w", err)
	}

	log.Printf("Created incident %s for resource %s (cause: %s)", incident.ID, r.ID, cause)

	// Step 1: Create "detected" event step
	detectedStep := &domain.IncidentEventStep{
		IncidentID: incident.ID,
		Step:       domain.IncidentEventStepDetected,
		Message:    stringPtr(fmt.Sprintf("Incident detected: %s", cause)),
	}

	if _, err := s.eventSteps.Create(ctx, detectedStep); err != nil {
		log.Printf("Warning: Failed to create detected event step: %v", err)
	}

	// Fetch notification channels associated with this resource
	channels, err := s.notificationChannels.FindByResourceID(ctx, r.ID)
	if err != nil {
		log.Printf("Warning: Failed to fetch notification channels for resource %s: %v", r.ID, err)
		return nil // Continue even if channel fetch fails
	}

	// If no channels configured, log and return
	if len(channels) == 0 {
		log.Printf("No notification channels configured for resource %s, skipping notifications", r.ID)
		return nil
	}

	// Dispatch notifications to all configured channels
	for _, channel := range channels {
		err := s.dispatchNotification(ctx, incident, channel)

		// Create event step for notification attempt (regardless of success/failure)
		statusMsg := "sent"
		if err != nil {
			statusMsg = fmt.Sprintf("failed: %v", err)
			log.Printf("Warning: Failed to dispatch notification via channel %s (%s): %v", channel.ID, channel.Type, err)
		}

		alertStep := &domain.IncidentEventStep{
			IncidentID: incident.ID,
			Step:       domain.IncidentEventStepDownAlert,
			Message:    stringPtr(fmt.Sprintf("Down notification %s via %s (%s)", statusMsg, channel.Type, channel.Name)),
		}
		if _, err := s.eventSteps.Create(ctx, alertStep); err != nil {
			log.Printf("Warning: Failed to create alert event step: %v", err)
		}

		// Create notification event record for tracking
		notificationEvent := &domain.NotificationEvent{
			IncidentID: incident.ID,
			Type:       domain.NotificationEventTypeDown,
		}
		if err := s.notifications.Create(ctx, notificationEvent); err != nil {
			log.Printf("Warning: Failed to create notification event: %v", err)
		}
	}

	return nil
}

// ResolveIncident resolves an active incident when a resource recovers.
// It updates the incident, creates event steps, and triggers recovery notifications.
// Notifications are sent via SMTP and webhook (if configured).
func (s *IncidentService) ResolveIncident(ctx context.Context, r *domain.Resource, result domain.CheckResult) error {
	if r == nil {
		return fmt.Errorf("resource cannot be nil")
	}

	// Find the active incident for this resource (ResolvedAt is nil)
	incidents, err := s.incidents.FindByResource(ctx, r.ID, 10, 0)
	if err != nil {
		return fmt.Errorf("failed to find incidents: %w", err)
	}

	// Look for the most recent unresolved incident
	var activeIncident *domain.Incident
	for _, incident := range incidents {
		if incident.ResolvedAt == nil {
			if activeIncident == nil || incident.StartedAt.After(activeIncident.StartedAt) {
				activeIncident = incident
			}
		}
	}

	// No active incident to resolve
	if activeIncident == nil {
		log.Printf("[INCIDENT] No active incident found for resource %s (recovery without prior incident)", r.ID)
		return nil
	}

	duration := time.Since(activeIncident.StartedAt)
	log.Printf("[INCIDENT] Resolving incident %s for resource %s (duration: %v)", activeIncident.ID, r.ID, duration)

	// Resolve the incident by setting ResolvedAt timestamp
	now := time.Now()
	activeIncident.ResolvedAt = &now

	if err := s.incidents.Update(ctx, activeIncident); err != nil {
		return fmt.Errorf("failed to resolve incident: %w", err)
	}

	log.Printf("[INCIDENT] Successfully resolved incident %s for resource %s", activeIncident.ID, r.ID)

	// Step 1: Create "resolved" event step
	resolvedStep := &domain.IncidentEventStep{
		IncidentID: activeIncident.ID,
		Step:       domain.IncidentEventStepResolved,
		Message:    stringPtr("Incident resolved: resource is back up"),
	}

	if _, err := s.eventSteps.Create(ctx, resolvedStep); err != nil {
		log.Printf("Warning: Failed to create resolved event step: %v", err)
	}

	// Fetch notification channels associated with this resource
	channels, err := s.notificationChannels.FindByResourceID(ctx, r.ID)
	if err != nil {
		log.Printf("Warning: Failed to fetch notification channels for resource %s: %v", r.ID, err)
		return nil // Continue even if channel fetch fails
	}

	// If no channels configured, log and return
	if len(channels) == 0 {
		log.Printf("No notification channels configured for resource %s, skipping notifications", r.ID)
		return nil
	}

	// Dispatch resolution notifications to all configured channels
	for _, channel := range channels {
		err := s.dispatchNotification(ctx, activeIncident, channel)

		// Create event step for notification attempt (regardless of success/failure)
		statusMsg := "sent"
		if err != nil {
			statusMsg = fmt.Sprintf("failed: %v", err)
			log.Printf("Warning: Failed to dispatch resolution notification via channel %s (%s): %v", channel.ID, channel.Type, err)
		}

		upAlertStep := &domain.IncidentEventStep{
			IncidentID: activeIncident.ID,
			Step:       domain.IncidentEventStepUpAlert,
			Message:    stringPtr(fmt.Sprintf("Up notification %s via %s (%s)", statusMsg, channel.Type, channel.Name)),
		}
		if _, err := s.eventSteps.Create(ctx, upAlertStep); err != nil {
			log.Printf("Warning: Failed to create up alert event step: %v", err)
		}

		// Create notification event record for tracking
		notificationEvent := &domain.NotificationEvent{
			IncidentID: activeIncident.ID,
			Type:       domain.NotificationEventTypeUp,
		}
		if err := s.notifications.Create(ctx, notificationEvent); err != nil {
			log.Printf("Warning: Failed to create notification event: %v", err)
		}
	}

	return nil
}

// extractCause extracts a structured cause from the monitoring result.
// This provides consistent failure categorization.
func extractCause(result domain.CheckResult) string {
	// Map status to structured causes
	switch result.Status {
	case "down":
		// Check the response data for more specific causes
		if len(result.ResponseData) > 0 {
			data := result.ResponseData
			if contains(data, "timeout") {
				return "connection_timeout"
			}
			if contains(data, "refused") {
				return "connection_refused"
			}
			if contains(data, "status") || contains(data, "code") {
				return "invalid_status_code"
			}
			if contains(data, "dns") || contains(data, "resolve") {
				return "dns_resolution_failure"
			}
			if contains(data, "ssl") || contains(data, "tls") || contains(data, "certificate") {
				return "ssl_certificate_error"
			}
		}
		return "health_check_failed"
	case "error":
		return "check_execution_error"
	default:
		return "unknown_failure"
	}
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > len(substr) &&
			(findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// dispatchNotification sends notifications via the given notification channel.
// It unmarshals the channel config and instantiates the appropriate notifier.
func (s *IncidentService) dispatchNotification(ctx context.Context, incident *domain.Incident, channel *domain.NotificationChannel) error {
	var n notifier.Notifier
	var err error

	// Instantiate the appropriate notifier based on channel type
	switch channel.Type {
	case "smtp":
		n, err = notifier.NewSMTPNotifierFromConfig(string(channel.Config))
		if err != nil {
			return fmt.Errorf("failed to build SMTP notifier: %w", err)
		}
	case "webhook":
		n, err = notifier.NewWebhookNotifierFromConfig(string(channel.Config))
		if err != nil {
			return fmt.Errorf("failed to build webhook notifier: %w", err)
		}
	default:
		return fmt.Errorf("unknown notification channel type: %s", channel.Type)
	}

	// Send the notification
	if err := n.Send(ctx, *incident); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	log.Printf("Successfully sent notification via %s channel %s", channel.Type, channel.ID)
	return nil
}

// stringPtr is a helper to create string pointers.
func stringPtr(s string) *string {
	return &s
}
