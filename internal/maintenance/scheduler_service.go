package maintenance

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository"
	"github.com/hibiken/asynq"
)

// SchedulerService schedules maintenance windows using Asynq.
type SchedulerService struct {
	client    *asynq.Client
	inspector *asynq.Inspector
	scheduler *asynq.Scheduler
	repo      repository.MaintenanceRepository
}

func NewSchedulerService(client *asynq.Client, inspector *asynq.Inspector, scheduler *asynq.Scheduler, repo repository.MaintenanceRepository) *SchedulerService {
	return &SchedulerService{client: client, inspector: inspector, scheduler: scheduler, repo: repo}
}

// EnsureScheduled registers periodic and one-time start tasks for maintenances.
func (s *SchedulerService) EnsureScheduled(ctx context.Context) error {
	maintenances, err := s.repo.List(ctx, "", 10000, 0)
	if err != nil {
		return err
	}
	for _, m := range maintenances {
		if m.Strategy == domain.Cron && m.CronExpr != nil && m.WindowMinutes != nil {
			payload := map[string]any{"maintenance_id": m.ID}
			task := asynq.NewTask("maintenance:start", mustJSON(payload))
			entryID := fmt.Sprintf("maintenance:start:%s", m.ID)
			_, err := s.scheduler.Register(*m.CronExpr, task, asynq.Queue("maintenance"), asynq.TaskID(entryID))
			if err != nil {
				// ignore duplicate or failure; continue
			}
			continue
		}
		if m.Strategy == domain.OneTime && m.Status == "scheduled" {
			// schedule start
			if m.StartAt != nil && m.StartAt.After(time.Now()) {
				payload := map[string]any{"maintenance_id": m.ID}
				startBytes, _ := json.Marshal(payload)
				task := asynq.NewTask("maintenance:start", startBytes)
				_, _ = s.client.Enqueue(task, asynq.Queue("maintenance"), asynq.ProcessAt(*m.StartAt))
			}
			// schedule end if EndAt provided
			if m.EndAt != nil && m.EndAt.After(time.Now()) {
				payload := map[string]any{"maintenance_id": m.ID}
				endBytes, _ := json.Marshal(payload)
				task := asynq.NewTask("maintenance:end", endBytes)
				_, _ = s.client.Enqueue(task, asynq.Queue("maintenance"), asynq.ProcessAt(*m.EndAt))
			}
		}
	}
	return nil
}

func mustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}
