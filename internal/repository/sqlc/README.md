# sqlc-Backed Repositories — Walkthrough

This package houses Ogoune's sqlc-generated query code and the per-dialect
helpers that surround it. The wrapper repositories (which satisfy the
existing `port.XxxRepository` interfaces) live next to the GORM impls under
`internal/repository/store/`.

The **tags** repository was the pilot (Spec 045). Everything documented
here is grounded in that ticket, and Waves 2/3 will copy this template.

---

## Layout

```
internal/repository/sqlc/
├── queries/
│   ├── postgres/<repo>.sql           # hand-written sqlc-annotated queries
│   └── sqlite/<repo>.sql             #   (same operations, ? placeholders)
├── pg/                               # sqlc-generated, package pgsqlc
│   ├── db.go, models.go, querier.go  # generator output
│   ├── <repo>.sql.go                 # one per queries file (generated)
│   └── tx.go                         # hand-written WithTx helper
├── sqlite/                           # sqlc-generated, package sqlitesqlc
│   ├── db.go, models.go, querier.go
│   ├── <repo>.sql.go
│   └── tx.go
└── README.md                         # this file
```

Mapping helpers and the wrapper repository live in the **store** package
(`internal/repository/store/<repo>_sqlc.go`) so they stay alongside the
GORM impl they shadow.

---

> **See also**: [PATTERNS.md](./PATTERNS.md) — Wave-1 (046) cross-cutting patterns (encryption call-sites, SQL-native expressions, singleton upserts, JSON columns, mapping helper reference table).

## How to add a sqlc-backed query (worked example: tags)

1. **Write the queries** — same logical operations, one file per dialect.
   Use sqlc annotations to mark return shapes (`:one`, `:many`, `:exec`,
   `:execrows`).

   `internal/repository/sqlc/queries/postgres/tags.sql`:

   ```sql
   -- name: CreateTag :one
   INSERT INTO tags (id, created_at, updated_at, name, color, description)
   VALUES ($1, $2, $3, $4, $5, $6)
   RETURNING *;

   -- name: ListTags :many
   SELECT * FROM tags ORDER BY created_at DESC LIMIT $1 OFFSET $2;

   -- name: UpdateTag :execrows
   UPDATE tags SET name = $2, color = $3, description = $4, updated_at = $5 WHERE id = $1;
   ```

   `internal/repository/sqlc/queries/sqlite/tags.sql`: same operations with
   `?` placeholders. For variadic slices, use `sqlc.slice('ids')`:

   ```sql
   -- name: FindTagsByIDs :many
   SELECT * FROM tags WHERE id IN (sqlc.slice('ids'));
   ```

2. **Regenerate** — Sqlc parses the migration tree (`internal/database/
   migrations/{postgres,sqlite}`) as the schema source and emits Go code
   under `internal/repository/sqlc/{pg,sqlite}/`.

   ```bash
   make sqlc-generate
   make sqlc-check     # CI also runs this; fails on drift
   ```

   Two things land:

   * `internal/repository/sqlc/pg/tags.sql.go` and the SQLite mirror.
   * Updated `querier.go` in both packages (the `Querier` interface gains
     the new methods — sqlc tracks them automatically).

3. **Write the wrapper repo** — one file in
   `internal/repository/store/<repo>_sqlc.go`. The pilot's structure:

   ```go
   type TagsRepositorySQLC struct {
       pgQ     *pgsqlc.Queries     // non-nil iff postgres
       sqliteQ *sqlitesqlc.Queries // non-nil iff sqlite
   }

   func NewTagsRepositorySQLC(rt SqlcRuntime) port.TagsRepository {
       r := &TagsRepositorySQLC{}
       if pool := rt.PgxPool(); pool != nil {
           r.pgQ = pgsqlc.New(pool)
       } else if db := rt.SQLiteDB(); db != nil {
           r.sqliteQ = sqlitesqlc.New(db)
       }
       return r
   }
   ```

   `SqlcRuntime` is a tiny local interface (`PgxPool() *pgxpool.Pool` +
   `SQLiteDB() *sql.DB`). `*database.Runtime` satisfies it via duck-typing
   — this keeps `store` from importing `internal/database` and avoids a
   cycle with the database package's tests.

4. **Add the compile-time check** to `internal/repository/store/verify.go`:

   ```go
   _ port.TagsRepository = (*TagsRepositorySQLC)(nil)
   ```

5. **Wire the bootstrap flag** — one env var per repo (`SQLC_<NAME>`):

   ```go
   // internal/platform/bootstrap/database.go
   tagsRepo, tagsImpl, err := selectTagsRepo(rt, db) // returns "gorm" | "sqlc"
   slog.Info("tags repository wired", "implementation", tagsImpl)
   ```

   `selectTagsRepo` parses `SQLC_TAGS` via `strconv.ParseBool` (so `true|1|t`
   → ON; anything else OFF), and fails fast when ON but the dialect-native
   handle is nil — never silently falls back.

