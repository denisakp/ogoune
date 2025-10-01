package monitoring

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/hibiken/asynq"
)

// SchedulerService manages scheduling and unscheduling of monitoring tasks using Asynq.
// It replaces the complex event-driven dispatcher/listener pattern with direct service calls.
type SchedulerService struct {
	client    *asynq.Client
	inspector *asynq.Inspector
}

// NewSchedulerService creates a new scheduler service with Asynq client and inspector.
func NewSchedulerService(client *asynq.Client, inspector *asynq.Inspector) *SchedulerService {
	return &SchedulerService{
		client:    client,
		inspector: inspector,
	}
}

// Schedule creates or updates a periodic monitoring task for the given resource.
// It first unschedules any existing task for the resource to handle updates correctly.
// The task is only scheduled if the resource is active.
func (s *SchedulerService) Schedule(ctx context.Context, r *domain.Resource) error {
	if r == nil {
		return fmt.Errorf("resource cannot be nil")
	}

	// Generate unique task name based on resource ID
	taskName := fmt.Sprintf("monitor:%s", r.ID)

	// First, try to unschedule any existing task for this resource
	if err := s.unscheduleTask(ctx, taskName); err != nil {
		// Log the error but don't fail the entire operation
		// The task might not exist, which is fine
	}

	// Only schedule the task if the resource is active
	if !r.IsActive {
		return nil // Successfully handled - no task needed for inactive resource
	}

	// Create the monitoring task payload
	payload := map[string]interface{}{
		"resource_id": r.ID,
		"type":        string(r.Type),
		"target":      r.Target,
		"timeout":     r.Timeout,
	}

	// Convert payload to JSON bytes
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	task := asynq.NewTask("monitoring:check", payloadBytes)

	// Enqueue the task with options
	_, err = s.client.Enqueue(task, asynq.TaskID(taskName), asynq.Queue("monitoring"))
	if err != nil {
		return fmt.Errorf("failed to schedule monitoring task for resource %s: %w", r.ID, err)
	}

	return nil
}

// Unschedule removes the periodic monitoring task for the given resource ID.
func (s *SchedulerService) Unschedule(ctx context.Context, resourceID string) error {
	if resourceID == "" {
		return fmt.Errorf("resource ID cannot be empty")
	}

	taskName := fmt.Sprintf("monitor:%s", resourceID)
	return s.unscheduleTask(ctx, taskName)
}

// unscheduleTask is a helper method to remove a specific task by name.
func (s *SchedulerService) unscheduleTask(ctx context.Context, taskName string) error {
	// Try to cancel the task using the inspector
	err := s.inspector.DeleteTask("monitoring", taskName)
	if err != nil {
		// If the task doesn't exist, that's fine - we consider it successfully unscheduled
		if err.Error() == "task not found" || err.Error() == "queue not found" {
			return nil
		}
		return fmt.Errorf("failed to unschedule task %s: %w", taskName, err)
	}

	return nil
}
