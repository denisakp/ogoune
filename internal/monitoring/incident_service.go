package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/repository"
	"github.com/hibiken/asynq"
)

// IncidentService manages incident creation and resolution based on resource status changes.
// It replaces the event-driven listener pattern with direct service calls.
type IncidentService struct {
	incidents repository.IncidentRepository
	client    *asynq.Client
}

// NewIncidentService creates a new incident service with the given dependencies.
func NewIncidentService(incidents repository.IncidentRepository, client *asynq.Client) *IncidentService {
	return &IncidentService{
		incidents: incidents,
		client:    client,
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
	if !hasActiveIncident {
		incident := &domain.Incident{
			ResourceID: r.ID,
			Reason:     "Resource status changed from UP to DOWN",
			IsResolved: false,
			StartedAt:  time.Now(),
			Details:    []byte(result.ResponseData),
		}

		if err := s.incidents.Create(ctx, incident); err != nil {
			return fmt.Errorf("failed to create incident: %w", err)
		}

		// Enqueue notification task
		if err := s.enqueueNotification(ctx, incident, "down"); err != nil {
			// Log the error but don't fail the incident creation
			return fmt.Errorf("incident created but failed to enqueue notification: %w", err)
		}
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

	// Resolve the incident if found
	if activeIncident != nil {
		now := time.Now()
		activeIncident.IsResolved = true
		activeIncident.ResolvedAt = &now

		if err := s.incidents.Update(ctx, activeIncident); err != nil {
			return fmt.Errorf("failed to resolve incident: %w", err)
		}

		// Enqueue notification task
		if err := s.enqueueNotification(ctx, activeIncident, "up"); err != nil {
			// Log the error but don't fail the incident resolution
			return fmt.Errorf("incident resolved but failed to enqueue notification: %w", err)
		}
	}

	return nil
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
