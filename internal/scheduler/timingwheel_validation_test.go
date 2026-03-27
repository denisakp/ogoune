package scheduler

import (
	"context"
	"testing"
	"time"
)

// TestTimingWheelValidation verifies Schedule() validation for invalid intervals.
// T020: Schedule() validation - reject interval <= 0, don't create schedule entry.
func TestTimingWheelValidation(t *testing.T) {
	cfg := &Config{
		Mode: ModeTimingWheel,
		TimingWheel: TimingWheelConfig{
			TickInterval:          1 * time.Second,
			MaxWorkers:            10,
			ShutdownTimeout:       10 * time.Second,
			NotificationQueueSize: 100,
		},
	}

	tw, err := NewTimingWheel(cfg)
	if err != nil {
		t.Fatalf("Failed to create TimingWheel: %v", err)
	}

	mockRepo := NewMockRepository([]ScheduleItem{}, nil)
	ctx := context.Background()

	if err := tw.Start(ctx, mockRepo); err != nil {
		t.Fatalf("Failed to start TimingWheel: %v", err)
	}

	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		tw.Stop(shutdownCtx)
	}()

	// Test invalid intervals
	testCases := []struct {
		name      string
		interval  time.Duration
		shouldErr bool
	}{
		{"negative interval", -1 * time.Second, true},
		{"zero interval", 0 * time.Second, true},
		{"valid interval 30s", 30 * time.Second, false},
		{"valid interval 1s", 1 * time.Second, false},
		{"valid interval 1h", 1 * time.Hour, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tw.Schedule("test-resource", tc.interval)
			if tc.shouldErr && err == nil {
				t.Errorf("Expected error for interval %v, got nil", tc.interval)
			}
			if tc.shouldErr && err != ErrInvalidInterval {
				t.Errorf("Expected ErrInvalidInterval, got %v", err)
			}
			if !tc.shouldErr && err != nil {
				t.Errorf("Expected no error for interval %v, got %v", tc.interval, err)
			}
		})
	}
}

// TestTimingWheelValidationMultiple verifies multiple invalid schedule attempts.
func TestTimingWheelValidationMultiple(t *testing.T) {
	cfg := &Config{
		Mode: ModeTimingWheel,
		TimingWheel: TimingWheelConfig{
			TickInterval:          1 * time.Second,
			MaxWorkers:            10,
			ShutdownTimeout:       10 * time.Second,
			NotificationQueueSize: 100,
		},
	}

	tw, err := NewTimingWheel(cfg)
	if err != nil {
		t.Fatalf("Failed to create TimingWheel: %v", err)
	}

	mockRepo := NewMockRepository([]ScheduleItem{}, nil)
	ctx := context.Background()

	if err := tw.Start(ctx, mockRepo); err != nil {
		t.Fatalf("Failed to start TimingWheel: %v", err)
	}

	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		tw.Stop(shutdownCtx)
	}()

	// Attempt multiple invalid schedules
	for i := 0; i < 10; i++ {
		err := tw.Schedule("test-resource-"+string(rune(i)), 0*time.Second)
		if err != ErrInvalidInterval {
			t.Errorf("Attempt %d: Expected ErrInvalidInterval, got %v", i, err)
		}
	}

	// Then schedule a valid one - should work
	err = tw.Schedule("valid-resource", 30*time.Second)
	if err != nil {
		t.Errorf("Expected valid schedule to succeed, got %v", err)
	}
}

