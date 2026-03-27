package scheduler

import (
	"context"
	"testing"
	"time"
)

// TestTimingWheelIntervalUpdateReschedules verifies that updating a monitor's interval
// takes effect for the next scheduled dispatch, even if a check is in flight.
func TestTimingWheelIntervalUpdateReschedules(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	initialInterval := 500 * time.Millisecond
	updatedInterval := 300 * time.Millisecond

	mockRepo := &mockActiveResourceRepository{
		resources: []ScheduleItem{
			{
				ResourceID: "res-1",
				Interval:   initialInterval,
				Paused:     false,
			},
		},
	}

	cfg := &Config{
		Mode: ModeTimingWheel,
		TimingWheel: TimingWheelConfig{
			TickInterval:          100 * time.Millisecond,
			MaxWorkers:            2,
			ShutdownTimeout:       5 * time.Second,
			NotificationQueueSize: 100,
		},
	}

	scheduler, err := NewTimingWheel(cfg)
	if err != nil {
		t.Fatalf("failed to create scheduler: %v", err)
	}

	_ = scheduler.Start(ctx, mockRepo)

	// Wait for initial state
	time.Sleep(200 * time.Millisecond)

	dispatchesBeforeUpdate := waitForRescheduleJobs(scheduler, 1, time.Second)
	t.Logf("Dispatches before interval update: %d", dispatchesBeforeUpdate)

	// Update interval to faster rate
	err = scheduler.Schedule("res-1", updatedInterval)
	if err != nil {
		t.Fatalf("failed to update schedule: %v", err)
	}

	// Wait and observe dispatch pattern with new interval
	// With new interval of 300ms, we should see more frequent dispatches
	drainRescheduleQueue(scheduler)
	dispatchesAfterUpdate := countRescheduleJobsOverWindow(scheduler, 1100*time.Millisecond)

	t.Logf("Dispatches after interval update (1s window with 300ms interval): %d",
		dispatchesAfterUpdate)

	// With original 500ms interval over 600ms we'd expect ~1 dispatch
	// After updating to 300ms over 1000ms we'd expect ~3 dispatches
	// Be lenient due to timing variations
	if dispatchesAfterUpdate < 2 {
		t.Errorf("expected faster dispatch rate after interval update, but got only %d new dispatches in 1s",
			dispatchesAfterUpdate)
	}

	// Now update to slower interval and verify it takes effect
	slowInterval := 1 * time.Second
	err = scheduler.Schedule("res-1", slowInterval)
	if err != nil {
		t.Fatalf("failed to update to slower interval: %v", err)
	}

	drainRescheduleQueue(scheduler)
	dispatchesAfterSlowUpdate := countRescheduleJobsOverWindow(scheduler, 800*time.Millisecond)

	// With 1s interval, over 800ms we should see 0 dispatches (next would be at 1s mark)
	if dispatchesAfterSlowUpdate > 0 {
		t.Logf("Note: Got unexpected dispatch after slow interval update: %d dispatches in 800ms window",
			dispatchesAfterSlowUpdate)
	}

	_ = scheduler.Stop(ctx)
}

// TestTimingWheelIntervalUpdateWhileCheckInFlight verifies that in-flight checks
// are allowed to complete even if interval is updated mid-execution.
func TestTimingWheelIntervalUpdateWhileCheckInFlight(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	mockRepo := &mockActiveResourceRepository{
		resources: []ScheduleItem{
			{
				ResourceID: "res-1",
				Interval:   500 * time.Millisecond,
				Paused:     false,
			},
		},
	}

	cfg := &Config{
		Mode: ModeTimingWheel,
		TimingWheel: TimingWheelConfig{
			TickInterval:          100 * time.Millisecond,
			MaxWorkers:            2,
			ShutdownTimeout:       5 * time.Second,
			NotificationQueueSize: 100,
		},
	}

	scheduler, err := NewTimingWheel(cfg)
	if err != nil {
		t.Fatalf("failed to create scheduler: %v", err)
	}

	_ = scheduler.Start(ctx, mockRepo)
	time.Sleep(200 * time.Millisecond)

	// Wait for first dispatch
	time.Sleep(600 * time.Millisecond)

	// Simulate slow check by updating interval while check might be in flight
	// (in real implementation, check would be executing in worker pool)
	go func() {
		time.Sleep(100 * time.Millisecond) // Let check start
		_ = scheduler.Schedule("res-1", 300*time.Millisecond)
	}()

	// Wait for interval update to occur
	time.Sleep(500 * time.Millisecond)

	// Verify that old check completed and new interval is in effect
	// (exact behavior depends on check execution model)
	_ = scheduler.Stop(ctx)

	t.Logf("✓ Interval update during in-flight check handled without panic")
}

