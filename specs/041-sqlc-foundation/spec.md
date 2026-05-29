# Feature Specification: sqlc Foundation

**Feature Branch**: `041-sqlc-foundation`
**Created**: 2026-05-29
**Status**: Draft
**Input**: User description: "sqlc foundation - read .prds/sqlc/001-foundation.md to understand the spec"

## Clarifications

### Session 2026-05-29

- Q: Pool sizing defaults for the new Runtime? → A: Postgres `pgxpool` max=25; SQLite `SetMaxOpenConns=1`, `SetMaxIdleConns=1` (single writer)
- Q: sqlc CLI install mechanism & version pin? → A: `go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.27.0` via Makefile when missing; version pinned in Makefile variable
- Q: Couple sqlc to `build-be`? → A: `sqlc-check` runs as pre-step of `build-be` (fail on drift, no auto-write); `sqlc-generate` stays explicit
- Q: How to verify single pool per dialect (SC-006)? → A: Runtime exposes `Stats()` accessor; test asserts single pool instance by identity per dialect

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Maintainer generates type-safe DB code from SQL (Priority: P1)

A backend maintainer adds or edits a `.sql` query file under the sqlc queries tree and runs a single make target. Generated Go code lands in versioned folders, compiles cleanly, and is reviewable in the PR alongside the SQL source. Existing GORM code paths keep working untouched.

**Why this priority**: Foundation ticket. Blocks every subsequent sqlc migration step (schema source, domain decoupling, repo migration). No business value ships until this exists.

**Independent Test**: Add one trivial query (`ping.sql`), run `make sqlc-generate`, confirm Go code appears under `internal/repository/sqlc/{pg,sqlite}/`, project builds, and all existing tests still pass.

**Acceptance Scenarios**:

1. **Given** a fresh checkout with no sqlc binary installed, **When** maintainer runs `make sqlc-generate`, **Then** sqlc installs (or is auto-bootstrapped) and produces compilable Go files under both Postgres and SQLite generated dirs.
2. **Given** the sqlc-generated tree is up to date, **When** maintainer runs `make build-be`, **Then** the backend binary builds successfully with no regression vs. main.
3. **Given** existing test suite, **When** maintainer runs `make test-be`, **Then** every GORM-backed test passes unchanged.

---

### User Story 2 - CI blocks drift between SQL queries and generated code (Priority: P1)

A contributor modifies a `.sql` query file but forgets to regenerate. CI fails fast on the regeneration drift check, preventing merge of inconsistent state.

**Why this priority**: Without this guard, generated code silently desyncs and downstream tickets debug ghost bugs. Cheap to add once `sqlc-generate` exists.

**Independent Test**: Modify a query file in a PR without regenerating; CI job `sqlc-check` fails. Regenerate and commit; CI passes.

**Acceptance Scenarios**:

1. **Given** a PR that edits a `.sql` queries file without committing regenerated Go, **When** CI runs, **Then** the `sqlc-check` step fails with a clear message indicating regen needed.
2. **Given** a PR with both SQL and regenerated Go in sync, **When** CI runs, **Then** `sqlc-check` passes.
3. **Given** existing CI matrix (SQLite + Postgres), **When** CI runs on this branch, **Then** both dialects still execute backend tests successfully.

---

### User Story 3 - Runtime exposes both legacy GORM and raw SQL handles from one connection (Priority: P2)

A future migration ticket needs access to a `*sql.DB` (or pgx pool) without opening a second physical connection. The bootstrap runtime exposes both GORM and raw handles backed by a single underlying connection pool.

**Why this priority**: Required by every later ticket that migrates a repo to sqlc, but no consumer exists yet in this ticket. Must be in place so 003+ can land without a second refactor.

**Independent Test**: Bootstrap the app in community mode (SQLite) and production mode (Postgres). Assert the runtime exposes a non-nil GORM DB plus the matching raw handle, and that they share the same underlying connection (single pool, no duplicate connect logs).

**Acceptance Scenarios**:

1. **Given** the app boots in community mode, **When** bootstrap completes, **Then** runtime exposes a working `*gorm.DB` and a SQLite raw `*sql.DB` sharing one pool.
2. **Given** the app boots in production mode, **When** bootstrap completes, **Then** runtime exposes a working `*gorm.DB` and a pgx pool sharing one connection pool.
3. **Given** existing code calls the legacy `Instance()` accessor, **When** the runtime change ships, **Then** legacy callers keep working (deprecated wrapper) until later tickets remove them.

---

### Edge Cases

