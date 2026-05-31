# Feature Specification: sqlc Schema Source

**Feature Branch**: `042-sqlc-schema-source`
**Created**: 2026-05-29
**Status**: Draft
**Input**: User description: "Schema source: migrations consommables par sqlc — read .prds/sqlc/002-schema-source.md"

## Clarifications

### Session 2026-05-29

- Q: Drift-check tool implementation form? → A: Go program at `cmd/migrations-drift-check`, runnable via `go run` and `make migrations-drift-check`
- Q: Type-compatibility check scope? → A: Name + nullability only; type pairs are reviewer responsibility guided by README mapping table
- Q: Audit findings deliverable? → A: `internal/database/migrations/README.md` — mapping table + "Non-DDL statements inventory" subsection (durable, versioned)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - sqlc consumes existing migrations as schema source without errors (Priority: P1)

A backend maintainer runs `make sqlc-generate` after pointing `sqlc.yaml` at the actual migration trees (`internal/database/migrations/postgres/` and `internal/database/migrations/sqlite/`). sqlc reads all 14 `.up.sql` files per dialect (or whatever count exists at run time) and produces compilable Go code with zero parser errors.

**Why this priority**: Foundation for every subsequent sqlc-backed repository ticket. Without a clean schema source, no real query can be generated. Today's `sqlc.yaml` (from 041) points at the migration trees but only ships a `Ping` query — schema parsing has not been exercised on the real DDL. Confirming it works (or fixing what does not) is the gate.

**Independent Test**: Run `make sqlc-generate` and `make sqlc-check` against the unchanged committed schema. Both succeed. Inspect any generated `models.go` that references real tables to confirm columns are typed and named as expected.

**Acceptance Scenarios**:

1. **Given** the current 14 Postgres migration files, **When** `sqlc generate` reads them as schema, **Then** it emits no `unsupported statement` / parser errors and produces compilable Go output for the existing pilot query plus any new queries added to test schema consumption.
2. **Given** the current 14 SQLite migration files, **When** `sqlc generate` reads them as schema, **Then** the same outcome holds for SQLite — no parser errors, compilable output.
3. **Given** parser-incompatible statements identified during audit (e.g. SQLite `PRAGMA`, triggers, non-DDL directives), **When** the implementation lands, **Then** the maintainer can either run the existing migration unchanged (sqlc skips/handles cleanly) or consume a curated aggregated schema artifact (`internal/database/schema/{postgres,sqlite}.sql`) — the choice is documented and the migrations themselves are never rewritten.

---

### User Story 2 - CI guards SQLite/Postgres migration drift (Priority: P1)

A contributor adds a new migration to one dialect tree and forgets the matching file in the other. CI fails fast on a drift-check job that enforces 1-to-1 file pairing and per-table column-name + nullability alignment.

**Why this priority**: A divergence between dialect trees today silently breaks shared domain behavior; with sqlc consuming both, it also corrupts generated code. Cheap to add a Go-based linter once and run it in CI.

**Independent Test**: In a throwaway branch, add `0099_test.up.sql` only to `internal/database/migrations/postgres/`; CI `migrations-drift-check` job fails with a clear message naming the missing SQLite file. Add the SQLite counterpart with matching column names + nullability; CI passes.

**Acceptance Scenarios**:

1. **Given** an additive migration file present in one tree but missing in the other, **When** the drift-check job runs, **Then** it exits non-zero and names the missing file path.
2. **Given** a `CREATE TABLE` / `ALTER TABLE ADD COLUMN` whose column appears in both dialects but with different nullability or a different column name, **When** the drift-check runs, **Then** it exits non-zero and names the dialect, table, and column where the divergence is.
3. **Given** clean, aligned trees, **When** the drift-check runs, **Then** it exits zero and the build proceeds.

---

### User Story 3 - Documented type-mapping conventions prevent future drift (Priority: P2)

