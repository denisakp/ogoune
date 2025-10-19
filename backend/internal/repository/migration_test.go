package repository

import (
	"testing"

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
		"integrations",
		// "notification_events", // TODO: Add when NotificationEvent model is complete
	}

	for _, table := range tables {
		t.Run("Table_"+table+"_exists", func(t *testing.T) {
			hasTable := db.Migrator().HasTable(table)
			assert.True(t, hasTable, "Table %s should exist after migration", table)
		})
	}

	// Test that key indexes exist (using raw SQL since GORM migrator index introspection is limited)
	t.Run("Key_indexes_exist", func(t *testing.T) {
		var count int64

		// Check created_at index on resources (common query pattern)
		err := db.Raw(`
			SELECT COUNT(*) 
			FROM pg_indexes 
			WHERE tablename = 'resources' 
			AND indexname LIKE '%created_at%'
		`).Scan(&count).Error

		if err != nil {
			t.Skip("Could not verify index existence (may not be in PostgreSQL)")
		} else {
			require.Greater(t, count, int64(0), "Should have created_at index on resources")
		}
	})
}
