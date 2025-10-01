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
	resources  repository.ResourceRepository
	activities repository.MonitoringActivityRepository
	executor   *monitoring.Executor
	incidents  *monitoring.IncidentService
}

// NewMonitoringTaskHandler creates a new monitoring task handler.
func NewMonitoringTaskHandler(
	resources repository.ResourceRepository,
	activities repository.MonitoringActivityRepository,
	executor *monitoring.Executor,
	incidents *monitoring.IncidentService,
) *MonitoringTaskHandler {
	return &MonitoringTaskHandler{
		resources:  resources,
		activities: activities,
		executor:   executor,
		incidents:  incidents,
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

	// Store the old status for comparison
	oldStatus := resource.Status

	// Execute the health check
	result, err := h.executor.ExecuteCheck(resource)
	if err != nil {
		return fmt.Errorf("failed to execute check for resource %s: %w", resource.ID, err)
	}

	// Update resource status and metadata
	resource.Status = domain.ResourceStatus(result.Status)
	resource.LastChecked = &[]time.Time{time.Now()}[0]

	// Update failure count based on status
	if resource.Status == domain.StatusDown {
		resource.FailureCount++
	} else if resource.Status == domain.StatusUp {
		resource.FailureCount = 0 // Reset failure count on successful check
	}

	// Save the updated resource
	if err := h.resources.Update(ctx, resource); err != nil {
		return fmt.Errorf("failed to update resource %s: %w", resource.ID, err)
	}

	// Persist monitoring activity for traceability
	message := fmt.Sprintf("Check %s - Status: %s", resource.Type, result.Status)
	activity := &domain.MonitoringActivity{
		ResourceID:   resource.ID,
		Message:      message,
		Success:      result.Status == string(domain.StatusUp),
		ResponseTime: int(result.ResponseTime.Milliseconds()),
		ResponseData: []byte(result.ResponseData),
	}
	if err := h.activities.Create(ctx, activity); err != nil {
		// Log error but don't fail the monitoring task
		return fmt.Errorf("monitoring completed but failed to persist activity: %w", err)
	}

	// Handle status changes by calling the incident service directly
	if oldStatus != resource.Status {
		if err := h.incidents.HandleStatusChange(ctx, resource, oldStatus, result); err != nil {
			// Log the error but don't fail the monitoring task
			return fmt.Errorf("monitoring completed but incident handling failed: %w", err)
		}
	}

	return nil
}