A future contributor adds a new column. They consult a documented table-of-correspondence under `internal/database/migrations/README.md` and CLAUDE.md, learn that ULIDs are `TEXT(26)`, timestamps are `TIMESTAMPTZ` (Postgres) / `TEXT` ISO-8601 (SQLite), JSON payloads are `JSONB` (Postgres) / `TEXT` (SQLite), money is `BIGINT` cents, and write the new migration consistently the first time.

**Why this priority**: Prevents the next drift, but doesn't gate this PR's correctness — the audit + linter (US1 + US2) carry the load. Documentation is the multiplier.

**Independent Test**: Open the updated docs, verify each documented type has both dialect mappings and a brief rationale. A reader unfamiliar with the project can pick the right Postgres + SQLite type for a hypothetical new column (date, money, flag, JSON, ULID, foreign-key id) using only the table.

**Acceptance Scenarios**:

1. **Given** a new contributor reading `internal/database/migrations/README.md`, **When** they need to add a JSON-shaped column, **Then** the doc tells them to use `JSONB` (Postgres) and `TEXT` (SQLite) with the same column name and nullability, plus a one-line reason.
2. **Given** the CLAUDE.md "Database migrations" section, **When** they need to know whether a `PRAGMA` or trigger is allowed, **Then** the doc states it requires tech-lead validation (and explains why — sqlc compatibility).

---

### Edge Cases

- A historical migration is unparseable by sqlc and rewriting it is forbidden → fall back to a curated `internal/database/schema/{postgres,sqlite}.sql` file pointed at by `sqlc.yaml`; the original migrations remain authoritative at runtime; this fallback is opt-in per dialect.
- Column name matches across dialects but the type semantics differ (e.g. `BIGINT` vs `INTEGER` in SQLite) → linter checks name + nullability only; type compatibility is owned by the type-mapping doc and reviewer judgement.
- A migration adds an index or constraint (no column delta) → drift-check passes silently as long as the file-pair exists.
- A migration is removed/renamed → drift-check catches the file-pair mismatch; renames require coordinated edits in both trees.
- `PRAGMA`, trigger, or function in a SQLite migration → audit identifies it; if sqlc tolerates it, leave it; if not, isolate via aggregated schema fallback.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Project MUST perform a one-time audit listing every divergence between matching files in `internal/database/migrations/postgres/` and `internal/database/migrations/sqlite/`. The audit is a manual sweep — no dedicated audit script is committed; ongoing enforcement is the Go drift-check tool (FR-009). Audit findings are recorded in `internal/database/migrations/README.md` per FR-010.
- **FR-002**: Audit MUST identify columns typed differently between dialects (e.g. `JSONB` vs `TEXT`, `TIMESTAMPTZ` vs `TEXT`) and record them in `internal/database/migrations/README.md` as the canonical type-mapping table.
- **FR-003**: Audit MUST flag every non-pure-DDL statement (triggers, `PRAGMA`, functions, etc.) that may disturb the sqlc parser, with a per-occurrence verdict (tolerated by sqlc / requires aggregated-schema fallback), recorded as a "Non-DDL statements inventory" subsection in `internal/database/migrations/README.md`.
- **FR-004**: After audit, `sqlc generate -f sqlc.yaml` MUST succeed against the actual migration trees with zero `unsupported statement` / parser errors.
- **FR-005**: System MUST NOT rewrite or alter any historical migration file to fix sqlc parsing. If a migration is unparseable, the project MUST introduce `internal/database/schema/{postgres,sqlite}.sql` curated files and repoint `sqlc.yaml` `schema:` at them. This fallback is opt-in per dialect.
- **FR-006**: Repository MUST gain (or preserve, when present) the convention that columns intended to carry JSON content use `JSONB` (Postgres) and `TEXT` (SQLite) with identical column name and nullability. The convention is documented (not retro-fitted to existing rows).
- **FR-007**: CI MUST run a `migrations-drift-check` job that, for every numeric migration prefix `NNNN_*`, verifies a file exists in both `postgres/` and `sqlite/` directories and fails the pipeline when a counterpart is missing.
- **FR-008**: CI `migrations-drift-check` MUST parse `CREATE TABLE` and `ALTER TABLE ADD COLUMN` statements from both dialects and fail when the same table-column pair has a different name or nullability across dialects. Type checking is explicitly OUT of scope for the linter (`JSONB`↔`TEXT`, `TIMESTAMPTZ`↔`TEXT`, and similar cross-dialect type pairs are intended). Type-pair correctness is reviewer responsibility, guided by the README mapping table.
- **FR-009**: The drift-check tool MUST be a Go program living at `cmd/migrations-drift-check`, invokable via `go run ./cmd/migrations-drift-check` and via a `make migrations-drift-check` target (CI and local use the same target). No shell/awk implementation; no dependency on `sqlc vet` for the file-pair and column lints.
- **FR-010**: `internal/database/migrations/README.md` MUST contain (a) a maintained type-mapping table covering at minimum: ULID, timestamp, money, boolean/flag, JSON, foreign-key id; and (b) the audit's "Non-DDL statements inventory" from FR-003. The README is the durable, versioned audit artifact — the PR description references it but is not authoritative.
- **FR-011**: `CLAUDE.md` "Database migrations" section MUST document: (a) one migration = two files with the same numeric prefix and the same intent; (b) `PRAGMA`, triggers, and functions require tech-lead validation; (c) JSON columns use `JSONB`/`TEXT` per the README's mapping.
- **FR-012**: System MUST NOT change runtime migration application logic, runtime schema validation, or any domain/repository code.

