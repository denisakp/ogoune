package database

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSQLiteSchemaSupportsRegisteredModels(t *testing.T) {
	runtime := openSQLiteTestRuntime(t)

	for _, model := range RegisteredModels {
		require.True(t, runtime.DB.Migrator().HasTable(model))
	}

	for _, tableName := range RegisteredJoinTables {
		require.True(t, runtime.DB.Migrator().HasTable(tableName))
	}

	require.True(t, runtime.DB.Migrator().HasColumn("incident_diagnostics", "request_headers"))
	require.True(t, runtime.DB.Migrator().HasColumn("notification_channels", "config"))
	require.True(t, runtime.DB.Migrator().HasColumn("users", "two_factor_backup_codes"))
	require.True(t, runtime.DB.Migrator().HasColumn("resources", "confirmation_checks"))
	require.True(t, runtime.DB.Migrator().HasColumn("resources", "confirmation_interval"))
	require.True(t, runtime.DB.Migrator().HasColumn("resources", "keyword"))
	require.True(t, runtime.DB.Migrator().HasColumn("resources", "keyword_mode"))
	require.True(t, runtime.DB.Migrator().HasColumn("incident_diagnostics", "keyword"))
	require.True(t, runtime.DB.Migrator().HasColumn("incident_diagnostics", "keyword_mode"))
	require.True(t, runtime.DB.Migrator().HasColumn("incident_diagnostics", "keyword_found"))
}
