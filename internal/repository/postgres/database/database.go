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
)

var (
	once sync.Once
	db   *gorm.DB
	// modelsToMigrate holds domain models for auto-migration
	modelsToMigrate []any
)

// Init initializes the database connection with the provided configuration.
// It configures connection pooling, logging, and runs auto-migration for registered models.
// This function should be called exactly once during application startup.
func Init(ctx context.Context, cfg *config.DBConfig) error {
	var initErr error
	once.Do(func() {
		if cfg == nil || cfg.DSN == "" {
			initErr = fmt.Errorf("db init: configuration is required")
			return
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
		database, err := gorm.Open(postgres.Open(cfg.DSN), gormConfig)
		if err != nil {
			initErr = fmt.Errorf("db init: failed to connect: %w", err)
			return
		}

		// Get underlying sql.DB for connection pool configuration
		sqlDB, err := database.DB()
		if err != nil {
			initErr = fmt.Errorf("db init: failed to get underlying db: %w", err)
			return
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
				return
			}
		} else {
			log.Printf("db_init=skipping_migration reason=no_models_registered")
		}

		log.Printf("db_init=completed pool_max_open=25 pool_max_idle=5")
	})
	return initErr
}

// Instance returns the singleton database instance.
// If Init has not been called, it attempts lazy initialization using environment variables.
func Instance() (*gorm.DB, error) {
	if db == nil {
		// Attempt lazy initialization
		log.Printf("db_instance=lazy_init action=loading_config")
		cfg, err := config.LoadDBConfig()
		if err != nil {
			return nil, fmt.Errorf("db instance: failed to load config: %w", err)
		}
		if err := Init(context.Background(), cfg); err != nil {
			return nil, fmt.Errorf("db instance: initialization failed: %w", err)
		}
		log.Printf("db_instance=lazy_init action=completed")
	}

	if db == nil {
		return nil, fmt.Errorf("db instance: database not initialized")
	}

	return db, nil
}

// Ping checks the database connection health by executing a simple query.
func Ping(ctx context.Context) error {
	if db == nil {
		return fmt.Errorf("db ping: database not initialized")
	}

	// Execute a simple query with context timeout
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("db ping: failed to get underlying db: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("db ping: connection failed: %w", err)
	}

	return nil
}
