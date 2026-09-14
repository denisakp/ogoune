// Package upgrade holds the upgrade-path gate: a database written by an older
// released Ogoune is restored from a text dump, this build's migrator is run
// against it, and every pre-existing row must come out exactly as it went in.
//
// This is the test that would have caught the beta.6 migrator defects (down
// files run forward, a duplicated version skipped on upgrade installs) before
// a release rather than after, and it is what a "stable" label leans on: an
// operator's data survives every upgrade, on both dialects, mechanically.
//
// Fixtures under testdata/ are plain-text dumps produced by BOOTING the tagged
// release (not by replaying its migration files): they carry whatever that
// release actually wrote -- including its mistakes, such as the ".down" names
// beta.4 recorded in schema_migrations -- which is the whole point.
package upgrade_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/database"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
)

// fixtures lists the released versions an upgrade is proven from. Add a
// version by booting that tag against a fresh database, seeding it through
// its own API, and dumping it as text (sqlite3 .dump / pg_dump --inserts).
var fixtures = []string{"v1.0.0-beta.4"}

// tableSnapshot is every row of a table, rendered as strings over a fixed
// column list, sorted. Comparing two of them over the PRE-upgrade column list
// is the "no row rewritten" assertion; new columns are invisible to it by
// construction.
type tableSnapshot struct {
	columns []string
	rows    []string
}

func snapshot(t *testing.T, db *sql.DB, driver database.Driver) map[string]tableSnapshot {
	t.Helper()
	out := map[string]tableSnapshot{}
	for _, table := range userTables(t, db, driver) {
		cols := columnsOf(t, db, driver, table)
		out[table] = tableSnapshot{columns: cols, rows: readRows(t, db, driver, table, cols)}
	}
	return out
}

func userTables(t *testing.T, db *sql.DB, driver database.Driver) []string {
	t.Helper()
	var q string
	switch driver {
	case database.DriverSQLite:
		q = `SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`
	default:
		q = `SELECT table_name FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE' ORDER BY table_name`
	}
	rows, err := db.Query(q)
	require.NoError(t, err)
	defer rows.Close()
	var names []string
	for rows.Next() {
		var n string
		require.NoError(t, rows.Scan(&n))
		names = append(names, n)
	}
	require.NoError(t, rows.Err())
	return names
}

func columnsOf(t *testing.T, db *sql.DB, driver database.Driver, table string) []string {
	t.Helper()
	var cols []string
	switch driver {
	case database.DriverSQLite:
		rows, err := db.Query(fmt.Sprintf(`PRAGMA table_info(%q)`, table))
		require.NoError(t, err)
		defer rows.Close()
		for rows.Next() {
			var cid int
			var name, typ string
			var notnull, pk int
			var dflt sql.NullString
			require.NoError(t, rows.Scan(&cid, &name, &typ, &notnull, &dflt, &pk))
			cols = append(cols, name)
		}
		require.NoError(t, rows.Err())
	default:
		rows, err := db.Query(`SELECT column_name FROM information_schema.columns WHERE table_schema='public' AND table_name=$1 ORDER BY ordinal_position`, table)
		require.NoError(t, err)
		defer rows.Close()
		for rows.Next() {
			var n string
			require.NoError(t, rows.Scan(&n))
			cols = append(cols, n)
		}
		require.NoError(t, rows.Err())
	}
	return cols
}

