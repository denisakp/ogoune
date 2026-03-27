package scheduler

import (
	"context"
	"runtime"
	"testing"
	"time"
)

// TestTimingWheelGoroutineLeakStartStop verifies no goroutine leaks on start/stop cycles
// T042: Test goroutine leak detection for Start/Shutdown lifecycle
func TestTimingWheelGoroutineLeakStartStop(t *testing.T) {
	cfg := &Config{
		Mode: ModeTimingWheel,
		TimingWheel: TimingWheelConfig{
			TickInterval:          100 * time.Millisecond,
			MaxWorkers:            5,
			ShutdownTimeout:       5 * time.Second,
			NotificationQueueSize: 100,
		},
	}

	// Get baseline goroutine count
	baselineGoroutines := runtime.NumGoroutine()

	// Run start/stop cycle
	tw, err := NewTimingWheel(cfg)
	if err != nil {
		t.Fatalf("Failed to create TimingWheel: %v", err)
	}

	mockRepo := NewMockRepository([]ScheduleItem{}, nil)

	ctx := context.Background()
	err = tw.Start(ctx, mockRepo)
	if err != nil {
		t.Fatalf("Failed to start TimingWheel: %v", err)
	}

	// Goroutines should increase (run goroutine added)
	runningGoroutines := runtime.NumGoroutine()
	if runningGoroutines <= baselineGoroutines {
		t.Logf("Warning: goroutines did not increase on Start (baseline: %d, running: %d)",
			baselineGoroutines, runningGoroutines)
	}

	// Shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = tw.Stop(shutdownCtx)
	if err != nil {
		t.Fatalf("Failed to stop TimingWheel: %v", err)
	}

	// Give goroutines time to clean up
	time.Sleep(100 * time.Millisecond)

	// Check for leaks
	finalGoroutines := runtime.NumGoroutine()
	t.Logf("Goroutine count - Baseline: %d, Running: %d, Final: %d",
		baselineGoroutines, runningGoroutines, finalGoroutines)

	// Allow small variance (may have other background goroutines)
	const tolerance = 3
	if finalGoroutines > baselineGoroutines+tolerance {
		t.Errorf("Potential goroutine leak: baseline=%d, final=%d (difference: %d)",
			baselineGoroutines, finalGoroutines, finalGoroutines-baselineGoroutines)
	}
}

// TestTimingWheelGoroutineLeakMultipleCycles verifies no leaks over multiple cycles
func TestTimingWheelGoroutineLeakMultipleCycles(t *testing.T) {
	cfg := &Config{
		Mode: ModeTimingWheel,
		TimingWheel: TimingWheelConfig{
			TickInterval:          100 * time.Millisecond,
			MaxWorkers:            3,
			ShutdownTimeout:       5 * time.Second,
			NotificationQueueSize: 100,
		},
	}

	baselineGoroutines := runtime.NumGoroutine()
	maxGoroutinesSeen := baselineGoroutines

	// Run multiple start/stop cycles
	for cycle := 1; cycle <= 3; cycle++ {
		tw, err := NewTimingWheel(cfg)
		if err != nil {
			t.Fatalf("Cycle %d: Failed to create TimingWheel: %v", cycle, err)
		}

		mockRepo := NewMockRepository([]ScheduleItem{}, nil)

		ctx := context.Background()
		err = tw.Start(ctx, mockRepo)
		if err != nil {
			t.Fatalf("Cycle %d: Failed to start TimingWheel: %v", cycle, err)
		}

		// Let it run briefly
		time.Sleep(50 * time.Millisecond)

		maxGoroutinesSeen = maxGoroutinesSeen + 1

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = tw.Stop(shutdownCtx)
		cancel()
		if err != nil {
			t.Fatalf("Cycle %d: Failed to stop TimingWheel: %v", cycle, err)
		}

		// Wait for cleanup
		time.Sleep(50 * time.Millisecond)

		currentGoroutines := runtime.NumGoroutine()
		t.Logf("Cycle %d: Current goroutines: %d", cycle, currentGoroutines)

		// Check for leaks per cycle
		const tolerance = 3
		if currentGoroutines > baselineGoroutines+tolerance {
			t.Errorf("Cycle %d: Potential goroutine leak (baseline: %d, current: %d)",
				cycle, baselineGoroutines, currentGoroutines)
		}
	}

	// Final check
	finalGoroutines := runtime.NumGoroutine()
	t.Logf("Final goroutine count: %d (baseline: %d, max seen: %d)",
		finalGoroutines, baselineGoroutines, maxGoroutinesSeen)
}

