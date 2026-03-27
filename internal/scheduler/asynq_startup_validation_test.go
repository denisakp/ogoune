package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
)

func TestAsynqStartupValidationRequiresRedis(t *testing.T) {
	_, err := New(&Config{Mode: ModeAsynq})
	if err != ErrRedisRequired {
		t.Fatalf("expected ErrRedisRequired, got %v", err)
	}
}

func TestAsynqStartupValidationWithRedisSucceeds(t *testing.T) {
	runtime, err := New(&Config{
		Mode: ModeAsynq,
		Asynq: AsynqConfig{
			RedisURL: "redis://localhost:6379",
			ResourceLoader: func(ctx context.Context, resourceID string) (*domain.Resource, error) {
				return &domain.Resource{Base: domain.Base{ID: resourceID}, Interval: 60, IsActive: true}, nil
			},
			SchedulerAdapter: &fakeAsynqSchedulerAdapter{},
		},
	})
	if err != nil {
		t.Fatalf("expected hosted Asynq runtime to be created, got %v", err)
	}

	if err := runtime.Start(context.Background(), nil); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := runtime.Stop(shutdownCtx); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
}
