package monitoring

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/hibiken/asynq"
)

// SchedulerService manages scheduling and unscheduling of monitoring tasks using Asynq.
// It uses Asynq's periodic task scheduler for recurring monitoring checks.
type SchedulerService struct {
	client    *asynq.Client
	inspector *asynq.Inspector
	scheduler *asynq.Scheduler
}

// NewSchedulerService creates a new scheduler service with Asynq client, inspector, and scheduler.
func NewSchedulerService(client *asynq.Client, inspector *asynq.Inspector, scheduler *asynq.Scheduler) *SchedulerService {
	return &SchedulerService{
		client:    client,
		inspector: inspector,
		scheduler: scheduler,
	}
}

// Schedule creates or updates a periodic monitoring task for the given resource.
// It first unschedules any existing task for the resource to handle updates correctly.
// The task is only scheduled if the resource is active.
// The task will run repeatedly at the interval specified in the resource (in seconds).
func (s *SchedulerService) Schedule(ctx context.Context, r *domain.Resource) error {
	if r == nil {
		return fmt.Errorf("resource cannot be nil")
	}

	// Generate unique entry ID based on resource ID
	entryID := fmt.Sprintf("monitor:%s", r.ID)

	// First, try to unschedule any existing task for this resource
	if err := s.unregisterPeriodicTask(entryID); err != nil {
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

	// Create the periodic task
	task := asynq.NewTask("monitoring:check", payloadBytes)

	// Convert interval from seconds to a cron expression
	// Use @every syntax for simplicity: @every 60s, @every 5m, etc.
	cronspec := fmt.Sprintf("@every %ds", r.Interval)

	// Register the periodic task
	_, err = s.scheduler.Register(
		cronspec,
		task,
		asynq.Queue("monitoring"),
		asynq.TaskID(entryID),
	)
	if err != nil {
		return fmt.Errorf("failed to register periodic monitoring task for resource %s: %w", r.ID, err)
	}

	return nil
}

// Unschedule removes the periodic monitoring task for the given resource ID.
func (s *SchedulerService) Unschedule(ctx context.Context, resourceID string) error {
	if resourceID == "" {
		return fmt.Errorf("resource ID cannot be empty")
	}

	entryID := fmt.Sprintf("monitor:%s", resourceID)
	return s.unregisterPeriodicTask(entryID)
}

// unregisterPeriodicTask is a helper method to remove a periodic task by entry ID.
func (s *SchedulerService) unregisterPeriodicTask(entryID string) error {
	// Unregister the periodic task from the scheduler
	err := s.scheduler.Unregister(entryID)
	if err != nil {
		// If the task doesn't exist, that's fine - we consider it successfully unscheduled
		// Asynq returns an error if trying to unregister a non-existent entry
		return nil // Don't fail on unregister errors
	}

	return nil
}
