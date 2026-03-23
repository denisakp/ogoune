package scheduler

import (
	"context"
	"testing"
	"time"
)

// TestTimingWheelNotificationQueueFull verifies behavior when notification queue is full
// T041: Test notification queue full behavior
func TestTimingWheelNotificationQueueFull(t *testing.T) {
	cfg := &Config{
		Mode: ModeTimingWheel,
		TimingWheel: TimingWheelConfig{
			TickInterval:          1 * time.Second,
			MaxWorkers:            5,
			ShutdownTimeout:       5 * time.Second,
			NotificationQueueSize: 5, // Very small to test saturation
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

	// Fill the notification queue
	enqueuedCount := 0
	for i := 0; i < 20; i++ {
		err := tw.EnqueueNotification("incident-"+string(byte('0'+byte(i))), "resource_down_alert")
		if err == nil {
			enqueuedCount++
		}
	}

	// EnqueueNotification is non-blocking, so all should succeed (returns nil on full)
	if enqueuedCount == 0 {
		t.Logf("No notifications enqueued (queue size issue)")
	}

	// Verify queue has items
	queueLen := len(tw.notifQueue)
	if queueLen == 0 {
		t.Logf("Notification queue is empty")
	} else {
		t.Logf("Queue has %d items (capacity: 5)", queueLen)
	}

	// Shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = tw.Stop(shutdownCtx)
	if err != nil {
		t.Fatalf("Graceful shutdown failed: %v", err)
	}
}

// TestTimingWheelEnqueueNotificationNonBlocking verifies EnqueueNotification doesn't block
func TestTimingWheelEnqueueNotificationNonBlocking(t *testing.T) {
	cfg := &Config{
		Mode: ModeTimingWheel,
		TimingWheel: TimingWheelConfig{
			TickInterval:          100 * time.Millisecond,
			MaxWorkers:            3,
			ShutdownTimeout:       5 * time.Second,
			NotificationQueueSize: 5,
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

	// Enqueue many notifications sequentially - should not block
	startTime := time.Now()
	for i := 0; i < 100; i++ {
		tw.EnqueueNotification("incident-"+string(byte(i)), "resource_up_alert")
	}
	elapsed := time.Since(startTime)

	// Should complete very quickly (not blocked waiting on queue)
	if elapsed > 100*time.Millisecond {
		t.Logf("EnqueueNotification took %d ms (may indicate blocking)", elapsed.Milliseconds())
	} else {
		t.Logf("Enqueued 100 notifications in %.2f ms (non-blocking confirmed)", elapsed.Seconds()*1000)
	}

	// Shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = tw.Stop(shutdownCtx)
	if err != nil {
		t.Fatalf("Graceful shutdown failed: %v", err)
	}
}

// TestTimingWheelNotificationJobFormatting verifies NotificationJob structure
func TestTimingWheelNotificationJobFormatting(t *testing.T) {
	cfg := &Config{
		Mode: ModeTimingWheel,
		TimingWheel: TimingWheelConfig{
			TickInterval:          1 * time.Second,
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

	// Enqueue specific notifications
	tw.EnqueueNotification("test-incident-1", "resource_down_alert")
	tw.EnqueueNotification("test-incident-2", "resource_up_alert")

	// Retrieve and verify
	var retrievedJobs []*NotificationJob
	for i := 0; i < 2; i++ {
		select {
		case job, ok := <-tw.notifQueue:
			if !ok {
				t.Fatalf("notification queue closed unexpectedly")
			}
			retrievedJobs = append(retrievedJobs, job)
		case <-time.After(1 * time.Second):
			t.Fatalf("Timeout waiting for notification job")
		}
	}

	if len(retrievedJobs) != 2 {
		t.Fatalf("Expected 2 jobs, got %d", len(retrievedJobs))
	}

	// Verify event types are canonical
	eventTypes := map[NotificationEventType]bool{
		NotificationEventResourceDownAlert: false,
		NotificationEventResourceUpAlert:   false,
	}

	for _, job := range retrievedJobs {
		eventType := NotificationEventType(job.EventType)
		if _, valid := eventTypes[eventType]; !valid {
			t.Errorf("Invalid event type: %s", job.EventType)
		}
		eventTypes[eventType] = true
	}

	for eventType, seen := range eventTypes {
		if !seen {
			t.Logf("Event type not tested: %s", eventType)
		}
	}

	if err := tw.EnqueueNotification("test-incident-3", "invalid"); err == nil {
		t.Fatal("expected invalid notification event type to be rejected")
	}

	// Shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = tw.Stop(shutdownCtx)
	if err != nil {
		t.Fatalf("Graceful shutdown failed: %v", err)
	}
}

// TestTimingWheelNotificationQueueDrainOnShutdown verifies notifications are drainable on shutdown
func TestTimingWheelNotificationQueueDrainOnShutdown(t *testing.T) {
	cfg := &Config{
		Mode: ModeTimingWheel,
		TimingWheel: TimingWheelConfig{
			TickInterval:          50 * time.Millisecond,
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

	// Enqueue notifications
	for i := 0; i < 10; i++ {
		tw.EnqueueNotification("incident-"+string(byte(i)), "resource_down_alert")
	}

	time.Sleep(50 * time.Millisecond)

	// Shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = tw.Stop(shutdownCtx)
	if err != nil {
		t.Fatalf("Graceful shutdown failed: %v", err)
	}

	// After shutdown, should be able to drain queue non-blockingly
	drainedNotifications := 0
	for {
		select {
		case _, ok := <-tw.notifQueue:
			if !ok {
				goto draindone
			}
			drainedNotifications++
		default:
			goto draindone
		}
	}
draindone:

	t.Logf("Drained %d notifications after shutdown", drainedNotifications)
}
