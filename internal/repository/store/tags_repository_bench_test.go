package store_test

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/database"
	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository/store"
)

// SC-005: per-call latency of sqlc impl within ±20% of GORM for Create,
// FindByID, List on the shared SQLite fixture.
//
// We deliberately don't reuse internaltest.SetupSQLite here because that
// helper takes *testing.T; benches use *testing.B. A small inline opener
// keeps the bench self-contained.

func benchOpenSQLite(b *testing.B) *database.Runtime {
	b.Helper()
	rt, err := database.Open(context.Background(), database.Config{
		Driver:     database.DriverSQLite,
		SQLitePath: filepath.Join(b.TempDir(), "bench.db"),
		LogLevel:   "silent",
	})
	require.NoError(b, err)
	b.Cleanup(func() {
		if sqlDB, derr := rt.GormDB().DB(); derr == nil && sqlDB != nil {
			_ = sqlDB.Close()
		}
	})
	return rt
}

func benchSeed(b *testing.B, repo port.TagsRepository, n int) []string {
	b.Helper()
	ctx := context.Background()
	ids := make([]string, n)
	for i := 0; i < n; i++ {
		id := fmt.Sprintf("01BENCH%020d", i)
		ids[i] = id
		err := repo.Create(ctx, &domain.Tags{
			Base: domain.Base{ID: id, CreatedAt: time.Now()},
			Name: fmt.Sprintf("seed-%d", i),
		})
		if err != nil {
			b.Fatalf("seed: %v", err)
		}
	}
	return ids
}

func BenchmarkTagsRepository_Create_GORM(b *testing.B) {
	rt := benchOpenSQLite(b)
	repo := store.NewTagsRepository(rt.GormDB())
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.Create(ctx, &domain.Tags{
			Base: domain.Base{ID: fmt.Sprintf("01BCG%021d", i), CreatedAt: time.Now()},
			Name: fmt.Sprintf("g-%d", i),
		})
	}
}

func BenchmarkTagsRepository_Create_SQLC(b *testing.B) {
	rt := benchOpenSQLite(b)
	repo := store.NewTagsRepositorySQLC(rt)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.Create(ctx, &domain.Tags{
			Base: domain.Base{ID: fmt.Sprintf("01BCS%021d", i), CreatedAt: time.Now()},
			Name: fmt.Sprintf("s-%d", i),
		})
	}
}

func BenchmarkTagsRepository_FindByID_GORM(b *testing.B) {
	rt := benchOpenSQLite(b)
	repo := store.NewTagsRepository(rt.GormDB())
	ids := benchSeed(b, repo, 200)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.FindByID(ctx, ids[i%len(ids)])
	}
}

func BenchmarkTagsRepository_FindByID_SQLC(b *testing.B) {
	rt := benchOpenSQLite(b)
	repo := store.NewTagsRepositorySQLC(rt)
	ids := benchSeed(b, repo, 200)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.FindByID(ctx, ids[i%len(ids)])
	}
}

func BenchmarkTagsRepository_List_GORM(b *testing.B) {
	rt := benchOpenSQLite(b)
	repo := store.NewTagsRepository(rt.GormDB())
	benchSeed(b, repo, 100)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.List(ctx, 50, 0)
	}
}

func BenchmarkTagsRepository_List_SQLC(b *testing.B) {
	rt := benchOpenSQLite(b)
	repo := store.NewTagsRepositorySQLC(rt)
	benchSeed(b, repo, 100)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.List(ctx, 50, 0)
	}
}
