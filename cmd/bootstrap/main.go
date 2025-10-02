package main

import (
	"context"
	"log"
	"time"

	"github.com/denisakp/pulseguard/internal/config"
	"github.com/denisakp/pulseguard/internal/monitoring"
	"github.com/denisakp/pulseguard/internal/repository/postgres"
	"github.com/denisakp/pulseguard/internal/repository/postgres/database"
	"github.com/hibiken/asynq"
)

func main() {
	log.Println("Starting Pulseguard Scheduler Bootstrap...")

	// Load configuration
	cfg := config.MustInit()

	// Initialize database connection
	ctx := context.Background()
	if err := database.Init(ctx, &cfg.DatabaseUrl); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Health check
	healthCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := database.Ping(healthCtx); err != nil {
		log.Fatalf("Database health check failed: %v", err)
	}

	log.Println("Database connection established successfully")

	// Get database instance
	db, err := database.Instance()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}

	// Initialize resource repository
	resourceRepo := postgres.NewResourceRepository(db)

	// Initialize Redis connection for Asynq
	redisOpt := asynq.RedisClientOpt{
		Addr: config.GetEnv("REDIS_URL", "localhost:6379"),
	}

	// Initialize Asynq client, inspector, and scheduler
	asynqClient := asynq.NewClient(redisOpt)
	defer asynqClient.Close()

	asynqInspector := asynq.NewInspector(redisOpt)
	defer asynqInspector.Close()

	asynqScheduler := asynq.NewScheduler(redisOpt, nil)

	// Initialize scheduler service
	schedulerService := monitoring.NewSchedulerService(asynqClient, asynqInspector, asynqScheduler)

	// Fetch all active resources from the database
	log.Println("Fetching active resources from database...")
	activeResources, err := resourceRepo.FindActive(ctx, 10000, 0) // Large limit to get all
	if err != nil {
		log.Fatalf("Failed to fetch active resources: %v", err)
	}

	log.Printf("Found %d active resources to schedule", len(activeResources))

	// Schedule each active resource
	successCount := 0
	failureCount := 0

	for _, resource := range activeResources {
		log.Printf("Scheduling resource: %s (ID: %s, Type: %s, Interval: %ds)",
			resource.Name, resource.ID, resource.Type, resource.Interval)

		if err := schedulerService.Schedule(ctx, resource); err != nil {
			log.Printf("  ⚠️  Failed to schedule resource %s: %v", resource.ID, err)
			failureCount++
		} else {
			log.Printf("  ✓ Successfully scheduled resource %s", resource.ID)
			successCount++
		}
	}

	// Log summary
	log.Println("========================================")
	log.Printf("Bootstrap completed successfully!")
	log.Printf("  Total resources processed: %d", len(activeResources))
	log.Printf("  Successfully scheduled: %d", successCount)
	log.Printf("  Failed to schedule: %d", failureCount)
	log.Println("========================================")

	if failureCount > 0 {
		log.Println("⚠️  Some resources failed to schedule. Check logs above for details.")
	}
}
