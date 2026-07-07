package internaltest

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/denisakp/ogoune/internal/database"
)

// SetupSQLite returns a freshly-migrated, per-test SQLite fixture. The DB
// file lives under t.TempDir() so test cleanup is automatic; the underlying
// connection is also closed via t.Cleanup.
func SetupSQLite(t *testing.T) *DialectFixture {
	t.Helper()
	cfg := database.Config{
		Driver:     database.DriverSQLite,
		SQLitePath: filepath.Join(t.TempDir(), "ogoune-test.db"),
		LogLevel:   "silent",
	}
	rt, err := database.Open(context.Background(), cfg)
	if err != nil {
		t.Fatalf("internaltest: SetupSQLite open: %v", err)
	}
	t.Cleanup(func() {
		if rt == nil || rt.SQLiteDB() == nil {
			return
		}
		_ = rt.SQLiteDB().Close()
	})
	return &DialectFixture{Dialect: "sqlite", Runtime: rt}
}
