package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// T0M1: startup applies migrations and initializes API key schema objects.
func TestStartupAppliesAPIKeyMigrations(t *testing.T) {
	cfg := newSQLiteTestConfig(t)
	runtime, err := Open(context.Background(), cfg)
	require.NoError(t, err)
	require.NotNil(t, runtime)

	migrator := runtime.DB.Migrator()
	require.True(t, migrator.HasTable("api_keys"))
	require.True(t, migrator.HasTable("schema_migrations"))
}
