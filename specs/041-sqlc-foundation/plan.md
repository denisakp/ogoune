# Implementation Plan: sqlc Foundation

**Branch**: `041-sqlc-foundation` | **Date**: 2026-05-29 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/041-sqlc-foundation/spec.md`

## Summary

Land the sqlc infrastructure without migrating any business repository. Add `sqlc.yaml`, dual dialect query trees with pilot queries, generated-code folders (committed), Makefile targets (`sqlc-generate`, `sqlc-check`), and CI drift-check. Extend the existing `database.Runtime` struct to expose raw handles (`*pgxpool.Pool` for Postgres, `*sql.DB` for SQLite via `modernc.org/sqlite`) backed by a **single** physical pool per dialect, while preserving the GORM handle and legacy `Instance()` accessor unchanged. No domain, no migration, no repo change.

## Technical Context

**Language/Version**: Go 1.23 (existing project version) / Node 20+ for unrelated frontend (untouched here)
**Primary Dependencies**:
- `github.com/sqlc-dev/sqlc` CLI v1.27.0 (pinned in Makefile variable)
- `github.com/jackc/pgx/v5` + `github.com/jackc/pgx/v5/pgxpool` (new — Postgres)
- `github.com/jackc/pgx/v5/stdlib` (new — bridges pgxpool to `*sql.DB` for GORM reuse)
- `modernc.org/sqlite` (new — pure-Go SQLite driver registered through `database/sql`)
- `github.com/glebarez/sqlite` GORM driver (new — pure-Go SQLite-compatible GORM dialect; replaces `gorm.io/driver/sqlite` which is CGO `mattn`)
- Existing: `gorm.io/gorm`, `gorm.io/driver/postgres`

**Storage**: Postgres (production) + SQLite (community); same as today
**Testing**: `go test -race ./...`; SQLite tests use existing `setupTestDB` helper; new tests cover Runtime pool identity (`Stats()` accessor) and Makefile target sanity
**Target Platform**: Linux/Darwin server + Linux ARM (Raspberry Pi self-hosters) for community single-binary; no CGO permitted on SQLite path
**Project Type**: Single Go service + Vue SPA (web/ untouched in this ticket)
**Performance Goals**:
- `build-be` time regression budget ≤ 5% (SC-004)
- SQLite write throughput accepted regression ~20–40% vs. CGO `mattn` (PRD-decided)
- No runtime hot-path regression on Postgres (only driver swap underneath GORM)
**Constraints**:
- Single static binary; no CGO on SQLite path (SC-005)
- One physical pool per dialect (FR-010, SC-006), proven via `Stats()` identity test (FR-010a)
- Legacy `Instance()` accessor preserved as deprecated wrapper (FR-011)
- No domain/migration/repo changes (FR-015)
**Scale/Scope**:
- Foundation only: ~10 new files (`sqlc.yaml`, 2 pilot SQL, 2 generated trees, Makefile edits, Runtime extension, CI step)
- ~300 LOC net Go change (excluding generated code)

## Constitution Check

| Principle | Verdict | Notes |
|-----------|---------|-------|
| I. Layered Boundary Integrity | PASS | Repository layer untouched. Runtime exposes handles only — no service/handler bypass. New raw handles will be consumed by future sqlc-backed repos (later tickets), not by handlers. |
| II. Community Simplicity, Hosted Continuity | PASS | SQLite path migrates GORM driver to pure-Go `glebarez/sqlite` → preserves "single binary, zero deps, ARM cross-compile" promise. Postgres path migrates GORM to `pgx`-backed `*sql.DB` via `stdlib.OpenDBFromPool` → hosted path keeps same SQL surface, same migrations, same repository code. No silent regression. Driver swap documented in research.md and quickstart.md. |
| III. Automated Verification for Runtime Changes | PASS | New tests: (1) `Runtime.Stats()` returns one pool per dialect (identity); (2) bootstrap smoke test boots both modes; (3) Makefile `sqlc-check` exit code test. Existing GORM-backed tests run unchanged on both dialects (matrix CI). |
| IV. Migration and Startup Safety | PASS | Migrations unchanged (FR-015). Startup behavior preserved: `Open` still runs `runMigrations` + `ValidateStartupSchema` + SQLite hardening before returning. Driver swap does not affect migration application order. Fail-fast on pool open errors retained. |
| V. Spec-to-Execution Traceability | PASS | Spec → plan → tasks (next) chain in place. Operator-visible changes documented in `quickstart.md` (new SQLite GORM driver name, no env var changes). |

No violations. Complexity Tracking section omitted.

## Project Structure

### Documentation (this feature)

```text
specs/041-sqlc-foundation/
├── plan.md                          # This file
├── spec.md                          # Feature specification
├── research.md                      # Phase 0 — driver/integration research
├── data-model.md                    # Phase 1 — Runtime contract
├── quickstart.md                    # Phase 1 — maintainer workflow
├── contracts/
│   ├── runtime-api.md               # Runtime struct contract
│   ├── makefile-targets.md          # Make target contract
│   └── sqlc-config.md               # sqlc.yaml schema contract
└── checklists/
    └── requirements.md              # From /speckit-specify
