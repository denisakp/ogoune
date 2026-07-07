package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Driver string

const (
	DriverPostgres Driver = "postgres"
	DriverSQLite   Driver = "sqlite"

	DefaultConfirmationChecks   = 2
	DefaultConfirmationInterval = 30
)

type PermissionStatus string

const (
	PermissionStatusNotApplicable PermissionStatus = "not_applicable"
	PermissionStatusEnforced      PermissionStatus = "enforced"
	PermissionStatusWarned        PermissionStatus = "warned"
)

// Config describes the supported database runtime options.
type Config struct {
	Driver      Driver
	DatabaseURL string
	SQLitePath  string
	LogLevel    string
}

// Runtime captures the active database runtime selected at startup.
//
// Invariants:
//   - Exactly one of pgxPool / sqliteDB is non-nil, matching Driver.
//   - For Postgres, pgxPool is the production handle for sqlc; pgSQL is a
//     stdlib.OpenDBFromPool *sql.DB used internally by migrations and the
//     startup schema validator. Both share the same underlying physical pool.
type Runtime struct {
	Driver           Driver
	PermissionStatus PermissionStatus

	pgxPool  *pgxpool.Pool
	pgSQL    *sql.DB // Postgres: stdlib.OpenDBFromPool(pgxPool). Migrations/validation only.
	sqliteDB *sql.DB // SQLite: the production handle.
}

// RuntimeStats exposes pool identity and live counts for observability and
// single-pool assertion in tests.
type RuntimeStats struct {
	Driver      Driver
	OpenConns   int
	InUseConns  int
	IdleConns   int
	PoolPointer uintptr
}

// PgxPool returns the underlying pgx pool. Nil when Driver != DriverPostgres.
func (r *Runtime) PgxPool() *pgxpool.Pool {
	if r == nil {
		return nil
	}
	return r.pgxPool
}

// SQLiteDB returns the underlying SQLite *sql.DB. Nil when Driver != DriverSQLite.
func (r *Runtime) SQLiteDB() *sql.DB {
	if r == nil {
		return nil
	}
	return r.sqliteDB
}

// migratorDB returns the *sql.DB used by the migrator + ValidateStartupSchema.
// For Postgres this is the stdlib wrapper around the pgx pool; for SQLite
// it is the production handle.
func (r *Runtime) migratorDB() *sql.DB {
	if r == nil {
		return nil
	}
	switch r.Driver {
	case DriverPostgres:
		return r.pgSQL
	case DriverSQLite:
		return r.sqliteDB
	}
	return nil
}

// Stats returns pool identity + live counts. PoolPointer is stable across
// calls for the lifetime of the Runtime and identifies the single underlying pool.
func (r *Runtime) Stats() RuntimeStats {
	if r == nil {
		return RuntimeStats{}
	}
	out := RuntimeStats{Driver: r.Driver}
	switch r.Driver {
	case DriverPostgres:
		if r.pgxPool != nil {
			out.PoolPointer = reflect.ValueOf(r.pgxPool).Pointer()
			s := r.pgxPool.Stat()
			out.OpenConns = int(s.TotalConns())
			out.InUseConns = int(s.AcquiredConns())
			out.IdleConns = int(s.IdleConns())
		}
	case DriverSQLite:
		if r.sqliteDB != nil {
			out.PoolPointer = reflect.ValueOf(r.sqliteDB).Pointer()
			s := r.sqliteDB.Stats()
			out.OpenConns = s.OpenConnections
			out.InUseConns = s.InUse
			out.IdleConns = s.Idle
		}
	}
	return out
}

type resolvedConfig struct {
	Driver     Driver
	DSN        string
	SQLitePath string
}

var (
	runtimeMu          sync.RWMutex
	activeRuntime      *Runtime
	initErr            error
	initializedRuntime bool
)

// Open constructs a driver-specific runtime and applies pending SQL migrations.
func Open(ctx context.Context, cfg Config) (*Runtime, error) {
	resolved, err := cfg.resolve()
	if err != nil {
		return nil, err
	}

	return openRuntime(ctx, resolved, embeddedMigrations)
}

func openRuntime(ctx context.Context, cfg resolvedConfig, migrationFS migrationFS) (*Runtime, error) {
	var (
		rt  = &Runtime{Driver: cfg.Driver, PermissionStatus: PermissionStatusNotApplicable}
		err error
	)

	switch cfg.Driver {
	case DriverPostgres:
		rt.pgxPool, rt.pgSQL, err = openPostgres(ctx, cfg)
	case DriverSQLite:
		rt.sqliteDB, rt.PermissionStatus, err = openSQLite(cfg)
	default:
		return nil, fmt.Errorf("db init: unsupported driver %q", cfg.Driver)
	}
	if err != nil {
		return nil, err
	}

	if err := runMigrations(ctx, rt.migratorDB(), cfg.Driver, migrationFS); err != nil {
		return nil, err
	}

	if err := ValidateStartupSchema(rt.migratorDB()); err != nil {
		return nil, err
	}

	if cfg.Driver == DriverSQLite {
		rt.PermissionStatus = mergePermissionStatus(rt.PermissionStatus, hardenSQLiteArtifacts(cfg.SQLitePath))
	}

	return rt, nil
}

