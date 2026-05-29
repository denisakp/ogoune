---

description: "Task list for 041 sqlc Foundation"
---

# Tasks: sqlc Foundation

**Input**: Design documents from `/specs/041-sqlc-foundation/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: REQUIRED — change touches persistence wiring and startup flow (constitution III + IV).

**Organization**: Tasks grouped by user story so each story can be implemented, tested, and demonstrated independently.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Different file, no incomplete deps
- **[Story]**: Maps to spec user story (US1, US2, US3)
- File paths absolute or repo-relative

## Path Conventions

Single Go service + Vue SPA (untouched). Paths are repo-relative under `/Users/yaovi/Projects/perso/ogoune/`.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Add new module dependencies and scaffold sqlc layout. No business logic.

- [X] T001 Add `SQLC_VERSION := v1.27.0` variable at top of `Makefile`
- [X] T002 [P] Create empty directories `internal/repository/sqlc/queries/postgres/`, `internal/repository/sqlc/queries/sqlite/`, `internal/repository/sqlc/pg/`, `internal/repository/sqlc/sqlite/`. Add `.gitkeep` ONLY in the two generated dirs (`pg/`, `sqlite/`); the queries dirs receive `.sql` files at T008/T009 so no `.gitkeep` needed
- [X] T003 [P] Add Go module dependencies via `go get`: `github.com/jackc/pgx/v5`, `github.com/jackc/pgx/v5/pgxpool`, `github.com/jackc/pgx/v5/stdlib`, `modernc.org/sqlite`, `github.com/glebarez/sqlite`. Commit `go.mod` and `go.sum`
- [X] T004 [P] Create `sqlc.yaml` at repo root per `specs/041-sqlc-foundation/contracts/sqlc-config.md`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Land the Makefile + CI machinery and the pilot queries so the regen/check loop works end-to-end before any runtime code changes.

**⚠️ CRITICAL**: No user-story work begins until this phase is green.

- [X] T005 Add `sqlc-bin` Makefile target that runs `command -v sqlc >/dev/null 2>&1 || go install github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION)` in `Makefile`
- [X] T006 Add `sqlc-generate` Makefile target depending on `sqlc-bin`; runs `sqlc generate -f sqlc.yaml` in `Makefile`
- [X] T007 Add `sqlc-check` Makefile target depending on `sqlc-bin`; runs `sqlc generate -f sqlc.yaml` then `git diff --exit-code -- internal/repository/sqlc/pg internal/repository/sqlc/sqlite`; on failure prints "sqlc drift: run 'make sqlc-generate'" and exits 1, in `Makefile`
- [X] T008 [P] Create pilot query `internal/repository/sqlc/queries/postgres/ping.sql` with `-- name: Ping :one\nSELECT 1::int AS ok;`
- [X] T009 [P] Create pilot query `internal/repository/sqlc/queries/sqlite/ping.sql` with `-- name: Ping :one\nSELECT 1 AS ok;`
- [X] T010 Run `make sqlc-generate` locally; commit the resulting files under `internal/repository/sqlc/pg/` and `internal/repository/sqlc/sqlite/`. Remove the placeholder `.gitkeep` files from those two dirs
- [X] T011 Identify CI workflow file (`.github/workflows/*.yml` or `.gitlab-ci.yml`) responsible for backend; document path in PR description for T029
- [X] T012 Identify documentation/config surfaces to update (CLAUDE.md gotchas, `README.md` / `QUICKSTART.md`, `.env.example`). Capture list as a note in the PR description for T037

**Checkpoint**: `make sqlc-generate` and `make sqlc-check` both succeed on a clean tree. Foundation is ready for runtime changes.

---

## Phase 3: User Story 1 — Maintainer generates type-safe DB code from SQL (Priority: P1) 🎯 MVP

**Goal**: A maintainer edits an `.sql` query, runs `make sqlc-generate`, gets compilable Go output; existing tests still pass.

**Independent Test**: Add a second pilot query, run `make sqlc-generate`, observe new `.sql.go` file appears; run `make build-be && make test-be`; both succeed with zero regression.

### Tests for User Story 1 ⚠️

> Write FIRST and confirm they FAIL before code changes land.

- [X] T013 [P] [US1] Add test `internal/database/runtime_stats_test.go` containing `TestRuntimeSinglePool` that calls `Open` for SQLite then asserts `Stats().PoolPointer` is stable across two consecutive calls and that `r.SQLiteDB()` returns a non-nil handle whose address matches the pointer
- [X] T014 [P] [US1] Extend the same `internal/database/runtime_stats_test.go` with a Postgres variant (skipped unless `TEST_POSTGRES_DSN` env present) that does the same for `r.PgxPool()`
- [X] T015 [P] [US1] Add Makefile-target smoke test `internal/database/sqlc_make_test.go` that shells out `make sqlc-check` from repo root and asserts exit 0 on a clean tree (use `t.Skip` if `make` not on PATH)

### Implementation for User Story 1

- [X] T016 [US1] In `internal/database/database.go`, extend `Runtime` struct with unexported `pgxPool *pgxpool.Pool` and `sqliteDB *sql.DB` fields. Add `RuntimeStats` type matching `contracts/runtime-api.md`. Add methods `PgxPool() *pgxpool.Pool`, `SQLiteDB() *sql.DB`, `GormDB() *gorm.DB`, `Stats() RuntimeStats`
- [X] T017 [US1] In `internal/database/postgres.go`, replace direct GORM open with: `pgxpool.New(ctx, dsn)` → `stdlib.OpenDBFromPool(pool)` → `gorm.Open(postgres.New(postgres.Config{Conn: db}))`. Apply `pool.Config().MaxConns = 25`. Return the pool plus the GORM handle to the caller in `openRuntime`
- [X] T018 [US1] In `internal/database/sqlite.go`, replace `gorm.io/driver/sqlite` import with `github.com/glebarez/sqlite`. Open `*sql.DB` via `sql.Open("sqlite", dsn)` using `modernc.org/sqlite`. Apply `SetMaxOpenConns(1)`, `SetMaxIdleConns(1)`. Pass the `*sql.DB` into GORM via `glebarez.New(glebarez.Config{Conn: db})`. Preserve existing WAL/PRAGMA enforcement and `hardenSQLiteArtifacts` step
- [X] T019 [US1] In `internal/database/database.go` `openRuntime`, populate the new `Runtime.pgxPool` (Postgres) or `Runtime.sqliteDB` (SQLite) so exactly one is non-nil. Preserve the existing migration + `ValidateStartupSchema` + SQLite hardening order
- [X] T020 [US1] In `internal/database/database.go`, add deprecation godoc comment on `Instance()` (`// Deprecated: use Runtime accessors`). Do NOT change behavior or signature
- [X] T021 [US1] In `internal/platform/bootstrap/database.go`, verify the bootstrap return value exposes the full `*database.Runtime` (read the current code — if it already does, no-op + add a short godoc comment confirming the contract; if it strips fields, lift them). Repository wiring stays GORM-only — no consumer change in this PR
- [X] T022 [US1] Add `sqlc-check` as a pre-dependency of `build-be` in `Makefile` (modify the `build-be` target line)
- [X] T023 [US1] Run `make test-be`; fix any test that depended on `mattn`-specific SQLite error strings or behaviors by switching to driver-agnostic assertions in the failing files under `internal/database/`, `internal/repository/store/`, or `internal/monitoring/`. Do NOT alter test intent
- [X] T024 [US1] Run `make build-be && make test-be && make lint`; all green

**Checkpoint**: US1 fully delivered. Maintainer flow works, runtime extends with raw handles, existing tests pass.

---

## Phase 4: User Story 2 — CI blocks SQL/generated drift (Priority: P1)

**Goal**: A PR that edits a `.sql` query without regenerating fails CI on `sqlc-check`.

**Independent Test**: In a throwaway branch, edit `internal/repository/sqlc/queries/sqlite/ping.sql`, push without regenerating; CI fails on the `sqlc-check` step with the expected message. Regenerate, push again; CI passes.

### Tests for User Story 2 ⚠️

- [X] T025 [P] [US2] Add test `internal/database/sqlc_drift_test.go` that copies `queries/sqlite/ping.sql` to a tempdir, mutates it, runs `make sqlc-check` via `exec.Command`, asserts non-zero exit and `"sqlc drift"` substring in stderr. Restore original file in `t.Cleanup`. Skip if `make` or `sqlc` missing

### Implementation for User Story 2

- [X] T026 [US2] In the CI workflow file identified at T011, add a step (before the test matrix) named `sqlc check` running `make sqlc-check` on a fresh checkout. Cache `~/go/bin` so sqlc install amortizes
- [ ] T027 [US2] Confirm CI matrix (SQLite + Postgres) still passes by triggering a no-op CI run; record the green run URL in the PR description
- [X] T028 [US2] _(merged into T035 — CLAUDE.md gotcha now covered there; keep as no-op marker or delete during commit)_

**Checkpoint**: US2 fully delivered. CI guards drift; matrix unchanged.

---

## Phase 5: User Story 3 — Runtime exposes GORM + raw handles from one pool (Priority: P2)

**Goal**: Bootstrap exposes both GORM and matching raw handles, sharing a single underlying pool per dialect.

**Independent Test**: Run `TestRuntimeSinglePool` (added in US1) for both dialects: handle identity and pool-pointer stability hold; only one physical pool per dialect.

> Most of this story's implementation lands in US1 (the Runtime extension itself). US3 adds the observability/exposure layer and operator-visible affordances.

### Tests for User Story 3 ⚠️

- [X] T029 [P] [US3] Extend `internal/database/runtime_stats_test.go` with `TestStatsCountsTrack`: after opening, perform N=3 trivial queries via the raw handle, assert `Stats().InUseConns` decreases to 0 and `OpenConns` ≥ 1 once queries complete
- [X] T030 [P] [US3] Add `internal/platform/bootstrap/database_runtime_test.go` asserting bootstrap returns a Runtime whose `GormDB()` is non-nil and matches the dialect-specific raw handle's parent pool identity

### Implementation for User Story 3

- [X] T031 [US3] In `internal/database/database.go`, implement `Stats()` to populate `RuntimeStats.OpenConns/InUseConns/IdleConns` from `*sql.DB.Stats()` (SQLite) or `*pgxpool.Pool.Stat()` (Postgres). Set `PoolPointer` via `reflect.ValueOf(r.sqliteDB).Pointer()` or `reflect.ValueOf(r.pgxPool).Pointer()` per dialect (no `unsafe` import)
- [X] T032 [US3] In `internal/database/database.go`, modify `Reset()` to close the new pool handle once (`pgxPool.Close()` or `sqliteDB.Close()`) after closing the GORM-derived `*sql.DB`, guarded against double-close
- [X] T033 [US3] Verify `internal/platform/bootstrap/database.go` exposes the full Runtime in the app context (no behavior change for repositories — they still receive `*gorm.DB`). Add a TODO comment pointing to ticket 002+ for the first raw-handle consumer

**Checkpoint**: US3 fully delivered. Raw handles available for downstream tickets; single-pool guarantee verified.

---

## Phase N: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, operator-visible changes, final verification evidence (constitution V).

- [X] T034 [P] Update `README.md` and/or `QUICKSTART.md`: note the new SQLite GORM driver (`glebarez/sqlite`) and the new `make sqlc-generate` / `make sqlc-check` targets. No new env vars to document
- [X] T035 [P] Update CLAUDE.md: add `make sqlc-generate` and `make sqlc-check` to the "## Commands" block AND add a "## Gotchas" entry describing the `sqlc-check` drift gate (covers T028 too)
- [X] T036 Verify `make run-ci` still passes locally (full gate: lint + race tests + build)
- [X] T037 Walk through `specs/041-sqlc-foundation/quickstart.md` end-to-end against a fresh checkout; record evidence (screenshots of `make sqlc-generate` output, `make build-be` success, test pass) in the PR description
- [ ] T038 [P] Run SonarQube scanner per CLAUDE.md "Code Quality" section; resolve any CRITICAL/BLOCKER introduced by this change
- [ ] T039 Cross-check spec FRs/SCs ↔ tasks coverage map in PR description (FR-001…FR-015, SC-001…SC-006). Flag any unmapped requirement

### Closing the coverage gaps surfaced by /speckit-analyze

- [X] T040 Measure `make build-be` wall-clock time on a clean working tree before and after this branch (3 runs each, median). Attach the two medians + percentage delta to the PR description. Verifies **SC-004** (≤5% regression budget)
- [X] T041 Cross-compile community-mode binary `GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o /tmp/ogoune-arm64 ./cmd/api`; run `file /tmp/ogoune-arm64` and `go tool nm /tmp/ogoune-arm64 | grep -c '^.* T '` sanity check; on Linux additionally run `ldd /tmp/ogoune-amd64` after building a linux/amd64 variant and assert "not a dynamic executable" (or equivalent). Attach output to PR. Verifies **SC-005** (zero new system-library deps)
- [X] T042 Scope-guard check: run `git diff --stat main..HEAD -- internal/domain/ internal/database/migrations/ internal/repository/store/` and assert the result is empty (no lines). If non-empty, the PR has scope creep and MUST be split. Verifies **FR-015** (no domain/migration/repo change)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 Setup**: no deps — start immediately
- **Phase 2 Foundational**: needs Phase 1 — blocks all user stories
- **Phase 3 US1 (P1)**: needs Phase 2 — MVP
- **Phase 4 US2 (P1)**: needs Phase 2; can run in parallel with US1 by a second engineer once T011 is captured
- **Phase 5 US3 (P2)**: needs the Runtime work from US1 (T016–T019) to land first — start once US1 checkpoint is green
- **Phase N Polish**: needs all targeted user stories complete

### Within Each User Story

- Tests written and observed failing **before** implementation tasks
- Models → services → wiring (here: struct extension → driver opens → bootstrap wiring)
- Verification (`build-be`, `test-be`, `lint`) before checkpoint declared

### Parallel Opportunities

- T002, T003, T004 (Phase 1) — independent files, run in parallel
- T008, T009 (pilot queries) — independent files
- T013, T014, T015 (US1 tests) — distinct files
- T025 (US2 test) parallelizable with T013–T015
- T029, T030 (US3 tests) — distinct files
- T034, T035, T038 (Polish docs/scan) — independent

---

## Parallel Example: User Story 1

```bash
# Tests for US1 in parallel:
Task: "T013 add TestRuntimeSinglePool in internal/database/runtime_stats_test.go"
Task: "T014 add Postgres variant in internal/database/runtime_stats_test.go"  # same file — actually SERIAL with T013
Task: "T015 add make smoke test in internal/database/sqlc_make_test.go"
```

> Note: T013 and T014 share a file → serialize them. T015 is in a distinct file → parallel.

---

## Implementation Strategy

### MVP First (US1 only)

1. Phase 1 Setup
2. Phase 2 Foundational
3. Phase 3 US1 — STOP and VALIDATE: `make sqlc-generate`, `make build-be`, `make test-be` all green, `TestRuntimeSinglePool` passes
4. Open PR as draft MVP

### Incremental Delivery

1. MVP (US1) shipped
2. Add US2 (CI guard) → re-test → push CI workflow → green
3. Add US3 (Stats wiring + bootstrap exposure) → re-test
4. Polish phase → merge

### Parallel Team Strategy

- Eng A: Phase 1 + Phase 2 + US1 (runtime + driver swap)
- Eng B: starts on US2 (CI + drift test) after T011 is captured
- Eng A picks up US3 once US1 checkpoint passes

---

## Notes

- [P] = different files, no incomplete-task dependencies
- Single-pool guarantee (FR-010, SC-006) is hard-verified by `TestRuntimeSinglePool` — do not weaken
- Driver swaps (Postgres → pgx-backed GORM, SQLite → glebarez GORM) are the highest-risk change; budget a full `make test-be` debug cycle in T023
- No domain/migration/repository change (FR-015) — reject scope creep
- Commit after each task or logical group (Conventional Commits)
- Stop at each Checkpoint and confirm independent test passes before advancing
