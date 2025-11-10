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
// Notifications are sent via:
// 1. SMTP (if configured via env variables)
// 2. Webhook (if configured via env variables)
type IncidentService struct {
	incidents     repository.IncidentRepository
	eventSteps    repository.IncidentEventStepRepository
	notifications repository.NotificationRepository
	client        *asynq.Client
	smtpIsEnabled bool
	smtpRecipient string
	smtpSender    string
	smtpHost      string
	smtpPort      string
	smtpUser      string
	smtpPassword  string
	webhookUrl    string
	webhookSecret *string
}

// NewIncidentService creates a new incident service with the given dependencies.
func NewIncidentService(
	incidents repository.IncidentRepository,
	eventSteps repository.IncidentEventStepRepository,
	notifications repository.NotificationRepository,
	client *asynq.Client,
	smtpIsEnabled bool,
	smtpRecipient string,
	smtpSender string,
	smtpHost string,
	smtpPort string,
	smtpUser string,
	smtpPassword string,
	webhookUrl string,
	webhookSecret *string,
) *IncidentService {
	return &IncidentService{
		incidents:     incidents,
		eventSteps:    eventSteps,
		notifications: notifications,
		client:        client,
		smtpIsEnabled: smtpIsEnabled,
		smtpRecipient: smtpRecipient,
		smtpSender:    smtpSender,
		smtpHost:      smtpHost,
		smtpPort:      smtpPort,
		smtpUser:      smtpUser,
		smtpPassword:  smtpPassword,
		webhookUrl:    webhookUrl,
		webhookSecret: webhookSecret,
	}
}

// CreateIncident creates a new incident when a resource reaches 3 consecutive failures.
// It checks for existing active incidents, creates event steps, and triggers notifications.
// Notifications are sent via SMTP and webhook (if configured).
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

	// ============================================================
	// Send SMTP Notification (if enabled)
	// ============================================================
	if s.smtpIsEnabled {
		if err := s.sendDownNotification(ctx, incident); err != nil {
			log.Printf("Warning: Failed to send SMTP down notification: %v", err)
			// Continue processing - notification failure should not stop incident creation
		}

		// Create "resource_down_alert" event step (only if SMTP notification was attempted)
		downAlertStep := &domain.IncidentEventStep{
			IncidentID: incident.ID,
			Step:       domain.IncidentEventStepDownAlert,
			Message:    stringPtr("Default SMTP resource down alert sent"),
		}

		if _, err := s.eventSteps.Create(ctx, downAlertStep); err != nil {
			log.Printf("Warning: Failed to create resource_down_alert event step: %v", err)
		}
	} else {
		log.Println("SMTP notifications disabled, skipping default DOWN notification")
	}

	// ============================================================
	// Send Webhook Notification (if enabled)
	// ============================================================
	if s.webhookUrl != "" {
		if err := s.sendWebhookNotification(ctx, incident); err != nil {
			log.Printf("Warning: Failed to send webhook down notification: %v", err)
			// Continue processing - notification failure should not stop incident creation
		}
	} else {
		log.Println("Webhook notifications disabled (no URL configured)")
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

	if _, err := s.eventSteps.Create(ctx, resolvedStep); err != nil {
		log.Printf("Warning: Failed to create resolved event step: %v", err)
	}

	// ============================================================
	// Send SMTP Notification (if enabled)
	// ============================================================
	if s.smtpIsEnabled {
		if err := s.sendUpNotification(ctx, activeIncident); err != nil {
			log.Printf("Warning: Failed to send SMTP up notification: %v", err)
			// Continue processing - notification failure should not stop incident resolution
		}

		// Create "resource_up_alert" event step (only if SMTP notification was attempted)
		upAlertStep := &domain.IncidentEventStep{
			IncidentID: activeIncident.ID,
			Step:       domain.IncidentEventStepUpAlert,
			Message:    stringPtr("Default SMTP resource up alert sent"),
		}

		if _, err := s.eventSteps.Create(ctx, upAlertStep); err != nil {
			log.Printf("Warning: Failed to create resource_up_alert event step: %v", err)
		}
	} else {
		log.Println("SMTP notifications disabled, skipping default UP notification")
	}

	// ============================================================
	// Send Webhook Notification (if enabled)
	// ============================================================
	if s.webhookUrl != "" {
		if err := s.sendWebhookNotification(ctx, activeIncident); err != nil {
			log.Printf("Warning: Failed to send webhook up notification: %v", err)
			// Continue processing - notification failure should not stop incident resolution
		}
	} else {
		log.Println("Webhook notifications disabled (no URL configured)")
	}

	return nil
}

