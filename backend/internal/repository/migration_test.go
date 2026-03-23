package repository

import (
	"testing"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/repository/internaltest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigration(t *testing.T) {
	db := internaltest.GetTestDB(t) // Skips if TEST_DB_URL not set

	// Test that core model tables exist
	tables := []string{
		"tags",
		"resources",
		"incidents",
		"incident_event_steps",
		// "notification_events",
	}

	for _, table := range tables {
		t.Run("Table_"+table+"_exists", func(t *testing.T) {
			hasTable := db.Migrator().HasTable(table)
			assert.True(t, hasTable, "Table %s should exist after migration", table)
		})
	}

	// Test that key indexes exist across supported drivers.
	t.Run("Key_indexes_exist", func(t *testing.T) {
		require.True(t, db.Migrator().HasIndex(&domain.Resource{}, "idx_resources_created_at"), "Should have created_at index on resources")
	})
}
