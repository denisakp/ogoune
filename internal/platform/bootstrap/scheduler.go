package bootstrap

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/denisakp/ogoune/internal/config"
	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/maintenance"
	"github.com/denisakp/ogoune/internal/monitoring"
	"github.com/denisakp/ogoune/internal/scheduler"
	"github.com/hibiken/asynq"
)

// InitScheduler configures and creates the scheduler based on runtime mode.
func InitScheduler(app *App) {
	cfg := app.Cfg

	schedulerCfg := &scheduler.Config{
		Mode: scheduler.ScheduleMode(config.GetEnv("SCHEDULER_MODE", "")),
		TimingWheel: scheduler.TimingWheelConfig{
			TickInterval:          1 * time.Second,
			MaxWorkers:            10,
			ShutdownTimeout:       15 * time.Second,
			NotificationQueueSize: 100,
		},
		Asynq: scheduler.AsynqConfig{
			RedisURL:    config.GetEnv("REDIS_URL", "localhost:6379"),
			Concurrency: 10,
		},
	}

	// Default to timingwheel for SQLite, asynq for PostgreSQL
	if schedulerCfg.Mode == "" {
		if strings.ToLower(cfg.DBDriver) == "sqlite" {
			schedulerCfg.Mode = scheduler.ModeTimingWheel
		} else {
			schedulerCfg.Mode = scheduler.ModeAsynq
		}
	}

	app.SchedulerCfg = schedulerCfg
	slog.Info("starting with scheduler", "mode", schedulerCfg.Mode)

	var err error

	switch schedulerCfg.Mode {
	case scheduler.ModeTimingWheel:
		app.RuntimeScheduler, err = scheduler.New(schedulerCfg)
		if err != nil {
			slog.Error("failed to create scheduler", "error", err)
			os.Exit(1)
		}

		slog.Info("using TimingWheel scheduler (Community Edition - no Redis required)")
		app.SchedulerAdapter = scheduler.NewRepositorySchedulerAdapter(app.RuntimeScheduler)
		if rs, ok := app.SchedulerAdapter.(interface {
			ScheduleWithInterval(ctx context.Context, resource *domain.Resource, interval time.Duration) error
		}); ok {
			app.ConfirmationScheduler = rs
		}

		app.MaintenanceScheduler = nil

	case scheduler.ModeAsynq:
		slog.Info("using Asynq scheduler (SaaS Edition with Redis)")

		redisOpt := asynq.RedisClientOpt{
			Addr: config.GetEnv("REDIS_URL", "localhost:6379"),
		}
		app.RedisOpt = redisOpt

		app.AsynqClient = asynq.NewClient(redisOpt)
		app.AsynqInspector = asynq.NewInspector(redisOpt)
		app.AsynqScheduler = asynq.NewScheduler(redisOpt, nil)

		// Start the scheduler in a goroutine
		go func() {
			if err := app.AsynqScheduler.Run(); err != nil {
				slog.Error("failed to start Asynq scheduler", "error", err)
				os.Exit(1)
			}
		}()

		// Initialize monitoring scheduler service (Asynq-based)
		schedulerService := monitoring.NewSchedulerService(app.AsynqClient, app.AsynqInspector, app.AsynqScheduler)
		schedulerCfg.Asynq.ResourceLoader = func(ctx context.Context, resourceID string) (*domain.Resource, error) {
			return app.ResourceRepo.FindByID(ctx, resourceID)
		}
		schedulerCfg.Asynq.SchedulerAdapter = schedulerService

		app.RuntimeScheduler, err = scheduler.New(schedulerCfg)
		if err != nil {
			slog.Error("failed to create scheduler", "error", err)
			os.Exit(1)
		}

		app.SchedulerAdapter = scheduler.NewRepositorySchedulerAdapter(app.RuntimeScheduler)
		if rs, ok := app.SchedulerAdapter.(interface {
			ScheduleWithInterval(ctx context.Context, resource *domain.Resource, interval time.Duration) error
		}); ok {
			app.ConfirmationScheduler = rs
		}

		// Initialize maintenance scheduler
		app.MaintenanceScheduler = maintenance.NewSchedulerService(
			&maintenance.AsynqClientAdapter{Client: app.AsynqClient},
			&maintenance.AsynqSchedulerAdapter{Scheduler: app.AsynqScheduler},
			app.MaintenanceRepo,
		)
		if err := app.MaintenanceScheduler.EnsureScheduled(context.Background()); err != nil {
			slog.Error("failed to ensure maintenance schedules", "error", err)
		}

	default:
		slog.Error("invalid scheduler mode", "mode", schedulerCfg.Mode)
		os.Exit(1)
	}
}

