package store_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/database"
	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/store"
)

func benchSeedIncident(b *testing.B, rt *database.Runtime, resourceID, incidentID string) {
	b.Helper()
	res := &domain.Resource{
		Base:   domain.Base{ID: resourceID},
		Name:   "bench",
		Type:   domain.ResourceHTTP,
		Target: "https://example.invalid/",
	}
	require.NoError(b, rt.GormDB().Create(res).Error)
	inc := &domain.Incident{
		Base:       domain.Base{ID: incidentID},
		ResourceID: resourceID,
		Cause:      "test",
		StartedAt:  time.Now(),
	}
	require.NoError(b, rt.GormDB().Create(inc).Error)
}

func BenchmarkIncidentDiagnostics_Create_GORM(b *testing.B) {
	rt := benchOpenSQLite(b)
	repo := store.NewIncidentDiagnosticsRepository(rt.GormDB())
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resID := fmt.Sprintf("01BRESG%020d", i)
		incID := fmt.Sprintf("01BINCG%020d", i)
		b.StopTimer()
		benchSeedIncident(b, rt, resID, incID)
		b.StartTimer()
		_, _ = repo.Create(ctx, &domain.IncidentDiagnostics{
			IncidentID:      incID,
			RequestMethod:   "GET",
			RequestHeaders:  map[string]string{},
			ResponseHeaders: map[string]string{},
		})
	}
}

func BenchmarkIncidentDiagnostics_Create_SQLC(b *testing.B) {
	rt := benchOpenSQLite(b)
	repo := store.NewIncidentDiagnosticsRepositorySQLC(rt)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resID := fmt.Sprintf("01BRESS%020d", i)
		incID := fmt.Sprintf("01BINCS%020d", i)
		b.StopTimer()
		benchSeedIncident(b, rt, resID, incID)
		b.StartTimer()
		_, _ = repo.Create(ctx, &domain.IncidentDiagnostics{
			IncidentID:      incID,
			RequestMethod:   "GET",
			RequestHeaders:  map[string]string{},
			ResponseHeaders: map[string]string{},
		})
	}
}

func BenchmarkIncidentDiagnostics_FindByIncidentID_GORM(b *testing.B) {
	rt := benchOpenSQLite(b)
	repo := store.NewIncidentDiagnosticsRepository(rt.GormDB())
	ctx := context.Background()
	resID := "res-diag-bench"
	incID := "inc-diag-bench"
	benchSeedIncident(b, rt, resID, incID)
	_, _ = repo.Create(ctx, &domain.IncidentDiagnostics{
		IncidentID:      incID,
		RequestMethod:   "GET",
		RequestHeaders:  map[string]string{},
		ResponseHeaders: map[string]string{},
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.FindByIncidentID(ctx, incID)
	}
}

func BenchmarkIncidentDiagnostics_FindByIncidentID_SQLC(b *testing.B) {
	rt := benchOpenSQLite(b)
	repo := store.NewIncidentDiagnosticsRepositorySQLC(rt)
	ctx := context.Background()
	resID := "res-diag-bench"
	incID := "inc-diag-bench"
	benchSeedIncident(b, rt, resID, incID)
	_, _ = repo.Create(ctx, &domain.IncidentDiagnostics{
		IncidentID:      incID,
		RequestMethod:   "GET",
		RequestHeaders:  map[string]string{},
		ResponseHeaders: map[string]string{},
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.FindByIncidentID(ctx, incID)
	}
}
