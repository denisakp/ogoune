# Quickstart — Maintainer Workflow

## Adding a new migration (post-042)

1. Pick the next prefix (e.g. `0014`).
2. Create **both** files in one commit:
   - `internal/database/migrations/postgres/0014_<name>.sql`
   - `internal/database/migrations/sqlite/0014_<name>.sql`
3. Use the dialect types from the type-mapping table in `internal/database/migrations/README.md`.
4. Keep column names and nullability **identical** across dialects.
5. Avoid `PRAGMA`, triggers, and stored functions in `.sql` files — they require tech-lead validation (sqlc compatibility).
6. Verify locally before pushing:

   ```bash
   make migrations-drift-check   # name + nullability lint
   make sqlc-check               # regen + drift gate (from 041)
   make test-be                  # full backend tests
   ```

## What the drift check enforces

- 1-to-1 file pairing between `postgres/` and `sqlite/` by `NNNN_` prefix.
- Same column names across dialects for each table.
- Same nullability (`NOT NULL` vs nullable) for each shared column.

What it does **not** enforce: SQL type tokens (deliberate — `JSONB`↔`TEXT`, `TIMESTAMPTZ`↔`TEXT` are valid by design), indexes, defaults, check constraints, foreign-key declarations, or migration intent beyond column shape.

## CI integration

Every PR runs `make migrations-drift-check` before tests. The job fails fast with a message naming the file or column at fault.

## sqlc parse-time errors

`sqlc generate` (run by `make sqlc-check` from 041) catches DDL that sqlc cannot parse. If a new migration breaks sqlc:

1. Try to express the same DDL in a sqlc-compatible form (preferred).
2. If unavoidable, introduce curated `internal/database/schema/{postgres,sqlite}.sql` and repoint `sqlc.yaml`. Original migrations stay authoritative at runtime.

## What this ticket does NOT change

- No existing migration file is rewritten.
- No domain model, repository, or runtime code is touched.
- No new env var, no new system package.
- `sqlc.yaml` and the foundation from 041 remain as-is.
