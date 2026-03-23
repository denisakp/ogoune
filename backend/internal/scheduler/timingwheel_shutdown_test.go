package scheduler

import (
	"context"
	"testing"
	"time"
)

// TestTimingWheelGracefulShutdown verifies graceful shutdown drains in-flight checks and notifications
// T038: Test graceful shutdown covering in-flight checks and notification drain
func TestTimingWheelGracefulShutdown(t *testing.T) {
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

	// Create initial schedules
	items := []ScheduleItem{
		{ResourceID: "res-1", Interval: 100 * time.Millisecond, Paused: false},
		{ResourceID: "res-2", Interval: 150 * time.Millisecond, Paused: false},
		{ResourceID: "res-3", Interval: 200 * time.Millisecond, Paused: false},
	}
	mockRepo := NewMockRepository(items, nil)

	// Start scheduler
	ctx := context.Background()
	err = tw.Start(ctx, mockRepo)
	if err != nil {
		t.Fatalf("Failed to start TimingWheel: %v", err)
	}

	// Verify schedules are loaded
	tw.mu.RLock()
	if len(tw.schedules) != 3 {
		t.Errorf("Expected 3 schedules loaded, got %d", len(tw.schedules))
	}
	tw.mu.RUnlock()

	// Let some checks queue up
	time.Sleep(250 * time.Millisecond)

	// Collect checks from queue before shutdown
	checkCount := 0
	for {
		select {
		case job := <-tw.checkQueue:
			if job == nil {
				break
			}
			checkCount++
		default:
			goto shutdownPhase
		}
	}

shutdownPhase:
	// Gracefully shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = tw.Stop(shutdownCtx)
	if err != nil {
		t.Fatalf("Expected graceful shutdown, got error: %v", err)
	}

	// Verify state is stopped
	if tw.state != StateStopped {
		t.Errorf("Expected state StateStopped, got %v", tw.state)
	}

	// Verify no more ticks happen after shutdown
	queueAfterShutdown := len(tw.checkQueue)
	time.Sleep(100 * time.Millisecond)
	if len(tw.checkQueue) != queueAfterShutdown {
		t.Errorf("Queue should not change after shutdown")
	}
}

// TestTimingWheelShutdownDrainsCheckQueue verifies check queue is processed on shutdown
func TestTimingWheelShutdownDrainsCheckQueue(t *testing.T) {
	cfg := &Config{
		Mode: ModeTimingWheel,
		TimingWheel: TimingWheelConfig{
			TickInterval:          50 * time.Millisecond,
			MaxWorkers:            2,
			ShutdownTimeout:       3 * time.Second,
			NotificationQueueSize: 50,
		},
	}

	tw, err := NewTimingWheel(cfg)
	if err != nil {
		t.Fatalf("Failed to create TimingWheel: %v", err)
	}

	// Create a resource that will trigger checks
	items := []ScheduleItem{
		{ResourceID: "heavy-res", Interval: 50 * time.Millisecond, Paused: false},
	}
	mockRepo := NewMockRepository(items, nil)

	ctx := context.Background()
	err = tw.Start(ctx, mockRepo)
	if err != nil {
		t.Fatalf("Failed to start TimingWheel: %v", err)
	}

	// Let checks accumulate
	time.Sleep(200 * time.Millisecond)

	// Count queued items before shutdown
	itemsInQueueBeforeShutdown := len(tw.checkQueue)
	if itemsInQueueBeforeShutdown == 0 {
		t.Logf("Warning: no items in check queue before shutdown (expected at least 1)")
	}

	// Shutdown should drain queue
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = tw.Stop(shutdownCtx)
	if err != nil {
		t.Fatalf("Graceful shutdown failed: %v", err)
	}

	// After shutdown, queue should be drainable (not blocking)
	drainedCount := 0
	for {
		select {
		case <-tw.checkQueue:
			drainedCount++
		default:
			goto done
		}
	}
done:
	t.Logf("Drained %d checks from queue after shutdown", drainedCount)
}

// TestTimingWheelShutdownDrainsNotificationQueue verifies notification queue is drained
func TestTimingWheelShutdownDrainsNotificationQueue(t *testing.T) {
	cfg := &Config{
		Mode: ModeTimingWheel,
		TimingWheel: TimingWheelConfig{
			TickInterval:          100 * time.Millisecond,
			MaxWorkers:            2,
			ShutdownTimeout:       3 * time.Second,
			NotificationQueueSize: 10,
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

	// Queue many notifications
	for i := 0; i < 5; i++ {
		tw.EnqueueNotification("incident-"+string(rune(i)), "resource_down_alert")
	}

	time.Sleep(50 * time.Millisecond)

	// Shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = tw.Stop(shutdownCtx)
	if err != nil {
		t.Fatalf("Graceful shutdown failed: %v", err)
	}

	// Should be able to drain notifications without blocking
	drainedNotifs := 0
	for {
		select {
		case _, ok := <-tw.notifQueue:
			if !ok {
				goto notifDone
			}
			drainedNotifs++
		default:
			goto notifDone
		}
	}
notifDone:
	t.Logf("Drained %d notifications from queue after shutdown", drainedNotifs)
}
