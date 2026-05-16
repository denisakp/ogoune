package scheduler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
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

// Scheduler is an alias for port.Scheduler.
type Scheduler = port.Scheduler

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

// AsynqSchedulerAdapter is an alias for port.AsynqSchedulerAdapter.
type AsynqSchedulerAdapter = port.AsynqSchedulerAdapter

// AsynqSchedulerAdapterWithInterval is an alias for port.AsynqSchedulerAdapterWithInterval.
type AsynqSchedulerAdapterWithInterval = port.AsynqSchedulerAdapterWithInterval

// ActiveResourceRepository is an alias for port.ActiveResourceRepository.
type ActiveResourceRepository = port.ActiveResourceRepository

// ScheduleItem is an alias for port.ScheduleItem.
type ScheduleItem = port.ScheduleItem

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
