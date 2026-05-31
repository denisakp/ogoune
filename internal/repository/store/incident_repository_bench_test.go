package store_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
)

// BenchmarkIncidentGetIncidentStats_Paired runs GetIncidentStats(24) via
// both impls on the same seeded incident fixture, gates p95 ratio ≤ 1.10.
// Spec 049 §FR-007 / SC-003.
//
// Smaller seed than the resource bench: incidents stats are O(N) over
// incidents-in-window, not over resources×relations. 1000 incidents over
// 24h is realistic and fits well under the memory cap.
func BenchmarkIncidentGetIncidentStats_Paired(b *testing.B) {
	rt := benchOpenSQLite(b)
	ctx := context.Background()

	// Seed 1 resource (FK target) + 1000 incidents in the last 24h.
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
		started := now.Add(-time.Duration(i) * time.Minute) // spread across last ~16h
		_, err := incRepo.Create(ctx, &domain.Incident{
			Base:       domain.Base{ID: fmt.Sprintf("bench-stats-inc-%04d", i), CreatedAt: started},
			ResourceID: "bench-stats-res",
			StartedAt:  started,
		})
		require.NoError(b, err)
	}

	gormRepo := store.NewIncidentRepository(rt.GormDB())
	sqlcRepo := store.NewIncidentRepositorySQLC(rt)

	internaltest.RunPairedBench(b, "BenchmarkIncidentGetIncidentStats_Paired",
		func() { _, _, _ = gormRepo.GetIncidentStats(ctx, 24) },
		func() { _, _, _ = sqlcRepo.GetIncidentStats(ctx, 24) },
	)
}
