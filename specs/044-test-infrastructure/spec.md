# Feature Specification: Test Infrastructure — Dual-Dialect

**Feature Branch**: `044-test-infrastructure`
**Created**: 2026-05-30
**Status**: Draft
**Input**: User description: "Test infrastructure dual-dialecte — read .prds/sqlc/004-test-infrastructure.md"

## Clarifications

### Session 2026-05-30

- Q: Postgres provisioning in CI + locally? → A: Testcontainers-go inside the test code, single launcher path used by both CI and `make test-be-pg`. CI caches the Postgres image; locally Docker is the only prerequisite.
- Q: Per-test Postgres database allocation? → A: One container per `go test` package; migrations applied once to a `template` database inside that container; each test creates its own isolated database via `CREATE DATABASE … TEMPLATE template_…`. Container lifecycle owned by `TestMain` of the consuming package via the helper.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - A maintainer can run a single test against SQLite AND Postgres with one helper (Priority: P1)

A maintainer writes a test that exercises repository behavior. By calling a single dual-dialect helper (e.g. `ForEachDialect(t, fn)`), the test runs twice: once against an in-memory / temp-file SQLite database, once against a real Postgres instance. Migrations are applied for each dialect before the test body runs. When `POSTGRES_TEST_DSN` is unset (typical local dev), the Postgres iteration is skipped automatically; in CI, the env var is set and both iterations run.

**Why this priority**: Every later ticket (005 pilot tags, 006/007/008 repo migrations) needs to assert that GORM and sqlc-backed implementations produce identical behavior on both dialects. Without a single source of test setup, every consumer reinvents the wheel and dialect coverage drifts.

**Independent Test**: Author a trivial repository test using the new helper. With `POSTGRES_TEST_DSN` unset, the test runs once (SQLite) and reports skipping Postgres. With `POSTGRES_TEST_DSN=...` set, the same test runs twice (SQLite + Postgres), both pass, both have migrations applied.

**Acceptance Scenarios**:

1. **Given** a test using `ForEachDialect`, **When** `POSTGRES_TEST_DSN` is unset, **Then** the SQLite iteration runs to completion and the Postgres iteration is skipped with a clear message.
2. **Given** a test using `ForEachDialect`, **When** `POSTGRES_TEST_DSN` is set to a reachable instance, **Then** both iterations run, each gets a freshly-migrated isolated database, and the test body sees a clean schema both times.
3. **Given** two distinct tests using the helper in the same package, **When** they run in parallel (`t.Parallel()`), **Then** each gets its own isolated database (no cross-test data leakage).
4. **Given** a test using the helper, **When** the test body returns or panics, **Then** the helper releases / drops the isolated database deterministically (via `t.Cleanup`).

---

### User Story 2 - The 6 existing repository contract tests exercise real SQLite and Postgres (Priority: P1)

Today, `internal/repository/store/*_contract_test.go` files (6 of them: tags, api_key, incident, incident_event_step, monitoring_activity, resource) instantiate the in-memory fake implementation. After this ticket, each contract test is parameterized by a factory `func(db) port.XxxRepository`, runs against the real GORM-backed implementation, and is executed by `ForEachDialect` so both SQLite and Postgres are covered. The contract tests become the baseline oracle that future sqlc-backed implementations must also pass.

**Why this priority**: A "contract test" that only tests the fake is a tautology. The investment in the helper from US1 has no payoff until real repositories run through it. This ticket flips the contract tests from fake-shape verification to true repository-contract verification.

**Independent Test**: Run `go test -race ./internal/repository/store/... -run Contract` — in local dev, SQLite iterations of all 6 contract tests pass; in CI with Postgres testcontainer up, both dialects of all 6 contract tests pass.

**Acceptance Scenarios**:

1. **Given** an existing fake-backed contract test for one of the 6 repositories, **When** the refactor lands, **Then** the test (a) accepts a `factory func(*gorm.DB) port.XxxRepository` argument, (b) runs through `ForEachDialect`, (c) passes against the real `store.NewXxxRepository(gormDB)` on both dialects.
2. **Given** all 6 refactored contract tests, **When** CI runs, **Then** they collectively complete in under 3 minutes (including testcontainer startup overhead, per PRD acceptance criterion).
3. **Given** a future sqlc-backed implementation of the same repository, **When** it is added to the factory set, **Then** the unchanged contract test can validate the sqlc impl with no code change beyond adding the factory.
4. **Given** the existing fake-based behavior assertions, **When** the refactor lands, **Then** the test intent is preserved — no assertion is weakened, only the implementation under test changes. The fake-backed assertions migrate to a sibling `*_fake_test.go` if their value is to keep the fake itself honest.

---

### User Story 3 - CI runs Postgres-backed tests with bounded startup overhead (Priority: P1)

