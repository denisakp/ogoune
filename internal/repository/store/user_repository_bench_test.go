package store_test

import (
	"context"
	"fmt"
	"testing"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository/store"
)

func benchUserSeed(b *testing.B, repo port.UserRepository, n int) {
	b.Helper()
	ctx := context.Background()
	for i := 0; i < n; i++ {
		_, err := repo.Create(ctx, &domain.User{
			Base:           domain.Base{ID: fmt.Sprintf("01BUSRLIST%016d", i)},
			Email:          fmt.Sprintf("seed-%d@bench.invalid", i),
			HashedPassword: "h",
		})
		if err != nil {
			b.Fatalf("seed user: %v", err)
		}
	}
}

func BenchmarkUserRepository_Create_GORM(b *testing.B) {
	rt := benchOpenSQLite(b)
	repo := store.NewUserRepository(rt.GormDB())
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.Create(ctx, &domain.User{
			Base:           domain.Base{ID: fmt.Sprintf("01BUSRG%020d", i)},
			Email:          fmt.Sprintf("g-%d@bench.invalid", i),
			HashedPassword: "h",
		})
	}
}

func BenchmarkUserRepository_Create_SQLC(b *testing.B) {
	rt := benchOpenSQLite(b)
	repo := store.NewUserRepositorySQLC(rt)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.Create(ctx, &domain.User{
			Base:           domain.Base{ID: fmt.Sprintf("01BUSRS%020d", i)},
			Email:          fmt.Sprintf("s-%d@bench.invalid", i),
			HashedPassword: "h",
		})
	}
}

func BenchmarkUserRepository_FindByEmail_GORM(b *testing.B) {
	rt := benchOpenSQLite(b)
	repo := store.NewUserRepository(rt.GormDB())
	benchUserSeed(b, repo, 100)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.FindByEmail(ctx, fmt.Sprintf("seed-%d@bench.invalid", i%100))
	}
}

func BenchmarkUserRepository_FindByEmail_SQLC(b *testing.B) {
	rt := benchOpenSQLite(b)
	repo := store.NewUserRepositorySQLC(rt)
	benchUserSeed(b, repo, 100)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.FindByEmail(ctx, fmt.Sprintf("seed-%d@bench.invalid", i%100))
	}
}