```

### Source Code (repository root)

```text
ogoune/
├── Makefile                          # MODIFIED: add SQLC_VERSION, sqlc-generate, sqlc-check; wire sqlc-check into build-be
├── sqlc.yaml                         # NEW: multi-dialect sqlc config
├── internal/
│   ├── database/
│   │   ├── database.go               # MODIFIED: extend Runtime with PgxPool, SQLiteDB, Stats(); single-pool wiring
│   │   ├── postgres.go               # MODIFIED: open pgxpool first, give GORM *sql.DB via stdlib.OpenDBFromPool
│   │   ├── sqlite.go                 # MODIFIED: switch GORM driver to glebarez/sqlite (modernc-based); open *sql.DB once
│   │   └── runtime_stats_test.go     # NEW: pool-identity test (SC-006/FR-010a)
│   ├── platform/bootstrap/
│   │   └── database.go               # MODIFIED: wire new Runtime; no repo changes
│   └── repository/sqlc/
│       ├── queries/
│       │   ├── postgres/
│       │   │   └── ping.sql          # NEW: pilot query
│       │   └── sqlite/
│       │       └── ping.sql          # NEW: pilot query
│       ├── pg/                       # NEW: generated, committed
│       │   ├── db.go
│       │   ├── models.go
│       │   └── ping.sql.go
│       └── sqlite/                   # NEW: generated, committed
│           ├── db.go
│           ├── models.go
│           └── ping.sql.go
├── .github/workflows/                # MODIFIED: add sqlc-check job to CI matrix
└── go.mod / go.sum                   # MODIFIED: add pgx/v5, pgxpool, stdlib, modernc.org/sqlite, glebarez/sqlite
```

**Structure Decision**: Reuse existing `database.Runtime` struct (already present at `internal/database/database.go:44`). Extend rather than replace. Place sqlc artifacts under `internal/repository/sqlc/{queries,pg,sqlite}/` as PRD dictates. No new top-level packages.

## Phase 0 — Research

See [`research.md`](./research.md). Resolved items:

1. Postgres single-pool strategy: `pgxpool.New()` first → `stdlib.OpenDBFromPool()` → hand `*sql.DB` to `gorm.io/driver/postgres` via `Conn:` option. Single physical pool, both APIs share it.
2. SQLite single-pool strategy: `sql.Open("sqlite", dsn)` (modernc driver) → keep `*sql.DB` → hand same handle to `glebarez/sqlite` GORM driver via `gorm.Open(glebarez.Dialector{Conn: db})`. Single handle, both APIs share it.
3. GORM SQLite driver swap rationale: stock `gorm.io/driver/sqlite` requires CGO. `glebarez/sqlite` is a drop-in pure-Go alternative using `modernc.org/sqlite` underneath. Risk: separate maintenance; mitigated by community adoption + test matrix.
4. sqlc CLI install: `go install github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION)` from Makefile when binary missing. Pinned to v1.27.0.
5. `sqlc-check` mechanism: regenerate to a tempdir, `diff -r` against committed tree, non-zero exit on mismatch.
6. Pool sizing (from Clarifications): Postgres `pgxpool` MaxConns=25; SQLite `SetMaxOpenConns(1)` + `SetMaxIdleConns(1)`.

## Phase 1 — Design & Contracts

Artifacts produced in this phase:

- [`data-model.md`](./data-model.md) — Runtime entity and lifecycle
- [`contracts/runtime-api.md`](./contracts/runtime-api.md) — Go API surface for `Runtime` post-change
- [`contracts/makefile-targets.md`](./contracts/makefile-targets.md) — Make target behavior contract
- [`contracts/sqlc-config.md`](./contracts/sqlc-config.md) — `sqlc.yaml` schema contract
- [`quickstart.md`](./quickstart.md) — Maintainer workflow (add a query, regenerate, verify)

Post-design Constitution re-check: all five principles still PASS — design does not introduce new layering violations, preserves both modes, includes test strategy, retains fail-fast startup, and ships traceable artifacts.

## Complexity Tracking

No violations to justify.
