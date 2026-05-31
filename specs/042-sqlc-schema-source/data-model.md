# Data Model

This feature does not persist any new entity. It introduces internal data structures inside the `cmd/migrations-drift-check` linter.

## Internal types (Go)

### `MigrationFile`

| Field | Type | Notes |
|-------|------|-------|
| `Prefix` | `string` | Four-digit numeric prefix (e.g. `"0013"`) |
| `Name` | `string` | Slug after the prefix (e.g. `"resource_credentials"`) |
| `Path` | `string` | Absolute path on disk |
| `Dialect` | `string` | `"postgres"` or `"sqlite"` |

### `MigrationPair`

| Field | Type | Notes |
|-------|------|-------|
| `Prefix` | `string` | Shared numeric prefix |
| `Postgres` | `*MigrationFile` | Nil → missing pair |
| `SQLite` | `*MigrationFile` | Nil → missing pair |

**Validation**: pair is invalid iff `Postgres == nil || SQLite == nil`.

### `ColumnDef`

| Field | Type | Notes |
|-------|------|-------|
| `Table` | `string` | Lowercased table name |
| `Column` | `string` | Lowercased column name |
| `NotNull` | `bool` | True if `NOT NULL` is present or column is `PRIMARY KEY` (R4) |
| `SourceFile` | `string` | The file the definition was found in (for error messages) |
| `Line` | `int` | Line number (for error messages) |

### `SchemaSnapshot`

```go
type SchemaSnapshot map[string]map[string]ColumnDef // table → column → def
```

Built once per dialect by scanning every `.sql` file in order.

## Validation rules (the linter)

The linter exits non-zero when **any** of the following hold:

1. **File-pair drift**: a numeric prefix exists in one dialect tree but not the other.
2. **Column-name drift**: a column declared in dialect A's snapshot is absent from dialect B's snapshot for the same table (after lowercase normalization).
3. **Nullability drift**: a column is present in both snapshots for the same table but `NotNull` differs.

Type tokens are intentionally **not** validated (Clarification Q2 / Research R5).

## Out-of-scope (no data changes)

- Domain models (`internal/domain/`)
- Runtime schema validation (`internal/database/database.go::ValidateStartupSchema`)
- Repository code
- Generated sqlc code

These are not touched and remain authoritative for application behavior.
