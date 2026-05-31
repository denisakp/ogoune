package store_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/database"
	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/store"
)

func benchSeedEventStepDeps(b *testing.B, rt *database.Runtime, resourceID, incidentID string) {
	b.Helper()
	require.NoError(b, rt.GormDB().Create(&domain.Resource{
		Base: domain.Base{ID: resourceID}, Name: "bench", Type: domain.ResourceHTTP, Target: "https://x.invalid",
	}).Error)
	require.NoError(b, rt.GormDB().Create(&domain.Incident{
		Base: domain.Base{ID: incidentID}, ResourceID: resourceID, Cause: "bench",
	}).Error)
}

func BenchmarkIncidentEventStep_Create_GORM(b *testing.B) {
	rt := benchOpenSQLite(b)
	benchSeedEventStepDeps(b, rt, "res-ies-bench", "inc-ies-bench")
	repo := store.NewIncidentEventStepRepository(rt.GormDB())
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.Create(ctx, &domain.IncidentEventStep{
			Base:       domain.Base{ID: fmt.Sprintf("01BIESG%020d", i)},
			IncidentID: "inc-ies-bench",
			Step:       domain.IncidentEventStepType("detected"),
		})
	}
}

func BenchmarkIncidentEventStep_Create_SQLC(b *testing.B) {
	rt := benchOpenSQLite(b)
	benchSeedEventStepDeps(b, rt, "res-ies-bench", "inc-ies-bench")
	repo := store.NewIncidentEventStepRepositorySQLC(rt)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.Create(ctx, &domain.IncidentEventStep{
			Base:       domain.Base{ID: fmt.Sprintf("01BIESS%020d", i)},
			IncidentID: "inc-ies-bench",
			Step:       domain.IncidentEventStepType("detected"),
		})
	}
}

func BenchmarkIncidentEventStep_FindByID_GORM(b *testing.B) {
	rt := benchOpenSQLite(b)
	benchSeedEventStepDeps(b, rt, "res-ies-bench", "inc-ies-bench")
	repo := store.NewIncidentEventStepRepository(rt.GormDB())
	ctx := context.Background()
	step, _ := repo.Create(ctx, &domain.IncidentEventStep{
		Base: domain.Base{ID: "step-ies-bench"}, IncidentID: "inc-ies-bench", Step: "detected",
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.FindByID(ctx, step.ID)
	}
}

func BenchmarkIncidentEventStep_FindByID_SQLC(b *testing.B) {
	rt := benchOpenSQLite(b)
	benchSeedEventStepDeps(b, rt, "res-ies-bench", "inc-ies-bench")
	repo := store.NewIncidentEventStepRepositorySQLC(rt)
	ctx := context.Background()
	step, _ := repo.Create(ctx, &domain.IncidentEventStep{
		Base: domain.Base{ID: "step-ies-bench"}, IncidentID: "inc-ies-bench", Step: "detected",
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.FindByID(ctx, step.ID)
	}
}
