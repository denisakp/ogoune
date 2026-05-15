package database

import (
	"context"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

func newSQLiteTestConfig(t *testing.T) Config {
	t.Helper()
	return Config{
		Driver:     DriverSQLite,
		SQLitePath: filepath.Join(t.TempDir(), "ogoune-test.db"),
		LogLevel:   "silent",
	}
}

func openSQLiteTestRuntime(t *testing.T) *Runtime {
	t.Helper()
	runtime, err := Open(context.Background(), newSQLiteTestConfig(t))
	require.NoError(t, err)
	require.NotNil(t, runtime)
	return runtime
}

func openSQLiteTestRuntimeWithFS(t *testing.T, cfg Config, migrationFS fstest.MapFS) (*Runtime, error) {
	t.Helper()
	resolved, err := cfg.resolve()
	require.NoError(t, err)
	return openRuntime(context.Background(), resolved, migrationFS)
}
