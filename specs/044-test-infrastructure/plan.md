# Implementation Plan: Test Infrastructure — Dual-Dialect

**Branch**: `044-test-infrastructure` | **Date**: 2026-05-30 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/044-test-infrastructure/spec.md`

## Summary

Build a dual-dialect test infrastructure under `internal/repository/internaltest/` and flip the 6 existing `*_contract_test.go` files from fake-implementation tests into true repository-contract tests that exercise real GORM repositories on both SQLite and Postgres. Postgres is provisioned by testcontainers-go inside the test code (Clarification Q1) with one container per `go test` package, and each test gets its own isolated database cloned from a migrated template via `CREATE DATABASE … TEMPLATE template_ogoune` (Clarification Q2). New Makefile target `make test-be-pg`. New CI job `test-be-postgres` with a 3-minute wall-clock budget. SQLite-only path stays unchanged.

## Technical Context

**Language/Version**: Go 1.25.1.
**Primary Dependencies**:
- New: `github.com/testcontainers/testcontainers-go` + `github.com/testcontainers/testcontainers-go/modules/postgres` (test-only — under build tag or `_test.go`-only imports; never compiled into production binary).
- Existing: `github.com/jackc/pgx/v5/pgxpool` (041), `gorm.io/driver/postgres`, `gorm.io/gorm`, `modernc.org/sqlite` (via `glebarez/sqlite`).
- Existing: `internal/database` package (041's `Runtime`, `Open`, migration application).
**Storage**: SQLite (per-test temp file) + Postgres 16 (containerized, per-package; per-test DB clone). On-disk format unchanged.
**Testing**: New tests under `internal/repository/internaltest/` validate the helper itself. The 6 refactored contract tests under `internal/repository/store/` exercise real repositories. Existing SQLite-only tests untouched.
**Target Platform**: Linux/Darwin CI runners + developer workstations with Docker.
**Project Type**: Single Go service + Vue SPA (frontend untouched).
**Performance Goals**:
- CI `test-be-postgres` job ≤ **180s** wall-clock (SC-004).
- Per-test Postgres DB allocation: < 50ms (template-clone is fast on Postgres).
- SQLite path no regression (SC-006).
**Constraints**:
- No `testcontainers` / Docker import in production code (FR-014).
- The 6 contract tests MUST NOT lose any assertion present today (FR-009).
- Postgres iteration is opt-in locally (via `make test-be-pg`); mandatory in CI.
- Production code stays untouched (no `internal/api/`, `internal/service/`, `internal/repository/store/` *non-test* edits).
**Scale/Scope**:
- ~400 LOC for the `internaltest` package (helpers + cleanup).
- ~300 LOC of net change across the 6 contract test files (mostly: replace `fake.NewXxxFake()` wiring with factory parameter, add `ForEachDialect`).
- ~150 LOC of helper-self-tests.
- 1 new Makefile target.
- 1 new CI job (`.github/workflows/test.yml` + GitLab mirror).

## Constitution Check

| Principle | Verdict | Notes |
|-----------|---------|-------|
| I. Layered Boundary Integrity | PASS | New code is test-only under `internal/repository/internaltest/`. Production layering untouched. |
| II. Community Simplicity, Hosted Continuity | PASS | Community single-binary unchanged (no testcontainers in build). Hosted/Postgres path gains real test coverage — strengthens it, never regresses. |
| III. Automated Verification for Runtime Changes | PASS — by design this IS the verification investment. Adds dual-dialect coverage on real repositories where today only fakes were tested. |
| IV. Migration and Startup Safety | PASS | Migrations applied via the existing `internal/database/migrations` path — same code, same migration tree. Helper just re-uses it on a fresh DB. |
| V. Spec-to-Execution Traceability | PASS | Spec → clarify → plan → tasks chain in place; refactor of existing tests is explicit, no silent rewrite. |

No violations. Complexity Tracking omitted.

## Project Structure

### Documentation (this feature)

```text
specs/044-test-infrastructure/
├── plan.md
├── spec.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── helper-api.md
│   └── contract-test-shape.md
└── checklists/requirements.md
```

### Source Code (repository root)

```text
ogoune/
├── internal/repository/internaltest/                # NEW package — test-only
│   ├── doc.go                                       # package-level docs + build constraint
│   ├── fixture.go                                   # DialectFixture struct + GORM/raw handle access
│   ├── sqlite.go                                    # SetupSQLite — temp-file per test, migrations applied
│   ├── postgres.go                                  # SetupPostgres — pgxpool over the per-test DB
│   ├── container.go                                 # testcontainers-go launcher: per-package container, template DB
│   ├── for_each.go                                  # ForEachDialect runner
│   ├── isolation_test.go                            # NEW — proves SC-007 (no data leak between parallel tests)
│   ├── helper_test.go                               # NEW — exercises SetupSQLite, SetupPostgres, cleanup
│   └── README.md                                    # NEW — how to use the helper, opt-in locally
├── internal/repository/store/
│   ├── tags_repository_contract_test.go             # MODIFIED — factory + ForEachDialect; tests real GORM
│   ├── api_key_repository_contract_test.go          # MODIFIED — same
│   ├── incident_repository_contract_test.go         # MODIFIED — same
│   ├── incident_event_step_repository_contract_test.go  # MODIFIED — same
│   ├── monitoring_activity_repository_contract_test.go  # MODIFIED — same
│   ├── resource_repository_contract_test.go         # MODIFIED — same
│   └── *_fake_test.go                               # NEW (per file as needed) — moved fake-specific assertions
├── Makefile                                          # MODIFIED — new target `test-be-pg`
├── .github/workflows/test.yml                       # MODIFIED — add `test-be-postgres` job
├── .github/workflows/ci.yml                         # MODIFIED — same on main branch
└── .gitlab-ci.yml                                   # MODIFIED — mirror the Postgres job
```

**Structure Decision**: All new code is test-only and confined to `internal/repository/internaltest/` plus modifications inside `*_contract_test.go` files. No new top-level dirs. No production import path touches testcontainers.

## Phase 0 — Research

See [`research.md`](./research.md). Resolved:

1. **Testcontainers-go module choice** (Clarification Q1): use `github.com/testcontainers/testcontainers-go/modules/postgres` — first-class Postgres module with built-in healthcheck + DSN exposure.
2. **Per-package container lifecycle** (Clarification Q2): each `*_test.go` package that needs Postgres calls a `internaltest.PostgresContainer(t *testing.T)` (or `internaltest.MainPostgres(m *testing.M)`) helper. The helper memoizes the container per process and registers `m.Cleanup` for the package.
3. **Template-DB pattern**: container start → connect as admin → `CREATE DATABASE template_ogoune` → apply migrations on `template_ogoune` → mark it template. Per-test: `CREATE DATABASE ogoune_test_<ULID> TEMPLATE template_ogoune`. `t.Cleanup` issues `DROP DATABASE`.
4. **Image caching**: pin `postgres:16-alpine` so CI's Docker layer cache hits across runs; document in `internaltest/README.md`.
5. **DSN injection contract**: spec mandates `POSTGRES_TEST_DSN` as the opt-in env var. The plan honors it: when set, the helper uses that DSN directly and skips the container (allows devs to point at a shared local PG). When unset, the helper boots a container per package. Both code paths end up at the same `DialectFixture` API.
6. **The 6 contract tests today test fakes** — the refactor flips them to real GORM, parameterized by a factory. Fake-specific assertions that don't translate (e.g. "fake increments an internal counter") move to a sibling `*_fake_test.go`.
7. **Production code MUST NOT import testcontainers**: enforced by package layout (testcontainers only in `_test.go` files + the `internaltest` package which is itself test-only by convention — guarded via a `doc.go` build constraint `//go:build test` if needed, or by virtue of being only imported from `_test.go` files).
8. **CI Postgres image cache strategy**: use `actions/cache` keyed on the image digest, OR rely on GitHub Actions' built-in Docker layer cache. Plan defers the exact mechanism to the tasks layer; budget enforces correctness.
9. **3-minute budget breakdown**: container startup ~10s (cached) + migrations ~2s + 6 contract tests × ~10s each ≈ 75s + SQLite path ≈ 30s + overhead → comfortably under 180s.

## Phase 1 — Design & Contracts

Artifacts:

- [`data-model.md`](./data-model.md) — `DialectFixture`, `PostgresContainer`, `Factory` types
- [`contracts/helper-api.md`](./contracts/helper-api.md) — `SetupSQLite`, `SetupPostgres`, `ForEachDialect` API
- [`contracts/contract-test-shape.md`](./contracts/contract-test-shape.md) — refactored contract test template
- [`quickstart.md`](./quickstart.md) — maintainer workflow (add a dual-dialect test in <10 lines)

Post-design Constitution re-check: 5/5 PASS.

## Complexity Tracking

No violations.
