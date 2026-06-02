package database

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestRuntimeSinglePool asserts the single-pool guarantee for SQLite:
// Stats().PoolPointer is stable across calls and equals the pointer of the
// raw *sql.DB exposed by SQLiteDB().
func TestRuntimeSinglePool(t *testing.T) {
	runtime := openSQLiteTestRuntime(t)
	t.Cleanup(func() {
		if runtime.sqliteDB != nil {
			_ = runtime.sqliteDB.Close()
		}
	})

	require.Equal(t, DriverSQLite, runtime.Driver)
	require.NotNil(t, runtime.SQLiteDB(), "SQLiteDB() must be non-nil for SQLite driver")
	require.Nil(t, runtime.PgxPool(), "PgxPool() must be nil for SQLite driver")

	stats1 := runtime.Stats()
	stats2 := runtime.Stats()
	require.NotZero(t, stats1.PoolPointer, "PoolPointer must be set")
	require.Equal(t, stats1.PoolPointer, stats2.PoolPointer, "PoolPointer must be stable across Stats() calls")

	rawPtr := reflect.ValueOf(runtime.SQLiteDB()).Pointer()
	require.Equal(t, rawPtr, stats1.PoolPointer, "Stats().PoolPointer must match SQLiteDB() pointer")
}

// TestStatsCountsTrack issues a handful of trivial queries via the raw handle
// and asserts Stats() reflects the activity: at least one OpenConns and zero
// InUseConns once the queries complete.
func TestStatsCountsTrack(t *testing.T) {
	runtime := openSQLiteTestRuntime(t)
	t.Cleanup(func() {
		if runtime.sqliteDB != nil {
			_ = runtime.sqliteDB.Close()
		}
	})

	raw := runtime.SQLiteDB()
	require.NotNil(t, raw)

	for i := 0; i < 3; i++ {
		var one int
		require.NoError(t, raw.QueryRow("SELECT 1").Scan(&one))
		require.Equal(t, 1, one)
	}

	stats := runtime.Stats()
	require.Equal(t, DriverSQLite, stats.Driver)
	require.GreaterOrEqual(t, stats.OpenConns, 1, "expected at least one open conn after queries")
	require.Equal(t, 0, stats.InUseConns, "no queries in flight after the loop")
}

// TestRuntimeSinglePoolPostgres mirrors TestRuntimeSinglePool for Postgres.
// Skipped unless TEST_POSTGRES_DSN is set in the environment.
func TestRuntimeSinglePoolPostgres(t *testing.T) {
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN not set; skipping postgres single-pool test")
	}

	runtime, err := Open(context.Background(), Config{
		Driver:      DriverPostgres,
		DatabaseURL: dsn,
		LogLevel:    "silent",
	})
	require.NoError(t, err)
	require.NotNil(t, runtime)
	t.Cleanup(func() {
		if runtime.pgxPool != nil {
			runtime.pgxPool.Close()
		}
	})

	require.Equal(t, DriverPostgres, runtime.Driver)
	require.NotNil(t, runtime.PgxPool(), "PgxPool() must be non-nil for Postgres driver")
	require.Nil(t, runtime.SQLiteDB(), "SQLiteDB() must be nil for Postgres driver")

	stats1 := runtime.Stats()
	stats2 := runtime.Stats()
	require.NotZero(t, stats1.PoolPointer)
	require.Equal(t, stats1.PoolPointer, stats2.PoolPointer)

	rawPtr := reflect.ValueOf(runtime.PgxPool()).Pointer()
	require.Equal(t, rawPtr, stats1.PoolPointer, "Stats().PoolPointer must match PgxPool() pointer")
}
