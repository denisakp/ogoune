package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/repository"
	"github.com/denisakp/pulseguard/pkg/notifier"
	"github.com/hibiken/asynq"
)

// IncidentService manages incident creation and resolution based on resource status changes.
// It replaces the event-driven listener pattern with direct service calls.
type IncidentService struct {
	incidents    repository.IncidentRepository
	eventSteps   repository.IncidentEventStepRepository
	integrations repository.IntegrationRepository
	client       *asynq.Client
}

// NewIncidentService creates a new incident service with the given dependencies.
func NewIncidentService(
	incidents repository.IncidentRepository,
	eventSteps repository.IncidentEventStepRepository,
	integrations repository.IntegrationRepository,
	client *asynq.Client,
) *IncidentService {
	return &IncidentService{
		incidents:    incidents,
		eventSteps:   eventSteps,
		integrations: integrations,
		client:       client,
	}
}

// HandleStatusChange processes resource status changes and manages incidents accordingly.
// This method consolidates the logic that was previously spread across event listeners.
func (s *IncidentService) HandleStatusChange(ctx context.Context, r *domain.Resource, oldStatus domain.ResourceStatus, result Result) error {
	if r == nil {
		return fmt.Errorf("resource cannot be nil")
	}

	// Convert result status to domain status
	newStatus := domain.ResourceStatus(result.Status)

	// Only process actual status changes
	if oldStatus == newStatus {
		return nil
	}

	switch {
	case oldStatus == domain.StatusUp && newStatus == domain.StatusDown:
		return s.handleDownStatusChange(ctx, r, result)
	case oldStatus == domain.StatusDown && newStatus == domain.StatusUp:
		return s.handleUpStatusChange(ctx, r, result)
	default:
		// For other status changes, we might want to log them but not create incidents
		return nil
	}
}

// handleDownStatusChange creates a new incident when a resource goes down.
func (s *IncidentService) handleDownStatusChange(ctx context.Context, r *domain.Resource, result Result) error {
	// Check if there's already an active incident for this resource
	incidents, err := s.incidents.FindByResource(ctx, r.ID, 1, 0)
	if err != nil {
		return fmt.Errorf("failed to check for existing incidents: %w", err)
	}

	// Look for unresolved incidents
	hasActiveIncident := false
	for _, incident := range incidents {
		if !incident.IsResolved {
			hasActiveIncident = true
			break
		}
	}

	// Only create a new incident if there isn't an active one
	if hasActiveIncident {
		return nil // Active incident already exists, no need to create another
	}

	// Create new incident
	incident := &domain.Incident{
		ResourceID: r.ID,
		Resource:   *r,
		Reason:     "Resource status changed from UP to DOWN",
		IsResolved: false,
		StartedAt:  time.Now(),
		Details:    []byte(result.ResponseData),
	}

	if err := s.incidents.Create(ctx, incident); err != nil {
		return fmt.Errorf("failed to create incident: %w", err)
	}

	// Create first event step: detected
	detectedStep := &domain.IncidentEventStep{
		IncidentID: incident.ID,
		Step:       domain.IncidentEventStepDetected,
		Message:    stringPtr("Incident detected: resource is down"),
	}

	if err := s.eventSteps.Create(ctx, detectedStep); err != nil {
		log.Printf("Failed to create detected event step: %v", err)
	}

	// Trigger notifications
	if err := s.sendNotifications(ctx, incident, "down"); err != nil {
		log.Printf("Failed to send notifications for incident %s: %v", incident.ID, err)
	}

	// Create second event step: alert_sent
	alertStep := &domain.IncidentEventStep{
		IncidentID: incident.ID,
		Step:       domain.IncidentEventStepAlert,
		Message:    stringPtr("Alert notifications sent"),
	}

	if err := s.eventSteps.Create(ctx, alertStep); err != nil {
		log.Printf("Failed to create alert_sent event step: %v", err)
	}

	return nil
}

