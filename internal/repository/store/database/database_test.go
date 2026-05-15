package database

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/config"
)

func TestSingletonReturnsSameInstance(t *testing.T) {
	// Reset singleton state for clean test
	Reset()
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("SQLITE_PATH", filepath.Join(t.TempDir(), "singleton.db"))

	// First call should lazily initialize and return an instance.
	instance1, err1 := Instance()
	if err1 != nil {
		t.Fatalf("Unexpected error from first Instance call: %v", err1)
	}

	// Subsequent calls should return the same singleton instance.
	instance2, err2 := Instance()
	if err2 != nil {
		t.Errorf("Unexpected error from second Instance call: %v", err2)
	}

	if instance1 != instance2 {
		t.Error("Instance calls returned different pointers, expected singleton behavior")
	}
	if instance1 == nil {
		t.Error("Instance returned nil after lazy initialization")
	}
}

func TestInitWithInvalidDSNReturnsError(t *testing.T) {
	// Reset singleton state
	Reset()

	// Use clearly invalid DSN
	cfg := &config.Config{DatabaseUrl: "invalid-dsn-format"}
	err := Init(context.Background(), &cfg.DatabaseUrl)
	if err == nil {
		t.Error("Expected error with invalid DSN, got nil")
	}

	// Should contain context in error message
	if !strings.Contains(err.Error(), "db init:") {
		t.Errorf("Expected error to contain 'db init:' context, got: %v", err)
	}
}

func TestPingSkippedWithoutRealDB(t *testing.T) {
	// Reset state
	Reset()

	// If no DATABASE_URL set, skip this test
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set, skipping ping test")
	}

	// Load config from environment
	cfg := config.MustInit()

	// Initialize
	if err := Init(context.Background(), &cfg.DatabaseUrl); err != nil {
		t.Skipf("Cannot initialize database: %v", err)
	}

	// Test ping with context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := Ping(ctx)
	if err != nil {
		t.Errorf("Ping failed: %v", err)
	}
}
