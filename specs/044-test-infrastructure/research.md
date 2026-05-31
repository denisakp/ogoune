# Phase 0 — Research

## R1. Testcontainers-go module (locked by Q1)

**Decision**: Use `github.com/testcontainers/testcontainers-go` + `github.com/testcontainers/testcontainers-go/modules/postgres`.

**Rationale**:
- Idiomatic Go API; works on Linux/Darwin with Docker Desktop, Colima, OrbStack.
- The `postgres` module ships built-in `WaitFor` strategies and a `ConnectionString()` helper.
- Active maintenance, used widely in Go test code.

**Alternatives**:
- `dockertest` (ory/dockertest). Rejected: less idiomatic API; spec/PRD lean towards testcontainers.
- `embedded-postgres`. Rejected: not officially supported beyond Linux; brings binary into module cache.

## R2. Container lifecycle — one per package (locked by Q2)

**Decision**: Each test package that wants Postgres calls `internaltest.PostgresContainer(tb testing.TB) *PgContainer`. The helper:

1. Inside `sync.Once`: starts a `postgres:16-alpine` container with `WaitFor=log_line_matched("ready to accept connections")`.
2. Connects as superuser, creates `template_ogoune`, applies the project's existing migrations against it, marks it `template` so it cannot be written to.
3. Returns a `*PgContainer` exposing `(ctx) (db *sql.DB, dsn string, drop func())`. Per-test `Acquire(t)`:
   - Generates a fresh DB name `ogoune_test_<ULID>`.
   - Runs `CREATE DATABASE … TEMPLATE template_ogoune` (≈ ms — Postgres copies the template's file tree).
   - Hands the DSN/connection to the test.
   - Registers `t.Cleanup`: close, then `DROP DATABASE … WITH (FORCE)`.

Container is dropped at process exit via `runtime.SetFinalizer` or `m.Cleanup` if `TestMain` is present. Simplest: rely on testcontainers' Ryuk to reap orphaned containers — no explicit teardown needed; Ryuk runs in CI by default.

**Rationale**: Amortizes the ~5-10s container start across all tests in the package while keeping per-test isolation at sub-100ms cost. Hits the 3-minute budget.

**Alternatives covered in Clarification Q2** (A/C/D) — rejected.

## R3. `POSTGRES_TEST_DSN` opt-out for containers

**Decision**: When `POSTGRES_TEST_DSN` is set in the env, `PostgresContainer` short-circuits and returns a `*PgContainer` backed by the external DSN. The helper still creates per-test databases (it has admin rights on the external DSN by assumption) but does NOT start Docker.

**Rationale**: Devs already running a local Postgres can opt out of Docker overhead. CI sets nothing; the container path runs.

**Alternatives**: Force testcontainers always. Rejected: extra friction for local dev.

## R4. SQLite path

**Decision**: `SetupSQLite(t *testing.T) *DialectFixture` creates a fresh `t.TempDir()/test.db`, opens it via the existing `internal/database` `Open` path (so migrations + `glebarez/sqlite` driver are wired identically to production), and registers `t.Cleanup` to close the underlying handles. No `:memory:` shared-cache complications.

**Rationale**: Mirrors the production SQLite open path so tests catch real driver-related bugs.

**Alternatives**: Use `file::memory:?cache=shared`. Rejected: complicates per-test isolation, parallel goroutines may share connections unexpectedly.

## R5. `DialectFixture` shape — single struct, both handles

**Decision**: `DialectFixture` holds:

```go
type DialectFixture struct {
    Dialect  string         // "sqlite" or "postgres"
    Runtime  *database.Runtime  // 041's Runtime, exposing GormDB(), PgxPool(), SQLiteDB(), Stats()
    DSN      string         // for Postgres: the connection string of THIS test's DB
}
```

**Rationale**: Reuse 041's `Runtime` so the helper's output is exactly what production code consumes. Tests can wire any repository factory with `fx.Runtime.GormDB()` or future sqlc consumers with `fx.Runtime.PgxPool()`.

## R6. `ForEachDialect` API

**Decision**:

```go
func ForEachDialect(t *testing.T, fn func(t *testing.T, fx *DialectFixture)) {
    t.Run("sqlite", func(t *testing.T) {
        fn(t, SetupSQLite(t))
    })
    t.Run("postgres", func(t *testing.T) {
        fx := SetupPostgres(t)  // calls t.Skip when neither testcontainers nor DSN available
        if fx == nil { return } // Skip already called
        fn(t, fx)
    })
}
```

**Rationale**: Sub-tests give clear failure localization in `go test -v` output. Skipping the Postgres iteration does NOT skip SQLite (FR-004 acceptance).

## R7. Contract-test refactor shape

**Decision**: Each refactored file gains a top-level test function that takes a factory and is invoked by `ForEachDialect`. The existing per-method `t.Run("Create", …)` structure is preserved.

```go
func TestTagsRepository_Contract(t *testing.T) {
    internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
        repo := store.NewTagsRepository(fx.Runtime.GormDB())
        runTagsContract(t, repo)
    })
}

func runTagsContract(t *testing.T, repo port.TagsRepository) { /* existing assertions */ }
```

Future tickets add additional drivers (`TestTagsRepository_SqlcContract(t)` reusing `runTagsContract`).

## R8. Fake-specific assertions migration

**Decision**: An assertion is fake-specific iff it would not hold for the GORM repository AND it has independent value (e.g. asserts the fake's in-memory map has N entries internally). Such assertions move to `internal/repository/fake/<entity>_fake_test.go` or stay in their existing `*_test.go` file under that package. Fake-shape tests of "did the fake compile / round-trip" are dropped if they offer no value beyond compile-time checks.

**Rationale**: Preserves test intent without polluting the contract test with fake-only knowledge.

## R9. Production-code separation

**Decision**: All testcontainers-go imports live exclusively inside files under `internal/repository/internaltest/` and `_test.go` files. The `internaltest` package is documented as "test-only" in `doc.go`; in addition, since Go does not enforce a "test-only package" notion natively, plan adds a guard in CI: `go build ./...` MUST not pull `testcontainers-go`. Verified via `go list -deps ./cmd/api/... | grep testcontainers` returning empty.

**Rationale**: Honors FR-014. Build/runtime never embeds testcontainers.

**Alternatives**: Build tag `//go:build test` everywhere. Rejected: noisy; the `go list -deps` invariant achieves the same guarantee with zero source-level noise.

## R10. CI image cache strategy

**Decision**: Add `testcontainers/ryuk` and `postgres:16-alpine` to the Docker layer cache via `docker pull` in a setup step before tests run. `actions/cache` keys on the digest if a finer-grained mechanism is needed. Final mechanism is task-level detail; the FR-011 3-min budget is the test.

## R11. Budget breakdown (defends SC-004)

| Phase | Cost (cached) |
|-------|---------------|
| Docker pull (cache hit) | ~1s |
| `postgres:16-alpine` container boot + ready | ~5–10s |
| Migrations against `template_ogoune` | ~1–2s |
| Per-test DB clone (`CREATE DATABASE … TEMPLATE`) × N tests | ~20ms each × ~50 tests = ~1s |
| 6 contract tests on Postgres (work in each) | ~30–60s |
| SQLite iteration of same | ~10–20s |
| Test runner overhead | ~5s |
| **Total** | **~55–100s** — well under 180s |

## R12. Risk: Ryuk + restricted CI runners

**Risk**: Some self-hosted runners disable container start of side-cars. Ryuk reaper may fail to start.

**Mitigation**: Document in `internaltest/README.md` how to disable Ryuk (`TESTCONTAINERS_RYUK_DISABLED=true`). CI runs on GitHub-hosted runners by default; assume Ryuk works.

## R13. Risk: Postgres template DB cannot be created concurrently

**Risk**: Two test packages starting their containers in parallel each try to create `template_ogoune` — that's fine, the containers are independent. But within one container, parallel template creates would race.

**Mitigation**: `sync.Once` inside the container helper ensures template creation runs once per container instance.
