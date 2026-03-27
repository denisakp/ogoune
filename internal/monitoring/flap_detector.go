package monitoring

import (
	"context"
	"time"

	"github.com/denisakp/pulseguard/internal/repository"
)

type FlapConfig struct {
	Enabled            bool
	Threshold          int
	WindowSeconds      int
	MaxDurationMinutes int
}

type FlapDetector struct {
	activities repository.MonitoringActivityRepository
	cfg        FlapConfig
}

func NewFlapDetector(activities repository.MonitoringActivityRepository, cfg FlapConfig) *FlapDetector {
	return &FlapDetector{activities: activities, cfg: cfg}
}

func (f *FlapDetector) Evaluate(ctx context.Context, resourceID string, windowStart time.Time) (int, error) {
	if !f.cfg.Enabled || f.activities == nil {
		return 0, nil
	}
	return f.activities.CountTransitionsInWindow(ctx, resourceID, windowStart)
}

func (f *FlapDetector) ShouldForceIncident(flapStartedAt *time.Time) bool {
	if flapStartedAt == nil || f.cfg.MaxDurationMinutes <= 0 {
		return false
	}
	return time.Since(*flapStartedAt) >= time.Duration(f.cfg.MaxDurationMinutes)*time.Minute
}
