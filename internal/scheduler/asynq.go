package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
)

// Asynq represents a Redis-based scheduler interface adapter.
type Asynq struct {
	config         *Config
	state          State
	resourceLoader AsynqResourceLoader
	adapter        AsynqSchedulerAdapter
}

// NewAsynq creates a new Asynq scheduler adapter instance.
func NewAsynq(cfg *Config) (*Asynq, error) {
	if cfg == nil {
		cfg = &Config{Mode: ModeAsynq}
	}

	if cfg.Asynq.RedisURL == "" {
		return nil, ErrRedisRequired
	}

	return &Asynq{
		config:         cfg,
		state:          StateStopped,
		resourceLoader: cfg.Asynq.ResourceLoader,
		adapter:        cfg.Asynq.SchedulerAdapter,
	}, nil
}

// Start initializes and starts the Asynq scheduler.
func (a *Asynq) Start(ctx context.Context, repo ActiveResourceRepository) error {
	if a.state != StateStopped {
		return ErrSchedulerAlreadyRunning
	}

	a.state = StateRunning
	return nil
}

// Stop gracefully shuts down the Asynq scheduler.
func (a *Asynq) Stop(ctx context.Context) error {
	if a.state != StateRunning {
		return ErrSchedulerNotRunning
	}

	a.state = StateStopped
	return nil
}

// Schedule adds or updates a resource's check schedule.
func (a *Asynq) Schedule(resourceID string, interval time.Duration) error {
	if interval <= 0 {
		return ErrInvalidInterval
	}

	resource, err := a.loadResource(resourceID)
	if err != nil {
		return err
	}

	resource.Interval = int(interval / time.Second)
	if resource.Interval <= 0 {
		resource.Interval = 1
	}

	if withInterval, ok := a.adapter.(AsynqSchedulerAdapterWithInterval); ok {
		return withInterval.ScheduleWithInterval(context.Background(), resource, interval)
	}

	return a.adapter.Schedule(context.Background(), resource)
}

// Unschedule removes a resource from the scheduling queue.
func (a *Asynq) Unschedule(resourceID string) error {
	if a.adapter == nil {
		return ErrAsynqAdapterUnavailable
	}

	return a.adapter.Unschedule(context.Background(), resourceID)
}

// Pause temporarily stops scheduling for a resource.
func (a *Asynq) Pause(resourceID string) error {
	return a.Unschedule(resourceID)
}

// Resume resumes scheduling for a paused resource.
func (a *Asynq) Resume(resourceID string) error {
	resource, err := a.loadResource(resourceID)
	if err != nil {
		return err
	}

	if resource.Interval <= 0 {
		return ErrInvalidInterval
	}

	resource.IsActive = true
	return a.adapter.Schedule(context.Background(), resource)
}

func (a *Asynq) loadResource(resourceID string) (*domain.Resource, error) {
	if resourceID == "" {
		return nil, fmt.Errorf("resource ID cannot be empty")
	}

	if a.adapter == nil || a.resourceLoader == nil {
		return nil, ErrAsynqAdapterUnavailable
	}

	resource, err := a.resourceLoader(context.Background(), resourceID)
	if err != nil {
		return nil, err
	}

	if resource == nil {
		return nil, fmt.Errorf("resource %s not found", resourceID)
	}

	return resource, nil
}
