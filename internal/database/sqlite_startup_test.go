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
	require.True(t, runtime.DB.Migrator().HasTable("schema_migrations"))
	require.True(t, runtime.DB.Migrator().HasTable("notification_channels"))
	require.True(t, runtime.DB.Migrator().HasTable("users"))
}
