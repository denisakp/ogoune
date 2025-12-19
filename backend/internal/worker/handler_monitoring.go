package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/monitoring"
	"github.com/denisakp/pulseguard/internal/repository"
	"github.com/hibiken/asynq"
)

// MonitoringTaskHandler processes monitoring tasks from the Asynq queue.
// It executes health checks and handles status changes with direct service calls.
type MonitoringTaskHandler struct {
	resources    repository.ResourceRepository
	activities   repository.MonitoringActivityRepository
	maintenances repository.MaintenanceRepository
	executor     *domain.CheckExecutor
	incidents    *monitoring.IncidentService
}

// NewMonitoringTaskHandler creates a new monitoring task handler.
func NewMonitoringTaskHandler(
	resources repository.ResourceRepository,
	activities repository.MonitoringActivityRepository,
	maintenances repository.MaintenanceRepository,
	executor *domain.CheckExecutor,
	incidents *monitoring.IncidentService,
) *MonitoringTaskHandler {
	return &MonitoringTaskHandler{
		resources:    resources,
		activities:   activities,
		maintenances: maintenances,
		executor:     executor,
		incidents:    incidents,
	}
}

// ProcessTask processes a monitoring task from the queue.
func (h *MonitoringTaskHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	// Parse the task payload
	var payload struct {
		ResourceID string `json:"resource_id"`
		Type       string `json:"type"`
		Target     string `json:"target"`
		Timeout    int    `json:"timeout"`
	}

	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task payload: %w", err)
	}

	// Get the current resource from the database
	resource, err := h.resources.FindByID(ctx, payload.ResourceID)
	if err != nil {
		return fmt.Errorf("failed to find resource %s: %w", payload.ResourceID, err)
	}

	// Skip monitoring if the resource is no longer active
	if !resource.IsActive {
		return nil
	}

	// Store the previous status for state transitions
	previousStatus := resource.Status

	// Execute the health check
	result, err := h.executor.ExecuteCheck(resource)
	if err != nil {
		return fmt.Errorf("failed to execute check for resource %s: %w", resource.ID, err)
	}

	message := fmt.Sprintf("Check %s - Status: %s", resource.Type, result.Status)
	activity := &domain.MonitoringActivity{
		ResourceID:   resource.ID,
		Message:      message,
		Success:      result.Status == string(domain.StatusUp),
		ResponseTime: int(result.ResponseTime.Milliseconds()),
		ResponseData: []byte(result.ResponseData),
	}
	// Maintenance override: if resource under maintenance, mark activity and skip business transitions
	activeMaintenances, _ := h.maintenances.FindActiveForResource(ctx, resource.ID, time.Now())
	if len(activeMaintenances) > 0 {
		activity.IsMaintenance = true
		if err := h.activities.Create(ctx, activity); err != nil {
			fmt.Printf("Warning: failed to persist monitoring activity (maintenance): %v\n", err)
		}
		// Do not update resource status, failure counts, or incidents during maintenance
		return nil
	}
	if err := h.activities.Create(ctx, activity); err != nil {
		// Log error but continue processing - we don't want to block incident logic
		fmt.Printf("Warning: failed to persist monitoring activity: %v\n", err)
	}

	currentResultStatus := domain.ResourceStatus(result.Status)

	switch currentResultStatus {
	case domain.StatusUp:
		// Resource is UP
		if previousStatus == domain.StatusDown {
			// This is a RECOVERY - resource was down, now it's up
			if err := h.incidents.ResolveIncident(ctx, resource, result); err != nil {
				fmt.Printf("Warning: failed to resolve incident: %v\n", err)
			}
		}

		// Reset failure count and update status to UP
		resource.FailureCount = 0
		resource.Status = domain.StatusUp
		resource.LastChecked = &[]time.Time{time.Now()}[0]

		if err := h.resources.Update(ctx, resource); err != nil {
			return fmt.Errorf("failed to update resource status: %w", err)
		}

	case domain.StatusDown:
		// Resource is DOWN - increment failure count
		resource.FailureCount++
		resource.Status = domain.StatusDown
		resource.LastChecked = &[]time.Time{time.Now()}[0]

		if err := h.resources.Update(ctx, resource); err != nil {
			return fmt.Errorf("failed to update resource after failure: %w", err)
		}

		// Only create incident on the 3rd consecutive failure
		if resource.FailureCount == 3 {
			if err := h.incidents.CreateIncident(ctx, resource, result); err != nil {
				fmt.Printf("Warning: failed to create incident on 3rd failure: %v\n", err)
			}
		}
	default:
		// Handle other statuses (error, unknown, etc.)
		resource.Status = currentResultStatus
		resource.LastChecked = &[]time.Time{time.Now()}[0]

		if err := h.resources.Update(ctx, resource); err != nil {
			return fmt.Errorf("failed to update resource status: %w", err)
		}
	}

	return nil
}
