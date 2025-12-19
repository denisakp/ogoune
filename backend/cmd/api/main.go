package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/denisakp/pulseguard/internal/api"
	"github.com/denisakp/pulseguard/internal/api/handler"
	"github.com/denisakp/pulseguard/internal/config"
	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/maintenance"
	"github.com/denisakp/pulseguard/internal/monitoring"
	"github.com/denisakp/pulseguard/internal/monitoring/strategy"
	"github.com/denisakp/pulseguard/internal/repository/postgres"
	"github.com/denisakp/pulseguard/internal/repository/postgres/database"
	"github.com/denisakp/pulseguard/internal/service"
	"github.com/denisakp/pulseguard/internal/worker"
	"github.com/go-chi/chi/v5"
	"github.com/hibiken/asynq"
)

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

func main() {
	log.Println("========================================")
	log.Println("Starting Pulse guard Application...")
	log.Println("========================================")

	// Load database configuration
	cfg := config.MustInit()

	// Initialize database connection
	if err := database.Init(context.Background(), &cfg.DatabaseUrl); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Health check
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := database.Ping(ctx); err != nil {
		log.Fatalf("Database health check failed: %v", err)
	}

	log.Println("✓ Database connection established successfully")

	log.Printf("[auth] Authentication enabled with email: %s", cfg.AuthEmail)
	log.Println("[auth] JWT-based authentication configured")

	// Get database instance
	db, err := database.Instance()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}

	// Initialize repositories
	resourceRepo := postgres.NewResourceRepository(db)
	incidentRepo := postgres.NewIncidentRepository(db)
	incidentEventStepRepo := postgres.NewIncidentEventStepRepository(db)
	notificationRepo := postgres.NewNotificationRepository(db)
	maintenanceRepo := postgres.NewMaintenanceRepository(db)
	notificationChannelRepo := postgres.NewNotificationChannelRepository(db)
	monitoringActivityRepo := postgres.NewMonitoringActivityRepository(db)
	tagsRepo := postgres.NewTagsRepository(db)

	// Initialize Redis connection for Asynq
	redisOpt := asynq.RedisClientOpt{
		Addr: config.GetEnv("REDIS_URL", "localhost:6379"),
	}

	// Initialize Asynq client, inspector, and scheduler for periodic tasks
	asynqClient := asynq.NewClient(redisOpt)
	defer asynqClient.Close()

	asynqInspector := asynq.NewInspector(redisOpt)
	defer asynqInspector.Close()

	// Create scheduler for periodic monitoring tasks
	asynqScheduler := asynq.NewScheduler(redisOpt, nil)

	// Start the scheduler in a goroutine
	go func() {
		if err := asynqScheduler.Run(); err != nil {
			log.Fatalf("Failed to start Asynq scheduler: %v", err)
		}
	}()
	defer asynqScheduler.Shutdown()

	// Initialize monitoring scheduler service
	schedulerService := monitoring.NewSchedulerService(asynqClient, asynqInspector, asynqScheduler)

	// Initialize maintenance scheduler and ensure schedules are registered
	maintenanceScheduler := maintenance.NewSchedulerService(asynqClient, asynqInspector, asynqScheduler, maintenanceRepo)
	if err := maintenanceScheduler.EnsureScheduled(context.Background()); err != nil {
		log.Printf("⚠️  Failed to ensure maintenance schedules: %v", err)
	}

	// Initialize enrichment service for resource metadata collection
	enrichmentService := service.NewEnrichmentService(30 * time.Second)

	// ========================================
	// STEP 1: BOOTSTRAP - Schedule all active resources
	// This runs sequentially and blocks until complete
	// ========================================
	log.Println("========================================")
	log.Println("BOOTSTRAP: Scheduling active resources...")
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

		if err := schedulerService.Schedule(bootstrapCtx, resource); err != nil {
			log.Printf("  ⚠️  Failed to schedule resource %s: %v", resource.ID, err)
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
		log.Println("⚠️  Some resources failed to schedule. Check logs above for details.")
	}

	// ========================================
	// STEP 2: INITIALIZE WORKER - Start Asynq worker in background
	// ========================================
	log.Println("Initializing background worker...")

	// Initialize monitoring services for worker
	strategies := map[domain.ResourceType]domain.CheckStrategy{
		domain.ResourceHTTP: strategy.NewHTTPStrategy(30 * time.Second),
		domain.ResourceTCP:  strategy.NewTCPStrategy(30 * time.Second),
	}
	executor := domain.NewCheckExecutor(strategies)

	// Initialize incident service with dynamic notification channel dispatch
	incidentService := monitoring.NewIncidentService(
		incidentRepo,
		incidentEventStepRepo,
		notificationRepo,
		notificationChannelRepo,
		asynqClient,
	)

	// Initialize task handlers
	monitoringHandler := worker.NewMonitoringTaskHandler(resourceRepo, monitoringActivityRepo, maintenanceRepo, executor, incidentService)
	maintenanceTaskHandler := maintenance.NewTaskHandler(maintenanceRepo, asynqClient)

	// Create the Asynq worker processor
	processor := worker.NewProcessor(redisOpt, monitoringHandler, maintenanceTaskHandler)

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

	// ========================================
	// STEP 3: INITIALIZE API SERVER
	// ========================================
	log.Println("Initializing API server...")

	// Initialize API services
	resourceService := service.NewResourceService(resourceRepo, incidentRepo, tagsRepo, schedulerService, monitoringActivityRepo, enrichmentService)
	activityService := service.NewMonitoringActivityService(monitoringActivityRepo)
	tagService := service.NewTagService(tagsRepo)
	statusPageService := service.NewStatusPageService(resourceRepo, incidentRepo, monitoringActivityRepo)
	incidentAPIService := service.NewIncidentService(incidentRepo, incidentEventStepRepo)
	notificationService := service.NewNotificationService(
		resourceRepo,
		notificationChannelRepo,
		cfg.SMTPIsEnabled,
		cfg.DefaultRecipientEmail,
		cfg.SMTPSender,
		cfg.SMTPHost,
		cfg.SMTPPort,
		cfg.SMTPUser,
		cfg.SMTPPassword,
	)
	maintenanceAPIService := service.NewMaintenanceService(maintenanceRepo, maintenanceScheduler)
	statsService := service.NewStatsService(monitoringActivityRepo, incidentRepo)
	authService := service.NewAuthService(cfg.AuthEmail, cfg.AuthPassword, cfg.JWTSecret)

	// Initialize JSON API handlers (no template dependencies)
	resourceHandler := handler.NewResourceHandler(resourceService)
	activityHandler := handler.NewMonitoringActivityHandler(activityService)
	tagHandler := handler.NewTagHandler(tagService)
	statusPageHandler := handler.NewStatusPageHandler(statusPageService)
	incidentHandler := handler.NewIncidentHandler(incidentAPIService)
	notificationHandler := handler.NewNotificationHandler(notificationService)
	maintenanceHandler := handler.NewMaintenanceHandler(maintenanceAPIService)
	statsHandler := handler.NewStatsHandler(statsService)
	authHandler := handler.NewAuthHandler(authService)

	// Create router with injected handlers
	apiHandler := api.NewRouter(resourceHandler, activityHandler, tagHandler, statusPageHandler, incidentHandler, notificationHandler, maintenanceHandler, statsHandler, authHandler, authService)

	// Root router: mount JSON API under /api
	rootRouter := chi.NewRouter()
	rootRouter.Mount("/api", apiHandler)
	// Expose health at root for compatibility
	rootRouter.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// ========================================
	// STEP 4: SERVE STATIC FILES (SPA Support)
	// ========================================
	// Serve Vue.js static files if available
	staticDir := cfg.StaticDir
	if info, err := os.Stat(staticDir); err == nil && info.IsDir() {
		log.Printf("✓ Serving static files from: %s", staticDir)
		serveStaticFiles(rootRouter, staticDir)
	} else {
		log.Printf("⚠️  Static directory not found at %s - frontend will not be served", staticDir)
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
	log.Println("Pulse guard application stopped successfully")
	log.Println("========================================")
}
