# Data Model

No persistent entity. The feature introduces test-only Go types under `internal/repository/internaltest/`.

## `DialectFixture` — `internal/repository/internaltest/fixture.go`

```go
type DialectFixture struct {
    Dialect string             // "sqlite" or "postgres"
    Runtime *database.Runtime  // 041's Runtime — GormDB(), PgxPool(), SQLiteDB(), Stats()
    DSN     string             // Postgres only: full DSN of THIS test's DB. Empty for SQLite.
}
```

**Invariants**:
- `Dialect` matches the runtime's `Runtime.Driver`.
- For `"postgres"`: `Runtime.PgxPool()` non-nil, `Runtime.SQLiteDB()` nil, `DSN` non-empty.
- For `"sqlite"`: `Runtime.SQLiteDB()` non-nil, `Runtime.PgxPool()` nil, `DSN` empty.
- Cleanup is registered via `t.Cleanup` at fixture creation time.

## `PgContainer` — `internal/repository/internaltest/container.go`

Internal helper, not exported beyond the package.

```go
type PgContainer struct {
    container testcontainers.Container // nil when using POSTGRES_TEST_DSN passthrough
    adminDSN  string                   // superuser DSN, used to CREATE/DROP per-test DBs
    template  string                   // "template_ogoune"
}

// Acquire returns a DSN for an isolated per-test database cloned from the template.
func (c *PgContainer) Acquire(t testing.TB) string
```

**Lifecycle** (`sync.Once` per process, one-per-package via `var once sync.Once`):
1. Start container (or use `POSTGRES_TEST_DSN` directly).
2. Connect as superuser.
3. `CREATE DATABASE template_ogoune ENCODING UTF8 IS_TEMPLATE true`.
4. Apply migrations against `template_ogoune` using `internal/database/migrations/postgres`.
5. Mark fully ready.

**Per-test (`Acquire`)**:
1. `name := "ogoune_test_" + ulid.New().String()`.
2. `CREATE DATABASE <name> TEMPLATE template_ogoune` (≈ ms).
3. Return DSN with `dbname=<name>`.
4. `t.Cleanup`: close connections, `DROP DATABASE <name> WITH (FORCE)`.

## `Factory` (logical type, not exported)

Each contract test file defines a local factory type:

```go
type tagsFactory func(*gorm.DB) port.TagsRepository
```

Default factory: `store.NewTagsRepository`. Future sqlc-backed factories may be added by later tickets without touching the contract body.

## Public package API surface (`internal/repository/internaltest`)

| Function | Signature | Purpose |
|----------|-----------|---------|
| `SetupSQLite` | `func(t *testing.T) *DialectFixture` | Per-test SQLite DB in `t.TempDir()`, migrations applied |
| `SetupPostgres` | `func(t *testing.T) *DialectFixture` | Per-test PG DB via container OR `POSTGRES_TEST_DSN`. Calls `t.Skip` if neither available. |
| `ForEachDialect` | `func(t *testing.T, fn func(t *testing.T, fx *DialectFixture))` | Runs `fn` once per supported dialect under named sub-tests |
| `DialectsAvailable` | `func() []string` | Lists which dialects will actually run (used by CI introspection scripts) |

## Out of scope (no data changes)

- Any domain entity.
- Any production repository.
- Any migration file.
- Any production code under `cmd/`, `internal/api/`, `internal/service/`, `internal/repository/store/`, or `internal/repository/store/database/`.
- Any change to `internal/database/test_helpers_test.go` (SQLite-only path stays).