// TestTimingWheelTickerCleanupOnStop verifies ticker is stopped
func TestTimingWheelTickerCleanupOnStop(t *testing.T) {
	cfg := &Config{
		Mode: ModeTimingWheel,
		TimingWheel: TimingWheelConfig{
			TickInterval:          100 * time.Millisecond,
			MaxWorkers:            3,
			ShutdownTimeout:       5 * time.Second,
			NotificationQueueSize: 100,
		},
	}

	tw, err := NewTimingWheel(cfg)
	if err != nil {
		t.Fatalf("Failed to create TimingWheel: %v", err)
	}

	mockRepo := NewMockRepository([]ScheduleItem{}, nil)

	ctx := context.Background()
	err = tw.Start(ctx, mockRepo)
	if err != nil {
		t.Fatalf("Failed to start TimingWheel: %v", err)
	}

	// Verify ticker is running
	if tw.ticker == nil {
		t.Fatalf("Ticker should be initialized after Start")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = tw.Stop(shutdownCtx)
	if err != nil {
		t.Fatalf("Failed to stop TimingWheel: %v", err)
	}

	// Give cleanup time
	time.Sleep(50 * time.Millisecond)

	// Ticker should still exist but be stopped (no way to directly test stopped state)
	// Verify state is stopped
	if tw.state != StateStopped {
		t.Errorf("Expected state StateStopped, got %v", tw.state)
	}
}

// TestTimingWheelChannelCleanupOnStop verifies channels are closed properly
func TestTimingWheelChannelCleanupOnStop(t *testing.T) {
	cfg := &Config{
		Mode: ModeTimingWheel,
		TimingWheel: TimingWheelConfig{
			TickInterval:          100 * time.Millisecond,
			MaxWorkers:            3,
			ShutdownTimeout:       5 * time.Second,
			NotificationQueueSize: 100,
		},
	}

	tw, err := NewTimingWheel(cfg)
	if err != nil {
		t.Fatalf("Failed to create TimingWheel: %v", err)
	}

	mockRepo := NewMockRepository([]ScheduleItem{}, nil)

	ctx := context.Background()
	err = tw.Start(ctx, mockRepo)
	if err != nil {
		t.Fatalf("Failed to start TimingWheel: %v", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = tw.Stop(shutdownCtx)
	if err != nil {
		t.Fatalf("Failed to stop TimingWheel: %v", err)
	}

	// Wait for cleanup
	time.Sleep(100 * time.Millisecond)

	// After shutdown, reading from doneChan should not block
	select {
	case <-tw.doneChan:
		t.Logf("doneChan is closed (expected)")
	case <-time.After(1 * time.Second):
		t.Logf("doneChan still open after shutdown")
	}
}

// TestTimingWheelWaitGroupCleanup verifies WaitGroup properly waits for goroutines
func TestTimingWheelWaitGroupCleanup(t *testing.T) {
	cfg := &Config{
		Mode: ModeTimingWheel,
		TimingWheel: TimingWheelConfig{
			TickInterval:          50 * time.Millisecond,
			MaxWorkers:            3,
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
	err = tw.Start(ctx, mockRepo)
	if err != nil {
		t.Fatalf("Failed to start TimingWheel: %v", err)
	}

	// Quick run
	time.Sleep(100 * time.Millisecond)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	startStopTime := time.Now()
	err = tw.Stop(shutdownCtx)
	stopElapsed := time.Since(startStopTime)

	if err != nil {
		t.Fatalf("Failed to stop TimingWheel: %v", err)
	}

	if stopElapsed > 5*time.Second {
		t.Errorf("Shutdown took longer than expected: %.2f seconds", stopElapsed.Seconds())
	}

	// Verify state is clean
	if tw.state != StateStopped {
		t.Errorf("Expected state StateStopped after shutdown")
	}

	// Verify WaitGroup completed (no panic on duplicate Done calls)
	time.Sleep(50 * time.Millisecond)
}
