package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimingWheelConfirmation_FastRetryIntervalDispatchesMoreFrequently(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	fastDispatches := countDispatchesForInterval(t, 120*time.Millisecond, 900*time.Millisecond)
	normalDispatches := countDispatchesForInterval(t, 400*time.Millisecond, 900*time.Millisecond)

	assert.Greater(t, fastDispatches, normalDispatches, "confirmation retry interval should dispatch more frequently than normal interval")
}

func TestTimingWheelConfirmation_NormalIntervalRestoredAfterRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	mockRepo := &mockActiveResourceRepository{resources: []ScheduleItem{{
		ResourceID: "res-confirm",
		Interval:   120 * time.Millisecond,
		Paused:     false,
	}}}

	cfg := &Config{
		Mode: ModeTimingWheel,
		TimingWheel: TimingWheelConfig{
			TickInterval:          20 * time.Millisecond,
			MaxWorkers:            4,
			ShutdownTimeout:       5 * time.Second,
			NotificationQueueSize: 32,
		},
	}

	tw, err := NewTimingWheel(cfg)
	require.NoError(t, err)
	require.NoError(t, tw.Start(ctx, mockRepo))
	defer func() { _ = tw.Stop(ctx) }()

	// Simulate confirming cadence first.
	time.Sleep(150 * time.Millisecond)
	drainConfirmationQueue(tw)
	confirmingDispatches := countConfirmationDispatches(tw, 500*time.Millisecond)

	// Simulate recovery by restoring normal monitor interval.
	require.NoError(t, tw.Schedule("res-confirm", 400*time.Millisecond))
	time.Sleep(40 * time.Millisecond)
	drainConfirmationQueue(tw)
	normalDispatches := countConfirmationDispatches(tw, 500*time.Millisecond)

	assert.Greater(t, confirmingDispatches, normalDispatches, "normal interval should reduce dispatch frequency after recovery")
}

func countDispatchesForInterval(t *testing.T, interval, window time.Duration) int {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	mockRepo := &mockActiveResourceRepository{resources: []ScheduleItem{{
		ResourceID: "res-1",
		Interval:   interval,
		Paused:     false,
	}}}

	cfg := &Config{
		Mode: ModeTimingWheel,
		TimingWheel: TimingWheelConfig{
			TickInterval:          20 * time.Millisecond,
			MaxWorkers:            4,
			ShutdownTimeout:       5 * time.Second,
			NotificationQueueSize: 32,
		},
	}

	tw, err := NewTimingWheel(cfg)
	require.NoError(t, err)
	require.NoError(t, tw.Start(ctx, mockRepo))
	defer func() { _ = tw.Stop(ctx) }()

	time.Sleep(150 * time.Millisecond)
	drainConfirmationQueue(tw)
	return countConfirmationDispatches(tw, window)
}

func countConfirmationDispatches(tw *TimingWheelScheduler, window time.Duration) int {
	deadline := time.Now().Add(window)
	count := 0
	for time.Now().Before(deadline) {
		count += drainConfirmationQueue(tw)
		time.Sleep(15 * time.Millisecond)
	}
	return count
}

func drainConfirmationQueue(tw *TimingWheelScheduler) int {
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
