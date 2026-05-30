---

description: "Task list for 044 Test Infrastructure — Dual-Dialect"
---

# Tasks: Test Infrastructure — Dual-Dialect

**Input**: Design documents from `/specs/044-test-infrastructure/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: REQUIRED — this ticket IS the test infrastructure investment (Constitution III).

**Organization**: Tasks grouped by user story so each can be implemented and shipped independently.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Different file, no incomplete deps
- **[Story]**: Maps to spec user story (US1, US2, US3)
- Paths repo-relative under `/Users/yaovi/Projects/perso/ogoune/`

## Path Conventions

Single Go service + Vue SPA (frontend untouched). New test-only code under `internal/repository/internaltest/`. Refactor edits inside existing `internal/repository/store/*_contract_test.go` files.

---

## Phase 1: Setup

**Purpose**: Add module dependencies and scaffold the empty package.

- [X] T001 Add Go module dependencies via `go get`: `github.com/testcontainers/testcontainers-go` and `github.com/testcontainers/testcontainers-go/modules/postgres`. Commit `go.mod` and `go.sum`
- [X] T002 Create directory `internal/repository/internaltest/` with a placeholder `doc.go` that declares `package internaltest` and a 3-line godoc comment: "Test-only helpers for dual-dialect repository tests. MUST NOT be imported from production code (cmd/, internal/api/, internal/service/, non-test store/). Production import is asserted by CI (see polish phase)."

---

## Phase 2: Foundational

**Purpose**: Land the type skeleton both US1 and US2 depend on.

**⚠️ CRITICAL**: US1 (helper) and US2 (contract refactor) both depend on `DialectFixture`.

- [X] T003 Create `internal/repository/internaltest/fixture.go` defining the `DialectFixture` struct per `data-model.md` (`Dialect string`, `Runtime *database.Runtime`, `DSN string`). No methods yet — just the type. Add a one-paragraph godoc

---

## Phase 3: User Story 1 — Dual-dialect helper exists and is testable (Priority: P1) 🎯 MVP

**Goal**: A maintainer can call `internaltest.SetupSQLite(t)` / `internaltest.SetupPostgres(t)` / `internaltest.ForEachDialect(t, fn)` from any `_test.go` file and get a working, isolated, migrated database per test on each dialect.

**Independent Test**: New tests under `internal/repository/internaltest/`: `TestSetupSQLite_AppliesMigrationsAndIsolates` and (Docker-gated) `TestSetupPostgres_TemplateCloneIsolates`, plus `TestForEachDialect_RunsBothAndIsolates`. All pass.

### Tests for User Story 1 ⚠️

> Write FIRST and observe FAIL.

- [X] T004 [P] [US1] Add `internal/repository/internaltest/sqlite_test.go` with `TestSetupSQLite_AppliesMigrationsAndIsolates`: calls `SetupSQLite(t)` twice in two parallel sub-tests, writes a distinct `Tag` row in each, asserts the other sub-test's row is NOT visible (per-test isolation)
- [X] T005 [P] [US1] Add `internal/repository/internaltest/postgres_test.go` with `TestSetupPostgres_TemplateCloneIsolates`: same shape as SQLite but on Postgres. Skip-aware (uses `SetupPostgres`; if it skips, test skips too)
- [X] T006 [P] [US1] Add `internal/repository/internaltest/for_each_test.go` with `TestForEachDialect_RunsBothAndIsolates`: counts how many sub-tests fired (sqlite + postgres-or-skip); asserts SQLite ran exactly once even if Postgres skipped (FR-004 acceptance)
- [X] T007 [P] [US1] Add `internal/repository/internaltest/isolation_test.go` with `TestIsolation_ParallelSamePrimaryKey`: spawns two parallel sub-tests via `ForEachDialect` that each insert a `Tag` with the SAME ID; asserts neither sees the other's row (defends SC-007). Skip-aware for Postgres

### Implementation for User Story 1

- [X] T008 [US1] Create `internal/repository/internaltest/sqlite.go` implementing `SetupSQLite(t *testing.T) *DialectFixture`: build a `database.Config{Driver: DriverSQLite, SQLitePath: t.TempDir()+"/test.db", LogLevel: "silent"}`, call `database.Open(ctx, cfg)`, wrap in `DialectFixture{Dialect: "sqlite", Runtime: rt}`, register `t.Cleanup` that closes the underlying `*sql.DB`
- [X] T009 [US1] Create `internal/repository/internaltest/container.go` implementing the `PgContainer` type per `data-model.md`: `sync.Once`-guarded container startup using `testcontainers-go/modules/postgres` with `WaitForListeningPort(5432/tcp)` (more reliable than `WaitForLog`, which Postgres emits twice during startup). Connect as superuser, `CREATE DATABASE template_ogoune`, then **apply migrations by calling `database.Open(ctx, Config{Driver: DriverPostgres, DatabaseURL: <template DSN>, LogLevel: "silent"})`** — `Open` already runs `runMigrations` + `ValidateStartupSchema`; close the returned Runtime once the template is migrated. Mark `template_ogoune` as a template (`ALTER DATABASE template_ogoune IS_TEMPLATE = true`). Expose `Acquire(t testing.TB) (dsn string)` that does `CREATE DATABASE ogoune_test_<ULID> TEMPLATE template_ogoune` and registers `t.Cleanup` to `DROP DATABASE … WITH (FORCE)`. Per-test cloned DBs do NOT re-run migrations — `database.Open` on the clone sees the migration table already populated and exits the migration loop as a no-op
- [X] T010 [US1] In the same `container.go`, add `POSTGRES_TEST_DSN` short-circuit path: when set, skip the container, treat the env var as the superuser DSN, perform the same template + per-test DB clone work against the external Postgres
- [X] T011 [US1] Create `internal/repository/internaltest/postgres.go` implementing `SetupPostgres(t *testing.T) *DialectFixture`: detect Docker availability OR `POSTGRES_TEST_DSN`; if neither, call `t.Skip("postgres backend unavailable (no Docker, no POSTGRES_TEST_DSN)")` and return nil; otherwise call `PgContainer.Acquire(t)`, open via `database.Open` with the per-test DSN, wrap in `DialectFixture{Dialect: "postgres", Runtime: rt, DSN: dsn}`
- [X] T012 [US1] Create `internal/repository/internaltest/for_each.go` implementing `ForEachDialect(t, fn)` per `contracts/helper-api.md`: SQLite sub-test always runs; Postgres sub-test runs unless `SetupPostgres` already called `t.Skip`. Sub-test names locked to `"sqlite"` and `"postgres"`
- [X] T013 [US1] Implement `DialectsAvailable() []string` in `for_each.go` — returns the dialects that will execute in this environment (sqlite always; postgres iff docker or POSTGRES_TEST_DSN)
- [X] T014 [US1] Run `go test -race ./internal/repository/internaltest/...` — observe T004/T006 PASS without Docker; with Docker running observe T005/T007 also PASS

**Checkpoint**: US1 delivered. Helper works, isolated, tested.

---

## Phase 4: User Story 2 — 6 existing contract tests run against real GORM on both dialects (Priority: P1)

**Goal**: Each `internal/repository/store/*_contract_test.go` is refactored per `contracts/contract-test-shape.md`: a top-level `TestXxxRepository_Contract` wires the GORM factory and delegates to a `runXxxContract(t, repo)` body that holds every preserved assertion. Fake-specific assertions move to a sibling `*_fake_test.go`.

**Independent Test**: `go test -race -run Contract ./internal/repository/store/...` passes against real GORM on SQLite locally. With Docker running locally: same test passes on both dialects. Each contract test's per-method sub-tests (`Create`, `GetByID`, …) keep their names.

### Implementation for User Story 2

> For each of the 6 files: (a) extract assertions into `runXxxContract`, (b) flip wrapper to `ForEachDialect` + factory, (c) move fake-specific assertions to a sibling file.

- [X] T015 [P] [US2] Refactor `internal/repository/store/tags_repository_contract_test.go`: extract body into `runTagsContract(t *testing.T, repo port.TagsRepository)`; top-level `TestTagsRepository_Contract` uses `internaltest.ForEachDialect` and `NewTagsRepository(fx.Runtime.GormDB())`. Move any fake-only assertions to `internal/repository/store/tags_repository_fake_test.go` (create if needed)
- [X] T016 [P] [US2] Refactor `internal/repository/store/api_key_repository_contract_test.go` — same recipe; helper `runAPIKeyContract(t, repo port.APIKeyRepository)`; fake-only assertions to `api_key_repository_fake_test.go`
- [X] T017 [P] [US2] Refactor `internal/repository/store/incident_repository_contract_test.go` — `runIncidentContract(t, repo port.IncidentRepository)`; fake-only to `incident_repository_fake_test.go`
- [X] T018 [P] [US2] Refactor `internal/repository/store/incident_event_step_repository_contract_test.go` — `runIncidentEventStepContract(t, repo port.IncidentEventStepRepository)`; fake-only to sibling
- [X] T019 [P] [US2] Refactor `internal/repository/store/monitoring_activity_repository_contract_test.go` — `runMonitoringActivityContract(t, repo port.MonitoringActivityRepository)`; fake-only to sibling
- [X] T020 [P] [US2] Refactor `internal/repository/store/resource_repository_contract_test.go` — `runResourceContract(t, repo port.ResourceRepository)`; fake-only to sibling
- [X] T021 [US2] Run `go test -race -run Contract ./internal/repository/store/...` — observe all 6 contracts pass on SQLite without Docker
- [X] T022 [US2] With Docker running: run `go test -race -run Contract ./internal/repository/store/...` — observe both dialect iterations pass for all 6 contracts. Capture timing per contract (used to defend SC-004 in polish)

**Checkpoint**: US2 delivered. Contract tests exercise real GORM on both dialects.

---

## Phase 5: User Story 3 — CI runs Postgres dual-dialect under 3 minutes (Priority: P1)

**Goal**: New CI job `test-be-postgres` exists in `.github/workflows/test.yml`, `.github/workflows/ci.yml`, and `.gitlab-ci.yml`. Local target `make test-be-pg` exists. Both run under 180s wall-clock.

**Independent Test**: Push the branch; observe a passing `test-be-postgres` job under 180s; on a Docker-equipped local machine, `make test-be-pg` exits 0.

### Implementation for User Story 3

- [ ] T023 [US3] In root `Makefile`, add target `test-be-pg`: runs `go test -race -timeout 180s ./internal/repository/store/... ./internal/repository/internaltest/...` with explicit empty `POSTGRES_TEST_DSN` to force the testcontainer path. If `docker info >/dev/null 2>&1` fails, print "Docker not available — skipping Postgres tests" and exit 0. Add to `.PHONY`
- [ ] T024 [P] [US3] In `.github/workflows/test.yml`, add a `test-be-postgres` job that: checks out, sets up Go, ensures Docker is available (GitHub-hosted ubuntu runners have it), runs `docker pull postgres:16-alpine` (cached layer), runs `make test-be-pg`. Job timeout 5 minutes hard ceiling; the 180s assertion is enforced by the inner `-timeout 180s` flag on the Go test invocation
- [ ] T025 [P] [US3] In `.github/workflows/ci.yml`, add the same `test-be-postgres` job (mirrors test.yml for main branch)
- [ ] T026 [P] [US3] In `.gitlab-ci.yml`, add a parallel `backend-tests-postgres` job under stage `test`: image `golang:1.25.1`, service `docker:27-dind` (or runner with Docker socket), pulls `postgres:16-alpine`, runs `make test-be-pg`. Rules mirror existing `backend-tests` job
- [ ] T027 [US3] Trigger one CI run on the branch; observe the new job pass and capture wall-clock. Record URL + timing in the PR description (US3 acceptance evidence)

**Checkpoint**: US3 delivered. CI enforces dual-dialect; local devs have the same path.

---

## Phase N: Polish & Cross-Cutting Concerns

- [ ] T028 [P] Run `go vet ./...` — clean
- [ ] T029 [P] Production-import guard (FR-014): run `go list -deps ./cmd/... | grep -E 'testcontainers|internal/repository/internaltest'`; assert empty. The `./cmd/...` pattern covers `cmd/api` and any future binary (e.g. `cmd/migrations-drift-check` from 042). Capture the empty result in PR description as evidence that the production binary has zero test-infra dependencies
- [ ] T030 [P] Create `internal/repository/internaltest/README.md` per FR-013: how to use `ForEachDialect`, how to opt into Postgres locally (Docker or `POSTGRES_TEST_DSN`), how the helper guarantees per-test isolation, how to disable Ryuk on restricted runners
- [ ] T031 [P] Update `CLAUDE.md`: under `## Commands` add `make test-be-pg`; under `### Testing` patterns subsection add a one-line pointer to `internal/repository/internaltest/README.md` and a sentence "Repository contract tests are dual-dialect (SQLite + Postgres) via `internaltest.ForEachDialect`"
- [ ] T032 [P] Update `QUICKSTART.md`: add a "Run tests against Postgres" subsection pointing to `make test-be-pg`
- [ ] T033 Run `make test-be` locally — assert exit 0, no regression vs main. Record tail in PR description (SC-006)
- [ ] T034 With Docker: run `make test-be-pg` 3 times consecutively; record median wall-clock. Assert median ≤ 180s (SC-004). If first run is much slower due to image pull, note the cached-vs-uncached split
- [ ] T035 Scope guard: run `git diff --stat 043-domain-decoupling -- internal/api/ internal/service/ web/ cmd/` and assert empty (production code untouched). Run `git diff --stat 043-domain-decoupling -- internal/repository/store/` and assert only `*_test.go` and `*_fake_test.go` files appear (no non-test changes to repos)
- [ ] T036 SonarQube scan per CLAUDE.md "Code Quality" section; resolve any CRITICAL/BLOCKER introduced under `internal/repository/internaltest/`
- [ ] T037 Cross-check FRs/SCs ↔ tasks coverage map in PR description (FR-001…FR-015, SC-001…SC-008). Flag any unmapped requirement
- [ ] T038 FR-015 verification: run `git grep -n 'BeforeCreate(nil)' -- internal/repository/fake/` and assert zero hits (fakes inherit 043's `EnsureID()` migration). If any hit appears, halt — a fake regressed

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 Setup**: no deps — start immediately
- **Phase 2 Foundational**: needs Phase 1 — blocks US1 and US2
- **Phase 3 US1 (P1)**: needs Phase 2 — MVP candidate (helper alone is shippable; US2 builds on it)
- **Phase 4 US2 (P1)**: needs US1 — can start once `ForEachDialect` is callable
- **Phase 5 US3 (P1)**: needs US1 + US2 — the CI job runs the refactored contracts
- **Phase N Polish**: needs everything else green

### Within Each User Story

- US1: T004–T007 (tests) before T008–T013 (impl); T014 (verify) last
- US2: T015–T020 are independent file edits — parallel-safe; T021/T022 (test runs) after
- US3: T023 (Makefile) before T024–T026 (CI uses it); T027 (CI trigger) last

### Parallel Opportunities

- T004 / T005 / T006 / T007 — distinct test files, run together
- T015 / T016 / T017 / T018 / T019 / T020 — six contract refactors, distinct files
- T024 / T025 / T026 — three CI files, distinct
- T028 / T029 / T030 / T031 / T032 — polish docs/guards, distinct

---

## Parallel Example: User Story 2

```bash
# Six contract refactors, parallel by file:
Task: "T015 tags contract refactor"
Task: "T016 api_key contract refactor"
Task: "T017 incident contract refactor"
Task: "T018 incident_event_step contract refactor"
Task: "T019 monitoring_activity contract refactor"
Task: "T020 resource contract refactor"

# Then sequential validation:
Task: "T021 sqlite-only contract run"
Task: "T022 dual-dialect contract run (Docker required)"
```

---

## Implementation Strategy

### MVP First (US1 only)

1. Phase 1 Setup
2. Phase 2 Foundational
3. Phase 3 US1 — helper exists, self-tested
4. STOP and VALIDATE: helper tests pass; helper README written; can be merged as a usable foundation before US2 lands

### Incremental Delivery

1. MVP US1 shipped → helper is usable from any `_test.go`
2. Add US2 (contract refactor) → 6 contracts now exercise real GORM dual-dialect
3. Add US3 (CI + Makefile) → dual-dialect runs every PR within budget
4. Polish → merge

### Parallel Team Strategy

- Eng A: US1 (helper)
- Eng B (after US1 checkpoint): US2 contract refactors — 6 files distributable
- Eng A or B: US3 once US1 + US2 land

---

## Notes

- [P] = different files, no incomplete-task dependencies
- US1 is the highest-risk piece (new package, testcontainers integration). T009 (container.go) is the longest single task
- US2 is mechanical but voluminous — 6 file refactors. Watch for fake-specific assertions that don't translate; move them, don't delete
- Reject scope creep: no production code edit, no domain change, no migration change. T035 scope guard catches it
- Production binary MUST NOT pull testcontainers (FR-014). T029 verifies
- Commit per task or per logical group (Conventional Commits)
- Stop at each Checkpoint and confirm the independent test before advancing
