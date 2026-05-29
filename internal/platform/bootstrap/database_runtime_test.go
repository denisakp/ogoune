package bootstrap

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	dbruntime "github.com/denisakp/ogoune/internal/database"
)

// TestBootstrapExposesRuntime asserts that opening a runtime through the
// database package surfaces both a non-nil GORM handle AND the dialect's raw
// handle, and that they share a single underlying pool identity.
//
// We exercise database.Open directly (rather than the full app bootstrap)
// because the latter wires repositories and calls os.Exit on failure.
func TestBootstrapExposesRuntime(t *testing.T) {
	cfg := dbruntime.Config{
		Driver:     dbruntime.DriverSQLite,
		SQLitePath: filepath.Join(t.TempDir(), "ogoune-runtime-test.db"),
		LogLevel:   "silent",
	}

	rt, err := dbruntime.Open(context.Background(), cfg)
	require.NoError(t, err)
	require.NotNil(t, rt)
	t.Cleanup(func() {
		if sqlDB, err := rt.GormDB().DB(); err == nil && sqlDB != nil {
			_ = sqlDB.Close()
		}
	})

	require.NotNil(t, rt.GormDB(), "bootstrap must expose GORM handle")
	require.NotNil(t, rt.SQLiteDB(), "bootstrap must expose raw SQLite handle")
	require.Nil(t, rt.PgxPool(), "PgxPool must be nil under SQLite driver")

	gormSQL, err := rt.GormDB().DB()
	require.NoError(t, err)
	require.Equal(t,
		reflect.ValueOf(gormSQL).Pointer(),
		reflect.ValueOf(rt.SQLiteDB()).Pointer(),
		"GORM and raw handles must share one underlying *sql.DB",
	)
}
