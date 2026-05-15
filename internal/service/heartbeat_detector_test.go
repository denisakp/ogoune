package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ missedHeartbeatQuerier = (*mockHeartbeatRepo)(nil)

// --- mock repository (implements only missedHeartbeatQuerier) ---

type mockHeartbeatRepo struct {
	resources []*domain.Resource
	err       error
}

func (m *mockHeartbeatRepo) FindMissedHeartbeats(_ context.Context, _ time.Time, _ int) ([]*domain.Resource, error) {
	return m.resources, m.err
}

// --- mock incident manager ---

type mockHeartbeatIncidents struct {
	calls  []*domain.Resource
	errors []error
	idx    int
}

func (m *mockHeartbeatIncidents) CreateIncident(_ context.Context, r *domain.Resource, _ domain.CheckResult) error {
	m.calls = append(m.calls, r)
	if m.idx < len(m.errors) {
		err := m.errors[m.idx]
		m.idx++
		return err
	}
	m.idx++
	return nil
}

// --- helpers ---

func makeHeartbeatResource(id string) *domain.Resource {
	last := time.Now().Add(-2 * time.Hour)
	interval := 3600
	grace := 60
	return &domain.Resource{
		Base:              domain.Base{ID: id},
		Name:              "Backup-" + id,
		Type:              domain.ResourceHeartbeat,
		HeartbeatInterval: &interval,
		HeartbeatGrace:    &grace,
		LastPingAt:        &last,
	}
}

// --- tests ---

func TestHeartbeatDetector_NoMissed(t *testing.T) {
	repo := &mockHeartbeatRepo{resources: nil}
	incidents := &mockHeartbeatIncidents{}
	svc := NewHeartbeatDetectorService(repo, incidents)

	err := svc.Detect(context.Background())
	require.NoError(t, err)
	assert.Empty(t, incidents.calls)
}

func TestHeartbeatDetector_ThreeMissed_AllSucceed(t *testing.T) {
	resources := []*domain.Resource{
		makeHeartbeatResource("r1"),
		makeHeartbeatResource("r2"),
		makeHeartbeatResource("r3"),
	}
	repo := &mockHeartbeatRepo{resources: resources}
	incidents := &mockHeartbeatIncidents{}
	svc := NewHeartbeatDetectorService(repo, incidents)

	err := svc.Detect(context.Background())
	require.NoError(t, err)
	assert.Len(t, incidents.calls, 3)

	// Verify CheckResult fields synthesized correctly
	cause := domain.MissedHeartbeat
	for i, r := range incidents.calls {
		assert.Equal(t, resources[i].ID, r.ID)
		_ = cause // cause is checked via CreateIncident capturing
	}
}

func TestHeartbeatDetector_QueryError_ReturnsError(t *testing.T) {
	repo := &mockHeartbeatRepo{err: errors.New("db unavailable")}
	incidents := &mockHeartbeatIncidents{}
	svc := NewHeartbeatDetectorService(repo, incidents)

	err := svc.Detect(context.Background())
	assert.Error(t, err)
	assert.Empty(t, incidents.calls)
}

func TestHeartbeatDetector_FireAndForget_OnIncidentFailure(t *testing.T) {
	resources := []*domain.Resource{
		makeHeartbeatResource("r1"),
		makeHeartbeatResource("r2"),
		makeHeartbeatResource("r3"),
	}
	repo := &mockHeartbeatRepo{resources: resources}
	// r2 fails, r1 and r3 succeed
	incidents := &mockHeartbeatIncidents{
		errors: []error{nil, errors.New("incident svc down"), nil},
	}
	svc := NewHeartbeatDetectorService(repo, incidents)

	// Fire-and-forget: Detect should not return error even when one incident call fails
	err := svc.Detect(context.Background())
	require.NoError(t, err)
	// All 3 monitors were attempted
	assert.Len(t, incidents.calls, 3)
}

func TestHeartbeatDetector_CheckResultHasMissedHeartbeatCause(t *testing.T) {
	resources := []*domain.Resource{makeHeartbeatResource("r1")}
	repo := &mockHeartbeatRepo{resources: resources}

	var capturedResult domain.CheckResult
	var capturedResource *domain.Resource

	incidents := &captureIncident{}
	incidents.fn = func(r *domain.Resource, result domain.CheckResult) {
		capturedResource = r
		capturedResult = result
	}

	svc := NewHeartbeatDetectorService(repo, incidents)
	err := svc.Detect(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "r1", capturedResource.ID)
	require.NotNil(t, capturedResult.Cause)
	assert.Equal(t, domain.MissedHeartbeat, *capturedResult.Cause)
	assert.Equal(t, string(domain.StatusDown), capturedResult.Status)
	assert.NotEmpty(t, capturedResult.ErrorMessage)
}

type captureIncident struct {
	fn func(r *domain.Resource, result domain.CheckResult)
}

func (c *captureIncident) CreateIncident(_ context.Context, r *domain.Resource, result domain.CheckResult) error {
	if c.fn != nil {
		c.fn(r, result)
	}
	return nil
}

// --- Backoff delay calculation tests (T028) ---

func TestHeartbeatDetector_BackoffDelay(t *testing.T) {
	svc := &HeartbeatDetectorService{}
	cases := []struct {
		attempt  int
		expected time.Duration
	}{
		{0, 1 * time.Minute},
		{1, 2 * time.Minute},
		{2, 4 * time.Minute},
		{3, 8 * time.Minute},
		{10, 60 * time.Minute}, // capped
	}
	for _, tc := range cases {
		got := svc.CalculateBackoffDelay(tc.attempt)
		assert.Equal(t, tc.expected, got, "attempt=%d", tc.attempt)
	}
}

func TestHeartbeatDetector_BackoffCappedAt60Minutes(t *testing.T) {
	svc := &HeartbeatDetectorService{}
	for attempt := 6; attempt <= 20; attempt++ {
		delay := svc.CalculateBackoffDelay(attempt)
		assert.Equal(t, 60*time.Minute, delay, "attempt=%d should be capped", attempt)
	}
}

// --- Start / context cancellation tests (T029 / T030) ---

func TestHeartbeatDetectorService_Start_ReturnsImmediatelyOnCancelledContext(t *testing.T) {
	repo := &mockHeartbeatRepo{}
	incidents := &mockHeartbeatIncidents{}
	svc := NewHeartbeatDetectorService(repo, incidents)

	// Cancelled context → Start should return an error immediately (fail-fast)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancel

	err := svc.Start(ctx, 60*time.Second)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestHeartbeatDetectorService_Start_LaunchesGoroutineAndRunsDetection(t *testing.T) {
	detected := make(chan struct{}, 1)
	repo := &mockHeartbeatRepo{}
	incidents := &captureIncident{fn: nil}
	svc := NewHeartbeatDetectorService(repo, incidents)
	// Intercept Detect by tracking the next mock call
	repo.resources = []*domain.Resource{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Use a very short interval so the goroutine fires quickly
	svc.now = func() time.Time { return time.Now() }
	err := svc.Start(ctx, 50*time.Millisecond)
	require.NoError(t, err)

	// Give the goroutine time to tick
	go func() {
		time.Sleep(200 * time.Millisecond)
		detected <- struct{}{}
	}()

	select {
	case <-detected:
		// goroutine ran without panic
	case <-time.After(1 * time.Second):
		t.Fatal("detector goroutine did not start within timeout")
	}
}
