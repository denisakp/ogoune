package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
)

func TestAsynqAdapterScheduleUnscheduleParity(t *testing.T) {
	adapter := &fakeAsynqSchedulerAdapter{}
	loader := func(ctx context.Context, resourceID string) (*domain.Resource, error) {
		return &domain.Resource{
			Base:     domain.Base{ID: resourceID},
			Name:     "Hosted Monitor",
			Type:     domain.ResourceHTTP,
			Target:   "https://example.com/health",
			Timeout:  15,
			Interval: 60,
			IsActive: true,
		}, nil
	}

	scheduler, err := New(&Config{
		Mode: ModeAsynq,
		Asynq: AsynqConfig{
			RedisURL:         "redis://localhost:6379",
			ResourceLoader:   loader,
			SchedulerAdapter: adapter,
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if err := scheduler.Start(context.Background(), nil); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	if err := scheduler.Schedule("res-1", 120*time.Second); err != nil {
		t.Fatalf("Schedule() error = %v", err)
	}

	if adapter.scheduled == nil {
		t.Fatal("expected resource to be delegated for scheduling")
	}
	if adapter.scheduled.ID != "res-1" {
		t.Fatalf("expected resource ID res-1, got %s", adapter.scheduled.ID)
	}
	if adapter.scheduled.Interval != 120 {
		t.Fatalf("expected delegated interval 120 seconds, got %d", adapter.scheduled.Interval)
	}
	if adapter.scheduled.Type != domain.ResourceHTTP || adapter.scheduled.Target != "https://example.com/health" || adapter.scheduled.Timeout != 15 {
		t.Fatalf("expected hosted payload parity to be preserved, got %+v", adapter.scheduled)
	}

	if err := scheduler.Unschedule("res-1"); err != nil {
		t.Fatalf("Unschedule() error = %v", err)
	}
	if adapter.unscheduledID != "res-1" {
		t.Fatalf("expected unscheduled ID res-1, got %s", adapter.unscheduledID)
	}
}

func TestAsynqAdapterPauseResumeParity(t *testing.T) {
	adapter := &fakeAsynqSchedulerAdapter{}
	loader := func(ctx context.Context, resourceID string) (*domain.Resource, error) {
		return &domain.Resource{
			Base:     domain.Base{ID: resourceID},
			Type:     domain.ResourceTCP,
			Target:   "127.0.0.1:443",
			Timeout:  5,
			Interval: 30,
			IsActive: false,
		}, nil
	}

	runtime, err := NewAsynq(&Config{
		Mode: ModeAsynq,
		Asynq: AsynqConfig{
			RedisURL:         "redis://localhost:6379",
			ResourceLoader:   loader,
			SchedulerAdapter: adapter,
		},
	})
	if err != nil {
		t.Fatalf("NewAsynq() error = %v", err)
	}

	if err := runtime.Pause("res-2"); err != nil {
		t.Fatalf("Pause() error = %v", err)
	}
	if adapter.unscheduledID != "res-2" {
		t.Fatalf("expected pause to unschedule res-2, got %s", adapter.unscheduledID)
	}

	if err := runtime.Resume("res-2"); err != nil {
		t.Fatalf("Resume() error = %v", err)
	}
	if adapter.scheduled == nil || !adapter.scheduled.IsActive || adapter.scheduled.Interval != 30 {
		t.Fatalf("expected resume to schedule active resource with preserved interval, got %+v", adapter.scheduled)
	}
}