// handleUpStatusChange resolves an active incident when a resource comes back up.
func (s *IncidentService) handleUpStatusChange(ctx context.Context, r *domain.Resource, result Result) error {
	// Find the active incident for this resource
	incidents, err := s.incidents.FindByResource(ctx, r.ID, 10, 0)
	if err != nil {
		return fmt.Errorf("failed to find incidents: %w", err)
	}

	// Look for the most recent unresolved incident
	var activeIncident *domain.Incident
	for _, incident := range incidents {
		if !incident.IsResolved {
			if activeIncident == nil || incident.StartedAt.After(activeIncident.StartedAt) {
				activeIncident = incident
			}
		}
	}

	// No active incident to resolve
	if activeIncident == nil {
		return nil
	}

	// Resolve the incident
	now := time.Now()
	activeIncident.IsResolved = true
	activeIncident.ResolvedAt = &now

	if err := s.incidents.Update(ctx, activeIncident); err != nil {
		return fmt.Errorf("failed to resolve incident: %w", err)
	}

	// Create resolved event step
	resolvedStep := &domain.IncidentEventStep{
		IncidentID: activeIncident.ID,
		Step:       domain.IncidentEventStepResolved,
		Message:    stringPtr("Incident resolved: resource is back up"),
	}

	if err := s.eventSteps.Create(ctx, resolvedStep); err != nil {
		log.Printf("Failed to create resolved event step: %v", err)
	}

	// Trigger "resource up" notifications
	if err := s.sendNotifications(ctx, activeIncident, "up"); err != nil {
		log.Printf("Failed to send up notifications for incident %s: %v", activeIncident.ID, err)
	}

	// Create alert sent event step for resolution
	alertStep := &domain.IncidentEventStep{
		IncidentID: activeIncident.ID,
		Step:       domain.IncidentEventStepAlert,
		Message:    stringPtr("Resolution notifications sent"),
	}

	if err := s.eventSteps.Create(ctx, alertStep); err != nil {
		log.Printf("Failed to create alert_sent event step for resolution: %v", err)
	}

	return nil
}

// sendNotifications sends notifications through all configured integrations.
func (s *IncidentService) sendNotifications(ctx context.Context, incident *domain.Incident, eventType string) error {
	// Create default in-app integration
	inAppIntegration := domain.Integration{
		Base:     domain.Base{ID: "inapp-default"},
		Name:     "In-App Notifications",
		Target:   "internal",
		Type:     "inapp",
		IsActive: true,
	}

	// Fetch all active integrations from database
	activeIntegrations, err := s.integrations.List(ctx, 100, 0)
	if err != nil {
		log.Printf("Failed to fetch integrations: %v", err)
		activeIntegrations = []*domain.Integration{}
	}

	// Filter for active integrations only
	var integrationsList []domain.Integration
	integrationsList = append(integrationsList, inAppIntegration) // Always include in-app

	for _, integration := range activeIntegrations {
		if integration.IsActive {
			integrationsList = append(integrationsList, *integration)
		}
	}

	// Send notification through each integration
	for _, integration := range integrationsList {
		// Get the appropriate notifier
		sender, err := notifier.NewNotifier(integration.Type)
		if err != nil {
			log.Printf("Failed to create notifier for type %s: %v", integration.Type, err)
			continue
		}

		// Send the notification
		if err := sender.Send(ctx, integration, *incident); err != nil {
			log.Printf("Failed to send notification via %s (%s): %v",
				integration.Name, integration.Type, err)
			continue
		}

		log.Printf("Successfully sent %s notification via %s", eventType, integration.Name)
	}

	return nil
}

// stringPtr is a helper to create string pointers.
func stringPtr(s string) *string {
	return &s
}

// enqueueNotification queues a notification task for the incident.
func (s *IncidentService) enqueueNotification(ctx context.Context, incident *domain.Incident, eventType string) error {
	payload := map[string]interface{}{
		"incident_id": incident.ID,
		"event_type":  eventType,
		"resource_id": incident.ResourceID,
		"timestamp":   time.Now().Unix(),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal notification payload: %w", err)
	}

	task := asynq.NewTask("notification:send", payloadBytes)
	_, err = s.client.Enqueue(task, asynq.Queue("notifications"))
	if err != nil {
		return fmt.Errorf("failed to enqueue notification task: %w", err)
	}

	return nil
}
