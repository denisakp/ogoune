package scheduler

import (
	"testing"
	"time"
)

// TestCommunityStartupDefaultMode verifies SQLite defaults to timingwheel mode.
// T022: SQLite default-to-timingwheel regression test.
// When using SQLite (Community Edition default), the scheduler should automatically
// default to timingwheel mode without requiring explicit configuration.
func TestCommunityStartupDefaultMode(t *testing.T) {
	// Test the mode detection logic
	testCases := []struct {
		name         string
		explicitMode string
		dbDriver     string
		expectedMode ScheduleMode
		description  string
	}{
		{
			name:         "SQLite default (no explicit mode)",
			explicitMode: "",
			dbDriver:     "sqlite",
			expectedMode: ModeTimingWheel,
			description:  "Should default to timingwheel when sqlite is selected",
		},
		{
			name:         "Explicit timingwheel",
			explicitMode: "timingwheel",
			dbDriver:     "postgres",
			expectedMode: ModeTimingWheel,
			description:  "Should use explicit timingwheel mode",
		},
		{
			name:         "Explicit asynq",
			explicitMode: "asynq",
			dbDriver:     "sqlite",
			expectedMode: ModeAsynq,
			description:  "Should use explicit asynq mode regardless of driver",
		},
		{
			name:         "Postgres default",
			explicitMode: "",
			dbDriver:     "postgres",
			expectedMode: ModeAsynq,
			description:  "Should default to hosted asynq mode outside sqlite community lane",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mode := DetectMode(tc.explicitMode, tc.dbDriver)
			if mode != tc.expectedMode {
				t.Errorf("Expected mode %v, got %v - %s",
					tc.expectedMode, mode, tc.description)
			}
		})
	}
}

// TestCommunityStartupExplicitModeFallback verifies explicit mode takes precedence.
func TestCommunityStartupExplicitModeFallback(t *testing.T) {
	// Explicit mode should override auto-detection

	testCases := []struct {
		name         string
		explicitMode string
		dbDriver     string
		expectedMode ScheduleMode
	}{
		{
			name:         "Explicit timingwheel overrides hosted default",
			explicitMode: "timingwheel",
			dbDriver:     "postgres",
			expectedMode: ModeTimingWheel,
		},
		{
			name:         "Explicit asynq without Redis should be handled in factory",
			explicitMode: "asynq",
			dbDriver:     "sqlite",
			expectedMode: ModeAsynq, // Detection allows it; factory should fail
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mode := DetectMode(tc.explicitMode, tc.dbDriver)
			if mode != tc.expectedMode {
				t.Errorf("Expected mode %v, got %v", tc.expectedMode, mode)
			}
		})
	}
}

// TestCommunityStartupModeConstruction verifies New() factory respects detected mode.
func TestCommunityStartupModeConstruction(t *testing.T) {
	// Test that New() creates the correct scheduler type

	testCases := []struct {
		name        string
		mode        ScheduleMode
		shouldErr   bool
		description string
	}{
		{
			name:        "Create TimingWheel",
			mode:        ModeTimingWheel,
			shouldErr:   false,
			description: "Should successfully create TimingWheel scheduler",
		},
		{
			name:        "Create Asynq",
			mode:        ModeAsynq,
			shouldErr:   false,
			description: "Should successfully create Asynq scheduler (factory will validate Redis later)",
		},
		{
			name:        "Invalid mode",
			mode:        ScheduleMode("invalid"),
			shouldErr:   true,
			description: "Should error on invalid mode",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &Config{
				Mode: tc.mode,
				TimingWheel: TimingWheelConfig{
					TickInterval:          1 * time.Second,
					MaxWorkers:            10,
					ShutdownTimeout:       10 * time.Second,
					NotificationQueueSize: 100,
				},
			}
			if tc.mode == ModeAsynq {
				cfg.Asynq.RedisURL = "redis://localhost:6379"
			}

			scheduler, err := New(cfg)

			if tc.shouldErr && err == nil {
				t.Errorf("Expected error for mode %v, got nil", tc.mode)
			}
			if !tc.shouldErr && err != nil {
				t.Errorf("Expected no error for mode %v, got %v", tc.mode, err)
			}
			if !tc.shouldErr && scheduler == nil {
				t.Errorf("Expected scheduler instance for mode %v, got nil", tc.mode)
			}
		})
	}
}

// TestCommunityStartupDefaultConfig verifies default config is used correctly.
func TestCommunityStartupDefaultConfig(t *testing.T) {
	// When creating a scheduler without explicit config
	scheduler, err := New(nil)
	if err != nil {
		t.Fatalf("Failed to create scheduler with nil config: %v", err)
	}

	if scheduler == nil {
		t.Fatal("Expected scheduler instance, got nil")
	}

	// Verify it's a TimingWheel (should be the default)
	tw, ok := scheduler.(*TimingWheel)
	if !ok {
		t.Errorf("Expected TimingWheel scheduler, got %T", scheduler)
	}

	if tw == nil {
		t.Fatal("Expected TimingWheel instance, got nil")
	}

	t.Log("Default configuration correctly creates TimingWheel scheduler")
}

// TestCommunityStartupRegression verifies no regressions in community setup.
func TestCommunityStartupRegression(t *testing.T) {
	// Simulate typical Community Edition startup: SQLite + no Redis

	mode := DetectMode("", "sqlite")
	if mode != ModeTimingWheel {
		t.Errorf("Community Edition should default to timingwheel, got %v", mode)
	}

	// Create the scheduler
	cfg := &Config{
		Mode: mode,
		TimingWheel: TimingWheelConfig{
			TickInterval:          1 * time.Second,
			MaxWorkers:            10,
			ShutdownTimeout:       10 * time.Second,
			NotificationQueueSize: 100,
		},
	}

	scheduler, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create Community scheduler: %v", err)
	}

	if scheduler == nil {
		t.Fatal("Expected scheduler instance")
	}

	t.Log("Community Edition startup path verified: SQLite -> timingwheel")
}

// TestCommunityStartupNoRedisRequired verifies timingwheel mode doesn't require Redis.
func TestCommunityStartupNoRedisRequired(t *testing.T) {
	// The whole point of Community Edition: no Redis required for timingwheel

	// Even with Redis URL empty, timingwheel should work
	mode := DetectMode("timingwheel", "sqlite")
	if mode != ModeTimingWheel {
		t.Errorf("Timingwheel mode should be selectable without Redis")
	}

	cfg := &Config{
		Mode: mode,
		TimingWheel: TimingWheelConfig{
			TickInterval:          1 * time.Second,
			MaxWorkers:            10,
			ShutdownTimeout:       10 * time.Second,
			NotificationQueueSize: 100,
		},
	}

	scheduler, err := New(cfg)
	if err != nil {
		t.Fatalf("Timingwheel mode should not require Redis: %v", err)
	}

	if scheduler == nil {
		t.Fatal("Expected scheduler instance for timingwheel mode")
	}

	t.Log("Timingwheel mode verified as Redis-independent")
}
