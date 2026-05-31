# Phase 0 — Research

## R1. Pre-plan sqlc parse sanity (SC-001 baseline)

**Decision**: Accept the current `sqlc.yaml` (from 041) as-is. No migration changes needed for sqlc to consume the schema.

**Evidence**: At plan time, `sqlc generate -f sqlc.yaml` exits 0 and produces 20 typed Go models per dialect from `internal/database/migrations/postgres/` and `internal/database/migrations/sqlite/`. Both `pg/models.go` and `sqlite/models.go` enumerate every expected table (`Resource`, `Incident`, `User`, `ApiKey`, `ResourceCredential`, …).

**Rationale**: SC-001 is therefore already met. The ticket reduces to documentation + drift linter. No urgent migration rewrite, no aggregated-schema fallback.

**Alternatives considered**: Author a curated `internal/database/schema/{postgres,sqlite}.sql`. Rejected: unnecessary given parse success; introduces a second source of truth which violates PRD intent.

## R2. Linter parsing strategy

**Decision**: Implement a small line-oriented scanner in Go stdlib (`bufio.Scanner` + `regexp`). State machine tracks whether we are inside a `CREATE TABLE (…)` block to extract per-column lines, and a flat regex matches `ALTER TABLE … ADD COLUMN …`.

**Rationale**:
- No third-party SQL parser dependency (constraint).
- Migrations are hand-written and follow a narrow style — full SQL grammar not required.
- A regex-based extractor that fails open (logs unknown lines) and a column extractor that fails closed (exits non-zero on missing pair / nullability divergence) gives the right ergonomics.

**Alternatives considered**:
- `pganalyze/pg_query_go` (full Postgres parser). Rejected: doesn't parse SQLite dialect; adds CGO; heavy.
- Embedded sqlc engine. Rejected: not designed as a library; tightly coupled to codegen.
- `xo/usql` parser. Rejected: heavier than needed.

## R3. Column extraction rules

**Decision**: For each migration file, extract column definitions from two contexts:

1. **`CREATE TABLE name (…)`** body — split inner content by top-level commas (ignoring commas inside parentheses), filter rows that look like `name TYPE [constraints]`, capture the leading identifier and presence/absence of `NOT NULL`.
2. **`ALTER TABLE name ADD COLUMN col TYPE [constraints]`** statements — capture table, column, and nullability.

Output a normalized map: `{table → {column → nullability}}` per dialect.

**Rationale**: Covers the two ways columns enter the schema in this codebase. Indexes, foreign-key constraints, `PRIMARY KEY` clauses are intentionally skipped.

**Edge cases**:
- Multi-line column definitions: scanner accumulates until matching paren or comma at depth 0.
- Comments (`--`): line stripped from `--` to EOL before parsing.
- Block comments (`/* */`): supported by a small flag in the state machine.

## R4. Nullability default rules

**Decision**: Apply standard SQL semantics — a column is **nullable unless `NOT NULL`** is explicitly present. `PRIMARY KEY` columns are treated as `NOT NULL` (both Postgres and SQLite enforce this implicitly).

**Rationale**: Avoids false drift when one dialect declares `… PRIMARY KEY` and the other declares `… NOT NULL PRIMARY KEY` — semantically identical.

**Alternatives**: Require explicit `NOT NULL` everywhere. Rejected: invasive, contradicts no-rewrite constraint.

## R5. Type pair NOT checked (Clarification Q2)

**Decision**: The linter ignores SQL type tokens entirely. `JSONB` ↔ `TEXT`, `TIMESTAMPTZ` ↔ `TEXT`, `BIGINT` ↔ `INTEGER` are all valid cross-dialect pairs intentionally.

**Rationale**: Mechanizing type-pair validation requires an allowlist that drifts with new types; high upkeep cost, low marginal safety. The README mapping table + reviewer judgement covers it.

## R6. Audit findings format (Clarification Q3)

**Decision**: Two additions to `internal/database/migrations/README.md`:

1. **Type mapping table** (1 row per logical type):

   | Logical type | Postgres | SQLite | Rationale |
   |---|---|---|---|
   | ULID | `CHAR(26)` or `TEXT` | `TEXT` | Sortable, deterministic 26-char string |
   | Timestamp | `TIMESTAMPTZ` | `TEXT` (ISO-8601 / RFC 3339) | TZ-aware on PG; text portable on SQLite |
   | Money | `BIGINT` cents | `INTEGER` cents | Integer math; avoid float drift |
   | Boolean | `BOOLEAN` | `INTEGER` 0/1 | SQLite lacks native bool |
   | JSON | `JSONB` | `TEXT` | Indexable on PG; same Go marshalling |
   | FK id | `CHAR(26)` or `TEXT` matching parent | `TEXT` matching parent | Mirror the parent ULID column |

2. **Non-DDL statements inventory** — list any `PRAGMA`, trigger, function found across all current migration files with a per-occurrence sqlc verdict.

**Pre-plan inventory (to be confirmed by audit task)**: SQLite migrations rely on runtime-applied `PRAGMA` (`foreign_keys = ON`, `journal_mode = WAL`, `busy_timeout`) but these live in `internal/database/sqlite.go`, **not** in the `.sql` migration files. Postgres migrations contain no triggers or functions. So the inventory is expected to be empty or near-empty.

## R7. Drift-check tool location + invocation

**Decision** (Clarification Q1 lock): Go program at `cmd/migrations-drift-check`. Two invocation paths:

- `go run ./cmd/migrations-drift-check` (developer, anywhere)
- `make migrations-drift-check` (Makefile target; CI uses this)

**CI integration**: Add a step `migrations drift check` to `.github/workflows/test.yml`, `.github/workflows/ci.yml` (backend-tests job), and `.gitlab-ci.yml` (backend-tests job) — before the test run.

**Performance budget**: <1s on current tree. Achievable with stdlib + regex.

## R8. Tests for the linter

**Decision**: Golden fixture trees under `cmd/migrations-drift-check/testdata/`:

| Fixture | Expected outcome |
|---|---|
| `ok/` (matched trees, two trivial tables) | exit 0 |
| `missing_pair/` (one file only in postgres) | exit non-zero, message names the missing sqlite file |
| `null_drift/` (same column, different nullability) | exit non-zero, message names dialect + table + column |

Plus a happy-path test that runs the linter against the **real** `internal/database/migrations/` tree and asserts exit 0 (regression guard).

## R9. CLAUDE.md updates

**Decision**: Append to existing "## Patterns to follow → Database migrations" section:

- One migration = two files, same `NNNN_` prefix, same intent.
- Column name + nullability MUST match across dialects (enforced by `make migrations-drift-check`).
- JSON columns: `JSONB` (PG) / `TEXT` (SQLite). See `internal/database/migrations/README.md` for the full type-mapping table.
- `PRAGMA`, triggers, and stored functions require tech-lead validation (sqlc compatibility risk).

**Rationale**: Hooks directly into existing structure; minimal cognitive load for future contributors.

## R10. Aggregated schema fallback (`internal/database/schema/`)

**Decision**: Do NOT introduce in this ticket. Document as a documented reactive option in the README: "If a future migration introduces sqlc-incompatible DDL, introduce `internal/database/schema/{postgres,sqlite}.sql` and repoint `sqlc.yaml` `schema:` — original migrations remain authoritative at runtime."

**Rationale**: YAGNI. Today's migrations parse. Introducing the fallback now creates a second source of truth and tempts drift.