- sqlc binary version drift between contributors → version pinned in Makefile/`sqlc.yaml`; mismatched local versions surfaced before generate.
- Generated Go committed but stale relative to `sqlc.yaml` config changes → `sqlc-check` catches diff.
- CI runner without network egress for `sqlc install` → fallback path documented (vendored binary or cached install step).
- SQLite driver swap (modernc → mattn) for perf experiment → runtime contract stays stable because both register through `database/sql`.
- Bootstrap reads existing `*sql.DB` from `*gorm.DB` for SQLite; for Postgres a fresh `pgxpool` opens — verify only one logical pool per dialect.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide an `sqlc.yaml` at repo root configuring two dialects (Postgres, SQLite) with separate query and generated-code paths.
- **FR-002**: System MUST organize sqlc inputs under `internal/repository/sqlc/queries/{postgres,sqlite}/` and generated outputs under `internal/repository/sqlc/{pg,sqlite}/`.
- **FR-003**: System MUST ship at least one dummy pilot query per dialect (e.g., `ping.sql`) so `sqlc-generate` produces real output.
- **FR-004**: System MUST commit generated Go code to the repository (open-core review requirement).
- **FR-005**: Build system MUST expose `make sqlc-generate` that auto-installs the pinned sqlc CLI via `go install github.com/sqlc-dev/sqlc/cmd/sqlc@<version>` when not on PATH, with the version pinned in a Makefile variable, then regenerates all dialects.
- **FR-006**: Build system MUST expose `make sqlc-check` that fails when generated code differs from what `sqlc.yaml` + queries would currently produce.
- **FR-006a**: `make build-be` MUST run `sqlc-check` as a pre-step (fail-fast on drift, no auto-regen). `sqlc-generate` remains explicit and manual.
- **FR-007**: CI MUST run `sqlc-check` and fail the pipeline on drift.
- **FR-008**: CI MUST continue running backend tests against both SQLite and Postgres dialects.
- **FR-009**: System MUST refactor `internal/database/database.go` to expose a `Runtime` struct providing a GORM handle, a raw SQL handle (Postgres-side using `pgxpool` from `pgx/v5`), and a raw SQLite handle (`*sql.DB` via `modernc.org/sqlite`).
- **FR-010**: System MUST guarantee a single underlying connection pool per dialect — the raw and GORM handles MUST share the same pool, not open a second physical connection.
- **FR-010a**: Runtime MUST expose a `Stats()` accessor (returning pool stats: open conns, idle, in-use) usable by tests to assert single-pool identity per dialect.
- **FR-011**: System MUST preserve the legacy global accessor (e.g., `Instance()`) as a deprecated wrapper so existing repos compile without change during the migration window.
- **FR-012**: Bootstrap (`internal/platform/bootstrap/database.go`) MUST consume the new `Runtime` and continue wiring existing GORM repositories with no behavioral regression.
- **FR-013**: System MUST set explicit pool sizing defaults: Postgres `pgxpool` max connections = 25; SQLite `SetMaxOpenConns(1)` + `SetMaxIdleConns(1)` (single-writer, avoids `SQLITE_BUSY`).
- **FR-014**: System MUST NOT introduce CGO dependencies for the SQLite path (Community Edition single-binary, cross-compile guarantee).
- **FR-015**: System MUST NOT modify domain models, business migrations, or any production repository implementation in this ticket.

### Key Entities *(include if feature involves data)*

- **Database Runtime**: New process-level abstraction owning the connection pool(s). Exposes legacy GORM handle plus raw handles for future sqlc-backed repos. One physical pool per active dialect.
- **sqlc Queries Tree**: Versioned SQL inputs partitioned by dialect; the contract source for generated Go code.
- **Generated Repository Code**: Versioned Go output under `internal/repository/sqlc/{pg,sqlite}/`; reviewable artifact; regenerated deterministically from queries + `sqlc.yaml`.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A maintainer can add a new SQL query and produce reviewable generated Go in under 2 minutes end-to-end (`make sqlc-generate` + commit).
- **SC-002**: 100% of PRs that change `.sql` queries without regenerating are blocked by CI before merge.
- **SC-003**: Existing backend test suite passes with zero new failures after the Runtime refactor (regression budget: 0).
- **SC-004**: Existing backend build time does not regress by more than 5% after sqlc-generate is integrated into `build-be`.
- **SC-005**: Community Edition binary remains a single static artifact with zero new system library dependencies (verified by `ldd` or equivalent yielding no new dynamic linkage on Linux).
- **SC-006**: Only one physical connection pool per dialect at runtime, verified by an automated test asserting pool-instance identity via the Runtime `Stats()` accessor.

## Assumptions

- sqlc CLI version is pinned in `sqlc.yaml` / Makefile; contributors install the same version locally and in CI.
- Postgres driver decision is final: `pgx/v5` with `pgxpool` (no `lib/pq`, no `database/sql` wrapper for the production path).
- SQLite driver decision is final: `modernc.org/sqlite`, accepting ~20-40% slower batch inserts vs. CGO `mattn` driver for the community profile.
- Generated code is committed (open-core review constraint); not generated on-the-fly at build time in production deploys.
- Legacy `Instance()` accessor remains intact as a deprecation shim; deprecation removal is owned by later tickets in the sqlc track.
- No domain model, migration, or production repository change is in scope — those are tickets 002 and 003+.
- CI runners can reach the public Go module proxy for sqlc install (or a vendored binary path is acceptable if egress is blocked).
- Existing dual migration trees (`internal/database/migrations/{sqlite,postgres}/`) are untouched in this ticket.
