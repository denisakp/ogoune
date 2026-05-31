# Contract — `cmd/migrations-drift-check` CLI

Package: `cmd/migrations-drift-check`

## Invocation

```bash
# Default: scan internal/database/migrations/{postgres,sqlite}
go run ./cmd/migrations-drift-check

# With explicit root (used by tests)
go run ./cmd/migrations-drift-check -root /path/to/migrations
```

Makefile wrapper:

```bash
make migrations-drift-check
```

## Flags

| Flag | Default | Purpose |
|------|---------|---------|
| `-root` | `internal/database/migrations` | Directory containing `postgres/` and `sqlite/` subdirs |
| `-verbose` | `false` | Log every scanned file + extracted column count |

## Behavior

1. List `*.sql` files in `<root>/postgres/` and `<root>/sqlite/`.
2. Build `MigrationPair`s keyed by 4-digit numeric prefix.
3. For any unpaired prefix, print:
   ```
   migrations-drift-check: missing pair for prefix 0042
     postgres: internal/database/migrations/postgres/0042_foo.sql
     sqlite:   (missing)
   ```
   and exit 1.
4. Scan each file and build a `SchemaSnapshot` per dialect (table → column → `{NotNull}`).
5. For every `(table, column)` appearing in either snapshot:
   - Missing in the other → print and accumulate error.
   - Nullability differs → print and accumulate error.
6. If any errors accumulated, exit 1 with a summary line. Otherwise exit 0.

## Exit codes

| Code | Meaning |
|------|---------|
| `0` | All file-pairs present, all columns aligned on name + nullability. |
| `1` | One or more drift issues; details printed to stderr. |
| `2` | Tool error (cannot read directory, malformed regex, etc.). |

## Output format (stderr)

Per issue, one of:

```
missing pair for prefix NNNN: postgres=<path>, sqlite=(missing)
missing pair for prefix NNNN: postgres=(missing), sqlite=<path>
column-name drift: table=<t>, column=<c> present in <dialect_a> (<path>:<line>) but missing in <dialect_b>
nullability drift: table=<t>, column=<c>, postgres=<NOT_NULL|nullable> (<path>:<line>), sqlite=<NOT_NULL|nullable> (<path>:<line>)
```

Final summary line on stderr when exiting 1:

```
migrations-drift-check: <N> drift issue(s) found
```

## Out of scope (explicit non-checks)

- Type tokens (Clarification Q2).
- Indexes, foreign-key constraints, default values, check constraints.
- Migration intent / semantic equivalence beyond column name + nullability.
- Schema parser errors (sqlc itself catches those at generate time).
