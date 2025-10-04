package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/denisakp/pulseguard/internal/config"
	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/monitoring"
	"github.com/denisakp/pulseguard/internal/monitoring/strategy"
	"github.com/denisakp/pulseguard/internal/repository/postgres"
	"github.com/denisakp/pulseguard/internal/repository/postgres/database"
	"github.com/denisakp/pulseguard/internal/worker"
	"github.com/hibiken/asynq"
)

func main() {
	log.Println("Starting Pulseguard Worker...")

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

	log.Println("Database initialized successfully")

	// Get database instance
	db, err := database.Instance()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}

	// Initialize repositories
	resourceRepo := postgres.NewResourceRepository(db)
	incidentRepo := postgres.NewIncidentRepository(db)
	incidentEventStepRepo := postgres.NewIncidentEventStepRepository(db)
	integrationRepo := postgres.NewIntegrationRepository(db)
	notificationRepo := postgres.NewNotificationRepository(db)
	monitoringActivityRepo := postgres.NewMonitoringActivityRepository(db)

	// Initialize Redis connection for Asynq
	redisOpt := asynq.RedisClientOpt{
		Addr: config.GetEnv("REDIS_URL", "localhost:6379"),
	}

	// Initialize Asynq client for the incident service
	asynqClient := asynq.NewClient(redisOpt)
	defer asynqClient.Close()

	// Initialize monitoring services
	strategies := map[domain.ResourceType]monitoring.Strategy{
		domain.ResourceHTTP: strategy.NewHTTPStrategy(30 * time.Second),
		domain.ResourceTCP:  strategy.NewTCPStrategy(30 * time.Second),
	}
	executor := monitoring.NewExecutor(strategies)
	incidentService := monitoring.NewIncidentService(
		incidentRepo,
		incidentEventStepRepo,
		integrationRepo,
		notificationRepo, // Added NotificationRepository for tracking notification attempts
		asynqClient,
	)

	// Initialize task handlers
	monitoringHandler := worker.NewMonitoringTaskHandler(resourceRepo, monitoringActivityRepo, executor, incidentService)
	notificationHandler := worker.NewNotificationTaskHandler(incidentRepo, integrationRepo, notificationRepo)

	// Initialize and start the worker processor
	processor := worker.NewProcessor(redisOpt, monitoringHandler, notificationHandler)

	// Handle graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start the processor in a goroutine
	go func() {
		if err := processor.Start(ctx); err != nil {
			log.Printf("Worker processor error: %v", err)
		}
	}()

	log.Println("Worker started successfully. Press Ctrl+C to stop.")

	// Wait for shutdown signal
	<-ctx.Done()

	log.Println("Shutting down worker...")
	processor.Stop()
	log.Println("Worker stopped")
}
