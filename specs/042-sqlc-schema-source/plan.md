# Implementation Plan: sqlc Schema Source

**Branch**: `042-sqlc-schema-source` | **Date**: 2026-05-29 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/042-sqlc-schema-source/spec.md`

## Summary

Migrations under `internal/database/migrations/{postgres,sqlite}/` already parse cleanly through sqlc (verified at plan time: 20 tables generated per dialect with zero parser errors). SC-001 is met by the 041 foundation. This ticket therefore focuses on:

1. **Document** the type-mapping table and "Non-DDL statements inventory" inside `internal/database/migrations/README.md`.
2. **Build** the Go-based `cmd/migrations-drift-check` linter (file-pair + per-column name + nullability) wired into Makefile + CI.
3. **Update** CLAUDE.md "Database migrations" section with conventions (file-pair rule, `PRAGMA`/triggers caveat, JSON mapping pointer).

No migration file is rewritten. No domain / repository / runtime code changes. No aggregated-schema fallback (`internal/database/schema/`) — current migrations all parse, so the fallback stays opt-in for the future.

## Technical Context

**Language/Version**: Go 1.25.1 (project default)
**Primary Dependencies**: stdlib only for the new linter (`os`, `path/filepath`, `regexp`, `bufio`, `strings`). No new module deps. Existing sqlc CLI (v1.27.0, pinned in 041's Makefile) is the schema consumer being validated.
**Storage**: N/A — this ticket touches no schema or runtime data.
**Testing**: `go test -race ./cmd/migrations-drift-check/...` plus a golden-fixtures test asserting (a) pair-missing fails, (b) nullability divergence fails, (c) aligned trees pass.
**Target Platform**: Linux/Darwin CI runners + developer workstations. Cross-compile not required for the linter.
**Project Type**: Single Go service + Vue SPA (frontend untouched).
**Performance Goals**: Linter MUST complete in <1s on the current 14-file tree, <5s at 100 migrations (scan is linear).
**Constraints**:
- No new module dependency for the linter (stdlib only) — reduces upkeep + supply-chain surface.
- Linter checks **name + nullability only**; type pairs are reviewer/doc responsibility (Clarification Q2).
- Audit artifact lives in `internal/database/migrations/README.md` (Clarification Q3).
- No migration file is edited (PRD hors-périmètre).
**Scale/Scope**:
- 14 file-pairs today (28 files), 20 tables. Will grow over time.
- ~250 LOC for the linter + tests.
- ~150 lines of new doc content (README + CLAUDE.md update).

## Constitution Check

| Principle | Verdict | Notes |
|-----------|---------|-------|
| I. Layered Boundary Integrity | PASS | New code is a CLI tool under `cmd/`. No service/repository/handler touched. No frontend touched. |
| II. Community Simplicity, Hosted Continuity | PASS | Pure additive: linter + docs. Both modes unchanged. No new env var. No new runtime dependency. |
| III. Automated Verification for Runtime Changes | PASS — by design no runtime change. The linter itself is unit-tested (golden fixtures + happy path on real trees). |
| IV. Migration and Startup Safety | PASS | Migrations are not touched. Startup logic untouched. The linter runs at CI time only, not at app startup. |
| V. Spec-to-Execution Traceability | PASS | Spec → clarify → plan → tasks → impl chain in place. CLAUDE.md + README documentation tasks tracked. |

No violations. Complexity Tracking omitted.

## Project Structure

### Documentation (this feature)

```text
specs/042-sqlc-schema-source/
├── plan.md                          # This file
├── spec.md
├── research.md                      # Phase 0
├── data-model.md                    # Phase 1 — minimal (linter entities)
├── quickstart.md                    # Phase 1 — maintainer workflow
├── contracts/
│   ├── drift-check-cli.md           # Linter CLI + exit codes
│   └── migrations-readme.md         # README structure contract
└── checklists/
    └── requirements.md              # From /speckit-specify
```

### Source Code (repository root)

```text
ogoune/
├── Makefile                         # MODIFIED: add migrations-drift-check target
├── cmd/migrations-drift-check/      # NEW: Go linter
│   ├── main.go                      # entry point, exit codes
│   ├── pair.go                      # file-pair check (NNNN_*.sql in both trees)
│   ├── column.go                    # CREATE TABLE / ALTER TABLE ADD COLUMN parsing + name/nullability cross-check
│   ├── pair_test.go                 # golden fixtures
│   ├── column_test.go               # golden fixtures
│   └── testdata/                    # tiny synthetic migration trees for tests
│       ├── ok/{postgres,sqlite}/
│       ├── missing_pair/{postgres,sqlite}/
│       └── null_drift/{postgres,sqlite}/
├── internal/database/migrations/
│   ├── README.md                    # MODIFIED: add type-mapping table + Non-DDL inventory
│   ├── postgres/                    # UNCHANGED
│   └── sqlite/                      # UNCHANGED
├── CLAUDE.md                        # MODIFIED: expand "## Patterns to follow → Database migrations"
└── .github/workflows/{ci,test}.yml  # MODIFIED: add `make migrations-drift-check` step
└── .gitlab-ci.yml                   # MODIFIED: same step
```

**Structure Decision**: New code is confined to a single new package `cmd/migrations-drift-check`. Stdlib only. Tests collocated with golden-fixture trees under `testdata/`. Docs land in two existing files (README + CLAUDE.md). No new top-level directories.

## Phase 0 — Research

See [`research.md`](./research.md). Resolved items:

1. **sqlc actually parses the current migrations** — pre-plan verification: `sqlc generate -f sqlc.yaml` returns exit 0 and produces 20 typed models per dialect. SC-001 is met without any audit-driven changes.
2. **Linter parser strategy** — regex + line scan for `CREATE TABLE name (...)` and `ALTER TABLE name ADD COLUMN col TYPE [NULL|NOT NULL]`. No full SQL parser dependency.
3. **Audit format** — type-mapping table + "Non-DDL statements inventory" inside `internal/database/migrations/README.md`.
4. **Aggregated-schema fallback** — not introduced. Migrations parse; reactive trigger remains documented for the future.

## Phase 1 — Design & Contracts

Artifacts produced in this phase:

- [`data-model.md`](./data-model.md) — Linter internal entities (MigrationPair, ColumnDef)
- [`contracts/drift-check-cli.md`](./contracts/drift-check-cli.md) — CLI contract: args, exit codes, output format
- [`contracts/migrations-readme.md`](./contracts/migrations-readme.md) — README structure contract
- [`quickstart.md`](./quickstart.md) — Maintainer workflow

Post-design Constitution re-check: all five principles still PASS. No new layering risks; pure-additive change.

## Complexity Tracking

No violations to justify.