A CI pipeline brings up a containerized Postgres, runs migrations, executes the dual-dialect test matrix, and tears down. The job completes in under 3 minutes end-to-end. Local developers do not need Docker to iterate on SQLite tests; setting `POSTGRES_TEST_DSN` opts them into the Postgres dialect locally.

**Why this priority**: A pipeline that takes 10 minutes for the Postgres path discourages contributors and gets disabled. Bounded overhead is part of "the feature works" for any test-infra ticket.

**Independent Test**: Trigger the new CI job on a no-op PR; the job reports a single status check `test-be-postgres` with total wall-clock under 3 minutes. Cancel any test that would push the job over 3 minutes — the budget is enforced.

**Acceptance Scenarios**:

1. **Given** the new CI job `test-be-postgres`, **When** it runs on a PR, **Then** it brings up a Postgres testcontainer, applies migrations, runs the dual-dialect contract test suite, and exits zero — all within 3 minutes wall-clock.
2. **Given** a developer with Docker installed, **When** they run `make test-be-pg` locally, **Then** the same testcontainer-driven flow runs and exits zero (or skips gracefully if Docker is not available, with a clear message).
3. **Given** the SQLite-only CI path that exists today, **When** the new Postgres job is added, **Then** the existing SQLite-only job keeps running unchanged.

---

### Edge Cases

- `POSTGRES_TEST_DSN` set but the Postgres instance is unreachable → the helper fails fast with a clear error naming the DSN (sanitized to hide the password), not a generic timeout.
- Multiple parallel tests using the helper simultaneously → each gets its own database (e.g. via `CREATE DATABASE ogoune_test_<ULID>` per test); cleanup drops them on `t.Cleanup`.
- Migrations fail on either dialect → the helper bubbles up the error, including which dialect failed and the migration filename, so the maintainer doesn't guess.
- A test takes longer than CI's 3-minute budget → CI surfaces it as a test failure (timeout), not a flaky pass.
- Docker daemon not running locally → `make test-be-pg` skips with an explicit "docker not available" message; `go test ./...` against SQLite continues unaffected.
- Test author calls `ForEachDialect` inside a test that was already inside another `t.Run` named like a dialect → no collision; sub-test names use the helper's known constants.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Project MUST introduce an `internaltest` package (location: `internal/repository/internaltest/`) housing the dual-dialect helpers; the package MUST be import-only from `_test.go` files (no production import).
- **FR-002**: The package MUST expose `SetupSQLite(t *testing.T) *DialectFixture` returning a freshly-migrated, isolated SQLite database. Tests get a brand-new schema per call; cleanup is registered via `t.Cleanup`.
- **FR-003**: The package MUST expose `SetupPostgres(t *testing.T) *DialectFixture` returning a freshly-migrated, isolated Postgres database backed by `POSTGRES_TEST_DSN`. If the env var is empty, the helper MUST call `t.Skip` with a clear message and return without panicking.
- **FR-004**: The package MUST expose `ForEachDialect(t *testing.T, fn func(t *testing.T, dialect string, fx *DialectFixture))` which runs `fn` once per supported dialect inside its own `t.Run(dialect, ...)` sub-test. Skipping Postgres when `POSTGRES_TEST_DSN` is empty MUST NOT skip the SQLite iteration.
- **FR-005**: The `DialectFixture` MUST expose at minimum: the GORM `*gorm.DB`, the raw `*sql.DB` (and, for Postgres, the `*pgxpool.Pool`), and the dialect name. This mirrors the runtime `Runtime` shape from 041.
- **FR-006**: Per-test isolation MUST be strict — two tests running in parallel under the same helper MUST NOT see each other's data. For Postgres the allocation strategy is: one container per `go test` package, migrations applied once to a `template_ogoune` database, then each test creates its own isolated database via `CREATE DATABASE ogoune_test_<random> TEMPLATE template_ogoune`. This amortizes container startup + migration cost across the package while giving each test a fresh schema-loaded DB in milliseconds. For SQLite, isolation is per-temp-file (one DB file per test under `t.TempDir()`).
- **FR-007**: Cleanup MUST be deterministic. On `t.Cleanup`, SQLite temp files are removed, Postgres databases are dropped, and connections are closed. Cleanup failures MUST be reported as test errors, not silently swallowed.
- **FR-008**: The 6 existing `*_contract_test.go` files under `internal/repository/store/` (tags, api_key, incident, incident_event_step, monitoring_activity, resource) MUST be refactored to:
  - Accept a factory `func(*gorm.DB) port.XxxRepository` (one per repository) so the same test body can validate any concrete implementation.
  - Be invoked by `ForEachDialect` so each dialect runs the full contract.
  - Default to wiring the existing GORM `store.NewXxxRepository(...)` factory — establishing the baseline oracle.
