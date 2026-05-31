# Contract — `internaltest` Helper API

Package: `internal/repository/internaltest`

## Public functions

```go
// SetupSQLite returns a freshly-migrated, per-test SQLite fixture.
// The DB file lives under t.TempDir(); cleanup is registered via t.Cleanup.
func SetupSQLite(t *testing.T) *DialectFixture

// SetupPostgres returns a per-test Postgres fixture. Provisioning:
//   - If POSTGRES_TEST_DSN is set, use that admin DSN and clone a per-test DB.
//   - Else, start a postgres:16-alpine container via testcontainers-go
//     (one container per Go test package, memoized via sync.Once).
//   - If neither testcontainers (no Docker) nor POSTGRES_TEST_DSN is available,
//     call t.Skip with a clear message and return nil.
func SetupPostgres(t *testing.T) *DialectFixture

// ForEachDialect runs fn once per supported dialect in named sub-tests:
//   t.Run("sqlite",   func(t){ fn(t, SetupSQLite(t)) })
//   t.Run("postgres", func(t){ fn(t, SetupPostgres(t)) })   // skip-aware
// Skipping the postgres iteration MUST NOT skip the sqlite iteration.
func ForEachDialect(t *testing.T, fn func(t *testing.T, fx *DialectFixture))

// DialectsAvailable returns the dialects that will actually execute in the
// current environment (for CI introspection / log clarity).
func DialectsAvailable() []string
```

## `DialectFixture` (data type)

See `data-model.md`. Stable struct returned by both `Setup*` functions.

## Behavior contract

| Concern | Contract |
|---------|----------|
| Per-test isolation | Each fixture wraps a freshly-cloned database. Two parallel tests MUST NOT see each other's data. |
| Migrations | Applied once per process per dialect (SQLite: per fixture; Postgres: once to `template_ogoune`). Per-test setup time after migration ≈ ms. |
| Cleanup | Registered via `t.Cleanup`. On failure: SQLite temp dir removed; Postgres DB dropped (`WITH (FORCE)`); connections closed. Errors during cleanup MUST report via `t.Errorf` (not silent). |
| Skip semantics | `SetupPostgres` calls `t.Skip` with: `"postgres backend unavailable (no Docker, no POSTGRES_TEST_DSN)"`. `ForEachDialect` continues with the sqlite sub-test. |
| Production import | Package MUST NOT be imported from non-test production code. Enforced by CI invariant: `go list -deps ./cmd/api/... | grep -E 'internaltest|testcontainers'` returns empty. |

## Error semantics

- Container start failure → fail the test with a wrapped error naming Docker availability.
- `CREATE DATABASE` failure → wrap and surface; mark cleanup of the partial state safely.
- Migration application failure → wrap with the failing migration filename.
- Postgres unreachable via passthrough DSN → fail fast with a sanitized DSN (no password) in the message.

## Stability contract

Once shipped, `SetupSQLite`, `SetupPostgres`, `ForEachDialect`, and `DialectFixture` are stable API for the lifetime of the sqlc migration track. Breaking changes require a follow-up ticket.
