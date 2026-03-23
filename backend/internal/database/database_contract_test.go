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
	require.Equal(t, "pulseguard.db", resolved.DSN)
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