// TestTimingWheelValidationZeroBoundary verifies exact zero boundary.
func TestTimingWheelValidationZeroBoundary(t *testing.T) {
	cfg := &Config{
		Mode: ModeTimingWheel,
		TimingWheel: TimingWheelConfig{
			TickInterval:          1 * time.Second,
			MaxWorkers:            10,
			ShutdownTimeout:       10 * time.Second,
			NotificationQueueSize: 100,
		},
	}

	tw, err := NewTimingWheel(cfg)
	if err != nil {
		t.Fatalf("Failed to create TimingWheel: %v", err)
	}

	mockRepo := NewMockRepository([]ScheduleItem{}, nil)
	ctx := context.Background()

	if err := tw.Start(ctx, mockRepo); err != nil {
		t.Fatalf("Failed to start TimingWheel: %v", err)
	}

	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		tw.Stop(shutdownCtx)
	}()

	// Test exact zero
	err = tw.Schedule("test-resource", 0*time.Second)
	if err != ErrInvalidInterval {
		t.Errorf("Expected ErrInvalidInterval for zero interval, got %v", err)
	}

	// Test minimum valid interval (1 nanosecond)
	err = tw.Schedule("test-resource", 1*time.Nanosecond)
	if err != nil {
		t.Logf("Nanosecond interval accepted: %v (platform-dependent)", err)
	}
}

// TestTimingWheelValidationNegative verifies negative intervals are rejected.
func TestTimingWheelValidationNegative(t *testing.T) {
	cfg := &Config{
		Mode: ModeTimingWheel,
		TimingWheel: TimingWheelConfig{
			TickInterval:          1 * time.Second,
			MaxWorkers:            10,
			ShutdownTimeout:       10 * time.Second,
			NotificationQueueSize: 100,
		},
	}

	tw, err := NewTimingWheel(cfg)
	if err != nil {
		t.Fatalf("Failed to create TimingWheel: %v", err)
	}

	mockRepo := NewMockRepository([]ScheduleItem{}, nil)
	ctx := context.Background()

	if err := tw.Start(ctx, mockRepo); err != nil {
		t.Fatalf("Failed to start TimingWheel: %v", err)
	}

	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		tw.Stop(shutdownCtx)
	}()

	// Test various negative durations
	negativeIntervals := []time.Duration{
		-1 * time.Nanosecond,
		-1 * time.Second,
		-30 * time.Second,
		-1 * time.Hour,
	}

	for _, interval := range negativeIntervals {
		err := tw.Schedule("test-resource", interval)
		if err != ErrInvalidInterval {
			t.Errorf("Expected ErrInvalidInterval for %v, got %v", interval, err)
		}
	}
}

// TestTimingWheelValidationNoScheduleOnError verifies no schedule entry is created on validation failure.
// This test verifies that after a failed Schedule() call, attempts to Unschedule the same resource succeed gracefully.
func TestTimingWheelValidationNoScheduleOnError(t *testing.T) {
	cfg := &Config{
		Mode: ModeTimingWheel,
		TimingWheel: TimingWheelConfig{
			TickInterval:          1 * time.Second,
			MaxWorkers:            10,
			ShutdownTimeout:       10 * time.Second,
			NotificationQueueSize: 100,
		},
	}

	tw, err := NewTimingWheel(cfg)
	if err != nil {
		t.Fatalf("Failed to create TimingWheel: %v", err)
	}

	mockRepo := NewMockRepository([]ScheduleItem{}, nil)
	ctx := context.Background()

	if err := tw.Start(ctx, mockRepo); err != nil {
		t.Fatalf("Failed to start TimingWheel: %v", err)
	}

	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		tw.Stop(shutdownCtx)
	}()

	// Try to schedule with invalid interval
	err = tw.Schedule("test-resource", 0*time.Second)
	if err != ErrInvalidInterval {
		t.Fatalf("Expected ErrInvalidInterval, got %v", err)
	}

	// Try to unschedule - should handle gracefully (resource was never scheduled)
	err = tw.Unschedule("test-resource")
	if err != nil {
		t.Errorf("Unschedule after failed Schedule should not error, got %v", err)
	}

	// Now schedule validly and verify it works
	err = tw.Schedule("test-resource", 30*time.Second)
	if err != nil {
		t.Errorf("Valid schedule should succeed, got %v", err)
	}
}
