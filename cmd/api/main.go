package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/denisakp/pulseguard/internal/api"
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

	// Initialize Redis connection for Asynq
	redisOpt := asynq.RedisClientOpt{
		Addr: config.GetEnv("REDIS_URL", "localhost:6379"),
	}

	// Initialize Asynq client and inspector for scheduling
	asynqClient := asynq.NewClient(redisOpt)
	defer asynqClient.Close()

	asynqInspector := asynq.NewInspector(redisOpt)
	defer asynqInspector.Close()

	// Initialize services
	schedulerService := monitoring.NewSchedulerService(asynqClient, asynqInspector)
	resourceService := service.NewResourceService(resourceRepo, incidentRepo, schedulerService)

	// Placeholder to use resourceService - remove when implementing actual routes
	_ = resourceService

	port := config.GetEnv("PORT", "8080")
	router := api.NewRouter()
	log.Printf("API server listening on port %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}
