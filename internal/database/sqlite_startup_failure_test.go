package database

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

func TestSQLiteStartupFailsOnInvalidPendingMigration(t *testing.T) {
	migrations := fstest.MapFS{
		"migrations/sqlite/0000_schema_migrations.sql": {Data: []byte("CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, name TEXT NOT NULL, applied_at DATETIME NOT NULL);")},
		"migrations/sqlite/0001_initial.sql":           {Data: []byte("CREATE TABLE IF NOT EXISTS widgets (id TEXT PRIMARY KEY);")},
		"migrations/sqlite/0002_bad.sql":               {Data: []byte("CREATE TABL broken syntax;")},
	}

	runtime, err := openSQLiteTestRuntimeWithFS(t, newSQLiteTestConfig(t), migrations)
	require.Error(t, err)
	require.Nil(t, runtime)
	require.Contains(t, err.Error(), "migration migrations/sqlite/0002_bad.sql failed")
}
