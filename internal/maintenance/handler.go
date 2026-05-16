package maintenance

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/denisakp/ogoune/internal/repository"
	"github.com/hibiken/asynq"
)

// TaskHandler processes maintenance start/end tasks
type TaskHandler struct {
	repo     repository.MaintenanceRepository
	enqueuer TaskEnqueuer
}

func NewTaskHandler(repo repository.MaintenanceRepository, enqueuer TaskEnqueuer) *TaskHandler {
	return &TaskHandler{repo: repo, enqueuer: enqueuer}
}

// ProcessStart activates a maintenance and schedules its end if cron-based.
func (h *TaskHandler) ProcessStart(ctx context.Context, task *asynq.Task) error {
	var payload struct {
		MaintenanceID string `json:"maintenance_id"`
	}
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("maintenance:start: invalid payload: %w", err)
	}
	m, err := h.repo.FindByID(ctx, payload.MaintenanceID)
	if err != nil {
		return err
	}
	now := time.Now()
	m.Status = "active"
	m.StartedAt = &now
	if err := h.repo.Update(ctx, m); err != nil {
		return err
	}
	// If cron-based, schedule an end task after window duration
	if m.WindowMinutes != nil {
		endPayload := map[string]any{"maintenance_id": m.ID}
		endBytes, _ := json.Marshal(endPayload)
		endTask := asynq.NewTask("maintenance:end", endBytes)
		delay := time.Duration(*m.WindowMinutes) * time.Minute
		_, _ = h.enqueuer.Enqueue(endTask, asynq.Queue("maintenance"), asynq.ProcessIn(delay))
	}
	return nil
}

// ProcessEnd marks a maintenance as finished.
func (h *TaskHandler) ProcessEnd(ctx context.Context, task *asynq.Task) error {
	var payload struct {
		MaintenanceID string `json:"maintenance_id"`
	}
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("maintenance:end: invalid payload: %w", err)
	}
	m, err := h.repo.FindByID(ctx, payload.MaintenanceID)
	if err != nil {
		return err
	}
	now := time.Now()
	m.Status = "finished"
	m.EndedAt = &now
	return h.repo.Update(ctx, m)
}

// no helper: use json.Marshal inline to avoid redeclaration
