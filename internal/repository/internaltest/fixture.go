package internaltest

import "github.com/denisakp/ogoune/internal/database"

// DialectFixture is a per-test handle to an isolated, freshly-migrated
// database in a known dialect. Returned by SetupSQLite / SetupPostgres and
// consumed by ForEachDialect.
//
// Invariants:
//   - Dialect matches Runtime.Driver.
//   - For "postgres": Runtime.PgxPool() non-nil, Runtime.SQLiteDB() nil, DSN non-empty.
//   - For "sqlite":   Runtime.SQLiteDB() non-nil, Runtime.PgxPool() nil, DSN empty.
//   - Cleanup is registered via t.Cleanup at fixture creation time.
type DialectFixture struct {
	// Dialect is the canonical dialect name ("sqlite" or "postgres").
	Dialect string

	// Runtime is the database runtime opened against this fixture's database.
	// Provides GormDB(), PgxPool(), SQLiteDB(), and Stats().
	Runtime *database.Runtime

	// DSN is the Postgres connection string for this test's isolated database.
	// Empty for SQLite fixtures.
	DSN string
}
