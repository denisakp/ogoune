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

func TestUptimeAggregator_RecomputesTodayFromActivity(t *testing.T) {
	resources := fake.NewResourceFake()
	activities := fake.NewMonitoringActivityFake()
	aggs := fake.NewUptimeDailyAggRepository()

	r := &domain.Resource{Name: "api", Target: "api.acme.com", IsActive: true, Status: domain.StatusUp}
	r.ID = "res-1"
	_, err := resources.Create(context.Background(), r)
	require.NoError(t, err)

	today := time.Date(2026, 6, 4, 0, 0, 0, 0, time.UTC)
	mid := today.Add(12 * time.Hour)

	// 8 ups + 2 downs on today.
	for i := 0; i < 10; i++ {
		act := &domain.MonitoringActivity{ResourceID: "res-1", Success: i >= 2}
		act.CreatedAt = mid.Add(time.Duration(i) * time.Minute)
		_ = activities.Create(context.Background(), act)
	}
	// 1 sample on yesterday — must not leak into today's bucket.
	yesterAct := &domain.MonitoringActivity{ResourceID: "res-1", Success: false}
	yesterAct.CreatedAt = today.AddDate(0, 0, -1).Add(6 * time.Hour)
	_ = activities.Create(context.Background(), yesterAct)

	agg := NewUptimeAggregator(resources, activities, aggs)
	agg.SetClock(func() time.Time { return mid.Add(time.Hour) })

	require.NoError(t, agg.RunOnce(context.Background()))

	rows, err := aggs.FindForResource(context.Background(), "res-1", today, today)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, 10, rows[0].Samples)
	assert.Equal(t, 8, rows[0].Up)
	assert.Equal(t, 2, rows[0].Down)
	assert.InDelta(t, 0.8, rows[0].UptimeRatio, 0.0001)
}

func TestUptimeAggregator_Idempotent(t *testing.T) {
	resources := fake.NewResourceFake()
	activities := fake.NewMonitoringActivityFake()
	aggs := fake.NewUptimeDailyAggRepository()

	r := &domain.Resource{Name: "api", Target: "api.acme.com", IsActive: true}
	r.ID = "res-1"
	_, _ = resources.Create(context.Background(), r)

	today := time.Date(2026, 6, 4, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 5; i++ {
		act := &domain.MonitoringActivity{ResourceID: "res-1", Success: true}
		act.CreatedAt = today.Add(time.Hour).Add(time.Duration(i) * time.Minute)
		_ = activities.Create(context.Background(), act)
	}

	agg := NewUptimeAggregator(resources, activities, aggs)
	agg.SetClock(func() time.Time { return today.Add(2 * time.Hour) })

	require.NoError(t, agg.RunOnce(context.Background()))
	require.NoError(t, agg.RunOnce(context.Background()))

	rows, err := aggs.FindForResource(context.Background(), "res-1", today, today)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, 5, rows[0].Samples)
	assert.InDelta(t, 1.0, rows[0].UptimeRatio, 0.0001)
}
