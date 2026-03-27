package database

import (
	"fmt"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func createSchemaValidationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:startup_notification_retry_%d?mode=memory&cache=shared", time.Now().UnixNano())), &gorm.Config{})
	require.NoError(t, err)
	return db
}

func createRequiredResourcesTable(t *testing.T, db *gorm.DB) {
	t.Helper()
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
}

func createNotificationEventsTable(t *testing.T, db *gorm.DB, includeLastError bool) {
	t.Helper()
	lastErrorCol := ""
	if includeLastError {
		lastErrorCol = ", last_error TEXT"
	}
	require.NoError(t, db.Exec(`
		CREATE TABLE notification_events (
			id TEXT PRIMARY KEY,
			incident_id TEXT NOT NULL,
			type TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			claim_owner TEXT,
			claimed_at TIMESTAMP,
			processed_at TIMESTAMP`+lastErrorCol+`
		)
	`).Error)
}

func TestValidateStartupSchema_SucceedsWithNotificationRetryColumns(t *testing.T) {
	db := createSchemaValidationTestDB(t)
	createRequiredResourcesTable(t, db)
	createNotificationEventsTable(t, db, true)

	err := ValidateStartupSchema(db)
	require.NoError(t, err)
}

func TestValidateStartupSchema_FailsWhenNotificationRetryColumnMissing(t *testing.T) {
	db := createSchemaValidationTestDB(t)
	createRequiredResourcesTable(t, db)
	createNotificationEventsTable(t, db, false)

	err := ValidateStartupSchema(db)
	require.Error(t, err)
	require.Contains(t, err.Error(), "notification_events.last_error")
	require.Contains(t, err.Error(), "run latest migrations")
}
