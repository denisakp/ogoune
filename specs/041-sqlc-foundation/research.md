# Phase 0 — Research

## R1. Postgres single-pool wiring with pgx/v5 + GORM

**Decision**: Open `*pgxpool.Pool` first via `pgxpool.New(ctx, dsn)`. Derive a `*sql.DB` from it using `github.com/jackc/pgx/v5/stdlib.OpenDBFromPool(pool)`. Hand the resulting `*sql.DB` to GORM via `postgres.New(postgres.Config{Conn: db})`. The `Runtime` then holds both `*pgxpool.Pool` (for future sqlc consumers) and `*gorm.DB` (for legacy consumers); they share the same physical pool.

**Rationale**:
- One pool, one set of connection-management settings.
- `pgxpool` is pgx's first-class API; sqlc-generated pg code targets `pgxpool.Pool` (or any `DBTX` interface) natively.
- `stdlib.OpenDBFromPool` is the documented pgx mechanism to expose a pool as `database/sql`.
- GORM's `postgres.Config{Conn: ...}` accepts any `*sql.DB`, bypassing the default DSN-based open.

**Alternatives considered**:
- `gorm.io/driver/postgres` opens its own pool from DSN. Rejected: yields two pools, violates FR-010.
- Open `*sql.DB` first via `sql.Open("pgx", dsn)` from `stdlib`, ignore `pgxpool`. Rejected: loses pgxpool ergonomics, and sqlc-generated `pgx/v5` code expects `*pgxpool.Pool` for typed paths.

## R2. SQLite single-pool wiring with modernc.org/sqlite + GORM

**Decision**: `sql.Open("sqlite", dsn)` using `modernc.org/sqlite` (registers itself with `database/sql` on import). Pass the resulting `*sql.DB` to GORM via `glebarez/sqlite` dialector: `gorm.Open(glebarez.New(glebarez.Config{Conn: db}))`. `Runtime` holds the single `*sql.DB` (also accessible as `SQLiteDB()`) plus the GORM wrapper.

**Rationale**:
- `modernc.org/sqlite` is pure Go → no CGO → community single-binary + ARM cross-compile guarantee (SC-005, Principle II).
- `glebarez/sqlite` is the established pure-Go GORM dialector backed by `modernc.org/sqlite`. Active maintenance, used in production by several Go projects.
- Sharing a `*sql.DB` between GORM and raw users is safe and intended.

**Alternatives considered**:
- Stay on `mattn/go-sqlite3` + `gorm.io/driver/sqlite` (CGO). Rejected: breaks "single binary" promise.
- Hand-roll a GORM dialector around modernc. Rejected: redundant; `glebarez/sqlite` does this.
- Open two separate `*sql.DB` (one for GORM, one for sqlc). Rejected: two pools, violates FR-010.

## R3. Pool sizing defaults

**Decision** (locked via Clarifications Q1):
- Postgres `pgxpool` `MaxConns = 25`.
- SQLite `*sql.DB`: `SetMaxOpenConns(1)`, `SetMaxIdleConns(1)`, `SetConnMaxLifetime(0)` (no expiry).

**Rationale**:
- 25 matches current effective concurrency observed via existing GORM default; safe upper bound under monitor worker load.
- SQLite single writer prevents `SQLITE_BUSY` under concurrent writes; WAL mode (already enforced) makes reads concurrent at the OS layer.

**Alternatives considered**: see spec Clarifications Q1 (options A/C/D rejected).

## R4. sqlc CLI install & version pin

**Decision** (locked via Clarifications Q2): Pinned version constant in Makefile:
```make
SQLC_VERSION := v1.27.0
sqlc-bin:
	@command -v sqlc >/dev/null 2>&1 || go install github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION)
```
Both `sqlc-generate` and `sqlc-check` depend on `sqlc-bin`. CI installs once per job via the same target (no extra step needed beyond cache for `GOPATH/bin`).

**Rationale**: Zero extra tooling beyond Go toolchain (already required). Pin guarantees deterministic codegen.

**Alternatives**: Docker, vendored binary, manual — rejected in clarification.

## R5. `sqlc-check` mechanism

**Decision**: Generate to a temp directory, diff against committed tree, fail with non-zero exit on any difference. Shell sketch:
```make
sqlc-check: sqlc-bin
	@TMP=$$(mktemp -d) && \
	  sqlc -f sqlc.yaml generate --output-dir "$$TMP" 2>/dev/null || true && \
	  diff -r internal/repository/sqlc/pg     "$$TMP/pg"     && \
	  diff -r internal/repository/sqlc/sqlite "$$TMP/sqlite" \
	    || { echo "sqlc drift: run 'make sqlc-generate'"; exit 1; }
```
*(Exact form refined in tasks — sqlc may not support `--output-dir` override; fallback: generate into checkout, run `git diff --exit-code internal/repository/sqlc/`.)*

**Rationale**: Idempotent regen + diff is the canonical sqlc CI guard. Git-diff fallback works on any sqlc version.

**Decision refinement**: Prefer `git diff --exit-code` form for simplicity and version-agnosticism.

## R6. `build-be` coupling

**Decision** (locked via Clarifications Q3): `make build-be` runs `sqlc-check` first. `sqlc-generate` stays explicit. CI runs both.

**Rationale**: Local builds fail loud on drift, mirroring CI, but never silently rewrite files mid-build.

## R7. `Stats()` accessor shape

**Decision**: 
```go
type RuntimeStats struct {
    Driver      Driver
    OpenConns   int
    InUseConns  int
    IdleConns   int
    PoolPointer uintptr  // for identity assertion in tests
}
func (r *Runtime) Stats() RuntimeStats
```
For Postgres, `PoolPointer` = address of the `*pgxpool.Pool`. For SQLite, address of the `*sql.DB`. Test asserts equal pointers when accessed twice via different handles.

**Rationale**: Programmatic identity proof, no log parsing (FR-010a, SC-006). Counts also useful for future observability.

**Alternatives**: log line + grep (rejected in clarification — flaky), expose pool directly to tests (rejected — leaks internals).

## R8. Legacy `Instance()` preservation

**Decision**: `Instance()` keeps its current signature `func Instance() (*gorm.DB, error)` and continues to return `activeRuntime.DB`. Marked `// Deprecated:` in godoc only — actual removal owned by later tickets in the sqlc track. No call-site touch in this PR.

**Rationale**: Zero churn for existing repos; this ticket is strictly additive on top of the existing Runtime.

## R9. SQLite GORM driver swap safety

**Risk**: Swapping `gorm.io/driver/sqlite` (CGO/mattn) → `glebarez/sqlite` (modernc) may change SQL dialect quirks (e.g., `RETURNING` support, datetime serialization, error messages used in tests).

**Mitigation**:
- Run full existing `make test-be` matrix on a draft PR; fix any test relying on mattn-specific error strings or behaviors.
- Document the swap in `quickstart.md` and CLAUDE.md gotchas (later doc ticket).
- Both drivers target SQLite 3.x; SQL surface area used by ogoune is conservative (no FTS, no JSON1 funcs beyond what modernc supports).
- modernc passes the SQLite test suite — semantic equivalence is strong.

## R10. CI matrix integration

**Decision**: Add a single `sqlc-check` step to the existing backend CI workflow (runs once, dialect-agnostic — it checks the queries → generated diff, not runtime). Keep existing SQLite + Postgres test matrix unchanged. No new matrix axis needed.

**Rationale**: `sqlc generate` reads `.sql` files; result independent of which DB tests against later.
