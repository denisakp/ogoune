package store_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/database"
	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/store"
	"github.com/denisakp/ogoune/pkg/crypto"
)

func benchSetupCredKey(b *testing.B) {
	b.Helper()
	b.Setenv("APP_SECRET_KEY", credentialContractTestKey)
	crypto.SetGlobalProvider(&crypto.EnvKeyProvider{})
}

func benchSeedResourcesBatch(b *testing.B, rt *database.Runtime, n int) {
	b.Helper()
	for i := 0; i < n; i++ {
		res := &domain.Resource{
			Base:   domain.Base{ID: fmt.Sprintf("res-cred-bench-%05d", i)},
			Name:   "bench",
			Type:   domain.ResourceHTTP,
			Target: "https://example.invalid/",
		}
		require.NoError(b, rt.GormDB().Create(res).Error)
	}
}

func BenchmarkResourceCredential_Upsert_GORM(b *testing.B) {
	benchSetupCredKey(b)
	rt := benchOpenSQLite(b)
	benchSeedResourcesBatch(b, rt, b.N)
	repo := store.NewResourceCredentialRepository(rt.GormDB())
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.Upsert(ctx, &domain.ResourceCredential{
			ResourceID: fmt.Sprintf("res-cred-bench-%05d", i),
			Username:   "u",
			Password:   []byte("pwd"),
		})
	}
}

func BenchmarkResourceCredential_Upsert_SQLC(b *testing.B) {
	benchSetupCredKey(b)
	rt := benchOpenSQLite(b)
	benchSeedResourcesBatch(b, rt, b.N)
	repo := store.NewResourceCredentialRepositorySQLC(rt)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.Upsert(ctx, &domain.ResourceCredential{
			ResourceID: fmt.Sprintf("res-cred-bench-%05d", i),
			Username:   "u",
			Password:   []byte("pwd"),
		})
	}
}

func BenchmarkResourceCredential_Get_GORM(b *testing.B) {
	benchSetupCredKey(b)
	rt := benchOpenSQLite(b)
	benchSeedResourcesBatch(b, rt, 1)
	repo := store.NewResourceCredentialRepository(rt.GormDB())
	ctx := context.Background()
	require.NoError(b, repo.Upsert(ctx, &domain.ResourceCredential{
		ResourceID: "res-cred-bench-00000",
		Username:   "u",
		Password:   []byte("pwd"),
	}))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.Get(ctx, "res-cred-bench-00000")
	}
}

func BenchmarkResourceCredential_Get_SQLC(b *testing.B) {
	benchSetupCredKey(b)
	rt := benchOpenSQLite(b)
	benchSeedResourcesBatch(b, rt, 1)
	repo := store.NewResourceCredentialRepositorySQLC(rt)
	ctx := context.Background()
	require.NoError(b, repo.Upsert(ctx, &domain.ResourceCredential{
		ResourceID: "res-cred-bench-00000",
		Username:   "u",
		Password:   []byte("pwd"),
	}))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.Get(ctx, "res-cred-bench-00000")
	}
}
