package store_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/database"
	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
)

// BenchmarkIncidentGetIncidentStats_Paired runs GetIncidentStats(24) via
// both impls on the same seeded incident fixture, gates p95 ratio ≤ 1.10.
// Spec 049 §FR-007 / SC-003. Runs on both dialects.
func BenchmarkIncidentGetIncidentStats_Paired(b *testing.B) {
	b.Run("sqlite", func(b *testing.B) {
		runIncidentStatsPaired(b, benchOpenSQLite(b), "BenchmarkIncidentGetIncidentStats_Paired/sqlite")
	})
	b.Run("postgres", func(b *testing.B) {
		fx := internaltest.SetupPostgres(b)
		if fx == nil {
			return
		}
		runIncidentStatsPaired(b, fx.Runtime, "BenchmarkIncidentGetIncidentStats_Paired/postgres")
	})
}

func runIncidentStatsPaired(b *testing.B, rt *database.Runtime, name string) {
	b.Helper()
	ctx := context.Background()

	resRepo := store.NewResourceRepository(rt.GormDB())
	_, err := resRepo.Create(ctx, &domain.Resource{
		Base:     domain.Base{ID: "bench-stats-res", CreatedAt: time.Now()},
		Name:     "bench-stats-res",
		Type:     domain.ResourceHTTP,
		Target:   "https://example.com",
		IsActive: true,
	})
	require.NoError(b, err)

	incRepo := store.NewIncidentRepository(rt.GormDB())
	now := time.Now()
	for i := 0; i < 1000; i++ {
		started := now.Add(-time.Duration(i) * time.Minute)
		_, err := incRepo.Create(ctx, &domain.Incident{
			Base:       domain.Base{ID: fmt.Sprintf("bench-stats-inc-%04d", i), CreatedAt: started},
			ResourceID: "bench-stats-res",
			StartedAt:  started,
		})
		require.NoError(b, err)
	}

	gormRepo := store.NewIncidentRepository(rt.GormDB())
	sqlcRepo := store.NewIncidentRepositorySQLC(rt)
	internaltest.RunPairedBench(b, name,
		func() { _, _, _ = gormRepo.GetIncidentStats(ctx, 24) },
		func() { _, _, _ = sqlcRepo.GetIncidentStats(ctx, 24) },
	)
}
