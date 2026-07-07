# internaltest — Dual-dialect repository test helpers

Test-only package. Provides isolated SQLite + Postgres fixtures with migrations applied, so a single repository contract body can be validated against both backends.

## Writing a dual-dialect test

```go
func TestMyRepository_Contract(t *testing.T) {
    internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
        repo := store.NewMyRepository(fx.Runtime.GormDB())
        // …assertions against repo…
    })
}
```

The helper:
- Applies migrations (via the project's `database.Open` path — same code production uses).
- Hands you an isolated, per-test database.
- Registers cleanup via `t.Cleanup`.

For tests that only need SQLite (or only Postgres), call `SetupSQLite(t)` or `SetupPostgres(t)` directly.

## Running locally — SQLite only (default)

```bash
make test-be
```

Postgres sub-tests skip with a clear message; SQLite sub-tests run. No Docker required.

## Running locally — both dialects

Pre-requisite: Docker reachable (Docker Desktop, Colima, OrbStack).

```bash
make test-be-pg
```

This boots a single `postgres:16-alpine` testcontainer per `go test` package via testcontainers-go.

If you already have a local Postgres you trust:

```bash
POSTGRES_TEST_DSN='postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable' make test-be-pg
```

The helper detects the DSN, skips Docker entirely, and creates one isolated test database per test directly against your local Postgres.

## How per-test isolation works

| Dialect | Strategy | Per-test cost |
|---------|----------|---------------|
| SQLite  | Fresh DB file under `t.TempDir()` per fixture | a few ms |
| Postgres | One container per `go test` package; migrations applied **once** to a `template_ogoune` database; per-test `CREATE DATABASE … TEMPLATE template_ogoune`; `DROP DATABASE … WITH (FORCE)` on cleanup | typically <50 ms |

Two parallel tests under `ForEachDialect` cannot see each other's data — verified by `TestIsolation_ParallelSamePrimaryKey`.

## CI

| Job | What runs | Budget |
|-----|-----------|--------|
| `test` / `backend-tests` (existing) | SQLite-only path | unchanged |
| `test-be-postgres` (new) | Both dialects via testcontainers | ≤ 180s wall-clock |

Both jobs run independently. If Postgres breaks, the SQLite job is still the fast feedback path.

## Public API

| Function | Purpose |
|----------|---------|
| `SetupSQLite(t) *DialectFixture` | Per-test SQLite DB |
| `SetupPostgres(t) *DialectFixture` | Per-test Postgres DB (skip-aware) |
| `ForEachDialect(t, fn)` | Run `fn` once per available dialect in named sub-tests |
| `DialectsAvailable() []string` | Lists which dialects will execute in this env |
| `GetTestDB(t) *gorm.DB` | Legacy single-dialect helper (preserved, not deprecated) |

`DialectFixture` exposes:
- `Dialect string` — `"sqlite"` or `"postgres"`
- `Runtime *database.Runtime` — 041's runtime; use `GormDB()` / `PgxPool()` / `SQLiteDB()` / `Stats()`
- `DSN string` — Postgres only; the connection string of THIS test's DB

## Troubleshooting

### `postgres backend unavailable: ... (set POSTGRES_TEST_DSN or run Docker)`

Expected when running locally without Docker. Set `POSTGRES_TEST_DSN` or start Docker.

### Restricted CI runners (no Ryuk reaper)

If your runner blocks side-car containers, the testcontainers Ryuk reaper may fail. Set:

```bash
TESTCONTAINERS_RYUK_DISABLED=true
```

The orphaned `ogoune_test_*` databases will be dropped on container teardown anyway.

### Production-import guard

This package MUST NOT be imported from non-test production code. CI asserts:

```bash
go list -deps ./cmd/... | grep -E 'testcontainers|internal/repository/internaltest'  # must return empty
```

## What this package does NOT do

- Run business-logic tests — only infrastructure for repository tests.
- Provide fakes — see `internal/repository/fake/` for those.
- Replace `internal/database/test_helpers_test.go` for tests in the `database` package.
