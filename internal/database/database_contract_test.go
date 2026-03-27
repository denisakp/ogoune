package database

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm/logger"
)

func TestSQLiteConfigResolveDefaults(t *testing.T) {
	resolved, err := (Config{Driver: DriverSQLite, LogLevel: "warn"}).resolve()
	require.NoError(t, err)
	require.Equal(t, DriverSQLite, resolved.Driver)
	require.Equal(t, "ogoune.db", resolved.DSN)
	require.Equal(t, logger.Warn, resolved.GormLogLevel)
}

func TestConfigResolveRejectsUnknownDriver(t *testing.T) {
	_, err := (Config{Driver: Driver("mysql")}).resolve()
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported DB_DRIVER")
}

func TestSQLiteRuntimeStartupLifecycleAppliesMigrations(t *testing.T) {
	runtime := openSQLiteTestRuntime(t)

	var migrationCount int64
	require.NoError(t, runtime.DB.Table("schema_migrations").Count(&migrationCount).Error)
	require.GreaterOrEqual(t, migrationCount, int64(3))
	require.True(t, runtime.DB.Migrator().HasTable("resources"))
	require.True(t, runtime.DB.Migrator().HasTable("incidents"))
}

// TestSQLiteMigration0004_ExpiryTables verifies that migration 0004 creates the
// expiry_notification_logs table and adds the expiry_alert_thresholds column to resources.
func TestSQLiteMigration0004_ExpiryTables(t *testing.T) {
	runtime := openSQLiteTestRuntime(t)

	require.True(t, runtime.DB.Migrator().HasTable("expiry_notification_logs"),
		"migration 0004 should create the expiry_notification_logs table")
	require.True(t, runtime.DB.Migrator().HasColumn(&struct{ ExpiryAlertThresholds *string }{}, "expiry_alert_thresholds") ||
		runtime.DB.Migrator().HasTable("resources"),
		"resources table should exist")

	// Direct column existence check via raw query
	var count int
	err := runtime.DB.Raw(
		"SELECT COUNT(*) FROM pragma_table_info('resources') WHERE name = 'expiry_alert_thresholds'",
	).Scan(&count).Error
	require.NoError(t, err)
	require.Equal(t, 1, count, "resources.expiry_alert_thresholds column must exist after migration 0004")
}
