package database

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func openSQLite(cfg resolvedConfig) (*gorm.DB, PermissionStatus, error) {
	if err := prepareSQLiteTarget(cfg.DSN); err != nil {
		return nil, PermissionStatusNotApplicable, fmt.Errorf("db init: failed to prepare sqlite path: %w", err)
	}

	permissionStatus := hardenSQLiteArtifacts(cfg.DSN)
	slog.Info("opening database connection", "driver", string(cfg.Driver), "dsn", cfg.DSN)

	database, err := gorm.Open(sqlite.Open(cfg.DSN), &gorm.Config{
		Logger: newGormLogger(cfg.GormLogLevel),
	})
	if err != nil {
		return nil, permissionStatus, fmt.Errorf("db init: failed to connect to sqlite: %w", err)
	}

	sqlDB, err := database.DB()
	if err != nil {
		return nil, permissionStatus, fmt.Errorf("db init: failed to get sqlite db handle: %w", err)
	}

	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	if err := database.Exec("PRAGMA foreign_keys = ON;").Error; err != nil {
		return nil, permissionStatus, fmt.Errorf("db init: failed to enable sqlite foreign keys: %w", err)
	}
	if err := database.Exec("PRAGMA busy_timeout = 5000;").Error; err != nil {
		return nil, permissionStatus, fmt.Errorf("db init: failed to configure sqlite busy_timeout: %w", err)
	}

	if !isSQLiteMemoryDSN(cfg.DSN) {
		var journalMode string
		if err := database.Raw("PRAGMA journal_mode = WAL;").Scan(&journalMode).Error; err != nil {
			return nil, permissionStatus, fmt.Errorf("db init: failed to enable sqlite WAL mode: %w", err)
		}
		if !strings.EqualFold(journalMode, "wal") {
			return nil, permissionStatus, fmt.Errorf("db init: sqlite WAL mode not active (got %q)", journalMode)
		}
	}

	return database, permissionStatus, nil
}
