# Data Model

This feature does not introduce new persisted entities. It introduces one runtime-only abstraction.

## Entity: `Runtime` (extended)

Location: `internal/database/database.go`

**Purpose**: Process-level handle owning the active database connection pool and providing access to GORM, raw SQL, and pool statistics through a single source of truth.

### Fields (post-change)

| Field | Type | When Set | Notes |
|-------|------|----------|-------|
| `Driver` | `Driver` (existing enum: `postgres` / `sqlite`) | At `Open()` | Unchanged |
| `DB` | `*gorm.DB` | At `Open()` | Unchanged. Backed by shared underlying pool. |
| `PermissionStatus` | `PermissionStatus` | At `Open()` (SQLite only) | Unchanged |
| `pgxPool` | `*pgxpool.Pool` (unexported) | At `Open()` when `Driver == DriverPostgres` | New. Nil when Driver != postgres. |
| `sqliteDB` | `*sql.DB` (unexported) | At `Open()` when `Driver == DriverSQLite` | New. Nil when Driver != sqlite. |

### Methods (post-change)

| Method | Signature | Behavior |
|--------|-----------|----------|
| `PgxPool()` | `func (r *Runtime) PgxPool() *pgxpool.Pool` | Returns pool, or nil if not Postgres. |
| `SQLiteDB()` | `func (r *Runtime) SQLiteDB() *sql.DB` | Returns `*sql.DB`, or nil if not SQLite. |
| `Stats()` | `func (r *Runtime) Stats() RuntimeStats` | Returns identity + counts (see `RuntimeStats` below). |
| `GormDB()` | `func (r *Runtime) GormDB() *gorm.DB` | Convenience getter for `r.DB`. |
| existing `Open`, `Init`, `Instance`, `Ping`, `Reset`, `DriverName` | unchanged | Behavior preserved (FR-011). |

### Companion type: `RuntimeStats`

```go
type RuntimeStats struct {
    Driver      Driver
    OpenConns   int
    InUseConns  int
    IdleConns   int
    PoolPointer uintptr // identity sentinel for single-pool assertion
}
```

### Lifecycle / state transitions

```
   uninitialized
        │
        │  Open(ctx, cfg)            ── opens pool (pgxpool OR sql.Open modernc)
        ▼                            ── derives *sql.DB / *gorm.DB from same pool
   initialized ──────────────────► running
        │   runMigrations + ValidateStartupSchema (existing)
        │
        ▼
     Reset()                          ── closes pool, clears activeRuntime
        │
        ▼
   uninitialized
```

Single physical pool per dialect throughout `running` state. Verified by `Stats().PoolPointer` returning identical value across repeated calls and across handles (GORM-side `DB().DB()` pointer ≡ `PgxPool()`/`SQLiteDB()` pointer).

### Validation rules

- `Open` MUST fail if pool open fails (existing behavior preserved).
- `Open` MUST fail if migrations fail (existing).
- `Open` MUST fail if `ValidateStartupSchema` fails (existing).
- New: `Open` MUST set exactly one of `pgxPool` / `sqliteDB` (the one matching `Driver`); the other MUST be nil.
- New: `Stats().PoolPointer` MUST be stable across calls within a single Runtime instance.

### Out of scope (no changes)

- Domain models (`internal/domain/models.go`)
- Migrations (`internal/database/migrations/{sqlite,postgres}/`)
- Any repository (`internal/repository/store/`)
- Service layer
- Handlers, DTOs, router
