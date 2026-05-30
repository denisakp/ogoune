# Quickstart — Maintainer Workflow

## Writing a dual-dialect test

```go
func TestMyRepository(t *testing.T) {
    internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
        repo := store.NewMyRepository(fx.Runtime.GormDB())
        // ... assertions against repo
    })
}
```

That is the full setup. The helper handles:
- Migrations (already applied when your `fn` runs).
- Per-test isolated database (no cross-test data leak).
- Cleanup (registered automatically via `t.Cleanup`).

## Running locally — SQLite only (default)

```bash
make test-be          # runs everything that doesn't require Docker
```

The Postgres sub-tests skip with a clear message; SQLite sub-tests run.

## Running locally — both dialects

Pre-requisite: Docker running (Docker Desktop, Colima, OrbStack — any will do).

```bash
make test-be-pg       # boots testcontainers-managed postgres:16-alpine
```

Or, if you already have a local Postgres you trust:

```bash
POSTGRES_TEST_DSN='postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable' make test-be-pg
```

The helper detects the DSN, skips Docker, and creates one isolated test DB per test directly against your local Postgres.

## CI behavior

| Job | What runs | Budget |
|-----|-----------|--------|
| `test-be` (existing) | SQLite-only path | unchanged |
| `test-be-postgres` (new) | Both dialects via testcontainers | ≤ 180s wall-clock |

The two jobs run independently — if Postgres breaks, the SQLite job is still your fast feedback.

## What testcontainers does for you

- One `postgres:16-alpine` container per Go test package (memoized in `sync.Once`).
- Container migrates a `template_ogoune` database once.
- Each test gets `CREATE DATABASE … TEMPLATE template_ogoune` — milliseconds.
- Ryuk reaper cleans up orphaned containers if your test panics or you ^C.

## What this ticket does NOT change

- The existing `make test-be` target stays SQLite-only and fast.
- Production code (`cmd/`, `internal/api/`, `internal/service/`, `internal/repository/store/` non-test) is untouched.
- No new env var in production deploy paths.
- Existing tests under `internal/database/` that use the SQLite-only helper continue to work — no consolidation here.
