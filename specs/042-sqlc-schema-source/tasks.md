---

description: "Task list for 042 sqlc Schema Source"
---

# Tasks: sqlc Schema Source

**Input**: Design documents from `/specs/042-sqlc-schema-source/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: REQUIRED — new Go program (`cmd/migrations-drift-check`) needs unit tests; CI integration needs a happy-path regression test on the real migration tree.

**Organization**: Tasks grouped by user story so each can be implemented and shipped independently.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Different file, no incomplete deps
- **[Story]**: Maps to spec user story (US1, US2, US3)
- Paths repo-relative under `/Users/yaovi/Projects/perso/ogoune/`

## Path Conventions

Single Go service + Vue SPA (frontend untouched). New code lives at `cmd/migrations-drift-check/`. Doc updates land in `internal/database/migrations/README.md` and `CLAUDE.md`.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Light prep — only the audit-evidence baseline + Makefile variable. No code yet.

- [X] T001 Record pre-plan sqlc parse baseline (already done at plan time): note in PR description that `sqlc generate -f sqlc.yaml` exits 0 on the unchanged migration trees and produces 20 typed models per dialect (proof: `grep -c '^type ' internal/repository/sqlc/{pg,sqlite}/models.go` returns 20 each)
- [X] T002 Identify CI workflow files to update: `.github/workflows/ci.yml`, `.github/workflows/test.yml`, `.gitlab-ci.yml`. Capture paths in PR description for T020

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: One-time audit data collection that feeds the README artifact (US1) and proves the linter rules are correctly scoped (US2).

**⚠️ CRITICAL**: US1 (README artifact) and US2 (linter) both depend on the audit output.

- [X] T003 Run a manual diff sweep on `internal/database/migrations/postgres/` vs `internal/database/migrations/sqlite/`: per file-pair, record column-name + nullability divergences in a scratch note (used to validate the linter's findings later in T015). Output stays in the PR description, not committed
- [X] T004 Grep every `.sql` file under `internal/database/migrations/{postgres,sqlite}/` for `CREATE TRIGGER`, `CREATE FUNCTION`, `PRAGMA`, `DO $$`. Record any hits in a scratch list for the "Non-DDL inventory" subsection of T006

**Checkpoint**: Audit findings captured. Foundation ready.

---

## Phase 3: User Story 1 — sqlc consumes migrations + audit artifact (Priority: P1) 🎯 MVP

**Goal**: Maintainer-facing proof that sqlc consumes the existing migration trees, plus the durable audit artifact in `README.md` (type-mapping table + Non-DDL inventory).

**Independent Test**: `make sqlc-generate` exits 0 against unchanged migrations (already true). Open `internal/database/migrations/README.md` — type-mapping table covers ULID/timestamp/money/boolean/JSON/FK id; Non-DDL inventory is present (even if empty); drift-check pointer and fallback note are present.

### Tests for User Story 1

> No code tests in this story (documentation deliverable). Validation is reviewer-driven against the contract in `specs/042-sqlc-schema-source/contracts/migrations-readme.md`.

### Implementation for User Story 1

- [X] T005 [US1] Update `internal/database/migrations/README.md`: preserve the existing introductory bullets; insert a new "## Type mapping" section per `contracts/migrations-readme.md` §2 with all six logical types (ULID, Timestamp, Money, Boolean, JSON, FK id) and concise per-row rationale (1 sentence each)
- [X] T006 [US1] In the same `internal/database/migrations/README.md`, insert a "## Non-DDL statements inventory" section per `contracts/migrations-readme.md` §3 populated from T004's findings. If T004 found nothing, render the table with a single italic note `_(none — only runtime-applied SQLite PRAGMAs in Go code; see internal/database/sqlite.go)_`
- [X] T007 [US1] In the same `internal/database/migrations/README.md`, append a "## Drift-check tool" one-paragraph pointer per `contracts/migrations-readme.md` §4 and a "## Aggregated-schema fallback" one-paragraph note per §5
- [X] T008 [US1] Sanity: run `sqlc generate -f sqlc.yaml` after the README edits land; confirm exit 0 (README changes cannot affect codegen, but the check costs nothing and documents the SC-001 baseline)

**Checkpoint**: US1 delivered. Audit artifact is committed in `README.md`. SC-001 baseline documented.

---

## Phase 4: User Story 2 — CI guards SQLite/Postgres migration drift (Priority: P1)

**Goal**: A Go-based linter (`cmd/migrations-drift-check`) enforces 1-to-1 file pairing + column-name + nullability alignment, wired into Makefile + 3 CI workflows.

**Independent Test**: In a throwaway branch, add `0099_test.up.sql` only to `postgres/` (no SQLite counterpart). Run `make migrations-drift-check` locally → exits 1 with `missing pair for prefix 0099`. Add the SQLite counterpart with matching column names + nullability; rerun → exits 0.

### Tests for User Story 2 ⚠️

> Write FIRST and observe FAIL before code lands.

- [X] T009 [P] [US2] Create test fixture trees under `cmd/migrations-drift-check/testdata/ok/{postgres,sqlite}/0001_init.sql` — same `CREATE TABLE foo (id TEXT NOT NULL PRIMARY KEY, name TEXT)` on both sides
- [X] T010 [P] [US2] Create test fixture trees under `cmd/migrations-drift-check/testdata/missing_pair/postgres/0001_init.sql` and `cmd/migrations-drift-check/testdata/missing_pair/sqlite/.gitkeep` — only postgres file present
- [X] T011 [P] [US2] Create test fixture trees under `cmd/migrations-drift-check/testdata/null_drift/{postgres,sqlite}/0001_init.sql` — same column name, opposite nullability across dialects
- [X] T012 [P] [US2] Add test `cmd/migrations-drift-check/pair_test.go` that invokes the linter (via `main.run(args)` or equivalent injectable entry point) against `testdata/ok` and asserts exit code 0; against `testdata/missing_pair` and asserts exit code 1 + stderr contains "missing pair for prefix 0001"
- [X] T013 [P] [US2] Add test `cmd/migrations-drift-check/column_test.go` that invokes against `testdata/null_drift` and asserts exit code 1 + stderr contains "nullability drift"
- [X] T014 [P] [US2] Add a real-tree regression test `cmd/migrations-drift-check/real_tree_test.go` that invokes the linter with `-root` pointing at the actual `internal/database/migrations` and asserts exit code 0 (catches accidental future regressions in the real tree)

### Implementation for User Story 2

- [X] T015 [US2] Create `cmd/migrations-drift-check/pair.go` implementing `MigrationFile`/`MigrationPair` per `data-model.md`: scan a `<root>/{postgres,sqlite}/` directory pair, build pairs keyed by 4-digit prefix, return a slice of unpaired prefixes with their dialect + path
- [X] T016 [US2] Create `cmd/migrations-drift-check/column.go` implementing `ColumnDef` + `SchemaSnapshot` per `data-model.md`: scan a list of `.sql` files, extract columns from `CREATE TABLE name (...)` blocks (line-oriented state machine tracking paren depth; strip `--` comments and `/* */` blocks) and from `ALTER TABLE name ADD COLUMN col TYPE [constraints]` statements. Apply nullability rule from research R4 (`NOT NULL` explicit OR `PRIMARY KEY` ⇒ `NotNull=true`; else nullable). Lowercase table + column names for cross-dialect comparison
- [X] T017 [US2] Create `cmd/migrations-drift-check/main.go`: parse `-root` (default `internal/database/migrations`) and `-verbose` flags; invoke `pair.go` then `column.go`; emit stderr messages per `contracts/drift-check-cli.md`; exit codes 0/1/2 per contract
- [X] T018 [US2] Run `go test -race ./cmd/migrations-drift-check/...` and observe T012, T013 PASS (and T014 PASS on the real tree)
- [X] T019 [US2] Add `migrations-drift-check` target to root `Makefile`: `migrations-drift-check: ; go run ./cmd/migrations-drift-check`. Add it to `.PHONY` line
- [X] T020 [US2] Add a step `migrations drift check` running `make migrations-drift-check` to `.github/workflows/test.yml` (placed before the test step), to `.github/workflows/ci.yml` `backend-tests` job (same position, between `sqlc check` and tests), and to `.gitlab-ci.yml` `backend-tests` `script:` block (line before `go test`)

**Checkpoint**: US2 delivered. Linter present, tested, wired into all three CI workflows.

---

## Phase 5: User Story 3 — Documented conventions prevent future drift (Priority: P2)

**Goal**: CLAUDE.md gains a tight, contributor-facing "Database migrations" section that points at the README and surfaces the file-pair + `PRAGMA` rules.

**Independent Test**: A new contributor reading only `CLAUDE.md` `## Patterns to follow → Database migrations` knows: (a) two files required per migration, (b) column name + nullability MUST match, (c) JSON columns use `JSONB`/`TEXT`, (d) `PRAGMA`/triggers/functions require tech-lead validation, (e) where to find the full mapping table.

