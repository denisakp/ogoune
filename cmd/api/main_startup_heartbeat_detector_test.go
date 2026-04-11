package main

import (
	"context"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubDetectorRepo satisfies missedHeartbeatQuerier (unexported interface in service pkg).
type stubDetectorRepo struct {
	runCh chan struct{}
}

func (s *stubDetectorRepo) FindMissedHeartbeats(_ context.Context, _ time.Time, _ int) ([]*domain.Resource, error) {
	if s.runCh != nil {
		select {
		case s.runCh <- struct{}{}:
		default:
		}
	}
	return nil, nil
}

// stubDetectorIncidents satisfies heartbeatIncidentManager (unexported interface in service pkg).
type stubDetectorIncidents struct{}

func (s *stubDetectorIncidents) CreateIncident(_ context.Context, _ *domain.Resource, _ domain.CheckResult) error {
	return nil
}

func newTestDetector() *service.HeartbeatDetectorService {
	return service.NewHeartbeatDetectorService(&stubDetectorRepo{}, &stubDetectorIncidents{})
}

// TestStartupHeartbeatDetector_RegistrationSucceeds verifies startHeartbeatDetector
// returns no error for a live context (T029).
func TestStartupHeartbeatDetector_RegistrationSucceeds(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := startHeartbeatDetector(ctx, newTestDetector(), 60*time.Second)
	require.NoError(t, err, "detector startup must succeed with a live context")
}

// TestStartupHeartbeatDetector_FailFastOnCancelledContext verifies startHeartbeatDetector
// returns context.Canceled immediately when context is pre-cancelled (T030).
func TestStartupHeartbeatDetector_FailFastOnCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancel → simulate startup failure

	err := startHeartbeatDetector(ctx, newTestDetector(), 60*time.Second)
	assert.Error(t, err, "detector startup must fail when context is already cancelled")
	assert.ErrorIs(t, err, context.Canceled)
}

// TestStartupHeartbeatDetector_NilDetector_ReturnsError verifies nil detector is rejected
// at startup (fail-fast: nil means misconfigured dependency injection).
func TestStartupHeartbeatDetector_NilDetector_ReturnsError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := startHeartbeatDetector(ctx, nil, 60*time.Second)
	assert.Error(t, err, "nil detector must be rejected at startup")
}

// TestHeartbeatDetector_CadenceReliability verifies the detector fires at the expected
// cadence (multiple times within a short window). Covers T069 cadence SLO.
func TestHeartbeatDetector_CadenceReliability(t *testing.T) {
	runCh := make(chan struct{}, 10)
	repo := &stubDetectorRepo{runCh: runCh}
	detector := service.NewHeartbeatDetectorService(repo, &stubDetectorIncidents{})

	ctx, cancel := context.WithCancel(context.Background())

	err := detector.Start(ctx, 20*time.Millisecond)
	require.NoError(t, err)

	// Expect at least 3 cycles within 200ms
	deadline := time.After(500 * time.Millisecond)
	fired := 0
	for fired < 3 {
		select {
		case <-runCh:
			fired++
		case <-deadline:
			cancel()
			detector.Wait()
			t.Fatalf("detector fired only %d time(s) before deadline, expected at least 3", fired)
		}
	}

	// Cancel and wait for the goroutine to finish before returning, so the log output
	// from "[heartbeat-detector] stopped" does not race with the next test's log.SetOutput.
	cancel()
	detector.Wait()
}
