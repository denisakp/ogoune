package scheduler

import (
	"context"
	"testing"
)

// mockActiveResourceRepository is a test helper for ActiveResourceRepository.
type mockActiveResourceRepository struct {
	resources []ScheduleItem
	err       error
}

// NewMockRepository creates a mock repository for testing.
func NewMockRepository(resources []ScheduleItem, err error) ActiveResourceRepository {
	return &mockActiveResourceRepository{
		resources: resources,
		err:       err,
	}
}

// FindScheduledResources returns the mock resources.
func (m *mockActiveResourceRepository) FindScheduledResources(ctx context.Context) ([]ScheduleItem, error) {
	return m.resources, m.err
}

// TestHelper is a utility for scheduler testing.
type TestHelper struct {
	t *testing.T
}

// NewTestHelper creates a new test helper instance.
func NewTestHelper(t *testing.T) *TestHelper {
	return &TestHelper{t: t}
}

// AssertSchedulerState verifies scheduler state matches expected.
func (th *TestHelper) AssertSchedulerState(scheduler Scheduler, expectedRunning bool) {
	tw, ok := scheduler.(*TimingWheel)
	if !ok {
		th.t.Fatal("Expected TimingWheel scheduler")
	}

	isRunning := tw.state == StateRunning
	if isRunning != expectedRunning {
		th.t.Errorf("Expected running=%v, got %v", expectedRunning, isRunning)
	}
}

// WaitForDone blocks until the scheduler is fully shut down or timeout occurs.
func (th *TestHelper) WaitForDone(scheduler Scheduler, ctx context.Context) error {
	tw, ok := scheduler.(*TimingWheel)
	if !ok {
		th.t.Fatal("Expected TimingWheel scheduler")
	}

	select {
	case <-tw.doneChan:
		return nil
	case <-ctx.Done():
		return ErrShutdownTimeout
	}
}
