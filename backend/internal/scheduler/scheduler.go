package scheduler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
)

// ScheduleMode represents the scheduler implementation mode.
type ScheduleMode string

type NotificationEventType string

const (
	ModeTimingWheel ScheduleMode = "timingwheel"
	ModeAsynq       ScheduleMode = "asynq"

	NotificationEventResourceDownAlert NotificationEventType = "resource_down_alert"
	NotificationEventResourceUpAlert   NotificationEventType = "resource_up_alert"
)

// Scheduler defines the interface for scheduling monitoring checks.
type Scheduler interface {
	// Start initializes and starts the scheduler.
	Start(ctx context.Context, repo ActiveResourceRepository) error
	// Schedule adds or updates a resource's check schedule.
	Schedule(resourceID string, interval time.Duration) error
	// Unschedule removes a resource from the scheduling queue.
	Unschedule(resourceID string) error
	// Pause temporarily stops scheduling for a resource.
	Pause(resourceID string) error
	// Resume resumes scheduling for a paused resource.
	Resume(resourceID string) error
	// Stop gracefully shuts down the scheduler.
	Stop(ctx context.Context) error
}

// Config holds the scheduler configuration.
type Config struct {
	Mode            ScheduleMode
	TimingWheel     TimingWheelConfig
	Asynq           AsynqConfig
	ShutdownTimeout time.Duration
}

// TimingWheelConfig holds TimingWheel-specific configuration.
type TimingWheelConfig struct {
	TickInterval          time.Duration
	MaxWorkers            int
	ShutdownTimeout       time.Duration
	NotificationQueueSize int
}

// AsynqConfig holds Asynq-specific configuration.
type AsynqConfig struct {
	RedisURL              string
	Concurrency           int
	NotificationQueueSize int
	ShutdownTimeout       time.Duration
	ResourceLoader        AsynqResourceLoader
	SchedulerAdapter      AsynqSchedulerAdapter
}

// AsynqResourceLoader resolves a monitored resource for hosted Asynq scheduling.
type AsynqResourceLoader func(ctx context.Context, resourceID string) (*domain.Resource, error)

// AsynqSchedulerAdapter bridges the scheduler runtime to the existing Asynq-backed scheduler service.
type AsynqSchedulerAdapter interface {
	Schedule(ctx context.Context, r *domain.Resource) error
	Unschedule(ctx context.Context, resourceID string) error
}

// ActiveResourceRepository defines the interface for accessing active resources.
type ActiveResourceRepository interface {
	FindScheduledResources(ctx context.Context) ([]ScheduleItem, error)
}

// ScheduleItem represents a schedulable resource.
type ScheduleItem struct {
	ResourceID string
	Interval   time.Duration
	Paused     bool
}

// Error types
var (
	ErrSchedulerAlreadyRunning = errors.New("scheduler is already running")
	ErrSchedulerNotRunning     = errors.New("scheduler is not running")
	ErrInvalidInterval         = errors.New("invalid interval: must be > 0")
	ErrRedisRequired           = errors.New("Redis is required for Asynq mode")
	ErrInvalidMode             = errors.New("invalid scheduler mode")
	ErrInvalidNotificationType = errors.New("invalid notification event type")
	ErrShutdownTimeout         = errors.New("scheduler shutdown timeout exceeded")
	ErrAsynqAdapterUnavailable = errors.New("asynq scheduler adapter is not configured")
)

// New creates a new scheduler instance based on configuration.
func New(cfg *Config) (Scheduler, error) {
	if cfg == nil {
		cfg = &Config{
			Mode: ModeTimingWheel,
			TimingWheel: TimingWheelConfig{
				TickInterval:          1 * time.Second,
				MaxWorkers:            10,
				ShutdownTimeout:       15 * time.Second,
				NotificationQueueSize: 100,
			},
		}
	}

	mode := cfg.Mode
	if mode == "" {
		mode = ModeTimingWheel // default
	}

	switch mode {
	case ModeTimingWheel:
		return NewTimingWheel(cfg)
	case ModeAsynq:
		// Validate Redis availability for Asynq mode
		if cfg.Asynq.RedisURL == "" {
			return nil, ErrRedisRequired
		}
		return NewAsynq(cfg)
	default:
		return nil, fmt.Errorf("%w: %s", ErrInvalidMode, mode)
	}
}

// DetectMode determines scheduler mode based on environment and configuration.
func DetectMode(explicitMode string, dbDriver string) ScheduleMode {
	// 1. Explicit mode takes precedence
	if explicitMode != "" {
		if ScheduleMode(explicitMode) == ModeAsynq {
			return ModeAsynq
		}
		return ModeTimingWheel
	}

	// 2. SQLite defaults to TimingWheel
	if dbDriver == "sqlite" {
		return ModeTimingWheel
	}

	// 3. Empty driver defaults to the safest in-process mode for direct constructor use.
	if dbDriver == "" {
		return ModeTimingWheel
	}

	// 4. Hosted deployments default to Asynq.
	return ModeAsynq
}

// ValidateSelector validates that the scheduler mode is compatible with environment settings.
func ValidateSelector(mode ScheduleMode, redisURL string) error {
	if mode == ModeAsynq && redisURL == "" {
		return ErrRedisRequired
	}
	return nil
}

func ValidateNotificationEventType(eventType string) error {
	switch NotificationEventType(eventType) {
	case NotificationEventResourceDownAlert, NotificationEventResourceUpAlert:
		return nil
	default:
		return fmt.Errorf("%w: %s", ErrInvalidNotificationType, eventType)
	}
}
