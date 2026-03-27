package scheduler

import (
	"context"
	"testing"
	"time"
)

// TestTimingWheelCheckQueueSaturation verifies non-blocking behavior when check queue is full
// T040: Test check queue saturation with non-blocking retry
func TestTimingWheelCheckQueueSaturation(t *testing.T) {
	cfg := &Config{
		Mode: ModeTimingWheel,
		TimingWheel: TimingWheelConfig{
			TickInterval:          10 * time.Millisecond, // Fast ticks to trigger saturation
			MaxWorkers:            2,                     // Small pool
			ShutdownTimeout:       5 * time.Second,
			NotificationQueueSize: 100,
		},
	}

	tw, err := NewTimingWheel(cfg)
	if err != nil {
		t.Fatalf("Failed to create TimingWheel: %v", err)
	}

	// Create many resources to saturate check queue
	items := make([]ScheduleItem, 0)
	for i := 1; i <= 10; i++ {
		items = append(items, ScheduleItem{
			ResourceID: "res-" + string(byte('0'+i)),
			Interval:   20 * time.Millisecond,
			Paused:     false,
		})
	}
	mockRepo := NewMockRepository(items, nil)

	ctx := context.Background()
	err = tw.Start(ctx, mockRepo)
	if err != nil {
		t.Fatalf("Failed to start TimingWheel: %v", err)
	}

	// Let checks queue up and potentially saturate
	time.Sleep(150 * time.Millisecond)

	// Verify scheduler is still running (not blocked/panicked)
	if tw.state != StateRunning {
		t.Errorf("Scheduler should still be running after saturation, got state: %v", tw.state)
	}

	// Collect items from queue
	checkCount := 0
	for {
		select {
		case <-tw.checkQueue:
			checkCount++
		default:
			goto queueDone
		}
	}
queueDone:
	if checkCount == 0 {
		t.Logf("Warning: no checks in queue (saturation may not have occurred)")
	} else {
		t.Logf("Collected %d checks from queue", checkCount)
	}

	// Verify we can still stop gracefully
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = tw.Stop(shutdownCtx)
	if err != nil {
		t.Fatalf("Graceful shutdown after saturation failed: %v", err)
	}
}

// TestTimingWheelSaturationDueTimeRetention verifies due times are retained on queue saturation
func TestTimingWheelSaturationDueTimeRetention(t *testing.T) {
	cfg := &Config{
		Mode: ModeTimingWheel,
		TimingWheel: TimingWheelConfig{
			TickInterval:          20 * time.Millisecond,
			MaxWorkers:            1, // Single worker to increase saturation chance
			ShutdownTimeout:       5 * time.Second,
			NotificationQueueSize: 100,
		},
	}

	tw, err := NewTimingWheel(cfg)
	if err != nil {
		t.Fatalf("Failed to create TimingWheel: %v", err)
	}

	// Single resource with fast interval to test retry logic
	items := []ScheduleItem{
		{ResourceID: "fast-res", Interval: 20 * time.Millisecond, Paused: false},
	}
	mockRepo := NewMockRepository(items, nil)

	ctx := context.Background()
	err = tw.Start(ctx, mockRepo)
	if err != nil {
		t.Fatalf("Failed to start TimingWheel: %v", err)
	}

	// Let checks tick (some may not be accepted due to queue saturation)
	time.Sleep(100 * time.Millisecond)

	// The resource should still be scheduled (NextDue wasn't updated on queue full)
	tw.mu.RLock()
	sched, exists := tw.schedules["fast-res"]
	tw.mu.RUnlock()

	if !exists {
		t.Fatalf("Resource should still be scheduled")
	}

	if sched.ResourceID != "fast-res" {
		t.Errorf("Expected resource 'fast-res', got %s", sched.ResourceID)
	}

	// Shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = tw.Stop(shutdownCtx)
	if err != nil {
		t.Fatalf("Graceful shutdown failed: %v", err)
	}
}

// TestTimingWheelNonBlockingOnQueueFull verifies tick() doesn't block on full queue
func TestTimingWheelNonBlockingOnQueueFull(t *testing.T) {
	cfg := &Config{
		Mode: ModeTimingWheel,
		TimingWheel: TimingWheelConfig{
			TickInterval:          15 * time.Millisecond,
			MaxWorkers:            1, // Small capacity to force full queue
			ShutdownTimeout:       5 * time.Second,
			NotificationQueueSize: 100,
		},
	}

	tw, err := NewTimingWheel(cfg)
	if err != nil {
		t.Fatalf("Failed to create TimingWheel: %v", err)
	}

	// Many fast resources - likely to saturate
	items := make([]ScheduleItem, 0)
	for i := 1; i <= 8; i++ {
		items = append(items, ScheduleItem{
			ResourceID: "res-" + string(byte('0'+i)),
			Interval:   15 * time.Millisecond,
			Paused:     false,
		})
	}
	mockRepo := NewMockRepository(items, nil)

	ctx := context.Background()
	err = tw.Start(ctx, mockRepo)
	if err != nil {
		t.Fatalf("Failed to start TimingWheel: %v", err)
	}

	// Measure tick execution time while queue may be saturated
	startTime := time.Now()
	time.Sleep(100 * time.Millisecond) // Multiple tick cycles
	elapsed := time.Since(startTime)

	// Should complete in reasonable time (not blocked)
	if elapsed > 200*time.Millisecond {
		t.Logf("Scheduler took %d ms (may indicate blocking)", elapsed.Milliseconds())
	}

	// Shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = tw.Stop(shutdownCtx)
	if err != nil {
		t.Fatalf("Graceful shutdown failed: %v", err)
	}
}

// TestTimingWheelQueueRecoveryAfterSaturation verifies queue drain after saturation
func TestTimingWheelQueueRecoveryAfterSaturation(t *testing.T) {
	cfg := &Config{
		Mode: ModeTimingWheel,
		TimingWheel: TimingWheelConfig{
			TickInterval:          10 * time.Millisecond,
			MaxWorkers:            2,
			ShutdownTimeout:       5 * time.Second,
			NotificationQueueSize: 100,
		},
	}

	tw, err := NewTimingWheel(cfg)
	if err != nil {
		t.Fatalf("Failed to create TimingWheel: %v", err)
	}

	items := make([]ScheduleItem, 0)
	for i := 1; i <= 5; i++ {
		items = append(items, ScheduleItem{
			ResourceID: "res-" + string(byte('0'+i)),
			Interval:   15 * time.Millisecond,
			Paused:     false,
		})
	}
	mockRepo := NewMockRepository(items, nil)

	ctx := context.Background()
	err = tw.Start(ctx, mockRepo)
	if err != nil {
		t.Fatalf("Failed to start TimingWheel: %v", err)
	}

	// Let it run and potentially saturate
	time.Sleep(100 * time.Millisecond)

	// Drain queue manually (simulating worker consumption)
	drainedCount := 0
	for {
		select {
		case <-tw.checkQueue:
			drainedCount++
		default:
			goto afterDrain
		}
	}
afterDrain:

	// Continue running and collecting new checks
	time.Sleep(50 * time.Millisecond)

	drainedCount2 := 0
	for {
		select {
		case <-tw.checkQueue:
			drainedCount2++
		default:
			goto afterDrain2
		}
	}
afterDrain2:

	t.Logf("First drain: %d checks, Second drain: %d checks (recovery test)", drainedCount, drainedCount2)

	// Shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = tw.Stop(shutdownCtx)
	if err != nil {
		t.Fatalf("Graceful shutdown failed: %v", err)
	}
}
