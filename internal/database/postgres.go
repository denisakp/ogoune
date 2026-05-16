package database

import (
	"fmt"
	"log/slog"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func openPostgres(cfg resolvedConfig) (*gorm.DB, error) {
	slog.Info("opening database connection", "driver", string(cfg.Driver))

	database, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{
		Logger: newGormLogger(cfg.GormLogLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("db init: failed to connect to postgres: %w", err)
	}

	sqlDB, err := database.DB()
	if err != nil {
		return nil, fmt.Errorf("db init: failed to get postgres db handle: %w", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	return database, nil
}
