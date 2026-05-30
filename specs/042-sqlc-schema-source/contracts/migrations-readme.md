# Contract — `internal/database/migrations/README.md`

The README is the durable, versioned audit artifact (Clarification Q3). It MUST contain the following sections, in this order:

## 1. Introduction (existing, preserved)

Conventions about migration file naming, authoritative behavior, additive-only rule.

## 2. Type mapping table (NEW)

Header row + at minimum these logical types:

| Logical type | Postgres | SQLite | Rationale |
|---|---|---|---|
| ULID | `CHAR(26)` or `TEXT` | `TEXT` | … |
| Timestamp | `TIMESTAMPTZ` | `TEXT` (ISO-8601 / RFC 3339) | … |
| Money | `BIGINT` cents | `INTEGER` cents | … |
| Boolean | `BOOLEAN` | `INTEGER` 0/1 | … |
| JSON | `JSONB` | `TEXT` | … |
| FK id | matches parent (`CHAR(26)`/`TEXT`) | matches parent (`TEXT`) | … |

Additions over time: each new logical type used in the schema MUST land here in the same PR that introduces it.

## 3. Non-DDL statements inventory (NEW)

Per-occurrence list of any non-pure-DDL statement found in `.sql` migration files (NOT runtime-applied PRAGMA in Go code), with a sqlc verdict:

| File | Line | Statement | Dialect | sqlc verdict |
|---|---|---|---|---|
| (e.g.) `sqlite/0099_x.sql` | 12 | `CREATE TRIGGER …` | SQLite | tolerated / needs fallback |

If the table is empty, leave the header and a single row `_(none — only runtime-applied SQLite PRAGMAs in Go code)_`.

## 4. Drift-check tool (NEW)

One-paragraph pointer:

> Column name + nullability across dialects are enforced by `make migrations-drift-check`. Type pairs are reviewer responsibility, guided by the table above.

## 5. Aggregated-schema fallback (NEW, reactive)

One-paragraph note documenting the documented escape hatch:

> If a future migration introduces sqlc-incompatible DDL, introduce `internal/database/schema/{postgres,sqlite}.sql` curated files and repoint `sqlc.yaml` `schema:` at them. Original migrations remain authoritative at runtime. Decision lives with tech lead.

## What this contract does NOT mandate

- The exact wording of each rationale cell.
- The order of rows in the mapping table (alphabetical or pedagogical — both fine).
- Markdown styling beyond the section structure above.
