package internaltest

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	shareddb "github.com/denisakp/pulseguard/internal/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GetTestDB returns a *gorm.DB connection for testing purposes.
// It prefers TEST_DB_URL when available and otherwise provisions a temporary SQLite database.
func GetTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	testDBURL := os.Getenv("TEST_DB_URL")
	if testDBURL == "" {
		runtime, err := shareddb.Open(context.Background(), shareddb.Config{
			Driver:     shareddb.DriverSQLite,
			SQLitePath: filepath.Join(t.TempDir(), "pulseguard-test.db"),
			LogLevel:   "silent",
		})
		if err != nil {
			t.Fatalf("failed to initialize sqlite test database: %v", err)
		}

		sqlDB, err := runtime.DB.DB()
		if err == nil {
			t.Cleanup(func() {
				_ = sqlDB.Close()
			})
		}

		return runtime.DB
	}

	db, err := gorm.Open(postgres.Open(testDBURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Quiet during tests
	})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	sqlDB, err := db.DB()
	if err == nil {
		t.Cleanup(func() {
			_ = sqlDB.Close()
		})
	}

	return db
}
