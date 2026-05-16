package service

import (
	"context"
	"log/slog"
	"math"
	"sync"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
)

const (
	detectorBackoffBase   = 1 * time.Minute
	detectorBackoffMax    = 60 * time.Minute
	detectorBackoffFactor = 2.0
	detectorQueryLimit    = 1000
)

// missedHeartbeatQuerier is the repository subset the detector requires.
type missedHeartbeatQuerier interface {
	FindMissedHeartbeats(ctx context.Context, now time.Time, limit int) ([]*domain.Resource, error)
}

// heartbeatIncidentManager is the subset of the incident service used by the detector.
type heartbeatIncidentManager interface {
	CreateIncident(ctx context.Context, r *domain.Resource, result domain.CheckResult) error
}

// HeartbeatDetectorService queries for missed heartbeat monitors every detection cycle
// and triggers incident creation using fire-and-forget semantics (H2).
type HeartbeatDetectorService struct {
	resources missedHeartbeatQuerier
	incidents heartbeatIncidentManager
	now       func() time.Time
	wg        sync.WaitGroup
}

// NewHeartbeatDetectorService creates a new HeartbeatDetectorService.
func NewHeartbeatDetectorService(
	resources missedHeartbeatQuerier,
	incidents heartbeatIncidentManager,
) *HeartbeatDetectorService {
	return &HeartbeatDetectorService{
		resources: resources,
		incidents: incidents,
		now:       time.Now,
	}
}

// CalculateBackoffDelay returns the exponential backoff delay for a given retry attempt.
// Attempt 0 → 1m, attempt 1 → 2m, attempt 2 → 4m, attempt 3+ → 60m.
func (d *HeartbeatDetectorService) CalculateBackoffDelay(attempt int) time.Duration {
	delay := time.Duration(float64(detectorBackoffBase) * math.Pow(detectorBackoffFactor, float64(attempt)))
	if delay > detectorBackoffMax {
		return detectorBackoffMax
	}
	return delay
}

// Detect runs one detection cycle: queries for missed heartbeat monitors, synthesizes
// a CheckResult for each, and calls CreateIncident. Failures are logged and skipped
// (fire-and-forget) so the pipeline continues for all remaining monitors.
func (d *HeartbeatDetectorService) Detect(ctx context.Context) error {
	now := d.now()

	missed, err := d.resources.FindMissedHeartbeats(ctx, now, detectorQueryLimit)
	if err != nil {
		slog.Error("heartbeat detection query failed", "error", err)
		return err
	}

	if len(missed) == 0 {
		slog.Info("heartbeat detection run complete", "missed", 0)
		return nil
	}

	cause := domain.MissedHeartbeat
	var succeeded, failed int

	for _, resource := range missed {
		result := domain.CheckResult{
			Status:       string(domain.StatusDown),
			Cause:        &cause,
			ErrorMessage: "No ping received within the expected interval + grace period.",
		}

		if err := d.incidents.CreateIncident(ctx, resource, result); err != nil {
			slog.Error("heartbeat incident creation failed",
				"resource_id", resource.ID, "resource_name", resource.Name,
				"error", err, "retry_after", d.CalculateBackoffDelay(0))
			failed++
		} else {
			slog.Info("heartbeat incident created", "resource_id", resource.ID, "resource_name", resource.Name)
			succeeded++
		}
	}

	slog.Info("heartbeat detection run complete",
		"missed", len(missed), "incidents_created", succeeded, "failures", failed)
	return nil
}

// Start launches the detector as a recurring goroutine on the given interval.
// It blocks until ctx is cancelled. The first detection cycle runs after one interval.
// Returns an error immediately if ctx is already cancelled, ensuring fail-fast startup detection.
func (d *HeartbeatDetectorService) Start(ctx context.Context, interval time.Duration) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				slog.Info("heartbeat detector stopped")
				return
			case <-ticker.C:
				if err := d.Detect(ctx); err != nil {
					slog.Error("heartbeat detection cycle error", "error", err)
				}
			}
		}
	}()

	slog.Info("heartbeat detector started", "interval", interval)
	return nil
}

// Wait blocks until the background goroutine started by Start has fully exited.
// Useful in tests to ensure the goroutine has stopped writing to the logger.
func (d *HeartbeatDetectorService) Wait() {
	d.wg.Wait()
}
