package database

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSQLiteStartupRunsMigrationsAutomatically(t *testing.T) {
	cfg := newSQLiteTestConfig(t)
	runtime, err := Open(context.Background(), cfg)
	require.NoError(t, err)
	require.NotNil(t, runtime)

	_, statErr := os.Stat(cfg.SQLitePath)
	require.NoError(t, statErr)
	// Probe schema via raw queries (post-decom: no GORM Migrator).
	sqlDB := runtime.SQLiteDB()
	require.NotNil(t, sqlDB)
	for _, table := range []string{"schema_migrations", "notification_channels", "users"} {
		_, err := sqlDB.Exec("SELECT 1 FROM " + table + " LIMIT 0")
		require.NoError(t, err, "expected table %s to exist post-migration", table)
	}
}
