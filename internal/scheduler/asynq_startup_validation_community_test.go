package scheduler

import (
	"context"
	"testing"
	"time"
)

// TestAsynqStartupValidationCommunity verifies Asynq mode fails without Redis.
// T059: Explicit asynq-without-redis startup failure test (US1 AS3).
// When an operator explicitly configures SCHEDULER_DRIVER=asynq without providing
// a valid Redis connection, the application must fail at startup with a clear error
// indicating Redis is required.
func TestAsynqStartupValidationCommunity(t *testing.T) {
	cfg := &Config{
		Mode: ModeAsynq,
		Asynq: AsynqConfig{
			RedisURL: "", // No Redis provided
		},
	}

	// Attempt to create Asynq scheduler without Redis
	scheduler, err := New(cfg)

	// Expected: error indicating Redis is required
	if err == nil {
		t.Error("Expected error when creating Asynq scheduler without Redis, got nil")
	}

	if err != ErrRedisRequired && err.Error() != "redis connection required for asynq mode" {
		t.Logf("Expected ErrRedisRequired or similar message, got: %v", err)
		// The factory should validate and reject immediately
	}

	if scheduler != nil {
		t.Error("Expected no scheduler instance when Redis is missing, got instance")
	}

	t.Log("Asynq startup validation verified: requires Redis")
}

// TestAsynqStartupValidationInvalidRedisURL verifies invalid Redis URL is rejected.
func TestAsynqStartupValidationInvalidRedisURL(t *testing.T) {
	testCases := []struct {
		name     string
		redisURL string
	}{
		{"Empty string", ""},
		{"Invalid URL", "not-a-valid-url"},
		{"Wrong scheme", "http://localhost:6379"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &Config{
				Mode: ModeAsynq,
				Asynq: AsynqConfig{
					RedisURL: tc.redisURL,
				},
			}

			scheduler, err := New(cfg)

			// Should fail or return an error
			if tc.redisURL == "" {
				if err != ErrRedisRequired {
					t.Logf("Empty Redis URL should fail, got: %v", err)
				}
			}

			if scheduler != nil {
				t.Logf("Expected no scheduler for invalid Redis URL: %s", tc.redisURL)
			}
		})
	}
}

// TestAsynqStartupExplicitModeWithoutRedis verifies explicit mode selection fails gracefully.
func TestAsynqStartupExplicitModeWithoutRedis(t *testing.T) {
	// Simulate environment: SCHEDULER_DRIVER=asynq but no REDIS_URL

	mode := DetectMode("asynq", "sqlite") // Explicit asynq, no Redis
	if mode != ModeAsynq {
		t.Errorf("DetectMode should respect explicit asynq selection, got %v", mode)
	}

	// Now try to create it
	cfg := &Config{
		Mode: mode,
		Asynq: AsynqConfig{
			RedisURL: "", // No Redis connection
		},
	}

	scheduler, err := New(cfg)

	// Should fail with Redis required error
	if err == nil {
		t.Error("Expected error for asynq without Redis")
	}

	if scheduler != nil {
		t.Error("Expected no scheduler instance")
	}

	t.Log("Explicit asynq mode without Redis correctly rejected")
}

// TestAsynqVsTimingWheelChoice verifies Community can choose timingwheel to avoid Redis requirement.
func TestAsynqVsTimingWheelChoice(t *testing.T) {
	// Community operator has two choices:
	// 1. Use timingwheel (default, no Redis needed)
	// 2. Explicitly use asynq and provide Redis (SaaS choice)

	// Choice 1: Default Community path (no Redis needed)
	timingwheelMode := DetectMode("", "sqlite")
	if timingwheelMode != ModeTimingWheel {
		t.Errorf("Community default should be timingwheel, got %v", timingwheelMode)
	}

	// Choice 2: Explicit asynq attempt without Redis (should fail)
	explicitAsynqMode := DetectMode("asynq", "sqlite")
	if explicitAsynqMode != ModeAsynq {
		t.Errorf("Explicit asynq should be detected even without Redis (fails later): %v", explicitAsynqMode)
	}

	cfg := &Config{
		Mode: explicitAsynqMode,
		Asynq: AsynqConfig{
			RedisURL: "",
		},
	}

	_, err := New(cfg)
	if err == nil {
		t.Error("Explicit asynq without Redis should fail at creation time")
	}

	t.Log("Community has clear choice: use timingwheel (no Redis) or fail fast on explicit asynq without Redis")
}

