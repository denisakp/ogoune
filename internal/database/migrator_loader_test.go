package database

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// migrateInto runs a synthetic migration tree against a throwaway SQLite file,
// bypassing Open's post-migration schema assertions -- those check the real
// application schema, which a three-table fixture deliberately does not have.
func migrateInto(t *testing.T, path string, migrations fstest.MapFS) (*sql.DB, error) {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return db, runMigrations(context.Background(), db, DriverSQLite, migrations)
}

func tempDBPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "migrator-test.db")
}

const bootstrapSQL = "CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, name TEXT NOT NULL, applied_at DATETIME NOT NULL);"

// Undo scripts are not migrations, and the loader used to run them forward.
//
// A same-version pair cannot demonstrate this: ".down.sql" sorts before
// ".up.sql", so the drop lands on a table that does not exist yet and the up
// recreates it — the old loader passes that test too. The damage shows when a
// later migration's down file drops something an earlier one created, which is
// destructive whatever the ordering.
//
// That is what this builds: 0002's undo script drops the table 0001 created.
func TestMigrator_NeverExecutesDownFiles(t *testing.T) {
	migrations := fstest.MapFS{
		"migrations/sqlite/0000_schema_migrations.sql": {Data: []byte(bootstrapSQL)},
		"migrations/sqlite/0001_widgets.up.sql":        {Data: []byte("CREATE TABLE IF NOT EXISTS widgets (id TEXT PRIMARY KEY);")},
		"migrations/sqlite/0001_widgets.down.sql":      {Data: []byte("DROP TABLE IF EXISTS widgets;")},
		"migrations/sqlite/0002_gadgets.up.sql":        {Data: []byte("CREATE TABLE IF NOT EXISTS gadgets (id TEXT PRIMARY KEY);")},
		"migrations/sqlite/0002_gadgets.down.sql":      {Data: []byte("DROP TABLE IF EXISTS gadgets; DROP TABLE IF EXISTS widgets;")},
	}

	db, err := migrateInto(t, tempDBPath(t), migrations)
	require.NoError(t, err)

	for _, table := range []string{"widgets", "gadgets"} {
		_, err := db.Exec("SELECT 1 FROM " + table + " LIMIT 0")
		require.NoErrorf(t, err, "%s was created by an up file and no undo script may take it away", table)
	}
}

// schema_migrations must say what was applied. It used to record the down file's
// name for 19 of 34 rows on a fresh database, because that file ran first and
// won the INSERT OR IGNORE.
func TestMigrator_RecordsTheUpFileName(t *testing.T) {
	migrations := fstest.MapFS{
		"migrations/sqlite/0000_schema_migrations.sql": {Data: []byte(bootstrapSQL)},
		"migrations/sqlite/0001_widgets.up.sql":        {Data: []byte("CREATE TABLE IF NOT EXISTS widgets (id TEXT PRIMARY KEY);")},
		"migrations/sqlite/0001_widgets.down.sql":      {Data: []byte("DROP TABLE IF EXISTS widgets;")},
	}

	db, err := migrateInto(t, tempDBPath(t), migrations)
	require.NoError(t, err)

	var name string
	require.NoError(t, db.
		QueryRow("SELECT name FROM schema_migrations WHERE version = '0001'").Scan(&name))
	assert.Equal(t, "widgets.up", name, "a record naming an undo script describes work that never happened")
}

// Applied state is keyed on the version alone, so two migrations sharing a
// prefix are indistinguishable: on a database that already recorded that
// version, the second is skipped forever — silently, and only on installs that
// upgrade rather than start fresh. Refusing to start is the honest answer, and
// the author gets it instead of an operator.
func TestMigrator_RefusesDuplicateVersions(t *testing.T) {
	migrations := fstest.MapFS{
		"migrations/sqlite/0000_schema_migrations.sql": {Data: []byte(bootstrapSQL)},
		"migrations/sqlite/0001_widgets.sql":           {Data: []byte("CREATE TABLE IF NOT EXISTS widgets (id TEXT PRIMARY KEY);")},
		"migrations/sqlite/0001_gadgets.sql":           {Data: []byte("CREATE TABLE IF NOT EXISTS gadgets (id TEXT PRIMARY KEY);")},
	}

	_, err := migrateInto(t, tempDBPath(t), migrations)
	require.Error(t, err, "startup fails before serving traffic")
	assert.Contains(t, err.Error(), "share version \"0001\"")
	assert.Contains(t, err.Error(), "every migration needs its own number",
		"the error has to say what the author should do about it")
}

// A down file must not be mistaken for a second migration at the same version.
func TestMigrator_UpAndDownAreNotADuplicate(t *testing.T) {
	migrations := fstest.MapFS{
		"migrations/sqlite/0000_schema_migrations.sql": {Data: []byte(bootstrapSQL)},
		"migrations/sqlite/0001_widgets.up.sql":        {Data: []byte("CREATE TABLE IF NOT EXISTS widgets (id TEXT PRIMARY KEY);")},
		"migrations/sqlite/0001_widgets.down.sql":      {Data: []byte("DROP TABLE IF EXISTS widgets;")},
	}

	_, err := migrateInto(t, tempDBPath(t), migrations)
	require.NoError(t, err)
}

// The upgrade path: a version already recorded is not run again, and a new one
// is. This is what makes renumbering a migration safe when its statements are
// IF NOT EXISTS.
func TestMigrator_AppliesOnlyWhatIsPending(t *testing.T) {
	path := tempDBPath(t)
	first := fstest.MapFS{
		"migrations/sqlite/0000_schema_migrations.sql": {Data: []byte(bootstrapSQL)},
		"migrations/sqlite/0001_widgets.up.sql":        {Data: []byte("CREATE TABLE IF NOT EXISTS widgets (id TEXT PRIMARY KEY);")},
	}
	_, err := migrateInto(t, path, first)
	require.NoError(t, err)

	// Same tree plus one migration, against the same database file.
	second := fstest.MapFS{
		"migrations/sqlite/0000_schema_migrations.sql": {Data: []byte(bootstrapSQL)},
		"migrations/sqlite/0001_widgets.up.sql":        {Data: []byte("CREATE TABLE IF NOT EXISTS widgets (id TEXT PRIMARY KEY);")},
		"migrations/sqlite/0002_gadgets.up.sql":        {Data: []byte("CREATE TABLE IF NOT EXISTS gadgets (id TEXT PRIMARY KEY);")},
	}
	db2, err := migrateInto(t, path, second)
	require.NoError(t, err)

	var versions int
	require.NoError(t, db2.
		QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&versions))
	assert.Equal(t, 3, versions, "bootstrap, the original, and the new one")

	for _, table := range []string{"widgets", "gadgets"} {
		_, err := db2.Exec("SELECT 1 FROM " + table + " LIMIT 0")
		require.NoErrorf(t, err, "expected %s to exist", table)
	}
}
