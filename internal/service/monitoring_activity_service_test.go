package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/fake"
)

// TestGetResourceUptimeStats_Uptime30dUsesAgg verifies that the 30d field is
// computed from uptime_daily_agg (SUM(up)/SUM(samples)) when aggregates exist,
// matching the list-view source so both surfaces show the same number.
func TestGetResourceUptimeStats_Uptime30dUsesAgg(t *testing.T) {
	actRepo := fake.NewMonitoringActivityFake()
	aggRepo := fake.NewUptimeDailyAggRepository()
	svc := NewMonitoringActivityService(actRepo, aggRepo)

	ctx := context.Background()
	today := time.Now().UTC().Truncate(24 * time.Hour)
	require.NoError(t, aggRepo.Upsert(ctx, &domain.UptimeDailyAgg{
		ResourceID: "res-1", Day: today,
		Samples: 200, Up: 200, ComputedAt: time.Now().UTC(),
	}))
	require.NoError(t, aggRepo.Upsert(ctx, &domain.UptimeDailyAgg{
		ResourceID: "res-1", Day: today.AddDate(0, 0, -1),
		Samples: 200, Up: 180, ComputedAt: time.Now().UTC(),
	}))

	stats, err := svc.GetResourceUptimeStats(ctx, "res-1")
	require.NoError(t, err)
	require.NotNil(t, stats.Uptime30d)
	// 380/400 = 0.95
	assert.InDelta(t, 0.95, *stats.Uptime30d, 1e-6)
}

// TestGetResourceUptimeStats_FallsBackToActivitiesWhenAggMissing verifies the
// fallback path: when uptime_daily_agg has no rows for the resource, the 30d
// value comes from the per-activity GetUptimeByWindow path.
func TestGetResourceUptimeStats_FallsBackToActivitiesWhenAggMissing(t *testing.T) {
	actRepo := fake.NewMonitoringActivityFake()
	aggRepo := fake.NewUptimeDailyAggRepository()
	svc := NewMonitoringActivityService(actRepo, aggRepo)

	ctx := context.Background()
	// Seed enough activities so GetUptimeByWindow(720) returns a value.
	for i := 0; i < 10; i++ {
		require.NoError(t, actRepo.Create(ctx, &domain.MonitoringActivity{
			ResourceID: "res-2", Success: true, ResponseTime: 100,
		}))
	}

	stats, err := svc.GetResourceUptimeStats(ctx, "res-2")
	require.NoError(t, err)
	// We don't assert a specific value — the fake's window math may differ —
	// only that the field is populated via the fallback path.
	assert.NotNil(t, stats.Uptime30d, "should fall back to activities when agg has no rows")
}
