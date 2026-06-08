package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
)

// TestResourceRepository_List_AttachesUptimeStats verifies that List populates
// Resource.Uptime30d (from uptime_daily_agg) and Resource.ResponseTimeAvg
// (from monitoring_activities) on the dual-dialect path.
func TestResourceRepository_List_AttachesUptimeStats(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		ctx := context.Background()
		resRepo := store.NewResourceRepositorySQLC(fx.Runtime)
		aggRepo := store.NewUptimeDailyAggRepositorySQLC(fx.Runtime)
		maRepo := store.NewMonitoringActivityRepositorySQLC(fx.Runtime)

		// res-with-stats: 2 daily agg rows (one fully up, one partial) + 3
		// successful monitoring activities. Expected uptime = 380/400 = 0.95,
		// expected avg = (100+200+300)/3 = 200.
		withStats := seedResource(t, fx, "res-stats-with", "with-stats")
		// res-no-stats: no agg rows, no activities → pointers nil.
		noStats := seedResource(t, fx, "res-stats-without", "without-stats")

		today := time.Now().UTC().Truncate(24 * time.Hour)
		require.NoError(t, aggRepo.Upsert(ctx, &domain.UptimeDailyAgg{
			ResourceID: withStats.ID, Day: today,
			Samples: 200, Up: 200, UptimeRatio: 1.0, ComputedAt: time.Now().UTC(),
		}))
		require.NoError(t, aggRepo.Upsert(ctx, &domain.UptimeDailyAgg{
			ResourceID: withStats.ID, Day: today.AddDate(0, 0, -1),
			Samples: 200, Up: 180, Down: 20, UptimeRatio: 0.9, ComputedAt: time.Now().UTC(),
		}))

		for _, rt := range []int{100, 200, 300} {
			require.NoError(t, maRepo.Create(ctx, &domain.MonitoringActivity{
				ResourceID:   withStats.ID,
				Message:      "ok",
				Success:      true,
				ResponseTime: rt,
			}))
		}
		// Failed activity must be excluded from avg.
		require.NoError(t, maRepo.Create(ctx, &domain.MonitoringActivity{
			ResourceID:   withStats.ID,
			Message:      "fail",
			Success:      false,
			ResponseTime: 9999,
		}))

		out, err := resRepo.List(ctx, 100, 0)
		require.NoError(t, err)

		byID := map[string]*domain.Resource{}
		for _, r := range out {
			byID[r.ID] = r
		}

		got := byID[withStats.ID]
		require.NotNil(t, got)
		require.NotNil(t, got.Uptime30d, "uptime_30d should be populated")
		assert.InDelta(t, 0.95, *got.Uptime30d, 1e-6)
		require.NotNil(t, got.Uptime7d, "uptime_7d should be populated")
		assert.InDelta(t, 0.95, *got.Uptime7d, 1e-6)
		require.NotNil(t, got.ResponseTimeAvg, "response_time avg should be populated")
		assert.Equal(t, 200, *got.ResponseTimeAvg)

		empty := byID[noStats.ID]
		require.NotNil(t, empty)
		assert.Nil(t, empty.Uptime30d, "uptime_30d should be nil when no agg rows")
		assert.Nil(t, empty.Uptime7d, "uptime_7d should be nil when no agg rows")
		assert.Nil(t, empty.ResponseTimeAvg, "response_time should be nil when no activities")
	})
}
