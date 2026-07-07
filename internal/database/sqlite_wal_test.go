package database

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSQLiteEnablesWALModeForFileBackedDatabases(t *testing.T) {
	runtime := openSQLiteTestRuntime(t)

	var journalMode string
	require.NoError(t, runtime.SQLiteDB().QueryRow("PRAGMA journal_mode;").Scan(&journalMode))
	require.Equal(t, "wal", journalMode)
}
