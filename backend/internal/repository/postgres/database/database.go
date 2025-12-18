package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/denisakp/pulseguard/internal/config"
	domain "github.com/denisakp/pulseguard/internal/domain"
)

var (
	mu          sync.RWMutex
	db          *gorm.DB
	initErr     error
	initialized bool

	// modelsToMigrate holds domain models for auto-migration
	modelsToMigrate = []any{
		// Core domain models introduced by feature 002
		&domain.Resource{},
		&domain.Incident{},
		&domain.IncidentEventStep{},
		&domain.NotificationEvent{},
		&domain.NotificationChannel{},
		&domain.Tags{},
		&domain.MonitoringActivity{},
	}
)

// Init initializes the database connection with the provided configuration.
// It configures connection pooling, logging, and runs auto-migration for registered models.
// This function should be called exactly once during application startup.
func Init(ctx context.Context, dsn *string) error {
	mu.Lock()
	defer mu.Unlock()

	// Return cached error or success if already initialized
	if initialized {
		return initErr
	}

	// Mark as initialized (even if it fails, we don't want to retry automatically)
	initialized = true

	if dsn == nil || *dsn == "" {
		initErr = fmt.Errorf("db init: configuration is required")
		return initErr
	}

	log.Printf("db_init=starting action=opening_connection")

	// Configure GORM with custom logger
	gormConfig := &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold: 200 * time.Millisecond,
				LogLevel:      logger.Warn,
				Colorful:      false,
			},
		),
	}

	// Open connection with postgres driver
	database, err := gorm.Open(postgres.Open(*dsn), gormConfig)
	if err != nil {
		initErr = fmt.Errorf("db init: failed to connect: %w", err)
		return initErr
	}

	// Get underlying sql.DB for connection pool configuration
	sqlDB, err := database.DB()
	if err != nil {
		initErr = fmt.Errorf("db init: failed to get underlying db: %w", err)
		return initErr
	}

	// Configure connection pool with conservative defaults
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	// Store the database instance
	db = database

	// Run auto-migration for registered models (safe with empty slice)
	if len(modelsToMigrate) > 0 {
		log.Printf("db_init=migrating models=%d", len(modelsToMigrate))
		if err := db.AutoMigrate(modelsToMigrate...); err != nil {
			initErr = fmt.Errorf("db init: migration failed: %w", err)
			return initErr
		}
	} else {
		log.Printf("db_init=skipping_migration reason=no_models_registered")
	}

	log.Printf("db_init=completed pool_max_open=25 pool_max_idle=5")
	initErr = nil
	return nil
}

// Instance returns the singleton database instance.
// If Init has not been called, it attempts lazy initialization using environment variables.
func Instance() (*gorm.DB, error) {
	mu.RLock()
	currentDB := db
	isInit := initialized
	mu.RUnlock()

	// If already initialized, return immediately
	if isInit {
		if currentDB == nil {
			return nil, fmt.Errorf("db instance: database initialization failed previously")
		}
		return currentDB, nil
	}

	// Attempt lazy initialization
	dsn := config.GetEnv("DATABASE_URL", "")
	if err := Init(context.Background(), &dsn); err != nil {
		return nil, fmt.Errorf("db instance: initialization failed: %w", err)
	}

	log.Printf("db_instance=lazy_init action=completed")

	mu.RLock()
	defer mu.RUnlock()

	if db == nil {
		return nil, fmt.Errorf("db instance: database not initialized")
	}

	return db, nil
}

// Ping checks the database connection health by executing a simple query.
func Ping(ctx context.Context) error {
	mu.RLock()
	currentDB := db
	mu.RUnlock()

	if currentDB == nil {
		return fmt.Errorf("db ping: database not initialized")
	}

	// Execute a simple query with context timeout
	sqlDB, err := currentDB.DB()
	if err != nil {
		return fmt.Errorf("db ping: failed to get underlying db: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("db ping: connection failed: %w", err)
	}

	return nil
}

// Reset resets the singleton state - ONLY FOR TESTING
func Reset() {
	mu.Lock()
	defer mu.Unlock()

	if db != nil {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}

	db = nil
	initErr = nil
	initialized = false
}
