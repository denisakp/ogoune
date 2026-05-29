# Quickstart — Maintainer Workflow

## One-time setup

No new env vars. No new system packages. Go toolchain is already required.

First `make sqlc-generate` invocation auto-installs `sqlc` at the pinned version (`SQLC_VERSION` in Makefile).

## Add or modify a query

1. Edit/add a `.sql` file under:
   - `internal/repository/sqlc/queries/postgres/` (for Postgres)
   - `internal/repository/sqlc/queries/sqlite/` (for SQLite)

   Use sqlc's `-- name: XxxName :exec|:one|:many` annotations.

2. Regenerate:
   ```bash
   make sqlc-generate
   ```

3. Commit the generated changes under `internal/repository/sqlc/{pg,sqlite}/` alongside your `.sql` edits.

4. Build & test:
   ```bash
   make build-be   # runs sqlc-check first; fails fast on drift
   make test-be
   ```

## Verifying single-pool guarantee

A test in `internal/database/runtime_stats_test.go` asserts `Runtime.Stats().PoolPointer` is stable and that the GORM and raw handles share the same backing pool. Run:

```bash
go test -race -run TestRuntimeSinglePool ./internal/database/...
```

## Driver changes (operator-visible)

- **Postgres**: Driver now `pgx/v5` via `pgxpool`. Same DSN format. `lib/pq`-specific quirks (none used) would not apply.
- **SQLite**: GORM dialector swapped to `glebarez/sqlite` (pure-Go, backed by `modernc.org/sqlite`). No CGO required. Single-binary cross-compile to ARM still works. Existing `SQLITE_PATH` and connection options unchanged.

Pool sizes (defaults):
- Postgres `pgxpool` MaxConns = **25**
- SQLite MaxOpenConns = **1**, MaxIdleConns = **1** (single-writer; WAL mode preserved)

## CI workflow

A `sqlc-check` step runs before the test matrix. It fails the pipeline if you edited a `.sql` file without committing the regenerated Go output.

## What this ticket does NOT do

- No repository migrated to sqlc yet (separate tickets 002+).
- No domain model change.
- No migration change.
- `database.Instance()` and all existing repository code keep working unchanged.
