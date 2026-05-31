package store_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/database"
	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository/store"
)

func benchSeedMAResource(b *testing.B, rt *database.Runtime) {
	b.Helper()
	require.NoError(b, rt.GormDB().Create(&domain.Resource{
		Base: domain.Base{ID: "res-ma-bench"}, Name: "bench", Type: domain.ResourceHTTP, Target: "https://x.invalid",
	}).Error)
}

func benchSeedMAs(b *testing.B, repo port.MonitoringActivityRepository, n int) {
	b.Helper()
	ctx := context.Background()
	now := time.Now()
	for i := 0; i < n; i++ {
		require.NoError(b, repo.Create(ctx, &domain.MonitoringActivity{
			Base:         domain.Base{ID: fmt.Sprintf("01BMASD%020d", i), CreatedAt: now.Add(-time.Duration(i*5) * time.Minute)},
			ResourceID:   "res-ma-bench",
			Message:      "seed",
			Success:      i%5 != 0,
			ResponseTime: 100 + i,
		}))
	}
}

func BenchmarkMonitoringActivity_Create_GORM(b *testing.B) {
	rt := benchOpenSQLite(b)
	benchSeedMAResource(b, rt)
	repo := store.NewMonitoringActivityRepository(rt.GormDB())
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.Create(ctx, &domain.MonitoringActivity{
			Base:         domain.Base{ID: fmt.Sprintf("01BMACG%020d", i)},
			ResourceID:   "res-ma-bench",
			Message:      "g",
			Success:      true,
			ResponseTime: 100,
		})
	}
}

func BenchmarkMonitoringActivity_Create_SQLC(b *testing.B) {
	rt := benchOpenSQLite(b)
	benchSeedMAResource(b, rt)
	repo := store.NewMonitoringActivityRepositorySQLC(rt)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.Create(ctx, &domain.MonitoringActivity{
			Base:         domain.Base{ID: fmt.Sprintf("01BMACS%020d", i)},
			ResourceID:   "res-ma-bench",
			Message:      "s",
			Success:      true,
			ResponseTime: 100,
		})
	}
}

func BenchmarkMonitoringActivity_GetUptimeStats_GORM(b *testing.B) {
	rt := benchOpenSQLite(b)
	benchSeedMAResource(b, rt)
	repo := store.NewMonitoringActivityRepository(rt.GormDB())
	benchSeedMAs(b, repo, 200)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.GetUptimeStats(ctx, "res-ma-bench")
	}
}

func BenchmarkMonitoringActivity_GetUptimeStats_SQLC(b *testing.B) {
	rt := benchOpenSQLite(b)
	benchSeedMAResource(b, rt)
	repo := store.NewMonitoringActivityRepositorySQLC(rt)
	benchSeedMAs(b, repo, 200)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.GetUptimeStats(ctx, "res-ma-bench")
	}
}
