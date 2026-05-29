package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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
//   - The raw handle and DB (GORM) share the same underlying physical pool.
type Runtime struct {
	Driver           Driver
	DB               *gorm.DB
	PermissionStatus PermissionStatus

	pgxPool  *pgxpool.Pool
	sqliteDB *sql.DB
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

// GormDB returns the GORM handle. Convenience getter mirroring the PRD wording.
func (r *Runtime) GormDB() *gorm.DB {
	if r == nil {
		return nil
	}
	return r.DB
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
	Driver       Driver
	DSN          string
	SQLitePath   string
	GormLogLevel logger.LogLevel
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
		db               *gorm.DB
		permissionStatus = PermissionStatusNotApplicable
		pgxPool          *pgxpool.Pool
		sqliteDB         *sql.DB
		err              error
	)

	switch cfg.Driver {
	case DriverPostgres:
		db, pgxPool, err = openPostgres(ctx, cfg)
	case DriverSQLite:
		db, permissionStatus, err = openSQLite(cfg)
		if err == nil {
			sqliteDB, err = db.DB()
		}
	default:
		return nil, fmt.Errorf("db init: unsupported driver %q", cfg.Driver)
	}
	if err != nil {
		return nil, err
	}

	if err := runMigrations(ctx, db, cfg.Driver, migrationFS); err != nil {
		return nil, err
	}

	if err := ValidateStartupSchema(db); err != nil {
		return nil, err
	}

	if cfg.Driver == DriverSQLite {
		permissionStatus = mergePermissionStatus(permissionStatus, hardenSQLiteArtifacts(cfg.SQLitePath))
	}

	return &Runtime{
		Driver:           cfg.Driver,
		DB:               db,
		PermissionStatus: permissionStatus,
		pgxPool:          pgxPool,
		sqliteDB:         sqliteDB,
	}, nil
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

// Instance returns the active gorm database handle, lazily initializing from env when needed.
//
// Deprecated: prefer obtaining the full *Runtime from bootstrap and using
// Runtime.GormDB(), Runtime.PgxPool(), or Runtime.SQLiteDB() as appropriate.
// This accessor is preserved during the sqlc migration and will be removed
// in a follow-up ticket.
func Instance() (*gorm.DB, error) {
	runtimeMu.RLock()
	currentRuntime := activeRuntime
	isInitialized := initializedRuntime
	runtimeMu.RUnlock()

	if isInitialized {
		if currentRuntime == nil || currentRuntime.DB == nil {
			return nil, fmt.Errorf("db instance: database initialization failed previously")
		}
		return currentRuntime.DB, nil
	}

	if err := Init(context.Background(), configFromEnv()); err != nil {
		return nil, fmt.Errorf("db instance: initialization failed: %w", err)
	}

	runtimeMu.RLock()
	defer runtimeMu.RUnlock()

	if activeRuntime == nil || activeRuntime.DB == nil {
		return nil, fmt.Errorf("db instance: database not initialized")
	}

	return activeRuntime.DB, nil
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
	db, err := Instance()
	if err != nil {
		return fmt.Errorf("db ping: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("db ping: failed to get underlying db: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("db ping: connection failed: %w", err)
	}

	return nil
}

// Reset clears the singleton state for tests.
func Reset() {
	runtimeMu.Lock()
	defer runtimeMu.Unlock()

	if activeRuntime != nil {
		if activeRuntime.DB != nil {
			sqlDB, err := activeRuntime.DB.DB()
			if err == nil && sqlDB != nil {
				_ = sqlDB.Close()
			}
		}
		if activeRuntime.pgxPool != nil {
			activeRuntime.pgxPool.Close()
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

	gormLogLevel, err := parseLogLevel(cfg.LogLevel)
	if err != nil {
		return resolvedConfig{}, err
	}

	switch driver {
	case DriverPostgres:
		dsn := strings.TrimSpace(cfg.DatabaseURL)
		if dsn == "" {
			return resolvedConfig{}, fmt.Errorf("db init: DATABASE_URL is required when DB_DRIVER=postgres")
		}
		return resolvedConfig{
			Driver:       driver,
			DSN:          dsn,
			GormLogLevel: gormLogLevel,
		}, nil
	case DriverSQLite:
		path := strings.TrimSpace(cfg.SQLitePath)
		if path == "" {
			path = "ogoune.db"
		}
		return resolvedConfig{
			Driver:       driver,
			DSN:          path,
			SQLitePath:   path,
			GormLogLevel: gormLogLevel,
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

func parseLogLevel(value string) (logger.LogLevel, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "error":
		return logger.Error, nil
	case "silent":
		return logger.Silent, nil
	case "warn":
		return logger.Warn, nil
	case "info":
		return logger.Info, nil
	default:
		return logger.Error, fmt.Errorf("db init: unsupported DB_LOG_LEVEL %q", value)
	}
}

func newGormLogger(level logger.LogLevel) logger.Interface {
	return logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  level,
			Colorful:                  false,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
		},
	)
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

// ValidateStartupSchema validates required schema columns before the application starts serving traffic.
func ValidateStartupSchema(db *gorm.DB) error {
	if !db.Migrator().HasColumn("resources", "confirmation_checks") {
		return fmt.Errorf("db init: missing required confirmation schema column resources.confirmation_checks; run latest migrations")
	}
	if !db.Migrator().HasColumn("resources", "confirmation_interval") {
		return fmt.Errorf("db init: missing required confirmation schema column resources.confirmation_interval; run latest migrations")
	}
	if !db.Migrator().HasColumn("notification_events", "status") {
		return fmt.Errorf("db init: missing required notification retry schema column notification_events.status; run latest migrations")
	}
	if !db.Migrator().HasColumn("notification_events", "claim_owner") {
		return fmt.Errorf("db init: missing required notification retry schema column notification_events.claim_owner; run latest migrations")
	}
	if !db.Migrator().HasColumn("notification_events", "claimed_at") {
		return fmt.Errorf("db init: missing required notification retry schema column notification_events.claimed_at; run latest migrations")
	}
	if !db.Migrator().HasColumn("notification_events", "processed_at") {
		return fmt.Errorf("db init: missing required notification retry schema column notification_events.processed_at; run latest migrations")
	}
	if !db.Migrator().HasColumn("notification_events", "last_error") {
		return fmt.Errorf("db init: missing required notification retry schema column notification_events.last_error; run latest migrations")
	}
	return nil
}
