package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	_ "github.com/denisakp/ogoune/docs"
	"github.com/denisakp/ogoune/internal/api"
	"github.com/denisakp/ogoune/internal/api/handler"
	v1handler "github.com/denisakp/ogoune/internal/api/handler/v1"
	"github.com/denisakp/ogoune/internal/config"
	dbruntime "github.com/denisakp/ogoune/internal/database"
	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/ee/license"
	icmppkg "github.com/denisakp/ogoune/internal/icmp"
	"github.com/denisakp/ogoune/internal/maintenance"
	"github.com/denisakp/ogoune/internal/metrics"
	"github.com/denisakp/ogoune/internal/monitoring"
	"github.com/denisakp/ogoune/internal/monitoring/strategy"
	"github.com/denisakp/ogoune/internal/repository"
	"github.com/denisakp/ogoune/internal/repository/store"
	"github.com/denisakp/ogoune/internal/scheduler"
	"github.com/denisakp/ogoune/internal/service"
	"github.com/denisakp/ogoune/internal/worker"
	"github.com/go-chi/chi/v5"
	"github.com/hibiken/asynq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

const appVersion = "1.0.0"

func logStartupEdition() {
	if license.IsEnterprise() {
		log.Println("✓ Ogoune Enterprise Edition")
	} else {
		log.Println("✓ Ogoune Community Edition")
	}
}

// resourceRepositoryAdapter adapts ResourceRepository to implement ActiveResourceRepository
type resourceRepositoryAdapter struct {
	repo repository.ResourceRepository
}

type pendingNotificationRetryRunner interface {
	RetryPendingNotifications(ctx context.Context, limit int) (service.PendingNotificationRetrySummary, error)
}

func (a *resourceRepositoryAdapter) FindScheduledResources(ctx context.Context) ([]scheduler.ScheduleItem, error) {
	resources, err := a.repo.FindScheduledResources(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]scheduler.ScheduleItem, 0, len(resources))
	for _, r := range resources {
		if r.Interval > 0 { // Only include resources with valid intervals
			items = append(items, scheduler.ScheduleItem{
				ResourceID: r.ID,
				Interval:   time.Duration(r.Interval) * time.Second,
				Paused:     false, // Pause state is managed at runtime, not persisted
			})
		}
	}
	return items, nil
}

func serveStaticFiles(router *chi.Mux, staticDir string) {
	// Serve files from the static directory
	fs := http.FileServer(http.Dir(staticDir))

	// Handle all non-API routes
	router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Don't serve static files for API routes
		if strings.HasPrefix(path, "/api/") {
			http.NotFound(w, r)
			return
		}

		if path == "/status" || strings.HasPrefix(path, "/status/") {
			statusHTML := filepath.Join(staticDir, "status.html")
			if _, err := os.Stat(statusHTML); err == nil {
				http.ServeFile(w, r, statusHTML)
				return
			}
		}

		// Check if the requested file exists
		fullPath := filepath.Join(staticDir, path)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			// File doesn't exist, serve index.html for Vue Router
			http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
			return
		}

		// File exists, serve it
		fs.ServeHTTP(w, r)
	})
}

// startHeartbeatDetector starts the recurring missed-heartbeat detector.
// It returns an error immediately if the detector is nil or the context is already cancelled,
// enforcing fail-fast startup behavior (T034).
func startHeartbeatDetector(ctx context.Context, detector *service.HeartbeatDetectorService, interval time.Duration) error {
	if detector == nil {
		return fmt.Errorf("heartbeat detector service is required")
	}
	return detector.Start(ctx, interval)
}

func runStartupPendingNotificationRetry(ctx context.Context, retryService pendingNotificationRetryRunner) {
	if retryService == nil {
		return
	}

	log.Println("[STARTUP] Checking for pending notifications...")
	summary, err := retryService.RetryPendingNotifications(ctx, 1000)
	if err != nil {
		log.Printf("[STARTUP] [WARNING] Pending notification retry failed: %v", err)
		return
	}

	log.Printf("[STARTUP] Pending notifications: scanned=%d retried=%d expired=%d failed=%d skipped_claimed=%d",
		summary.ScannedCount,
		summary.RetriedCount,
		summary.ExpiredCount,
		summary.FailedCount,
		summary.SkippedClaimedCount,
	)
}

