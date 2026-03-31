package service

import (
	"context"
	"log"
	"math"
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
		log.Printf("[heartbeat-detector] query failed: %v", err)
		return err
	}

	if len(missed) == 0 {
		log.Printf("[heartbeat-detector] run complete: 0 missed monitors")
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
			log.Printf("[heartbeat-detector] incident call failed for monitor %s (%s): %v — retry scheduled after %s",
				resource.ID, resource.Name, err, d.CalculateBackoffDelay(0))
			failed++
		} else {
			log.Printf("[heartbeat-detector] incident created for monitor %s (%s)", resource.ID, resource.Name)
			succeeded++
		}
	}

	log.Printf("[heartbeat-detector] run complete: missed=%d incidents_created=%d failures=%d",
		len(missed), succeeded, failed)
	return nil
}

// Start launches the detector as a recurring goroutine on the given interval.
// It blocks until ctx is cancelled. The first detection cycle runs after one interval.
// Returns an error immediately if ctx is already cancelled, ensuring fail-fast startup detection.
func (d *HeartbeatDetectorService) Start(ctx context.Context, interval time.Duration) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Println("[heartbeat-detector] stopped")
				return
			case <-ticker.C:
				if err := d.Detect(ctx); err != nil {
					log.Printf("[heartbeat-detector] detection cycle error: %v", err)
				}
			}
		}
	}()

	log.Printf("[heartbeat-detector] started (interval=%s)", interval)
	return nil
}