### Implementation for User Story 3

- [X] T021 [US3] In `CLAUDE.md`, locate the existing "## Patterns to follow → Database migrations" subsection. Append (do not replace existing bullets) the following bullets:
  - "One migration = two files with the same `NNNN_` prefix and the same intent. Drift between trees is enforced by `make migrations-drift-check`."
  - "Column **name + nullability MUST match** across dialects. Type tokens are intentionally NOT enforced cross-dialect (`JSONB`↔`TEXT`, `TIMESTAMPTZ`↔`TEXT`)."
  - "JSON columns: `JSONB` (Postgres) / `TEXT` (SQLite). See `internal/database/migrations/README.md` for the full type-mapping table."
  - "`PRAGMA`, triggers, and stored functions in `.sql` migration files require tech-lead validation (sqlc compatibility risk)."
- [X] T022 [US3] In `CLAUDE.md` "## Gotchas" section, add a one-line entry: "After editing any migration `.sql`, run `make migrations-drift-check` locally. CI runs it before tests and fails on drift."

**Checkpoint**: US3 delivered. Convention is discoverable from CLAUDE.md.

---

## Phase N: Polish & Cross-Cutting Concerns

- [X] T023 [P] Run `make lint` and `go vet ./cmd/migrations-drift-check/...` — clean
- [X] T024 [P] Run `make run-ci` end-to-end locally (lint + race tests + build) — exit 0
- [X] T025 [P] Walk through `specs/042-sqlc-schema-source/quickstart.md` end-to-end and record evidence in the PR description
- [X] T026 SonarQube scan per CLAUDE.md "Code Quality" section; resolve any CRITICAL/BLOCKER introduced by new code under `cmd/migrations-drift-check/`
- [X] T027 Trigger CI run on the branch to confirm `migrations drift check` step is green across `.github/workflows/{test,ci}.yml` and `.gitlab-ci.yml`; record the green run URL in the PR description
- [X] T028 Cross-check spec FRs/SCs ↔ tasks coverage in PR description (FR-001…FR-012, SC-001…SC-005). Flag any unmapped requirement