func logICMPCapabilityState(enableICMP bool, capability icmppkg.CapabilityResult) {
	if enableICMP {
		if capability.Available {
			log.Println("✓ [ICMP] ICMP probing enabled and capability available")
		} else {
			log.Printf("[ICMP] ICMP probing enabled but capability unavailable: %s", capability.Reason)
		}
		return
	}

	log.Println("[ICMP] ICMP probing disabled (set ENABLE_ICMP=true to enable)")
}

// @title Ogoune Public API
// @version 1.0
// @description Ogoune uptime monitoring — Public REST API v1.
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	log.Println("========================================")
	log.Println("Starting Ogoune Application...")
	logStartupEdition()
	log.Println("========================================")

	// Load database configuration
	cfg := config.MustInit()

	// Initialize database connection
	if err := dbruntime.Init(context.Background(), dbruntime.Config{
		Driver:      dbruntime.Driver(cfg.DBDriver),
		DatabaseURL: cfg.DatabaseUrl,
		SQLitePath:  cfg.SQLitePath,
		LogLevel:    cfg.DBLogLevel,
	}); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Health check
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := dbruntime.Ping(ctx); err != nil {
		log.Fatalf("Database health check failed: %v", err)
	}

	log.Println("✓ Database connection established successfully")

	// Detect ICMP capability at startup — never fails startup.
	icmpCapability := icmppkg.Detect()
	logICMPCapabilityState(cfg.EnableICMP, icmpCapability)

	log.Printf("[auth] Authentication enabled with email: %s", cfg.AuthEmail)
	log.Println("[auth] JWT-based authentication configured")

	// Get database instance
	db, err := dbruntime.Instance()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}

	// Initialize repositories
	resourceRepo := store.NewResourceRepository(db)
	incidentRepo := store.NewIncidentRepository(db)
	incidentEventStepRepo := store.NewIncidentEventStepRepository(db)
	incidentDiagnosticsRepo := store.NewIncidentDiagnosticsRepository(db)
	notificationRepo := store.NewNotificationRepository(db)
	maintenanceRepo := store.NewMaintenanceRepository(db)
	notificationChannelRepo := store.NewNotificationChannelRepository(db)
	monitoringActivityRepo := store.NewMonitoringActivityRepository(db)
	tagsRepo := store.NewTagsRepository(db)
	statusPageSettingsRepo := store.NewStatusPageSettingsRepository(db)
	componentRepo := store.NewComponentRepository(db)
	userRepo := store.NewUserRepository(db)
	apiKeyRepo := store.NewAPIKeyRepository(db)

	// ========================================
	// Initialize metrics (recorder + optional endpoint)
	// ========================================
	var metricsRecorder domain.MetricsRecorder = metrics.NewNoopRecorder()
	var metricsReg *prometheus.Registry
	if cfg.MetricsEnabled {
		if cfg.MetricsToken == "" {
			log.Println("[metrics] WARN metrics endpoint is unauthenticated — set METRICS_TOKEN or restrict access at the network level")
		}
		metricsReg = prometheus.NewRegistry()
		metricsReg.MustRegister(
			collectors.NewGoCollector(),
			collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		)
		metricsRecorder = metrics.NewPrometheusRecorder(metricsReg)
		ogouneCollector := metrics.NewOgouneCollector(resourceRepo, incidentRepo, monitoringActivityRepo)
		metricsReg.MustRegister(ogouneCollector)
		log.Println("[metrics] Prometheus metrics endpoint enabled")
	}

	// ========================================
	// Initialize scheduler based on configuration
	// ========================================
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

	log.Printf("Starting with scheduler mode: %s", schedulerCfg.Mode)

	// Create adapter for services (implements repository.Scheduler interface)
	var runtimeScheduler scheduler.Scheduler
	var schedulerAdapter repository.Scheduler
	var confirmationScheduler interface {
		ScheduleWithInterval(ctx context.Context, resource *domain.Resource, interval time.Duration) error
	}
	var asynqClient *asynq.Client
	var asynqInspector *asynq.Inspector
	var asynqScheduler *asynq.Scheduler
	var redisOpt asynq.RedisClientOpt
	var processor *worker.Processor
	var maintenanceScheduler *maintenance.SchedulerService
	var detectorIncidentSvc *monitoring.IncidentService

	switch schedulerCfg.Mode {
	case scheduler.ModeTimingWheel:
		runtimeScheduler, err = scheduler.New(schedulerCfg)
		if err != nil {
			log.Fatalf("Failed to create scheduler: %v", err)
		}

		log.Println("✓ Using TimingWheel scheduler (Community Edition - no Redis required)")
		schedulerAdapter = scheduler.NewRepositorySchedulerAdapter(runtimeScheduler)
		if rs, ok := schedulerAdapter.(interface {
			ScheduleWithInterval(ctx context.Context, resource *domain.Resource, interval time.Duration) error
		}); ok {
			confirmationScheduler = rs
		}

		// For TimingWheel, no Asynq setup needed, but we need maintenance scheduler
		// Create a no-op Asynq client for compatibility
		// Actually, we should handle maintenance differently for non-Asynq
		maintenanceScheduler = nil

	case scheduler.ModeAsynq:
		log.Println("✓ Using Asynq scheduler (SaaS Edition with Redis)")

		// Initialize Redis connection
		redisOpt = asynq.RedisClientOpt{
			Addr: config.GetEnv("REDIS_URL", "localhost:6379"),
		}

		// Initialize Asynq client, inspector, and scheduler
		asynqClient = asynq.NewClient(redisOpt)
		defer asynqClient.Close()

		asynqInspector = asynq.NewInspector(redisOpt)
		defer asynqInspector.Close()

		// Create scheduler for periodic monitoring tasks
		asynqScheduler = asynq.NewScheduler(redisOpt, nil)

		// Start the scheduler in a goroutine
		go func() {
			if err := asynqScheduler.Run(); err != nil {
				log.Fatalf("Failed to start Asynq scheduler: %v", err)
			}
		}()
		defer asynqScheduler.Shutdown()

		// Initialize monitoring scheduler service (Asynq-based)
		schedulerService := monitoring.NewSchedulerService(asynqClient, asynqInspector, asynqScheduler)
		schedulerCfg.Asynq.ResourceLoader = func(ctx context.Context, resourceID string) (*domain.Resource, error) {
			return resourceRepo.FindByID(ctx, resourceID)
		}
		schedulerCfg.Asynq.SchedulerAdapter = schedulerService

		runtimeScheduler, err = scheduler.New(schedulerCfg)
		if err != nil {
			log.Fatalf("Failed to create scheduler: %v", err)
		}

		schedulerAdapter = scheduler.NewRepositorySchedulerAdapter(runtimeScheduler)
		if rs, ok := schedulerAdapter.(interface {
			ScheduleWithInterval(ctx context.Context, resource *domain.Resource, interval time.Duration) error
		}); ok {
			confirmationScheduler = rs
		}

		// Initialize maintenance scheduler
		maintenanceScheduler = maintenance.NewSchedulerService(asynqClient, asynqInspector, asynqScheduler, maintenanceRepo)
		if err := maintenanceScheduler.EnsureScheduled(context.Background()); err != nil {
			log.Printf("Failed to ensure maintenance schedules: %v", err)
		}

	default:
		log.Fatalf("Invalid scheduler mode: %s", schedulerCfg.Mode)
	}

	// Initialize enrichment service for resource metadata collection
	enrichmentService := service.NewEnrichmentService(30 * time.Second)

	// Initialize component service (used by both worker and API handlers)
	componentService := service.NewComponentService(componentRepo, resourceRepo, notificationChannelRepo)
	pendingRetryService := service.NewPendingNotificationRetryService(
		notificationRepo,
		incidentRepo,
		notificationChannelRepo,
		componentRepo,
		"",
		24*time.Hour,
	)

	// ========================================
	// STEP 1: START SCHEDULER AND BOOTSTRAP
	// ========================================

	if err := runtimeScheduler.Start(context.Background(), &resourceRepositoryAdapter{repo: resourceRepo}); err != nil {
		log.Fatalf("Failed to start scheduler runtime: %v", err)
	}

	if schedulerCfg.Mode == scheduler.ModeTimingWheel {
		// Start TimingWheel in-process scheduler
		log.Println("✓ TimingWheel scheduler started")

		// For TimingWheel, resources are loaded during Start()
		activeResources, err := resourceRepo.FindActive(context.Background(), 10000, 0)
		if err == nil {
			log.Printf("✓ TimingWheel loaded %d active resources at startup", len(activeResources))
		}

		// Start in-process dispatcher that consumes timingwheel check jobs.
		if tw, ok := runtimeScheduler.(*scheduler.TimingWheelScheduler); ok {
			strategies := map[domain.ResourceType]domain.CheckStrategy{
				domain.ResourceHTTP:     strategy.NewHTTPStrategy(30 * time.Second),
				domain.ResourceTCP:      strategy.NewTCPStrategy(30 * time.Second),
				domain.ResourceDNS:      strategy.NewDNSStrategy(30 * time.Second),
				domain.ResourceICMP:     strategy.NewICMPStrategy(),
				domain.ResourceKeyword:  strategy.NewKeywordStrategy(30 * time.Second),
				domain.ResourceProtocol: strategy.NewProtocolStrategy(30 * time.Second),
			}
			executor := domain.NewCheckExecutor(strategies, metricsRecorder)

			incidentService := monitoring.NewIncidentService(
				incidentRepo,
				incidentEventStepRepo,
				notificationRepo,
				notificationChannelRepo,
				incidentDiagnosticsRepo,
				nil,
			)
			detectorIncidentSvc = incidentService

			monitoringHandler := worker.NewMonitoringTaskHandler(resourceRepo, monitoringActivityRepo, maintenanceRepo, incidentDiagnosticsRepo, executor, incidentService, componentService, confirmationScheduler)

			workers := schedulerCfg.TimingWheel.MaxWorkers
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
							log.Printf("Failed to marshal timingwheel payload for resource %s: %v", job.ResourceID, err)
							continue
						}

						task := asynq.NewTask("monitoring:check", payload)
						if err := monitoringHandler.ProcessTask(context.Background(), task); err != nil {
							log.Printf("TimingWheel check failed for resource %s: %v", job.ResourceID, err)
						}
					}
				}()
			}

			log.Printf("✓ TimingWheel check dispatcher started with %d workers", workers)

			// Daily expiry check in TimingWheel mode (no Redis).
			twExpiryNotificationLogRepo := store.NewExpiryNotificationLogRepository(db)
			twExpiryService := service.NewExpiryNotificationService(
				twExpiryNotificationLogRepo,
				notificationChannelRepo,
				service.ParseGlobalThresholds(cfg.ExpiryAlertThresholds),
			)
			twExpiryHandler := worker.NewExpiryTaskHandler(resourceRepo, notificationChannelRepo, enrichmentService, twExpiryService)
			go func() {
				ticker := time.NewTicker(24 * time.Hour)
				defer ticker.Stop()
				for range ticker.C {
					if err := twExpiryHandler.ProcessTask(context.Background(), asynq.NewTask(worker.TypeExpiryCheck, nil)); err != nil {
						log.Printf("TimingWheel expiry:check failed: %v", err)
					}
				}
			}()
			log.Println("✓ TimingWheel daily expiry check scheduled")
		}

	} else {
		// Asynq path: bootstrap as before
		log.Println("========================================")
		log.Println("BOOTSTRAP: Scheduling active resources with Asynq...")
		log.Println("========================================")

		bootstrapCtx := context.Background()
		activeResources, err := resourceRepo.FindActive(bootstrapCtx, 10000, 0) // Large limit to get all
		if err != nil {
			log.Fatalf("Failed to fetch active resources during bootstrap: %v", err)
		}

		log.Printf("Found %d active resources to schedule", len(activeResources))

		successCount := 0
		failureCount := 0

		for _, resource := range activeResources {
			log.Printf("Scheduling resource: %s (ID: %s)", resource.Name, resource.ID)

			if err := schedulerAdapter.Schedule(bootstrapCtx, resource); err != nil {
				log.Printf(" Failed to schedule resource %s: %v", resource.ID, err)
				failureCount++
			} else {
				log.Printf("  ✓ Successfully scheduled resource %s", resource.ID)
				successCount++
			}
		}

		log.Println("========================================")
		log.Printf("Bootstrap completed!")
		log.Printf("  Total resources processed: %d", len(activeResources))
		log.Printf("  Successfully scheduled: %d", successCount)
		log.Printf("  Failed to schedule: %d", failureCount)
		log.Println("========================================")

		if failureCount > 0 {
			log.Println(" Some resources failed to schedule. Check logs above for details.")
		}
	}

	runStartupPendingNotificationRetry(context.Background(), pendingRetryService)

	// Start heartbeat detector — runs every 60s in both scheduler modes (T033/T034).
	// Fail-fast: if detector initialization fails, abort startup rather than silently skip.
	heartbeatDetector := service.NewHeartbeatDetectorService(resourceRepo, detectorIncidentSvc)
	if err := startHeartbeatDetector(context.Background(), heartbeatDetector, 60*time.Second); err != nil {
		log.Fatalf("[STARTUP] Failed to start heartbeat detector: %v", err)
	}
	log.Println("✓ Heartbeat detector started (interval=60s)")

	// ========================================
	// STEP 2: INITIALIZE WORKER (Asynq only)
	// ========================================
	if schedulerCfg.Mode == scheduler.ModeAsynq {
		log.Println("Initializing background worker for Asynq...")

		// Initialize monitoring services for worker
		strategies := map[domain.ResourceType]domain.CheckStrategy{
			domain.ResourceHTTP:     strategy.NewHTTPStrategy(30 * time.Second),
			domain.ResourceTCP:      strategy.NewTCPStrategy(30 * time.Second),
			domain.ResourceDNS:      strategy.NewDNSStrategy(30 * time.Second),
			domain.ResourceICMP:     strategy.NewICMPStrategy(),
			domain.ResourceKeyword:  strategy.NewKeywordStrategy(30 * time.Second),
			domain.ResourceProtocol: strategy.NewProtocolStrategy(30 * time.Second),
		}
		executor := domain.NewCheckExecutor(strategies, metricsRecorder)

		// Initialize incident service with dynamic notification channel dispatch
		incidentService := monitoring.NewIncidentService(
			incidentRepo,
			incidentEventStepRepo,
			notificationRepo,
			notificationChannelRepo,
			incidentDiagnosticsRepo,
			asynqClient,
		)
		detectorIncidentSvc = incidentService

		// Initialize task handlers
		monitoringHandler := worker.NewMonitoringTaskHandler(resourceRepo, monitoringActivityRepo, maintenanceRepo, incidentDiagnosticsRepo, executor, incidentService, componentService, confirmationScheduler)
		maintenanceTaskHandler := maintenance.NewTaskHandler(maintenanceRepo, asynqClient)

		// Initialize expiry notification service and handler
		expiryNotificationLogRepo := store.NewExpiryNotificationLogRepository(db)
		expiryService := service.NewExpiryNotificationService(
			expiryNotificationLogRepo,
			notificationChannelRepo,
			service.ParseGlobalThresholds(cfg.ExpiryAlertThresholds),
		)
		expiryTaskHandler := worker.NewExpiryTaskHandler(resourceRepo, notificationChannelRepo, enrichmentService, expiryService)

		// Register daily expiry check with the Asynq scheduler
		if _, err := asynqScheduler.Register("@daily", asynq.NewTask(worker.TypeExpiryCheck, nil)); err != nil {
			log.Printf(" Failed to register expiry:check scheduler: %v", err)
		}

		// Create the Asynq worker processor
		processor = worker.NewProcessor(redisOpt, monitoringHandler, maintenanceTaskHandler, expiryTaskHandler, worker.Config{
			Concurrency: schedulerCfg.Asynq.Concurrency,
		})
		defer processor.Stop()

		// Start worker in a goroutine (non-blocking)
		go func() {
			log.Println("✓ Starting Asynq worker server...")
			if err := processor.Start(context.Background()); err != nil {
				log.Fatalf("Could not run Asynq worker server: %v", err)
			}
		}()

		log.Printf("Waiting 10 seconds for the worker to start...")
		// Wait briefly to ensure worker starts before proceeding | this is a simple approach to ensure the worker is ready before handling tasks
		time.Sleep(10 * time.Second)
		log.Println("✓ Background worker started successfully")
	} else {
		log.Println("Skipping Asynq worker initialization (TimingWheel mode)")
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Printf("Received shutdown signal, gracefully closing %s scheduler...", schedulerCfg.Mode)
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := runtimeScheduler.Stop(shutdownCtx); err != nil {
			log.Printf(" Error during scheduler shutdown: %v", err)
		}
		if processor != nil {
			processor.Stop()
		}
	}()

	// ========================================
	// STEP 3: INITIALIZE API SERVER
	// ========================================
	log.Println("Initializing API server...")

	// Initialize API services
	resourceService := service.NewResourceService(resourceRepo, incidentRepo, tagsRepo, schedulerAdapter, monitoringActivityRepo, enrichmentService, componentService)
	activityService := service.NewMonitoringActivityService(monitoringActivityRepo)
	tagService := service.NewTagService(tagsRepo)
	statusPageSettingsService := service.NewStatusPageSettingsService(statusPageSettingsRepo)
	statusPageService := service.NewStatusPageService(resourceRepo, incidentRepo, monitoringActivityRepo, maintenanceRepo, statusPageSettingsRepo, componentRepo)
	incidentAPIService := service.NewIncidentService(incidentRepo, incidentEventStepRepo)
	liveSnapshotService := service.NewLiveSnapshotService(resourceService, activityService, incidentAPIService)
	notificationService := service.NewNotificationService(
		resourceRepo,
		notificationChannelRepo,
	)
	maintenanceAPIService := service.NewMaintenanceService(maintenanceRepo, maintenanceScheduler)
	statsService := service.NewStatsService(monitoringActivityRepo, incidentRepo)

	// New auth service with database support
	jwtManager := service.NewJWTManager(cfg.JWTSecret, "ogoune", 24*time.Hour)
	apiKeyService := service.NewAPIKeyService(apiKeyRepo, userRepo)
	authService := service.NewAuthService(userRepo, jwtManager)

	// Create default admin user on first startup
	if cfg.AuthEmail == "" {
		cfg.AuthEmail = "admin@ogoune.test"
	}
	if cfg.AuthPassword == "" {
		cfg.AuthPassword = "ogu3n3@rd"
	}
	_, _ = authService.CreateDefaultUser(context.Background(), cfg.AuthEmail, cfg.AuthPassword)

	// Initialize JSON API handlers (no template dependencies)
	resourceHandler := handler.NewResourceHandler(resourceService, liveSnapshotService)
	pingHandler := handler.NewPingHandler(resourceService)
	activityHandler := handler.NewMonitoringActivityHandler(activityService)
	tagHandler := handler.NewTagHandler(tagService)
	statusPageHandler := handler.NewStatusPageHandler(statusPageService)
	statusPageSettingsHandler := handler.NewStatusPageSettingsHandler(statusPageSettingsService)
	incidentHandler := handler.NewIncidentHandler(incidentAPIService)
	notificationHandler := handler.NewNotificationHandler(notificationService)
	maintenanceHandler := handler.NewMaintenanceHandler(maintenanceAPIService)
	statsHandler := handler.NewStatsHandler(statsService)
	systemHandler := handler.NewSystemHandler()
	authHandler := handler.NewAuthHandler(authService, jwtManager)
	accountHandler := handler.NewAccountHandler(authService, apiKeyService)
	componentHandler := handler.NewComponentHandler(componentService)

	// V1 handlers
	monitorV1Handler := v1handler.NewMonitorHandler(resourceService)
	incidentV1Handler := v1handler.NewIncidentHandler(incidentAPIService)
	channelV1Handler := v1handler.NewNotificationChannelHandler(notificationService)
	componentV1Handler := v1handler.NewComponentHandler(componentRepo)
	tagV1Handler := v1handler.NewTagHandler(tagService)
	statusPageV1Handler := v1handler.NewStatusPageV1Handler(componentRepo)
	heartbeatV1Handler := v1handler.NewHeartbeatV1Handler(resourceService)

	// Create router with injected handlers
	os.Setenv("APP_VERSION", appVersion)
	apiHandler := api.NewRouter(resourceHandler, pingHandler, activityHandler, tagHandler, componentHandler, statusPageHandler, statusPageSettingsHandler, incidentHandler, notificationHandler, maintenanceHandler, statsHandler, systemHandler, authHandler, accountHandler, authService, apiKeyService, monitorV1Handler, incidentV1Handler, channelV1Handler, componentV1Handler, tagV1Handler, statusPageV1Handler, heartbeatV1Handler, cfg.EnableSwagger)

	// Root router: mount JSON API under /api
	rootRouter := chi.NewRouter()
	rootRouter.Mount("/api", apiHandler)
	// Expose health at root for compatibility
	rootRouter.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	if cfg.MetricsEnabled && metricsReg != nil {
		rootRouter.Handle("/metrics", metrics.NewHandler(cfg.MetricsToken, metricsReg))
		log.Println("[metrics] /metrics route registered")
	}

	// ========================================
	// STEP 4: SERVE STATIC FILES (SPA Support)
	// ========================================
	// Serve Vue.js static files if available
	staticDir := cfg.StaticDir
	if info, err := os.Stat(staticDir); err == nil && info.IsDir() {
		log.Printf("✓ Serving static files from: %s", staticDir)
		serveStaticFiles(rootRouter, staticDir)
	} else {
		log.Printf(" Static directory not found at %s - frontend will not be served", staticDir)
	}

	// Create HTTP server with explicit configuration
	addr := ":" + cfg.Port
	srv := &http.Server{
		Addr:         addr,
		Handler:      rootRouter,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// ========================================
	// STEP 5: GRACEFUL SHUTDOWN SETUP
	// ========================================
	// Channel to listen for interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start API server in a goroutine
	go func() {
		log.Println("========================================")
		log.Printf("✓ API server listening on %s", addr)
		log.Println("✓ All systems operational!")
		log.Println("========================================")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Block until we receive a signal
	<-quit
	log.Println("========================================")
	log.Println("Received shutdown signal...")
	log.Println("========================================")

	// Create a context with timeout for shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown the HTTP server gracefully
	log.Println("Shutting down API server...")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server forced to shutdown: %v", err)
	} else {
		log.Println("✓ API server stopped gracefully")
	}

	// Shutdown the worker processor
	log.Println("Shutting down background worker...")
	processor.Stop()
	log.Println("✓ Background worker stopped gracefully")

	log.Println("========================================")
	log.Println("Ogoune application stopped successfully")
	log.Println("========================================")
}
