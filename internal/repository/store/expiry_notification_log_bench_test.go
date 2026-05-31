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

func benchSeedResource(b *testing.B, rt *database.Runtime, id string) {
	b.Helper()
	res := &domain.Resource{
		Base:   domain.Base{ID: id},
		Name:   "bench-" + id,
		Type:   domain.ResourceHTTP,
		Target: "https://example.invalid/" + id,
	}
	require.NoError(b, rt.GormDB().Create(res).Error)
}

func benchENLSeed(b *testing.B, repo port.ExpiryNotificationLogRepository, resourceID string, n int) {
	b.Helper()
	ctx := context.Background()
	for i := 0; i < n; i++ {
		require.NoError(b, repo.Create(ctx, &domain.ExpiryNotificationLog{
			Base:       domain.Base{ID: fmt.Sprintf("01BENL%020d", i)},
			ResourceID: resourceID,
			ExpiryType: "ssl",
			Threshold:  i + 1,
			SentAt:     time.Now(),
		}))
	}
}

func BenchmarkExpiryNotificationLog_Create_GORM(b *testing.B) {
	rt := benchOpenSQLite(b)
	benchSeedResource(b, rt, "res-benl")
	repo := store.NewExpiryNotificationLogRepository(rt.GormDB())
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.Create(ctx, &domain.ExpiryNotificationLog{
			Base:       domain.Base{ID: fmt.Sprintf("01BENLG%020d", i)},
			ResourceID: "res-benl",
			ExpiryType: "ssl",
			Threshold:  i,
			SentAt:     time.Now(),
		})
	}
}

func BenchmarkExpiryNotificationLog_Create_SQLC(b *testing.B) {
	rt := benchOpenSQLite(b)
	benchSeedResource(b, rt, "res-benl")
	repo := store.NewExpiryNotificationLogRepositorySQLC(rt)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.Create(ctx, &domain.ExpiryNotificationLog{
			Base:       domain.Base{ID: fmt.Sprintf("01BENLS%020d", i)},
			ResourceID: "res-benl",
			ExpiryType: "ssl",
			Threshold:  i,
			SentAt:     time.Now(),
		})
	}
}

func BenchmarkExpiryNotificationLog_CountByKey_GORM(b *testing.B) {
	rt := benchOpenSQLite(b)
	benchSeedResource(b, rt, "res-benl")
	repo := store.NewExpiryNotificationLogRepository(rt.GormDB())
	benchENLSeed(b, repo, "res-benl", 100)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.CountByKey(ctx, "res-benl", "ssl", i%100)
	}
}

func BenchmarkExpiryNotificationLog_CountByKey_SQLC(b *testing.B) {
	rt := benchOpenSQLite(b)
	benchSeedResource(b, rt, "res-benl")
	repo := store.NewExpiryNotificationLogRepositorySQLC(rt)
	benchENLSeed(b, repo, "res-benl", 100)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.CountByKey(ctx, "res-benl", "ssl", i%100)
	}
}
