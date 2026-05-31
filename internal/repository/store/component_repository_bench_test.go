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

func benchSeedComponents(b *testing.B, repo port.ComponentRepository, n int) {
	b.Helper()
	ctx := context.Background()
	for i := 0; i < n; i++ {
		_, err := repo.Create(ctx, &domain.Component{
			Base: domain.Base{ID: fmt.Sprintf("01BCMPLIST%016d", i)},
			Name: fmt.Sprintf("seed-%d", i),
		})
		require.NoError(b, err)
	}
}

func BenchmarkComponentRepository_Create_GORM(b *testing.B) {
	rt := benchOpenSQLite(b)
	repo := store.NewComponentRepository(rt.GormDB())
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.Create(ctx, &domain.Component{
			Base: domain.Base{ID: fmt.Sprintf("01BCMPG%020d", i)},
			Name: fmt.Sprintf("g-%d", i),
		})
	}
}

func BenchmarkComponentRepository_Create_SQLC(b *testing.B) {
	rt := benchOpenSQLite(b)
	repo := store.NewComponentRepositorySQLC(rt)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.Create(ctx, &domain.Component{
			Base: domain.Base{ID: fmt.Sprintf("01BCMPS%020d", i)},
			Name: fmt.Sprintf("s-%d", i),
		})
	}
}

func BenchmarkComponentRepository_List_GORM(b *testing.B) {
	rt := benchOpenSQLite(b)
	repo := store.NewComponentRepository(rt.GormDB())
	benchSeedComponents(b, repo, 50)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.List(ctx, 25, 0)
	}
}

func BenchmarkComponentRepository_List_SQLC(b *testing.B) {
	rt := benchOpenSQLite(b)
	repo := store.NewComponentRepositorySQLC(rt)
	benchSeedComponents(b, repo, 50)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.List(ctx, 25, 0)
	}
}