### Key Entities *(include if feature involves data)*

- **Migration File Pair**: A `NNNN_<name>.sql` file present in both `postgres/` and `sqlite/` with matching intent. The pair is the unit guarded by the drift-check.
- **Type-Mapping Table**: Documented, single-source-of-truth mapping between logical column types (ULID, timestamp, JSON, …) and dialect-specific SQL types. Owns column-name and nullability conventions.
- **Aggregated Schema File (optional fallback)**: Curated `internal/database/schema/{postgres,sqlite}.sql` consumed by sqlc when a historical migration is unparseable. Introduced only on proven need; original migrations remain authoritative at runtime.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `make sqlc-generate` succeeds against the actual migration trees with zero parser errors logged.
- **SC-002**: 100% of PRs that add or modify a migration in only one dialect are blocked by CI before merge (verified by adding a deliberate single-tree migration to a draft branch and observing the failure).
- **SC-003**: 100% of PRs that introduce a same-named column with different nullability across dialects are blocked by CI before merge.
- **SC-004**: A new contributor can choose the correct Postgres + SQLite type for a new column of each of the documented kinds (ULID, timestamp, money, boolean, JSON, foreign-key id) using only `internal/database/migrations/README.md`, with no follow-up question to a maintainer (validated qualitatively in PR review on the next migration-touching PR).
- **SC-005**: Runtime behavior of the application is unchanged: migration application order, schema validation, and existing automated tests pass with zero regression.

## Assumptions

- Migration counts (14 per dialect today) may evolve before this ticket lands; the audit and linter operate on whatever pairs exist at execution time, not on a frozen count.
- sqlc CLI version is the one pinned in 041 (`v1.27.0`); upgrade is out of scope.
- The drift-check tool is a Go program (consistent with the rest of the codebase) rather than a shell script, so it can be invoked via `go run` in CI without an extra runtime.
- "Same intent" between two paired migrations is enforced by the linter only at the column-name + nullability level. Deeper semantic alignment (defaults, check constraints, index strategies) remains a reviewer judgement.
- The aggregated-schema fallback (`internal/database/schema/*.sql`) is not introduced unless an audit-identified migration actually fails sqlc parsing.
- No new env vars, no new external services.
- Documentation updates land in this PR; future drift prevention relies on contributors reading the README and CLAUDE.md.
- 041 (sqlc foundation) is merged or at minimum present on the same branch — its `sqlc.yaml`, `Makefile`, and runtime layout are prerequisites.
