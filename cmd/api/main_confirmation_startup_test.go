package main

import (
	"testing"

	"github.com/denisakp/ogoune/internal/database"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestStartupSchemaValidationFailsWhenConfirmationChecksMissing(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=private"), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, db.Exec(`
		CREATE TABLE resources (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			interval INTEGER NOT NULL,
			failure_count INTEGER NOT NULL DEFAULT 0
		)
	`).Error)

	err = database.ValidateStartupSchema(db)
	require.Error(t, err)
	require.Contains(t, err.Error(), "resources.confirmation_checks")
	require.Contains(t, err.Error(), "run latest migrations")
}

func TestStartupSchemaValidationFailsWhenConfirmationIntervalMissing(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=private"), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, db.Exec(`
		CREATE TABLE resources (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			interval INTEGER NOT NULL,
			failure_count INTEGER NOT NULL DEFAULT 0,
			confirmation_checks INTEGER NOT NULL DEFAULT 2
		)
	`).Error)

	err = database.ValidateStartupSchema(db)
	require.Error(t, err)
	require.Contains(t, err.Error(), "resources.confirmation_interval")
}
