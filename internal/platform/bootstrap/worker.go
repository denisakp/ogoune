package bootstrap

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/maintenance"
	"github.com/denisakp/ogoune/internal/monitoring"
	"github.com/denisakp/ogoune/internal/repository/store"
	"github.com/denisakp/ogoune/internal/scheduler"
	"github.com/denisakp/ogoune/internal/service"
	"github.com/denisakp/ogoune/internal/worker"
	"github.com/hibiken/asynq"
)

// InitWorker starts the scheduler runtime, dispatchers, heartbeat detector, and Asynq processor.
func InitWorker(app *App) {
	// Initialize enrichment and component services needed by worker
	enrichmentService := service.NewEnrichmentService(30 * time.Second)
	app.ComponentService = service.NewComponentService(app.ComponentRepo, app.ResourceRepo, app.NotificationChannelRepo)
	pendingRetryService := service.NewPendingNotificationRetryService(
		app.NotificationRepo,
		app.IncidentRepo,
		app.NotificationChannelRepo,
		app.ComponentRepo,
		"",
		24*time.Hour,
	)

	// Start scheduler runtime
	if err := app.RuntimeScheduler.Start(context.Background(), NewResourceRepositoryAdapter(app.ResourceRepo)); err != nil {
		slog.Error("failed to start scheduler runtime", "error", err)
		os.Exit(1)
	}

	if app.SchedulerCfg.Mode == scheduler.ModeTimingWheel {
		slog.Info("TimingWheel scheduler started")

		activeResources, err := app.ResourceRepo.FindActive(context.Background(), 10000, 0)
		if err == nil {
			slog.Info("TimingWheel loaded active resources", "count", len(activeResources))
		}

		// Start in-process dispatcher
		if tw, ok := app.RuntimeScheduler.(*scheduler.TimingWheelScheduler); ok {
			strategies := BuildStrategies()
			executor := domain.NewCheckExecutor(strategies, app.MetricsRecorder)

			incidentService := monitoring.NewIncidentService(
				app.IncidentRepo,
				app.IncidentEventStepRepo,
				app.NotificationRepo,
				app.NotificationChannelRepo,
				app.IncidentDiagnosticsRepo,
				nil,
			)
			app.DetectorIncidentSvc = incidentService

			monitoringHandler := worker.NewMonitoringTaskHandler(app.ResourceRepo, app.MonitoringActivityRepo, app.MaintenanceRepo, app.IncidentDiagnosticsRepo, executor, incidentService, app.ComponentService, app.ConfirmationScheduler)

			workers := app.SchedulerCfg.TimingWheel.MaxWorkers
			if workers <= 0 {
				workers = 1
			}

			for i := 0; i < workers; i++ {
				go func() {
					for job := range tw.CheckJobs() {
						if job == nil || job.ResourceID == "" {
							continue
						}

						payload, err := json.Marshal(map[string]string{"resource_id": job.ResourceID})
						if err != nil {
							slog.Error("failed to marshal timingwheel payload", "resource_id", job.ResourceID, "error", err)
							continue
						}

						task := asynq.NewTask("monitoring:check", payload)
						if err := monitoringHandler.ProcessTask(context.Background(), task); err != nil {
							slog.Error("TimingWheel check failed", "resource_id", job.ResourceID, "error", err)
						}
					}
				}()
			}

			slog.Info("TimingWheel check dispatcher started", "workers", workers)

			// Daily expiry check in TimingWheel mode
			twExpiryNotificationLogRepo := store.NewExpiryNotificationLogRepository(app.DB)
			twExpiryService := service.NewExpiryNotificationService(
				twExpiryNotificationLogRepo,
				app.NotificationChannelRepo,
				service.ParseGlobalThresholds(app.Cfg.ExpiryAlertThresholds),
			)
			twExpiryHandler := worker.NewExpiryTaskHandler(app.ResourceRepo, app.NotificationChannelRepo, enrichmentService, twExpiryService)
			go func() {
				ticker := time.NewTicker(24 * time.Hour)
				defer ticker.Stop()
				for range ticker.C {
					if err := twExpiryHandler.ProcessTask(context.Background(), asynq.NewTask(worker.TypeExpiryCheck, nil)); err != nil {
						slog.Error("TimingWheel expiry:check failed", "error", err)
					}
				}
			}()
			slog.Info("TimingWheel daily expiry check scheduled")
		}

	} else {
		slog.Info("bootstrapping: scheduling active resources with Asynq")

		bootstrapCtx := context.Background()
		activeResources, err := app.ResourceRepo.FindActive(bootstrapCtx, 10000, 0)
		if err != nil {
			slog.Error("failed to fetch active resources during bootstrap", "error", err)
			os.Exit(1)
		}

		slog.Info("found active resources to schedule", "count", len(activeResources))

		successCount := 0
		failureCount := 0

		for _, resource := range activeResources {
			slog.Debug("scheduling resource", "name", resource.Name, "resource_id", resource.ID)

			if err := app.SchedulerAdapter.Schedule(bootstrapCtx, resource); err != nil {
				slog.Error("failed to schedule resource", "resource_id", resource.ID, "error", err)
				failureCount++
			} else {
				slog.Debug("successfully scheduled resource", "resource_id", resource.ID)
				successCount++
			}
		}

		slog.Info("bootstrap completed",
			"total", len(activeResources),
			"success", successCount,
			"failed", failureCount,
		)

		if failureCount > 0 {
			slog.Warn("some resources failed to schedule", "failed", failureCount)
		}
	}

	runStartupPendingNotificationRetry(context.Background(), pendingRetryService)

	// Start heartbeat detector
	heartbeatDetector := service.NewHeartbeatDetectorService(app.ResourceRepo, app.DetectorIncidentSvc)
	if err := startHeartbeatDetector(context.Background(), heartbeatDetector, 60*time.Second); err != nil {
		slog.Error("failed to start heartbeat detector", "error", err)
		os.Exit(1)
	}
	slog.Info("heartbeat detector started", "interval", "60s")

	// Initialize Asynq worker processor (Asynq mode only)
	if app.SchedulerCfg.Mode == scheduler.ModeAsynq {
		slog.Info("initializing background worker for Asynq")

		strategies := BuildStrategies()
		executor := domain.NewCheckExecutor(strategies, app.MetricsRecorder)

		incidentService := monitoring.NewIncidentService(
			app.IncidentRepo,
			app.IncidentEventStepRepo,
			app.NotificationRepo,
			app.NotificationChannelRepo,
			app.IncidentDiagnosticsRepo,
			app.AsynqClient,
		)
		app.DetectorIncidentSvc = incidentService

		monitoringHandler := worker.NewMonitoringTaskHandler(app.ResourceRepo, app.MonitoringActivityRepo, app.MaintenanceRepo, app.IncidentDiagnosticsRepo, executor, incidentService, app.ComponentService, app.ConfirmationScheduler)
		maintenanceTaskHandler := maintenance.NewTaskHandler(app.MaintenanceRepo, &maintenance.AsynqClientAdapter{Client: app.AsynqClient})

		expiryNotificationLogRepo := store.NewExpiryNotificationLogRepository(app.DB)
		expiryService := service.NewExpiryNotificationService(
			expiryNotificationLogRepo,
			app.NotificationChannelRepo,
			service.ParseGlobalThresholds(app.Cfg.ExpiryAlertThresholds),
		)
		expiryTaskHandler := worker.NewExpiryTaskHandler(app.ResourceRepo, app.NotificationChannelRepo, enrichmentService, expiryService)

		// Register daily expiry check
		if _, err := app.AsynqScheduler.Register("@daily", asynq.NewTask(worker.TypeExpiryCheck, nil)); err != nil {
			slog.Error("failed to register expiry:check scheduler", "error", err)
		}

		// Create Asynq worker processor
		app.Processor = worker.NewProcessor(app.RedisOpt, monitoringHandler, maintenanceTaskHandler, expiryTaskHandler, worker.Config{
			Concurrency: app.SchedulerCfg.Asynq.Concurrency,
		})

		go func() {
			slog.Info("starting Asynq worker server")
			if err := app.Processor.Start(context.Background()); err != nil {
				slog.Error("could not run Asynq worker server", "error", err)
				os.Exit(1)
			}
		}()

		slog.Info("waiting for worker to start", "delay", "10s")
		time.Sleep(10 * time.Second)
		slog.Info("background worker started")
	} else {
		slog.Info("skipping Asynq worker initialization (TimingWheel mode)")
	}
}
