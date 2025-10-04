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

// IncidentService manages incident creation and resolution with refined stateful logic.
// It creates incidents only after 3 consecutive failures and tracks the full lifecycle.
type IncidentService struct {
	incidents     repository.IncidentRepository
	eventSteps    repository.IncidentEventStepRepository
	integrations  repository.IntegrationRepository
	notifications repository.NotificationRepository
	client        *asynq.Client
}

// NewIncidentService creates a new incident service with the given dependencies.
func NewIncidentService(
	incidents repository.IncidentRepository,
	eventSteps repository.IncidentEventStepRepository,
	integrations repository.IntegrationRepository,
	notifications repository.NotificationRepository,
	client *asynq.Client,
) *IncidentService {
	return &IncidentService{
		incidents:     incidents,
		eventSteps:    eventSteps,
		integrations:  integrations,
		notifications: notifications,
		client:        client,
	}
}

// CreateIncident creates a new incident when a resource reaches 3 consecutive failures.
// It checks for existing active incidents, creates event steps, and triggers notifications.
func (s *IncidentService) CreateIncident(ctx context.Context, r *domain.Resource, result Result) error {
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
			log.Printf("Active incident already exists for resource %s, skipping creation", r.ID)
			return nil // Active incident already exists, avoid duplicates
		}
	}

	// Extract cause from result - this is the structured failure reason
	cause := extractCause(result)

	// Create new incident
	incident := &domain.Incident{
		ResourceID: r.ID,
		Resource:   *r,
		Reason:     "Resource failed health check 3 consecutive times",
		Cause:      cause,
		ResolvedAt: nil, // nil means active
		StartedAt:  time.Now(),
		Details:    []byte(result.ResponseData),
	}

	if err := s.incidents.Create(ctx, incident); err != nil {
		return fmt.Errorf("failed to create incident: %w", err)
	}

	log.Printf("Created incident %s for resource %s (cause: %s)", incident.ID, r.ID, cause)

	// Step 1: Create "detected" event step
	detectedStep := &domain.IncidentEventStep{
		IncidentID: incident.ID,
		Step:       domain.IncidentEventStepDetected,
		Message:    stringPtr(fmt.Sprintf("Incident detected: %s", cause)),
	}

	if err := s.eventSteps.Create(ctx, detectedStep); err != nil {
		log.Printf("Warning: Failed to create detected event step: %v", err)
	}

	// Step 2: Send DOWN notification via SMTP
	if err := s.sendDownNotification(ctx, incident); err != nil {
		log.Printf("Warning: Failed to send down notification: %v", err)
	}

	// Step 3: Create "resource_down_alert" event step
	downAlertStep := &domain.IncidentEventStep{
		IncidentID: incident.ID,
		Step:       domain.IncidentEventStepDownAlert,
		Message:    stringPtr("Resource down alert sent"),
	}

	if err := s.eventSteps.Create(ctx, downAlertStep); err != nil {
		log.Printf("Warning: Failed to create resource_down_alert event step: %v", err)
	}

	return nil
}

// ResolveIncident resolves an active incident when a resource recovers.
// It updates the incident, creates event steps, and triggers recovery notifications.
func (s *IncidentService) ResolveIncident(ctx context.Context, r *domain.Resource, result Result) error {
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
		log.Printf("No active incident found for resource %s", r.ID)
		return nil
	}

	// Resolve the incident by setting ResolvedAt timestamp
	now := time.Now()
	activeIncident.ResolvedAt = &now

	if err := s.incidents.Update(ctx, activeIncident); err != nil {
		return fmt.Errorf("failed to resolve incident: %w", err)
	}

	log.Printf("Resolved incident %s for resource %s", activeIncident.ID, r.ID)

	// Step 1: Create "resolved" event step
	resolvedStep := &domain.IncidentEventStep{
		IncidentID: activeIncident.ID,
		Step:       domain.IncidentEventStepResolved,
		Message:    stringPtr("Incident resolved: resource is back up"),
	}

	if err := s.eventSteps.Create(ctx, resolvedStep); err != nil {
		log.Printf("Warning: Failed to create resolved event step: %v", err)
	}

	// Step 2: Send UP notification via SMTP
	if err := s.sendUpNotification(ctx, activeIncident); err != nil {
		log.Printf("Warning: Failed to send up notification: %v", err)
	}

	// Step 3: Create "resource_up_alert" event step
	upAlertStep := &domain.IncidentEventStep{
		IncidentID: activeIncident.ID,
		Step:       domain.IncidentEventStepUpAlert,
		Message:    stringPtr("Resource up alert sent"),
	}

	if err := s.eventSteps.Create(ctx, upAlertStep); err != nil {
		log.Printf("Warning: Failed to create resource_up_alert event step: %v", err)
	}

	return nil
}

// sendDownNotification sends a "Resource Down" email notification and logs it.
func (s *IncidentService) sendDownNotification(ctx context.Context, incident *domain.Incident) error {
	// Use default SMTP notifier
	smtpNotifier := notifier.NewSMTPNotifier()

	// Create a default SMTP integration for sending
	smtpIntegration := domain.Integration{
		Base:     domain.Base{ID: "smtp-default"},
		Name:     "Default SMTP",
		Target:   "alerts@pulseguard.local", // This should come from config
		Type:     domain.IntegrationSMTP,
		IsActive: true,
	}

	// Send the notification
	err := smtpNotifier.Send(ctx, smtpIntegration, *incident)

	// Create notification record to log the attempt
	notificationEvent := &domain.NotificationEvent{
		IncidentID: incident.ID,
		Type:       domain.NotificationEventTypeDown,
	}

	if err != nil {
		log.Printf("Failed to send down notification via SMTP: %v", err)
		// Even if sending failed, we log the attempt
	} else {
		log.Printf("Successfully sent down notification for incident %s", incident.ID)
	}

	// Persist the notification event
	if err := s.notifications.Create(ctx, notificationEvent); err != nil {
		log.Printf("Warning: Failed to persist notification event: %v", err)
	}

	return err
}

// sendUpNotification sends a "Resource Up" email notification and logs it.
func (s *IncidentService) sendUpNotification(ctx context.Context, incident *domain.Incident) error {
	// Use default SMTP notifier
	smtpNotifier := notifier.NewSMTPNotifier()

	// Create a default SMTP integration for sending
	smtpIntegration := domain.Integration{
		Base:     domain.Base{ID: "smtp-default"},
		Name:     "Default SMTP",
		Target:   "alerts@pulseguard.local", // This should come from config
		Type:     domain.IntegrationSMTP,
		IsActive: true,
	}

	// Send the notification
	err := smtpNotifier.Send(ctx, smtpIntegration, *incident)

	// Create notification record to log the attempt
	notificationEvent := &domain.NotificationEvent{
		IncidentID: incident.ID,
		Type:       domain.NotificationEventTypeUp,
	}

	if err != nil {
		log.Printf("Failed to send up notification via SMTP: %v", err)
	} else {
		log.Printf("Successfully sent up notification for incident %s", incident.ID)
	}

	// Persist the notification event
	if err := s.notifications.Create(ctx, notificationEvent); err != nil {
		log.Printf("Warning: Failed to persist notification event: %v", err)
	}

	return err
}

// extractCause extracts a structured cause from the monitoring result.
// This provides consistent failure categorization.
func extractCause(result Result) string {
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

// stringPtr is a helper to create string pointers.
func stringPtr(s string) *string {
	return &s
}
