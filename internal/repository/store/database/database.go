package database

import (
	"context"
	"fmt"

	"github.com/denisakp/ogoune/internal/config"
	shareddb "github.com/denisakp/ogoune/internal/database"
)

// Init initializes the database connection with the provided configuration.
// It preserves the legacy postgres-only signature while delegating to the shared runtime.
// This function should be called exactly once during application startup.
func Init(ctx context.Context, dsn *string) error {
	if dsn == nil {
		return fmt.Errorf("db init: configuration is required")
	}

	return shareddb.Init(ctx, shareddb.Config{
		Driver:      shareddb.DriverPostgres,
		DatabaseURL: *dsn,
		SQLitePath:  config.GetEnv("SQLITE_PATH", "ogoune.db"),
		LogLevel:    config.GetEnv("DB_LOG_LEVEL", "error"),
	})
}

// Ping checks the database connection health by executing a simple query.
func Ping(ctx context.Context) error {
	return shareddb.Ping(ctx)
}

// Reset resets the singleton state - ONLY FOR TESTING
func Reset() {
	shareddb.Reset()
}