### Closing the coverage gaps surfaced by /speckit-analyze

- [X] T029 [P] Run `make test-be` after all changes land and assert exit 0 with zero new failures vs. main. Attach the tail of the output to the PR description. Verifies **SC-005** (zero runtime regression)
- [X] T030 Scope-guard check: run `git diff --stat 041-sqlc-foundation..HEAD -- internal/domain/ internal/database/migrations/postgres/ internal/database/migrations/sqlite/ internal/repository/store/ internal/database/database.go` and assert it is empty. If non-empty the PR has scope creep and MUST be split. Verifies **FR-012** (no runtime / migration / repository change). Note: `internal/database/migrations/README.md` is intentionally excluded — it is the audit artifact this ticket ships

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 Setup**: no deps — start immediately
- **Phase 2 Foundational**: needs Phase 1 — blocks US1 (README artifact) and US2 (linter test scoping)
- **Phase 3 US1 (P1)**: needs Phase 2 (audit findings) — MVP candidate (doc-only)
- **Phase 4 US2 (P1)**: needs Phase 2; can be parallelized with US1 by a second engineer
- **Phase 5 US3 (P2)**: needs US1's README live (it points to it); start after T005–T007 land
- **Phase N Polish**: needs all user stories complete

### Within Each User Story

- US2 tests (T009–T014) MUST be written before US2 implementation (T015–T017); observe T012/T013 FAIL on the empty `cmd/migrations-drift-check/`, then PASS after T015–T017
- Within US2 impl: `pair.go` (T015) and `column.go` (T016) before `main.go` (T017)
- Within US1: T005 → T006 → T007 → T008 (each adds to the same file; sequential)
- Within US3: T021 then T022 (same file, distinct sections)

### Parallel Opportunities

- T009 / T010 / T011 — distinct fixture dirs, run together
- T012 / T013 / T014 — distinct test files, run together
- Polish T023 / T024 / T025 — independent

---

## Parallel Example: User Story 2

```bash
# Fixtures in parallel:
Task: "T009 testdata/ok/* fixtures"
Task: "T010 testdata/missing_pair/* fixtures"
Task: "T011 testdata/null_drift/* fixtures"

# Then tests in parallel (after fixtures land):
Task: "T012 pair_test.go"
Task: "T013 column_test.go"
Task: "T014 real_tree_test.go"

# Then sequential implementation:
Task: "T015 pair.go"  →  "T016 column.go"  →  "T017 main.go"
```

---

## Implementation Strategy

### MVP First (US1 only — doc-only)

1. Phase 1 Setup
2. Phase 2 Foundational (audit data collection)
3. Phase 3 US1 — README artifact published
4. STOP and VALIDATE: open README, reviewer checks contract compliance

Doc-only MVP ships value (single source of truth for type mapping) before any CI gate exists.

### Incremental Delivery

1. MVP US1 shipped → maintainers have the mapping table
2. Add US2 (linter + CI) → drift becomes pipeline-blocked
3. Add US3 (CLAUDE.md) → contributors discover conventions without finding the README
4. Polish → merge

### Parallel Team Strategy

- Eng A: US1 (README writing — fast, doc-only)
- Eng B (in parallel after T004): US2 (linter — bulk of new code)
- Eng A picks up US3 once US1 lands

---

## Notes

- [P] = different files, no incomplete-task dependencies
- US1 ships zero new Go code — it is a documentation deliverable
- US2 is stdlib-only Go — no new module dependency to review (constraint locked in plan)
- Type-pair allowlist is intentionally OUT of scope (Clarification Q2). Do not add it under polish "for safety"
- Reject scope creep: no migration file edit, no aggregated-schema fallback in this ticket (it stays documented as a future option in README)
- Commit per task or per logical group (Conventional Commits)
- Stop at each Checkpoint and confirm the independent test before advancing
