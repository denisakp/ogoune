package database

import (
	"context"
	"os"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// T0M1: startup applies migrations and initializes API key schema objects.
func TestStartupAppliesAPIKeyMigrations(t *testing.T) {
	cfg := newSQLiteTestConfig(t)
	runtime, err := Open(context.Background(), cfg)
	require.NoError(t, err)
	require.NotNil(t, runtime)

	migrator := runtime.DB.Migrator()
	require.True(t, migrator.HasTable("api_keys"))
	require.True(t, migrator.HasTable("schema_migrations"))
}

func TestValidateStartupSchemaFailsWhenNotificationRetryColumnMissing(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:startup_missing_notification_last_error?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, db.Exec(`
		CREATE TABLE resources (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			interval INTEGER NOT NULL,
			failure_count INTEGER NOT NULL DEFAULT 0,
			confirmation_checks INTEGER NOT NULL DEFAULT 2,
			confirmation_interval INTEGER NOT NULL DEFAULT 30
		)
	`).Error)

	require.NoError(t, db.Exec(`
		CREATE TABLE notification_events (
			id TEXT PRIMARY KEY,
			incident_id TEXT NOT NULL,
			type TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			claim_owner TEXT,
			claimed_at TIMESTAMP,
			processed_at TIMESTAMP
		)
	`).Error)

	err = ValidateStartupSchema(db)
	require.Error(t, err)
	require.Contains(t, err.Error(), "notification_events.last_error")
}

// T030: Verify migration 0012 adds protocol_type and protocol_port columns to resources table.
func TestMigration0012AddsProtocolColumns(t *testing.T) {
	runtime := openSQLiteTestRuntime(t)

	type colInfo struct {
		Name string `gorm:"column:name"`
	}
	var cols []colInfo
	require.NoError(t, runtime.DB.Raw("PRAGMA table_info(resources)").Scan(&cols).Error)

	found := map[string]bool{}
	for _, c := range cols {
		found[c.Name] = true
	}

	require.True(t, found["protocol_type"], "protocol_type column must exist after migration 0012")
	require.True(t, found["protocol_port"], "protocol_port column must exist after migration 0012")
}

// T031: Verify migration 0012 adds protocol_type and protocol_port columns on PostgreSQL.
func TestMigration0012AddsProtocolColumnsPostgres(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set; skipping PostgreSQL migration test")
	}
	runtime, err := Open(context.Background(), Config{
		Driver:      DriverPostgres,
		DatabaseURL: dsn,
		LogLevel:    "silent",
	})
	require.NoError(t, err)
	require.NotNil(t, runtime)

	type colInfo struct {
		ColumnName string `gorm:"column:column_name"`
	}
	var cols []colInfo
	require.NoError(t, runtime.DB.Raw(`
		SELECT column_name FROM information_schema.columns
		WHERE table_name = 'resources'
	`).Scan(&cols).Error)

	found := map[string]bool{}
	for _, c := range cols {
		found[c.ColumnName] = true
	}

	require.True(t, found["protocol_type"], "protocol_type column must exist after migration 0012")
	require.True(t, found["protocol_port"], "protocol_port column must exist after migration 0012")
}
