package worker

import (
	"time"

	"github.com/denisakp/ogoune/internal/domain"
)

// noopRecorder is a test-only no-op MetricsRecorder for worker tests.
type noopRecorder struct{}

func (n *noopRecorder) RecordCheck(resourceID, name string, resourceType domain.ResourceType, duration time.Duration, status string) {
}
