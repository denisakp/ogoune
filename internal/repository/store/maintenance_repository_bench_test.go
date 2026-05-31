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

func benchSeedResourceMtc(b *testing.B, rt *database.Runtime, id string) {
	b.Helper()
	require.NoError(b, rt.GormDB().Create(&domain.Resource{
		Base: domain.Base{ID: id}, Name: id, Type: domain.ResourceHTTP, Target: "https://x.invalid/" + id,
	}).Error)
}

func benchSeedMaintenances(b *testing.B, repo port.MaintenanceRepository, resID string, n int) {
	b.Helper()
	ctx := context.Background()
	now := time.Now()
	start := now.Add(-time.Hour)
	end := now.Add(time.Hour)
	for i := 0; i < n; i++ {
		_, err := repo.Create(ctx, &domain.Maintenance{
			Base:     domain.Base{ID: fmt.Sprintf("01BMTCLST%016d", i)},
			Title:    "seed",
			Strategy: domain.OneTime,
			Status:   "active",
			StartAt:  &start,
			EndAt:    &end,
			Resources: []*domain.Resource{{Base: domain.Base{ID: resID}}},
		})
		require.NoError(b, err)
	}
}

func BenchmarkMaintenanceRepository_Create_GORM(b *testing.B) {
	rt := benchOpenSQLite(b)
	benchSeedResourceMtc(b, rt, "res-mtc-bench")
	repo := store.NewMaintenanceRepository(rt.GormDB())
	ctx := context.Background()
	now := time.Now()
	start := now.Add(-time.Hour)
	end := now.Add(time.Hour)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.Create(ctx, &domain.Maintenance{
			Base: domain.Base{ID: fmt.Sprintf("01BMTCG%020d", i)},
			Title: "g", Strategy: domain.OneTime, Status: "scheduled",
			StartAt: &start, EndAt: &end,
			Resources: []*domain.Resource{{Base: domain.Base{ID: "res-mtc-bench"}}},
		})
	}
}

func BenchmarkMaintenanceRepository_Create_SQLC(b *testing.B) {
	rt := benchOpenSQLite(b)
	benchSeedResourceMtc(b, rt, "res-mtc-bench")
	repo := store.NewMaintenanceRepositorySQLC(rt)
	ctx := context.Background()
	now := time.Now()
	start := now.Add(-time.Hour)
	end := now.Add(time.Hour)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.Create(ctx, &domain.Maintenance{
			Base: domain.Base{ID: fmt.Sprintf("01BMTCS%020d", i)},
			Title: "s", Strategy: domain.OneTime, Status: "scheduled",
			StartAt: &start, EndAt: &end,
			Resources: []*domain.Resource{{Base: domain.Base{ID: "res-mtc-bench"}}},
		})
	}
}

func BenchmarkMaintenanceRepository_FindActiveForResource_GORM(b *testing.B) {
	rt := benchOpenSQLite(b)
	benchSeedResourceMtc(b, rt, "res-mtc-bench")
	repo := store.NewMaintenanceRepository(rt.GormDB())
	benchSeedMaintenances(b, repo, "res-mtc-bench", 20)
	ctx := context.Background()
	now := time.Now()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.FindActiveForResource(ctx, "res-mtc-bench", now)
	}
}

func BenchmarkMaintenanceRepository_FindActiveForResource_SQLC(b *testing.B) {
	rt := benchOpenSQLite(b)
	benchSeedResourceMtc(b, rt, "res-mtc-bench")
	repo := store.NewMaintenanceRepositorySQLC(rt)
	benchSeedMaintenances(b, repo, "res-mtc-bench", 20)
	ctx := context.Background()
	now := time.Now()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.FindActiveForResource(ctx, "res-mtc-bench", now)
	}
}
