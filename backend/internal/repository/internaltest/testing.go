package internaltest

import (
	"os"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GetTestDB returns a *gorm.DB connection for testing purposes.
// If TEST_DB_URL environment variable is not set, the test is skipped.
// This helper is used by integration tests to obtain a real database connection.
func GetTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	testDBURL := os.Getenv("TEST_DB_URL")
	if testDBURL == "" {
		t.Skip("TEST_DB_URL not set, skipping integration test")
	}

	db, err := gorm.Open(postgres.Open(testDBURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Quiet during tests
	})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	return db
}
