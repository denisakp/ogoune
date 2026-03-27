package database

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/denisakp/pulseguard/internal/config"
)

func TestSingletonReturnsSameInstance(t *testing.T) {
	// Reset singleton state for clean test
	Reset()

	// First call to Instance should return error since Init not called
	instance1, err1 := Instance()
	if err1 == nil {
		t.Error("Expected error when Instance called before Init, got nil")
	}
	if instance1 != nil {
		t.Error("Expected nil instance when Instance called before Init")
	}

	// Initialize with a valid but non-connecting DSN for testing singleton logic
	dsn := "postgres://test:test@localhost:5432/testdb?sslmode=disable"
	if err := Init(context.Background(), &dsn); err != nil {
		t.Skipf("Skipping test due to DB connection error: %v", err)
	}

	// Now Instance calls should return same pointer
	instance2, err2 := Instance()
	if err2 != nil {
		t.Errorf("Unexpected error from Instance after Init: %v", err2)
	}

	instance3, err3 := Instance()
	if err3 != nil {
		t.Errorf("Unexpected error from second Instance call: %v", err3)
	}

	// Should be same pointer
	if instance2 != instance3 {
		t.Error("Instance calls returned different pointers, expected singleton behavior")
	}
	if instance2 == nil {
		t.Error("Instance returned nil after successful Init")
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
