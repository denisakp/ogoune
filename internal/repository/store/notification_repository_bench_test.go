package store_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/database"
	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository/store"
)

func benchSeedNotifDeps(b *testing.B, rt *database.Runtime) {
	b.Helper()
	require.NoError(b, rt.GormDB().Create(&domain.Resource{
		Base: domain.Base{ID: "res-notif-bench"}, Name: "bench", Type: domain.ResourceHTTP, Target: "https://x.invalid",
	}).Error)
	require.NoError(b, rt.GormDB().Create(&domain.Incident{
		Base: domain.Base{ID: "inc-notif-bench"}, ResourceID: "res-notif-bench", Cause: "bench",
	}).Error)
}

func benchSeedNotifPending(b *testing.B, repo port.NotificationRepository, n int) {
	b.Helper()
	ctx := context.Background()
	for i := 0; i < n; i++ {
		require.NoError(b, repo.Create(ctx, &domain.NotificationEvent{
			Base: domain.Base{ID: fmt.Sprintf("01BNTFP%020d", i)},
			IncidentID: "inc-notif-bench",
			Type:       domain.NotificationEventTypeDown,
			Status:     domain.NotificationEventStatusPending,
		}))
	}
}

func BenchmarkNotificationRepository_Create_GORM(b *testing.B) {
	rt := benchOpenSQLite(b)
	benchSeedNotifDeps(b, rt)
	repo := store.NewNotificationRepository(rt.GormDB())
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.Create(ctx, &domain.NotificationEvent{
			Base: domain.Base{ID: fmt.Sprintf("01BNTFG%020d", i)},
			IncidentID: "inc-notif-bench",
			Type:       domain.NotificationEventTypeDown,
			Status:     domain.NotificationEventStatusPending,
		})
	}
}

func BenchmarkNotificationRepository_Create_SQLC(b *testing.B) {
	rt := benchOpenSQLite(b)
	benchSeedNotifDeps(b, rt)
	repo := store.NewNotificationRepositorySQLC(rt)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.Create(ctx, &domain.NotificationEvent{
			Base: domain.Base{ID: fmt.Sprintf("01BNTFS%020d", i)},
			IncidentID: "inc-notif-bench",
			Type:       domain.NotificationEventTypeDown,
			Status:     domain.NotificationEventStatusPending,
		})
	}
}

func BenchmarkNotificationRepository_FindPending_GORM(b *testing.B) {
	rt := benchOpenSQLite(b)
	benchSeedNotifDeps(b, rt)
	repo := store.NewNotificationRepository(rt.GormDB())
	benchSeedNotifPending(b, repo, 50)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.FindPending(ctx, 25, 0)
	}
}

func BenchmarkNotificationRepository_FindPending_SQLC(b *testing.B) {
	rt := benchOpenSQLite(b)
	benchSeedNotifDeps(b, rt)
	repo := store.NewNotificationRepositorySQLC(rt)
	benchSeedNotifPending(b, repo, 50)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.FindPending(ctx, 25, 0)
	}
}