6. **Reuse the contract test** — the pilot reuses 044's
   `runTagsContract(t, repo)` body unchanged:

   ```go
   func TestTagsRepository_SqlcContract(t *testing.T) {
       internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
           repo := store.NewTagsRepositorySQLC(fx.Runtime)
           runTagsContract(t, repo)
       })
   }
   ```

   No fork of the contract body. Behavioral parity is mechanically
   enforced on every PR (`SQLC_TAGS=true` lane).

---

## Struct ↔ domain mapping convention

Mappers are **manual** (no codegen). One per dialect, named
`<entity>FromPG` and `<entity>FromSQLite`, both living in the same wrapper
file. They translate sqlc-generated row types into `*domain.<Entity>` —
absorbing the type differences:

| Field shape   | Postgres (pgx)        | SQLite (database/sql) | Domain                |
|---------------|----------------------|----------------------|-----------------------|
| `time.Time`   | `pgtype.Timestamptz` | `time.Time`          | `time.Time`           |
| nullable text | `pgtype.Text`        | `sql.NullString`     | `*string`             |
| `text[]`      | `[]string`           | `sqlc.slice('ids')` → `[]string` | `[]string` |
| `int4`        | `int32`              | `int64` (sqlite SQL `INTEGER`) | `int`           |

Helpers `pgTextFromPtr(*string) pgtype.Text` and `nullStringFromPtr(*string)
sql.NullString` go the other way (writes/updates). Keep them tiny and
co-located with the wrapper they serve. The tags example uses both in
`tags_repository_sqlc.go`.

---

## Transactions — when (and which) to use

| Scenario                                                                 | Use                                              |
|---------------------------------------------------------------------------|--------------------------------------------------|
| CRUD on one entity, atomic by the storage                                 | no tx                                            |
| Multi-row insert into one repo (e.g. resource + resource_tags)            | `dialect.WithTx` inside the repo                 |
| Multi-repo insert from a service (Wave 3)                                 | service starts the tx, passes `*Queries` to each repo via the Querier-injection method |
| Postgres-specific `SELECT FOR UPDATE SKIP LOCKED`                         | `pg.WithTx` only — don't try to genericize       |

There is **no shared cross-dialect transaction interface**. Each dialect
exposes its own `WithTx` next to its generated `Queries`:

```go
// internal/repository/sqlc/pg/tx.go
func WithTx(ctx context.Context, pool *pgxpool.Pool, fn func(*Queries) error) error

// internal/repository/sqlc/sqlite/tx.go
func WithTx(ctx context.Context, db *sql.DB, fn func(*Queries) error) error
```

Both rollback on error and on panic (via `defer`), commit on success.
Both surface the underlying begin / commit error wrapped with a package
prefix. Tests cover commit / rollback-on-error / rollback-on-panic /
concurrent callers — copy that test shape for new helpers you add.

### Querier injection (the wave-3 template)

When a service needs an insert into your repo to participate in *its*
transaction, expose an unexported method that takes a pre-bound `*Queries`:

```go
// Postgres path; SQLite mirror is createWithSqlQ on the same struct.
func (r *TagsRepositorySQLC) createWithQ(ctx context.Context, q *pgsqlc.Queries, t *domain.Tags) error
```

The service then does:

```go
err := pgsqlc.WithTx(ctx, pool, func(q *pgsqlc.Queries) error {
    if err := tagsRepo.createWithQ(ctx, q, &tag); err != nil { return err }
    if err := resourceRepo.createWithQ(ctx, q, &res); err != nil { return err }
    return resourceTagsRepo.linkWithQ(ctx, q, res.ID, tag.ID)
})
```

All inserts share one Postgres transaction; one error rolls everything
back. **This is the template Wave 3 will lift verbatim.**

---

## Bootstrap pointer

The pilot's flag wiring lives at:

```
internal/platform/bootstrap/database.go  →  selectTagsRepo()
```

One env var per repo (`SQLC_TAGS`, future: `SQLC_RESOURCES`, etc.). All
default OFF. A new CI lane runs the full backend suite with the flag ON
on every PR (`test-be-sqlc-tags` on GitHub, `backend-tests-sqlc-tags` on
GitLab) — this provides automated parity evidence.

---

## Anti-patterns

* **No cross-dialect tx interface.** Each helper takes its native handle
  and yields its native `*Queries`. Don't introduce a `TxRunner` interface
  unless you have a concrete reason — you'll fight pgx vs. database/sql
  semantics differences for very little ergonomic gain.

* **No auto-generated mappers.** Manual `tagFromPG` / `tagFromSQLite` are
  small enough to read and explicit enough to debug. Codegen for mappers
  is a separate decision deferred to a later ticket.

* **No silent fallback** in bootstrap. If `SQLC_<NAME>=true` and the
  dialect handle is nil, fail at boot with a message naming the dialect
  — don't let the operator think they're running on the sqlc path when
  they're not.

* **Don't edit the GORM impl** when you add the sqlc one. Both coexist
  during the migration. Production stays on GORM until the flag flips.