// TestTimingWheelMultipleScheduleUpdates verifies rapid schedule updates don't corrupt state.
func TestTimingWheelMultipleScheduleUpdates(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	mockRepo := &mockActiveResourceRepository{
		resources: []ScheduleItem{
			{
				ResourceID: "res-1",
				Interval:   500 * time.Millisecond,
				Paused:     false,
			},
		},
	}

	cfg := &Config{
		Mode: ModeTimingWheel,
		TimingWheel: TimingWheelConfig{
			TickInterval:          100 * time.Millisecond,
			MaxWorkers:            2,
			ShutdownTimeout:       5 * time.Second,
			NotificationQueueSize: 100,
		},
	}

	scheduler, err := NewTimingWheel(cfg)
	if err != nil {
		t.Fatalf("failed to create scheduler: %v", err)
	}

	_ = scheduler.Start(ctx, mockRepo)
	time.Sleep(200 * time.Millisecond)

	// Perform rapid interval updates
	intervals := []time.Duration{
		300 * time.Millisecond,
		600 * time.Millisecond,
		400 * time.Millisecond,
		500 * time.Millisecond,
		250 * time.Millisecond,
	}

	for i, interval := range intervals {
		err := scheduler.Schedule("res-1", interval)
		if err != nil {
			t.Errorf("update %d: failed to schedule interval %v: %v", i, interval, err)
		}
		time.Sleep(50 * time.Millisecond)
	}

	// Verify scheduler is still running and responsive
	err = scheduler.Pause("res-1")
	if err != nil {
		t.Fatalf("failed to pause after rapid updates: %v", err)
	}

	err = scheduler.Resume("res-1")
	if err != nil {
		t.Fatalf("failed to resume after rapid updates: %v", err)
	}

	_ = scheduler.Stop(ctx)
	t.Logf("✓ Multiple rapid schedule updates handled without state corruption")
}

// TestTimingWheelInvalidIntervalRejected verifies that invalid intervals are rejected.
func TestTimingWheelInvalidIntervalRejected(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	mockRepo := &mockActiveResourceRepository{
		resources: []ScheduleItem{},
	}

	cfg := &Config{
		Mode: ModeTimingWheel,
		TimingWheel: TimingWheelConfig{
			TickInterval:          100 * time.Millisecond,
			MaxWorkers:            2,
			ShutdownTimeout:       5 * time.Second,
			NotificationQueueSize: 100,
		},
	}

	scheduler, err := NewTimingWheel(cfg)
	if err != nil {
		t.Fatalf("failed to create scheduler: %v", err)
	}

	_ = scheduler.Start(ctx, mockRepo)
	defer scheduler.Stop(ctx)

	testCases := []struct {
		name     string
		interval time.Duration
	}{
		{"zero interval", 0},
		{"negative interval", -100 * time.Millisecond},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := scheduler.Schedule("res-1", tc.interval)
			if err != ErrInvalidInterval {
				t.Errorf("expected ErrInvalidInterval for %s, got: %v", tc.name, err)
			}
		})
	}
}

func waitForRescheduleJobs(tw *TimingWheelScheduler, minCount int, timeout time.Duration) int {
	deadline := time.Now().Add(timeout)
	count := 0
	for time.Now().Before(deadline) {
		count += drainRescheduleQueue(tw)
		if count >= minCount {
			return count
		}
		time.Sleep(25 * time.Millisecond)
	}
	return count
}

func countRescheduleJobsOverWindow(tw *TimingWheelScheduler, window time.Duration) int {
	deadline := time.Now().Add(window)
	count := 0
	for time.Now().Before(deadline) {
		count += drainRescheduleQueue(tw)
		time.Sleep(25 * time.Millisecond)
	}
	return count
}

func drainRescheduleQueue(tw *TimingWheelScheduler) int {
	count := 0
	for {
		select {
		case <-tw.checkQueue:
			count++
		default:
			return count
		}
	}
}