// sendDownNotification sends a "Resource Down" email notification via SMTP.
func (s *IncidentService) sendDownNotification(ctx context.Context, incident *domain.Incident) error {
	// Create SMTP notifier with configured credentials
	smtpNotifier := notifier.NewSMTPNotifier(
		s.smtpRecipient,
		s.smtpSender,
		s.smtpHost,
		s.smtpPort,
		s.smtpUser,
		s.smtpPassword,
	)

	// Send the notification
	err := smtpNotifier.Send(ctx, *incident)

	// Create notification record to log the attempt
	notificationEvent := &domain.NotificationEvent{
		IncidentID: incident.ID,
		Type:       domain.NotificationEventTypeDown,
	}

	if err != nil {
		// Log error verbosely - this is critical for debugging SMTP issues
		log.Printf("[SMTP ERROR] Failed to send DOWN notification for incident %s: %v", incident.ID, err)
		log.Printf("[SMTP ERROR] Configuration - Recipient: %s, Sender: %s", s.smtpRecipient, s.smtpSender)
		log.Printf("[SMTP ERROR] Resource: %s (%s), Cause: %s", incident.Resource.Name, incident.Resource.Target, incident.Cause)

		// Continue to persist the notification event with failure status
	} else {
		log.Printf("[SMTP SUCCESS] Sent DOWN notification for incident %s to %s", incident.ID, s.smtpRecipient)
	}

	// Persist the notification event (regardless of send success/failure)
	if err := s.notifications.Create(ctx, notificationEvent); err != nil {
		log.Printf("Warning: Failed to persist notification event: %v", err)
	}

	return err
}

// sendUpNotification sends a "Resource Up" email notification via SMTP.
func (s *IncidentService) sendUpNotification(ctx context.Context, incident *domain.Incident) error {
	// Create SMTP notifier with configured credentials
	smtpNotifier := notifier.NewSMTPNotifier(
		s.smtpRecipient,
		s.smtpSender,
		s.smtpHost,
		s.smtpPort,
		s.smtpUser,
		s.smtpPassword,
	)

	// Send the notification
	err := smtpNotifier.Send(ctx, *incident)

	// Create notification record to log the attempt
	notificationEvent := &domain.NotificationEvent{
		IncidentID: incident.ID,
		Type:       domain.NotificationEventTypeUp,
	}

	if err != nil {
		// Log error verbosely - this is critical for debugging SMTP issues
		log.Printf("[SMTP ERROR] Failed to send UP notification for incident %s: %v", incident.ID, err)
		log.Printf("[SMTP ERROR] Configuration - Recipient: %s, Sender: %s", s.smtpRecipient, s.smtpSender)
		log.Printf("[SMTP ERROR] Resource: %s (%s), Cause: %s", incident.Resource.Name, incident.Resource.Target, incident.Cause)

		// Continue to persist the notification event with failure status
	} else {
		log.Printf("[SMTP SUCCESS] Sent UP notification for incident %s to %s", incident.ID, s.smtpRecipient)
	}

	// Persist the notification event (regardless of send success/failure)
	if err := s.notifications.Create(ctx, notificationEvent); err != nil {
		log.Printf("Warning: Failed to persist notification event: %v", err)
	}

	return err
}

// sendWebhookNotification sends a notification via webhook.
func (s *IncidentService) sendWebhookNotification(ctx context.Context, incident *domain.Incident) error {
	// Create webhook notifier with configured URL and secret
	webhookNotifier := notifier.NewWebHookNotifier(s.webhookUrl, s.webhookSecret)

	// Send the notification
	err := webhookNotifier.Send(ctx, *incident)

	// Determine notification event type based on incident status
	eventType := domain.NotificationEventTypeDown
	if incident.ResolvedAt != nil {
		eventType = domain.NotificationEventTypeUp
	}

	// Create notification record to log the attempt
	notificationEvent := &domain.NotificationEvent{
		IncidentID: incident.ID,
		Type:       eventType,
	}

	if err != nil {
		log.Printf("[WEBHOOK ERROR] Failed to send %s notification for incident %s: %v",
			eventType, incident.ID, err)
	} else {
		log.Printf("[WEBHOOK SUCCESS] Sent %s notification for incident %s to %s",
			eventType, incident.ID, s.webhookUrl)
	}

	// Persist the notification event (regardless of send success/failure)
	if err := s.notifications.Create(ctx, notificationEvent); err != nil {
		log.Printf("Warning: Failed to persist webhook notification event: %v", err)
	}

	return err
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

// stringPtr is a helper to create string pointers.
func stringPtr(s string) *string {
	return &s
}
