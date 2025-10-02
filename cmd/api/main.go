package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/denisakp/pulseguard/internal/api"
	"github.com/denisakp/pulseguard/internal/api/handler"
	"github.com/denisakp/pulseguard/internal/config"
	"github.com/denisakp/pulseguard/internal/monitoring"
	"github.com/denisakp/pulseguard/internal/repository/postgres"
	"github.com/denisakp/pulseguard/internal/repository/postgres/database"
	"github.com/denisakp/pulseguard/internal/service"
	"github.com/hibiken/asynq"
)

func main() {
	log.Println("Starting Pulseguard API server...")

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

	// Get database instance
	db, err := database.Instance()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}

	// Initialize repositories
	resourceRepo := postgres.NewResourceRepository(db)
	incidentRepo := postgres.NewIncidentRepository(db)
	monitoringActivityRepo := postgres.NewMonitoringActivityRepository(db)

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

	// Initialize services
	schedulerService := monitoring.NewSchedulerService(asynqClient, asynqInspector, asynqScheduler)
	resourceService := service.NewResourceService(resourceRepo, incidentRepo, schedulerService)
	activityService := service.NewMonitoringActivityService(monitoringActivityRepo)

	// Initialize handlers with dependency injection
	resourceHandler := handler.NewResourceHandler(resourceService)
	activityHandler := handler.NewMonitoringActivityHandler(activityService)

	// Create router with injected handlers
	router := api.NewRouter(resourceHandler, activityHandler)

	// Create HTTP server with explicit configuration
	addr := ":" + cfg.Port
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel to listen for interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		log.Printf("API server listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Block until we receive a signal
	<-quit
	log.Println("Received shutdown signal, gracefully shutting down...")

	// Create a context with timeout for shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown the HTTP server gracefully
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server forced to shutdown: %v", err)
	}

	log.Println("API server stopped gracefully")
}
