# Database Migrations

- Migration files are authoritative for schema changes. Runtime does not use GORM `AutoMigrate`.
- Each driver keeps its own ordered SQL files under `postgres/` and `sqlite/`.
- Filenames use zero-padded numeric versions: `0000_schema_migrations.sql`, `0001_initial.sql`, `0002_indexes.sql`.
- New migrations must be additive, deterministic, and safe to re-run through version tracking.
- Keep PostgreSQL and SQLite schemas semantically aligned for shared domain behavior.

## File-pair rule

One migration = two files with the **same `NNNN_` prefix** and the **same intent**, one under `postgres/` and one under `sqlite/`. Column names and nullability MUST match across dialects; SQL type tokens are intentionally allowed to differ (see mapping below). The pair rule is enforced by `make migrations-drift-check`.

## Type mapping

When choosing the SQL type for a new column, pick from this table. Column **name** and **nullability** are identical across dialects; the SQL **type** differs per dialect by design.

| Logical type | Postgres | SQLite | Rationale |
|---|---|---|---|
| ULID | `CHAR(26)` or `TEXT` | `TEXT` | Sortable, deterministic 26-char string identifier (`oklog/ulid/v2`). |
| Timestamp | `TIMESTAMPTZ` | `TEXT` (ISO-8601 / RFC 3339) | Timezone-aware on Postgres; portable ISO text on SQLite. Go marshals both to `time.Time`. |
| Money | `BIGINT` cents | `INTEGER` cents | Integer math; avoid floating-point drift. Store smallest currency unit. |
| Boolean | `BOOLEAN` | `INTEGER` 0/1 | SQLite lacks a native boolean — store 0/1 and let Go convert. |
| JSON | `JSONB` | `TEXT` | Indexable + queryable on Postgres; same Go marshalling via `json.Marshal`/`json.Unmarshal` on both sides. Use `GORM serializer:json`, not `type:jsonb` directly, in the domain layer. |
| FK id | matches parent (`CHAR(26)` / `TEXT`) | matches parent (`TEXT`) | Mirror the parent ULID column's dialect type exactly. |

If a new logical type appears in the schema, add a row in the same PR.

## Non-DDL statements inventory

Any non-pure-DDL statement (`PRAGMA`, `CREATE TRIGGER`, `CREATE FUNCTION`, `DO $$ … $$`) found inside the `.sql` migration files. Statements applied at runtime from Go code (e.g. SQLite `PRAGMA foreign_keys = ON` in `internal/database/sqlite.go`) are NOT listed here — only DDL embedded in migration files.

| File | Line | Statement | Dialect | sqlc verdict |
|---|---|---|---|---|
| _(none — only runtime-applied SQLite PRAGMAs live in `internal/database/sqlite.go`; no triggers or functions in any migration file)_ | | | | |

When adding a new non-DDL statement to a migration, append a row here and confirm `make sqlc-check` still passes.

## Drift-check tool

Column **name** and **nullability** across dialects are enforced by `make migrations-drift-check` (Go program at `cmd/migrations-drift-check`). It also enforces 1-to-1 file pairing by `NNNN_` prefix. Run locally before pushing; CI runs the same target.

The linter intentionally does NOT check SQL type tokens — cross-dialect type pairs (`JSONB` ↔ `TEXT`, `TIMESTAMPTZ` ↔ `TEXT`, `BIGINT` ↔ `INTEGER`) are by design. Type correctness is reviewer responsibility, guided by the mapping table above.

## Aggregated-schema fallback

`sqlc` consumes the migration trees directly via `sqlc.yaml`. If a future migration introduces sqlc-incompatible DDL that cannot be expressed compatibly, the documented escape hatch is to introduce curated `internal/database/schema/{postgres,sqlite}.sql` files and repoint `sqlc.yaml` `schema:` at them. Original migrations remain authoritative at runtime. This fallback is opt-in per dialect and requires tech-lead approval. As of 042, it has not been needed.