// Init initializes the shared database singleton exactly once for the process.
func Init(ctx context.Context, cfg Config) error {
	runtimeMu.Lock()
	defer runtimeMu.Unlock()

	if initializedRuntime {
		return initErr
	}

	initializedRuntime = true

	runtime, err := Open(ctx, cfg)
	if err != nil {
		initErr = err
		return initErr
	}

	activeRuntime = runtime
	initErr = nil
	slog.Info("database initialization completed", "driver", string(runtime.Driver), "permission_status", string(runtime.PermissionStatus))
	return nil
}

// ActiveRuntime returns the active *Runtime once the singleton has been
// initialized. Returns an error if Init has not been called or failed.
func ActiveRuntime() (*Runtime, error) {
	runtimeMu.RLock()
	defer runtimeMu.RUnlock()

	if activeRuntime == nil {
		return nil, fmt.Errorf("db runtime: not initialized")
	}
	return activeRuntime, nil
}

// DriverName returns the active driver name once the singleton has been initialized.
func DriverName() (Driver, error) {
	runtimeMu.RLock()
	defer runtimeMu.RUnlock()

	if activeRuntime == nil {
		return "", fmt.Errorf("db driver: runtime not initialized")
	}

	return activeRuntime.Driver, nil
}

// Ping validates the underlying SQL connection.
func Ping(ctx context.Context) error {
	rt, err := ActiveRuntime()
	if err != nil {
		return fmt.Errorf("db ping: %w", err)
	}
	switch rt.Driver {
	case DriverPostgres:
		if err := rt.pgxPool.Ping(ctx); err != nil {
			return fmt.Errorf("db ping: postgres: %w", err)
		}
	case DriverSQLite:
		if err := rt.sqliteDB.PingContext(ctx); err != nil {
			return fmt.Errorf("db ping: sqlite: %w", err)
		}
	default:
		return fmt.Errorf("db ping: unsupported driver %q", rt.Driver)
	}
	return nil
}

// Reset clears the singleton state for tests.
func Reset() {
	runtimeMu.Lock()
	defer runtimeMu.Unlock()

	if activeRuntime != nil {
		if activeRuntime.pgSQL != nil {
			_ = activeRuntime.pgSQL.Close()
		}
		if activeRuntime.pgxPool != nil {
			activeRuntime.pgxPool.Close()
		}
		if activeRuntime.sqliteDB != nil {
			_ = activeRuntime.sqliteDB.Close()
		}
	}

	activeRuntime = nil
	initErr = nil
	initializedRuntime = false
}

func (cfg Config) resolve() (resolvedConfig, error) {
	driver := Driver(strings.ToLower(strings.TrimSpace(string(cfg.Driver))))
	if driver == "" {
		driver = DriverPostgres
	}

	switch driver {
	case DriverPostgres:
		dsn := strings.TrimSpace(cfg.DatabaseURL)
		if dsn == "" {
			return resolvedConfig{}, fmt.Errorf("db init: DATABASE_URL is required when DB_DRIVER=postgres")
		}
		return resolvedConfig{
			Driver: driver,
			DSN:    dsn,
		}, nil
	case DriverSQLite:
		path := strings.TrimSpace(cfg.SQLitePath)
		if path == "" {
			path = "ogoune.db"
		}
		return resolvedConfig{
			Driver:     driver,
			DSN:        path,
			SQLitePath: path,
		}, nil
	default:
		return resolvedConfig{}, fmt.Errorf("db init: unsupported DB_DRIVER %q", driver)
	}
}

func configFromEnv() Config {
	return Config{
		Driver:      Driver(getEnv("DB_DRIVER", string(DriverSQLite))),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://ogoune:EE94PPHGz3TZ@postgres:5432/pulse?sslmode=disable"),
		SQLitePath:  getEnv("SQLITE_PATH", "ogoune.db"),
		LogLevel:    getEnv("DB_LOG_LEVEL", "error"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func mergePermissionStatus(current, next PermissionStatus) PermissionStatus {
	if current == PermissionStatusWarned || next == PermissionStatusWarned {
		return PermissionStatusWarned
	}
	if current == PermissionStatusEnforced || next == PermissionStatusEnforced {
		return PermissionStatusEnforced
	}
	return PermissionStatusNotApplicable
}

// ValidateStartupSchema validates required schema columns before the application
// starts serving traffic. Uses a dialect-agnostic probe (`SELECT col FROM table LIMIT 0`)
// — if the column is missing, the SQL driver returns an error pointing at it.
func ValidateStartupSchema(db *sql.DB) error {
	required := []struct {
		Table, Column string
	}{
		{"resources", "confirmation_checks"},
		{"resources", "confirmation_interval"},
		{"notification_events", "status"},
		{"notification_events", "claim_owner"},
		{"notification_events", "claimed_at"},
		{"notification_events", "processed_at"},
		{"notification_events", "last_error"},
	}
	for _, r := range required {
		// Defence in depth: identifier interpolation is safe here because both
		// table and column are hardcoded constants in this slice (not user-derived).
		q := fmt.Sprintf("SELECT %s FROM %s LIMIT 0", r.Column, r.Table)
		if _, err := db.Exec(q); err != nil {
			return fmt.Errorf("db init: missing required schema column %s.%s; run latest migrations: %w", r.Table, r.Column, err)
		}
	}
	return nil
}
