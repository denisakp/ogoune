package monitoring

import (
	"context"
	"time"

	"github.com/denisakp/ogoune/internal/port"
)

type FlapConfig struct {
	Enabled            bool
	Threshold          int
	WindowSeconds      int
	MaxDurationMinutes int
}

type FlapDetector struct {
	activities port.MonitoringActivityRepository
	cfg        FlapConfig
}

func NewFlapDetector(activities port.MonitoringActivityRepository, cfg FlapConfig) *FlapDetector {
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
