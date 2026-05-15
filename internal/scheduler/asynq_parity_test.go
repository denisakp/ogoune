package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
)

func TestAsynqHostedParityPreservesDispatchInputs(t *testing.T) {
	adapter := &fakeAsynqSchedulerAdapter{}
	loader := func(ctx context.Context, resourceID string) (*domain.Resource, error) {
		return &domain.Resource{
			Base:     domain.Base{ID: resourceID},
			Name:     "Parity Monitor",
			Type:     domain.ResourceHTTP,
			Target:   "https://parity.example.com",
			Timeout:  20,
			Interval: 45,
			IsActive: true,
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

	if err := runtime.Schedule("parity-1", 90*time.Second); err != nil {
		t.Fatalf("Schedule() error = %v", err)
	}

	if adapter.scheduled == nil {
		t.Fatal("expected delegated resource")
	}
	if adapter.scheduled.Type != domain.ResourceHTTP {
		t.Fatalf("expected resource type parity, got %s", adapter.scheduled.Type)
	}
	if adapter.scheduled.Target != "https://parity.example.com" {
		t.Fatalf("expected target parity, got %s", adapter.scheduled.Target)
	}
	if adapter.scheduled.Timeout != 20 {
		t.Fatalf("expected timeout parity, got %d", adapter.scheduled.Timeout)
	}
	if adapter.scheduled.Interval != 90 {
		t.Fatalf("expected interval override parity, got %d", adapter.scheduled.Interval)
	}
}

func TestAsynqHostedParityPreservesInactiveUnscheduleSemantics(t *testing.T) {
	adapter := &fakeAsynqSchedulerAdapter{}
	runtime, err := NewAsynq(&Config{
		Mode: ModeAsynq,
		Asynq: AsynqConfig{
			RedisURL: "redis://localhost:6379",
			ResourceLoader: func(ctx context.Context, resourceID string) (*domain.Resource, error) {
				return &domain.Resource{Base: domain.Base{ID: resourceID}, Interval: 60, IsActive: false}, nil
			},
			SchedulerAdapter: adapter,
		},
	})
	if err != nil {
		t.Fatalf("NewAsynq() error = %v", err)
	}

	if err := runtime.Schedule("inactive-1", 60*time.Second); err != nil {
		t.Fatalf("Schedule() error = %v", err)
	}
	if adapter.scheduled == nil || adapter.scheduled.IsActive {
		t.Fatalf("expected inactive resource to be delegated unchanged, got %+v", adapter.scheduled)
	}

	if err := runtime.Pause("inactive-1"); err != nil {
		t.Fatalf("Pause() error = %v", err)
	}
	if adapter.unscheduledID != "inactive-1" {
		t.Fatalf("expected hosted unschedule parity for pause, got %s", adapter.unscheduledID)
	}
}

type fakeAsynqSchedulerAdapter struct {
	scheduled     *domain.Resource
	unscheduledID string
}

func (f *fakeAsynqSchedulerAdapter) Schedule(ctx context.Context, r *domain.Resource) error {
	clone := *r
	f.scheduled = &clone
	return nil
}

func (f *fakeAsynqSchedulerAdapter) Unschedule(ctx context.Context, resourceID string) error {
	f.unscheduledID = resourceID
	return nil
}
