package store_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository/store"
)

func benchSPSSeed(b *testing.B, repo port.StatusPageSettingsRepository) {
	b.Helper()
	require.NoError(b, repo.Upsert(context.Background(), &domain.StatusPageSettings{
		Name:                 "bench page",
		EnableDetailsPage:    true,
		ShowUptimePercentage: true,
		HidePausedMonitors:   true,
		ShowIncidentHistory:  true,
	}))
}

func BenchmarkStatusPageSettings_Upsert_GORM(b *testing.B) {
	rt := benchOpenSQLite(b)
	repo := store.NewStatusPageSettingsRepository(rt.GormDB())
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.Upsert(ctx, &domain.StatusPageSettings{
			Name:                 fmt.Sprintf("g-%d", i),
			EnableDetailsPage:    true,
			ShowUptimePercentage: true,
			HidePausedMonitors:   true,
			ShowIncidentHistory:  true,
		})
	}
}

func BenchmarkStatusPageSettings_Upsert_SQLC(b *testing.B) {
	rt := benchOpenSQLite(b)
	repo := store.NewStatusPageSettingsRepositorySQLC(rt)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.Upsert(ctx, &domain.StatusPageSettings{
			Name:                 fmt.Sprintf("s-%d", i),
			EnableDetailsPage:    true,
			ShowUptimePercentage: true,
			HidePausedMonitors:   true,
			ShowIncidentHistory:  true,
		})
	}
}

func BenchmarkStatusPageSettings_Get_GORM(b *testing.B) {
	rt := benchOpenSQLite(b)
	repo := store.NewStatusPageSettingsRepository(rt.GormDB())
	benchSPSSeed(b, repo)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.Get(ctx)
	}
}

func BenchmarkStatusPageSettings_Get_SQLC(b *testing.B) {
	rt := benchOpenSQLite(b)
	repo := store.NewStatusPageSettingsRepositorySQLC(rt)
	benchSPSSeed(b, repo)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.Get(ctx)
	}
}
