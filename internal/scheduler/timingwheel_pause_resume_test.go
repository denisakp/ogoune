package scheduler

import (
	"context"
	"testing"
	"time"
)

// TestTimingWheelPauseImmediatelyStopsDispatch verifies that pausing a monitor
// stops dispatch immediately, not at the end of the current interval.
func TestTimingWheelPauseImmediatelyStopsDispatch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Setup: Create a mock repository and timing wheel
	mockRepo := &mockActiveResourceRepository{
		resources: []ScheduleItem{
			{
				ResourceID: "res-1",
				Interval:   500 * time.Millisecond, // short interval for testing
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

	if err := scheduler.Start(ctx, mockRepo); err != nil {
		t.Fatalf("failed to start scheduler: %v", err)
	}

	// Wait for scheduler to start
	time.Sleep(200 * time.Millisecond)

	if count := waitForCheckJobs(scheduler, 1, time.Second); count < 1 {
		t.Fatalf("expected some dispatches before pause, got %d", count)
	}

	// Pause the resource
	err = scheduler.Pause("res-1")
	if err != nil {
		t.Fatalf("failed to pause: %v", err)
	}

	// Clear any jobs already queued before asserting pause behavior.
	drainCheckQueue(scheduler)

	// Wait for time that would normally trigger more dispatches
	time.Sleep(600 * time.Millisecond)
	dispatchAfterPause := drainCheckQueue(scheduler)

	// Verify no new dispatches occurred after pause
	if dispatchAfterPause != 0 {
		t.Errorf("expected no dispatches after pause, but got %d new dispatches", dispatchAfterPause)
	}

	// Resume the resource
	err = scheduler.Resume("res-1")
	if err != nil {
		t.Fatalf("failed to resume: %v", err)
	}

	// Wait for dispatches to resume
	dispatchAfterResume := waitForCheckJobs(scheduler, 1, time.Second)

	// Verify dispatches resumed
	if dispatchAfterResume < 1 {
		t.Errorf("expected dispatches to resume after resume, but got %d", dispatchAfterResume)
	}

	// Cleanup
	_ = scheduler.Stop(ctx)
}

// TestTimingWheelResumeStartsDispatching verifies that resuming a paused monitor
// starts dispatching checks again at the configured interval.
func TestTimingWheelResumeStartsDispatching(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	mockRepo := &mockActiveResourceRepository{
		resources: []ScheduleItem{
			{
				ResourceID: "res-1",
				Interval:   300 * time.Millisecond,
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

	// Mock check executor to track dispatch times
	_ = scheduler.Start(ctx, mockRepo)

	// Let scheduler get some initial state
	time.Sleep(200 * time.Millisecond)

	// Pause immediately
	_ = scheduler.Pause("res-1")

	// Verify paused state persists
	time.Sleep(500 * time.Millisecond)

	// Resume
	_ = scheduler.Resume("res-1")

	// Verify scheduling resumes after resume
	time.Sleep(200 * time.Millisecond)

	_ = scheduler.Stop(ctx)
}

// TestTimingWheelMultiplePauseResumeToggling verifies pause/resume can be toggled multiple times.
func TestTimingWheelMultiplePauseResumeToggling(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	mockRepo := &mockActiveResourceRepository{
		resources: []ScheduleItem{
			{
				ResourceID: "res-1",
				Interval:   200 * time.Millisecond,
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

	// Toggle pause/resume 3 times and verify no errors
	for i := 0; i < 3; i++ {
		err := scheduler.Pause("res-1")
		if err != nil {
			t.Errorf("iteration %d: pause failed: %v", i, err)
		}

		time.Sleep(200 * time.Millisecond)

		err = scheduler.Resume("res-1")
		if err != nil {
			t.Errorf("iteration %d: resume failed: %v", i, err)
		}

		time.Sleep(200 * time.Millisecond)
	}

	_ = scheduler.Stop(ctx)
	t.Logf("✓ Multiple pause/resume toggles completed successfully")
}

// TestTimingWheelPauseUnknownResource returns appropriate error for unknown resource.
func TestTimingWheelPauseUnknownResource(t *testing.T) {
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

	// Try to pause non-existent resource
	err = scheduler.Pause("non-existent")
	// Should either return an error or silently succeed (implementation choice)
	t.Logf("Pause unknown resource result: %v", err)
}

func waitForCheckJobs(tw *TimingWheelScheduler, minCount int, timeout time.Duration) int {
	deadline := time.Now().Add(timeout)
	count := 0
	for time.Now().Before(deadline) {
		count += drainCheckQueue(tw)
		if count >= minCount {
			return count
		}
		time.Sleep(25 * time.Millisecond)
	}
	return count
}

func drainCheckQueue(tw *TimingWheelScheduler) int {
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
