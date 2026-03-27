package database

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSQLiteEnablesWALModeForFileBackedDatabases(t *testing.T) {
	runtime := openSQLiteTestRuntime(t)

	var journalMode string
	require.NoError(t, runtime.DB.Raw("PRAGMA journal_mode;").Scan(&journalMode).Error)
	require.Equal(t, "wal", journalMode)
}
