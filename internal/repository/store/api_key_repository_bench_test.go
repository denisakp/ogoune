package store_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/database"
	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository/store"
)

func benchSeedUser(b *testing.B, rt *database.Runtime, userID string) {
	b.Helper()
	u := &domain.User{
		Base:           domain.Base{ID: userID},
		Email:          userID + "@bench.invalid",
		HashedPassword: "h",
	}
	require.NoError(b, rt.GormDB().Create(u).Error)
}

func benchRandHex(b *testing.B) string {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		b.Fatalf("randHex: %v", err)
	}
	return hex.EncodeToString(buf[:])
}

func benchAPIKeySeed(b *testing.B, repo port.APIKeyRepository, userID string, n int) {
	b.Helper()
	ctx := context.Background()
	for i := 0; i < n; i++ {
		require.NoError(b, repo.Create(ctx, &domain.APIKey{
			Base:      domain.Base{ID: fmt.Sprintf("01BAKLIST%017d", i)},
			UserID:    userID,
			Name:      fmt.Sprintf("seed-%d", i),
			KeyHash:   benchRandHex(b),
			KeyPrefix: "test",
			Scope:     domain.APIKeyScopeRead,
			IsActive:  true,
		}))
	}
}

func BenchmarkAPIKeyRepository_Create_GORM(b *testing.B) {
	rt := benchOpenSQLite(b)
	benchSeedUser(b, rt, "user-bench")
	repo := store.NewAPIKeyRepository(rt.GormDB())
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.Create(ctx, &domain.APIKey{
			Base:      domain.Base{ID: fmt.Sprintf("01BAKG%021d", i)},
			UserID:    "user-bench",
			Name:      "g",
			KeyHash:   benchRandHex(b),
			KeyPrefix: "g",
			Scope:     domain.APIKeyScopeRead,
			IsActive:  true,
		})
	}
}

func BenchmarkAPIKeyRepository_Create_SQLC(b *testing.B) {
	rt := benchOpenSQLite(b)
	benchSeedUser(b, rt, "user-bench")
	repo := store.NewAPIKeyRepositorySQLC(rt)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.Create(ctx, &domain.APIKey{
			Base:      domain.Base{ID: fmt.Sprintf("01BAKS%021d", i), CreatedAt: time.Now(), UpdatedAt: time.Now()},
			UserID:    "user-bench",
			Name:      "s",
			KeyHash:   benchRandHex(b),
			KeyPrefix: "s",
			Scope:     domain.APIKeyScopeRead,
			IsActive:  true,
		})
	}
}

func BenchmarkAPIKeyRepository_ListByUserID_GORM(b *testing.B) {
	rt := benchOpenSQLite(b)
	benchSeedUser(b, rt, "user-bench")
	repo := store.NewAPIKeyRepository(rt.GormDB())
	benchAPIKeySeed(b, repo, "user-bench", 50)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.ListByUserID(ctx, "user-bench")
	}
}

func BenchmarkAPIKeyRepository_ListByUserID_SQLC(b *testing.B) {
	rt := benchOpenSQLite(b)
	benchSeedUser(b, rt, "user-bench")
	repo := store.NewAPIKeyRepositorySQLC(rt)
	benchAPIKeySeed(b, repo, "user-bench", 50)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.ListByUserID(ctx, "user-bench")
	}
}
