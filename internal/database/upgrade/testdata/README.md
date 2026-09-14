# Upgrade-path fixtures

Each pair `<tag>.sqlite.sql` / `<tag>.postgres.sql` is a plain-text dump of a database
**written by that released version** — booted from the tag, seeded through its own API,
dumped. Not a replay of its migration files: the dump carries what the release actually
wrote, mistakes included (beta.4 recorded `.down` names in `schema_migrations`; that is
the history a real install has).

The test in `../upgrade_test.go` restores each dump into an empty database, runs this
build's migrator, and asserts that every table survives with the same rows, byte-for-byte
over the columns that existed before, and that `schema_migrations` reaches the latest
migration on disk. It is dialect-agnostic and runs on SQLite always, on Postgres when
Docker is present (`make test-be-pg`, and the `test-be-postgres` CI job).

## Adding a version

```bash
git worktree add /tmp/rel vX.Y.Z && (cd /tmp/rel && go build -o /tmp/ogoune-api-rel ./cmd/api)

# SQLite
DB_DRIVER=sqlite SQLITE_PATH=/tmp/rel.db SCHEDULER_MODE=timingwheel APP_PORT=18094 \
  AUTH_PASSWORD=fixture /tmp/ogoune-api-rel &          # then seed via the API, then stop it
sqlite3 /tmp/rel.db .dump > testdata/vX.Y.Z.sqlite.sql

# PostgreSQL (any reachable server; the dump is text)
DB_DRIVER=postgres DATABASE_URL=postgres://…/ogoune_rel … /tmp/ogoune-api-rel &   # seed, stop
pg_dump -d ogoune_rel --inserts --no-owner --no-privileges --no-comments --no-tablespaces \
  > testdata/vX.Y.Z.postgres.sql
```

Seed at least: one host, one monitor linked to it, one incident, one tag. The test
refuses a fixture with no `resources` rows, and one whose schema is already current.
Then add the tag to `fixtures` in `upgrade_test.go`.

The `users` row carries the bcrypt hash of a throwaway fixture password. Nothing here
is a secret.
