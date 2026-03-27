package monitoring

import (
	"context"
	"testing"
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/repository/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlap_DetectedAfterNTransitions(t *testing.T) {
	activityRepo := fake.NewMonitoringActivityFake()
	resourceID := "res-flap-detector"
	statuses := []bool{false, true, false}
	for _, success := range statuses {
		activity := &domain.MonitoringActivity{
			Base:       domain.Base{CreatedAt: time.Now()},
			ResourceID: resourceID,
			Success:    success,
		}
		require.NoError(t, activityRepo.Create(context.Background(), activity))
	}

	detector := NewFlapDetector(activityRepo, FlapConfig{
		Enabled:       true,
		Threshold:     2,
		WindowSeconds: 600,
	})

	transitions, err := detector.Evaluate(context.Background(), resourceID, time.Now().Add(-10*time.Minute))
	require.NoError(t, err)
	assert.GreaterOrEqual(t, transitions, 2)
}