func quoteIdent(driver database.Driver, s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

func readRows(t *testing.T, db *sql.DB, driver database.Driver, table string, cols []string) []string {
	t.Helper()
	if len(cols) == 0 {
		return nil
	}
	quoted := make([]string, len(cols))
	for i, c := range cols {
		// Cast to text on both dialects so the comparison is on the stored
		// value, not on how each driver decodes a type.
		quoted[i] = "CAST(" + quoteIdent(driver, c) + " AS TEXT)"
	}
	rows, err := db.Query(fmt.Sprintf(`SELECT %s FROM %s`, strings.Join(quoted, ", "), quoteIdent(driver, table)))
	require.NoError(t, err, table)
	defer rows.Close()
	var out []string
	for rows.Next() {
		vals := make([]sql.NullString, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		require.NoError(t, rows.Scan(ptrs...), table)
		parts := make([]string, len(cols))
		for i, v := range vals {
			if !v.Valid {
				parts[i] = "<NULL>"
			} else {
				parts[i] = v.String
			}
		}
		out = append(out, strings.Join(parts, "\x1f"))
	}
	require.NoError(t, rows.Err(), table)
	sort.Strings(out)
	return out
}

// latestMigrationVersion is the highest version on disk for the dialect --
// what schema_migrations must reach after the upgrade.
func latestMigrationVersion(t *testing.T, dialect string) string {
	t.Helper()
	entries, err := filepath.Glob(filepath.Join("..", "migrations", dialect, "*.sql"))
	require.NoError(t, err)
	require.NotEmpty(t, entries)
	re := regexp.MustCompile(`^(\d{4})_`)
	latest := ""
	for _, e := range entries {
		m := re.FindStringSubmatch(filepath.Base(e))
		if m != nil && m[1] > latest {
			latest = m[1]
		}
	}
	return latest
}

func schemaVersions(t *testing.T, db *sql.DB) (count int, max string) {
	t.Helper()
	require.NoError(t, db.QueryRow(`SELECT COUNT(*), MAX(version) FROM schema_migrations`).Scan(&count, &max))
	return
}

// assertUpgrade is the gate itself, dialect-agnostic: given the snapshot taken
// before the migrator ran and a connection to the upgraded database, every
// table that existed still exists with the same number of rows, and every row
// reads identically over the columns that existed before. schema_migrations
// alone may grow, and must reach the latest version on disk.
func assertUpgrade(t *testing.T, driver database.Driver, dialect string, before map[string]tableSnapshot, beforeCount int, db *sql.DB) {
	t.Helper()
	after := snapshot(t, db, driver)

	for table, was := range before {
		now, ok := after[table]
		require.Truef(t, ok, "table %s disappeared during the upgrade", table)
		if table == "schema_migrations" {
			continue
		}
		for _, c := range was.columns {
			assert.Containsf(t, now.columns, c, "table %s lost column %s", table, c)
		}
		assert.Equalf(t, len(was.rows), len(now.rows), "table %s: row count changed", table)
		// Re-read over the OLD column list only, so added columns do not
		// disturb the comparison and a rewritten value in an old column does.
		reread := readRows(t, db, driver, table, was.columns)
		assert.Equalf(t, was.rows, reread, "table %s: a pre-existing row was rewritten", table)
	}

	count, max := schemaVersions(t, db)
	assert.Greater(t, count, beforeCount, "the migrator applied nothing")
	assert.Equal(t, latestMigrationVersion(t, dialect), max, "schema_migrations did not reach the latest migration on disk")
}

// --- SQLite ------------------------------------------------------------------

func restoreSQLite(t *testing.T, dumpPath string) string {
	t.Helper()
	dump, err := os.ReadFile(dumpPath)
	require.NoError(t, err)
	path := filepath.Join(t.TempDir(), "upgrade.db")
	db, err := sql.Open("sqlite", path)
	require.NoError(t, err)
	_, err = db.Exec(string(dump))
	require.NoError(t, err, "restore %s", dumpPath)
	require.NoError(t, db.Close())
	return path
}

func TestUpgrade_SQLite(t *testing.T) {
	for _, version := range fixtures {
		t.Run(version, func(t *testing.T) {
			path := restoreSQLite(t, filepath.Join("testdata", version+".sqlite.sql"))

			pre, err := sql.Open("sqlite", path)
			require.NoError(t, err)
			before := snapshot(t, pre, database.DriverSQLite)
			beforeCount, beforeMax := schemaVersions(t, pre)
			require.NoError(t, pre.Close())
			require.NotEmpty(t, before["resources"].rows, "fixture must carry data, or the gate proves nothing")
			require.Less(t, beforeMax, latestMigrationVersion(t, "sqlite"), "fixture is already current; nothing to upgrade")

			rt, err := database.Open(context.Background(), database.Config{
				Driver: database.DriverSQLite, SQLitePath: path, LogLevel: "silent",
			})
			require.NoError(t, err, "this build refused to start on a %s database", version)
			t.Cleanup(func() { _ = rt.SQLiteDB().Close() })

			assertUpgrade(t, database.DriverSQLite, "sqlite", before, beforeCount, rt.SQLiteDB())
		})
	}
}

// --- PostgreSQL --------------------------------------------------------------

// restorePostgres feeds a pg_dump text dump to the database. Two kinds of
// line are not the schema and are dropped: psql meta-commands (pg_dump 17+
// emits a leading `\restrict`), and the session `SET` preamble, which names
// parameters of the server that produced the dump (`transaction_timeout` is
// PG 17+) rather than the one restoring it. Every statement in the dump is
// schema-qualified, so the search_path line is not needed either.
func restorePostgres(t *testing.T, dsn, dumpPath string) {
	t.Helper()
	raw, err := os.ReadFile(dumpPath)
	require.NoError(t, err)
	var sqlLines []string
	for _, line := range strings.Split(string(raw), "\n") {
		if strings.HasPrefix(line, `\`) || strings.HasPrefix(line, "SET ") {
			continue
		}
		sqlLines = append(sqlLines, line)
	}
	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	defer db.Close()
	_, err = db.Exec(strings.Join(sqlLines, "\n"))
	require.NoError(t, err, "restore %s", dumpPath)
}

func TestUpgrade_Postgres(t *testing.T) {
	for _, version := range fixtures {
		t.Run(version, func(t *testing.T) {
			dsn := internaltest.EmptyPostgresDSN(t)
			restorePostgres(t, dsn, filepath.Join("testdata", version+".postgres.sql"))

			pre, err := sql.Open("pgx", dsn)
			require.NoError(t, err)
			before := snapshot(t, pre, database.DriverPostgres)
			beforeCount, beforeMax := schemaVersions(t, pre)
			require.NoError(t, pre.Close())
			require.NotEmpty(t, before["resources"].rows, "fixture must carry data, or the gate proves nothing")
			require.Less(t, beforeMax, latestMigrationVersion(t, "postgres"), "fixture is already current; nothing to upgrade")

			rt, err := database.Open(context.Background(), database.Config{
				Driver: database.DriverPostgres, DatabaseURL: dsn, LogLevel: "silent",
			})
			require.NoError(t, err, "this build refused to start on a %s database", version)
			t.Cleanup(func() { rt.PgxPool().Close() })

			post, err := sql.Open("pgx", dsn)
			require.NoError(t, err)
			defer post.Close()
			assertUpgrade(t, database.DriverPostgres, "postgres", before, beforeCount, post)
		})
	}
}