- **FR-009**: Refactored contract tests MUST NOT lose any assertion present today. Any assertion that specifically targets the fake's in-memory behavior (not a repository contract) MUST move to a separate `*_fake_test.go` file rather than be deleted.
- **FR-010**: Project MUST add a Makefile target `make test-be-pg` that runs the backend test suite with Postgres enabled. Provisioning is owned by testcontainers-go inside the test code itself (Clarification Q1) — the Make target sets no explicit `POSTGRES_TEST_DSN`; the test code provides the DSN once the container is ready. When Docker is unavailable the target MUST skip with a clear message and exit zero.
- **FR-011**: CI MUST run a job `test-be-postgres` that brings up Postgres via testcontainers-go inside the test code (CI does not declare a Postgres `services:` block), applies migrations via the helper, runs the dual-dialect test matrix, and completes in **under 3 minutes** wall-clock per PR. The Postgres image MUST be cached (Docker layer cache or actions/cache) to keep startup amortized.
- **FR-012**: The existing SQLite-only CI test job MUST remain unchanged so the fast path stays fast. The Postgres job is additive.
- **FR-013**: The new package MUST ship a README at `internal/repository/internaltest/README.md` documenting: how to use `ForEachDialect`, how to opt into Postgres locally, and how the helper guarantees per-test isolation.
- **FR-014**: The helper MUST NOT introduce a production-code dependency on Docker, testcontainers, or any Postgres-specific package — these dependencies live exclusively under the `internaltest` package and the test files.
- **FR-015**: Fakes MUST already use `EnsureID()` (from 043) — this ticket re-verifies and does not regress.

### Key Entities *(include if feature involves data)*

- **DialectFixture**: Test-only handle owning one isolated database (SQLite tempfile or Postgres named DB). Exposes GORM + raw handle, dialect name, and a deterministic cleanup contract.
- **Dialect**: Logical constant set: `"sqlite"`, `"postgres"`. Iteration order is fixed (SQLite first) for deterministic CI logs.
- **Contract Test Factory**: Function type `func(*gorm.DB) port.XxxRepository` — the indirection that lets one contract body validate multiple implementations (today: GORM only; future: sqlc).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A new test using `ForEachDialect` can be written in ≤ 10 lines of helper-related boilerplate. The example in `internal/repository/internaltest/README.md` MUST fit on a single screen.
- **SC-002**: All 6 existing contract tests pass against the real GORM-backed implementation on SQLite locally (no Postgres required) — 100% pass rate after refactor.
- **SC-003**: All 6 contract tests pass against the real GORM-backed implementation on BOTH SQLite and Postgres in CI — 100% pass rate.
- **SC-004**: CI job `test-be-postgres` wall-clock time ≤ **180 seconds** (3 minutes) per PR, measured as median across 5 consecutive runs on a clean cache.
- **SC-005**: A developer with Docker installed can run `make test-be-pg` locally and see both dialect iterations of all 6 contract tests pass.
- **SC-006**: A developer without Docker can run `make test-be` (existing target) and observe zero regression from before this ticket; SQLite-only flow is untouched.
- **SC-007**: Zero data leakage between parallel tests using `ForEachDialect` — verified by an explicit isolation test that runs two parallel sub-tests writing rows with the same primary key and asserts neither sees the other's row.
- **SC-008**: When a future sqlc-backed implementation of any of the 6 repositories ships, adding it to the contract test matrix is a one-line factory addition — the contract test body itself does not need to change.

## Assumptions

- 043 (domain decoupling) is merged or present on the same branch — fakes use `EnsureID()`; the typed crypto wrappers exist but are not exercised by this infra ticket.
- 041 (sqlc foundation) is merged or present on the same branch — `database.Runtime` exposes raw handles (`PgxPool`, `SQLiteDB`) needed by `DialectFixture`.
- Postgres is provisioned in CI via a containerized service (testcontainers OR a docker-compose service) — the exact choice is a plan-level decision, not a spec-level one. The acceptance criterion is the 3-minute budget.
- `POSTGRES_TEST_DSN` is the single env var that switches Postgres on/off across both local and CI contexts.
- "Per-test isolation" for Postgres means a fresh database per test (cloned from a migrated template inside the package's container), not a fresh schema per test, to avoid `search_path` complications and migration replays inside one connection.
- The existing SQLite test helpers in `internal/database/test_helpers_test.go` continue to work for tests that don't need the dual-dialect runner; this ticket does not consolidate them.
- The 6 contract test bodies need refactoring **and so do the fakes-as-implementation-under-test patterns** they encode — current contract tests instantiate `fake.NewTagsFake()` etc.; after this ticket the default is GORM.
- Production code (`cmd/`, `internal/api/`, `internal/service/`, `internal/repository/store/` non-test files) MUST NOT depend on testcontainers / Docker / `internaltest`.
- No domain change, no migration change, no env var change in production deploy paths.
- The new CI job runs on the same OS/Go version as the existing test job for parity.
