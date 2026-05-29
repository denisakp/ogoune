# Contract — `database.Runtime` Public API (post-change)

Package: `internal/database`

## Existing surface (preserved verbatim)

```go
type Driver string
const (
    DriverPostgres Driver = "postgres"
    DriverSQLite   Driver = "sqlite"
)

type Runtime struct {
    Driver           Driver
    DB               *gorm.DB
    PermissionStatus PermissionStatus
    // + unexported pgxPool, sqliteDB (new)
}

func Open(ctx context.Context, cfg Config) (*Runtime, error)
func Init(ctx context.Context, cfg Config) error
func Instance() (*gorm.DB, error)   // Deprecated, kept as wrapper
func DriverName() (Driver, error)
func Ping(ctx context.Context) error
func Reset()
```

## New surface

```go
// Returns the underlying pgx pool. Nil when Driver != DriverPostgres.
func (r *Runtime) PgxPool() *pgxpool.Pool

// Returns the underlying SQLite *sql.DB. Nil when Driver != DriverSQLite.
func (r *Runtime) SQLiteDB() *sql.DB

// Convenience getter for r.DB (matches PRD wording).
func (r *Runtime) GormDB() *gorm.DB

// Pool identity + counts. Backing pointer is stable for the lifetime of the Runtime.
func (r *Runtime) Stats() RuntimeStats

type RuntimeStats struct {
    Driver      Driver
    OpenConns   int
    InUseConns  int
    IdleConns   int
    PoolPointer uintptr
}
```

## Invariants

1. Exactly one of `PgxPool()` / `SQLiteDB()` returns non-nil for any initialized `*Runtime`.
2. `Stats().PoolPointer` is invariant across calls on the same `*Runtime` instance.
3. `r.DB` (GORM) is backed by the same physical pool as the matching raw accessor.
4. `Instance()` continues to return `activeRuntime.DB` with unchanged error semantics.
5. `Reset()` closes the underlying pool exactly once.

## Backwards compatibility

- All existing call sites of `database.Instance()`, `database.DriverName()`, `database.Ping()`, `database.Reset()` continue to compile and behave identically.
- Repository wiring in `internal/platform/bootstrap/database.go` adjusted only where it constructs the GORM DB (now consumes the new Runtime opener internally — no API change visible to repos).