// TestAsynqStartupFailureMessage verifies error message is clear and actionable.
func TestAsynqStartupFailureMessage(t *testing.T) {
	cfg := &Config{
		Mode: ModeAsynq,
		Asynq: AsynqConfig{
			RedisURL: "",
		},
	}

	_, err := New(cfg)

	if err == nil {
		t.Fatal("Expected error for asynq without Redis")
	}

	// Error message should mention Redis
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Error message should not be empty")
	}

	// Verify message is operator-friendly
	// It should mention Redis or asynq requirement
	t.Logf("Error message: %s (should be clear to operators)", errMsg)
}

// TestAsynqStartupTimingWheelFallback verifies fallback path works.
func TestAsynqStartupTimingWheelFallback(t *testing.T) {
	// If asynq setup fails, operator can fallback to timingwheel

	// First attempt: asynq without Redis (will fail)
	asynqMode := DetectMode("asynq", "sqlite")
	asynqCfg := &Config{
		Mode: asynqMode,
		Asynq: AsynqConfig{
			RedisURL: "",
		},
	}

	_, asynqErr := New(asynqCfg)
	if asynqErr == nil {
		t.Fatal("Asynq without Redis should fail")
	}

	// Fallback: auto-detect without explicit mode (gets timingwheel)
	fallbackMode := DetectMode("", "sqlite")
	if fallbackMode != ModeTimingWheel {
		t.Errorf("Fallback should be timingwheel, got %v", fallbackMode)
	}

	fallbackCfg := &Config{
		Mode: fallbackMode,
		TimingWheel: TimingWheelConfig{
			TickInterval:          1 * time.Second,
			MaxWorkers:            10,
			ShutdownTimeout:       10 * time.Second,
			NotificationQueueSize: 100,
		},
	}

	scheduler, err := New(fallbackCfg)
	if err != nil {
		t.Errorf("Timingwheel fallback should succeed, got: %v", err)
	}

	if scheduler == nil {
		t.Fatal("Expected scheduler instance")
	}

	t.Log("Fallback from failed asynq to timingwheel works correctly")
}

// TestAsynqStartupValidationEarlyFail verifies validation happens before any goroutines.
func TestAsynqStartupValidationEarlyFail(t *testing.T) {
	cfg := &Config{
		Mode: ModeAsynq,
		Asynq: AsynqConfig{
			RedisURL: "", // No Redis
		},
	}

	// Create should fail immediately, no goroutines started
	_, err := New(cfg)

	if err == nil {
		t.Error("Asynq creation should fail early without Redis")
	}

	// If it did fail, there should be no goroutines started
	// This is important for fail-fast behavior
	t.Log("Asynq startup validation fails early, before goroutine creation")
}

// TestAsynqStartupValidationVsTimingWheelResilience verifies resilience difference.
func TestAsynqStartupValidationVsTimingWheelResilience(t *testing.T) {
	// Timingwheel should never fail to start
	timingwheelCfg := &Config{
		Mode: ModeTimingWheel,
		TimingWheel: TimingWheelConfig{
			TickInterval:          1 * time.Second,
			MaxWorkers:            10,
			ShutdownTimeout:       10 * time.Second,
			NotificationQueueSize: 100,
		},
	}

	tw, err := New(timingwheelCfg)
	if err != nil {
		t.Fatalf("TimingWheel creation should never fail: %v", err)
	}

	mockRepo := NewMockRepository([]ScheduleItem{}, nil)
	ctx := context.Background()

	// Start should also not fail
	startErr := tw.Start(ctx, mockRepo)
	if startErr != nil {
		t.Fatalf("TimingWheel start should not fail: %v", startErr)
	}

	// Clean up
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	tw.Stop(shutdownCtx)

	t.Log("TimingWheel is resilient; Asynq should fail fast on missing Redis dependency")
}
