package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
)

func TestAPIKeyRepository_Contract(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		// Seed the user rows referenced by the FK column. Real DBs enforce
		// the FK; the in-memory fake did not.
		seedUsers(t, fx, "user-1", "user-dup", "user-2", "user-3", "user-4",
			"user-list", "user-count", "user-revoke", "user-revoke2", "user-lastused")
		repo := store.NewAPIKeyRepository(fx.Runtime.GormDB())
		runAPIKeyContract(t, repo)
	})
}

// seedUsers ensures the given user IDs exist in the users table so that
// api_keys.user_id FK references resolve.
func seedUsers(t *testing.T, fx *internaltest.DialectFixture, ids ...string) {
	t.Helper()
	for _, id := range ids {
		u := &domain.User{
			Base:           domain.Base{ID: id},
			Email:          id + "@example.invalid",
			HashedPassword: "hash",
		}
		if err := fx.Runtime.GormDB().Create(u).Error; err != nil {
			t.Fatalf("seed user %q: %v", id, err)
		}
	}
}

func runAPIKeyContract(t *testing.T, repo port.APIKeyRepository) {
	t.Helper()
	ctx := context.Background()

	t.Run("Create_Success", func(t *testing.T) {
		key := &domain.APIKey{
			UserID: "user-1", Name: "CI Pipeline",
			KeyHash: "hash-abc", KeyPrefix: "pk_live_abc",
			Scope: domain.APIKeyScopeReadWrite, IsActive: true,
		}
		require.NoError(t, repo.Create(ctx, key))
		assert.NotEmpty(t, key.ID)
	})

	t.Run("Create_DuplicateHash_ReturnsError", func(t *testing.T) {
		key := &domain.APIKey{
			UserID: "user-dup", Name: "Duplicate",
			KeyHash: "hash-duplicate", KeyPrefix: "pk_live_dup",
			Scope: domain.APIKeyScopeRead, IsActive: true,
		}
		require.NoError(t, repo.Create(ctx, key))

		dup := &domain.APIKey{
			UserID: "user-dup", Name: "Duplicate2",
			KeyHash: "hash-duplicate", KeyPrefix: "pk_live_dup",
			Scope: domain.APIKeyScopeRead, IsActive: true,
		}
		assert.Error(t, repo.Create(ctx, dup))
	})

	t.Run("FindByID_Success", func(t *testing.T) {
		key := &domain.APIKey{
			UserID: "user-2", Name: "Find Me",
			KeyHash: "hash-find", KeyPrefix: "pk_live_fin",
			Scope: domain.APIKeyScopeRead, IsActive: true,
		}
		require.NoError(t, repo.Create(ctx, key))
		found, err := repo.FindByID(ctx, key.ID, "user-2")
		require.NoError(t, err)
		assert.Equal(t, key.ID, found.ID)
		assert.Equal(t, "Find Me", found.Name)
	})

	t.Run("FindByID_WrongUser_NotFound", func(t *testing.T) {
		key := &domain.APIKey{
			UserID: "user-3", Name: "Secret Key",
			KeyHash: "hash-secret", KeyPrefix: "pk_live_sec",
			Scope: domain.APIKeyScopeRead, IsActive: true,
		}
		require.NoError(t, repo.Create(ctx, key))
		_, err := repo.FindByID(ctx, key.ID, "wrong-user")
		assert.Error(t, err)
	})

	t.Run("FindByKeyHash_Success", func(t *testing.T) {
		key := &domain.APIKey{
			UserID: "user-4", Name: "Hash Lookup",
			KeyHash: "hash-lookup", KeyPrefix: "pk_live_loo",
			Scope: domain.APIKeyScopeReadWrite, IsActive: true,
		}
		require.NoError(t, repo.Create(ctx, key))
		found, err := repo.FindByKeyHash(ctx, "hash-lookup")
		require.NoError(t, err)
		assert.Equal(t, key.ID, found.ID)
	})

	t.Run("FindByKeyHash_NotFound", func(t *testing.T) {
		_, err := repo.FindByKeyHash(ctx, "nonexistent-hash")
		assert.Error(t, err)
	})

	t.Run("ListByUserID_ReturnsOwnedKeys", func(t *testing.T) {
		userID := "user-list"
		for i := 0; i < 3; i++ {
			key := &domain.APIKey{
				UserID: userID, Name: "Key",
				KeyHash: "hash-list-" + string(rune('a'+i)), KeyPrefix: "pk_live_ls",
				Scope: domain.APIKeyScopeRead, IsActive: true,
			}
			require.NoError(t, repo.Create(ctx, key))
		}
		keys, err := repo.ListByUserID(ctx, userID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(keys), 3)
	})

	t.Run("CountByUserID", func(t *testing.T) {
		userID := "user-count"
		for i := 0; i < 2; i++ {
			key := &domain.APIKey{
				UserID: userID, Name: "Key",
				KeyHash: "hash-count-" + string(rune('a'+i)), KeyPrefix: "pk_live_ct",
				Scope: domain.APIKeyScopeRead, IsActive: true,
			}
			require.NoError(t, repo.Create(ctx, key))
		}
		count, err := repo.CountByUserID(ctx, userID)
		require.NoError(t, err)
		assert.EqualValues(t, 2, count)
	})

	t.Run("Revoke_Success", func(t *testing.T) {
		key := &domain.APIKey{
			UserID: "user-revoke", Name: "To Revoke",
			KeyHash: "hash-revoke", KeyPrefix: "pk_live_rev",
			Scope: domain.APIKeyScopeRead, IsActive: true,
		}
		require.NoError(t, repo.Create(ctx, key))
		require.NoError(t, repo.Revoke(ctx, key.ID, "user-revoke"))
		found, err := repo.FindByID(ctx, key.ID, "user-revoke")
		require.NoError(t, err)
		assert.False(t, found.IsActive)
	})

	t.Run("Revoke_WrongUser_NotFound", func(t *testing.T) {
		key := &domain.APIKey{
			UserID: "user-revoke2", Name: "Protected Key",
			KeyHash: "hash-protected", KeyPrefix: "pk_live_pro",
			Scope: domain.APIKeyScopeRead, IsActive: true,
		}
		require.NoError(t, repo.Create(ctx, key))
		assert.Error(t, repo.Revoke(ctx, key.ID, "attacker"))
	})

	t.Run("UpdateLastUsed_Success", func(t *testing.T) {
		key := &domain.APIKey{
			UserID: "user-lastused", Name: "Track Me",
			KeyHash: "hash-track", KeyPrefix: "pk_live_tra",
			Scope: domain.APIKeyScopeReadWrite, IsActive: true,
		}
		require.NoError(t, repo.Create(ctx, key))
		usedAt := time.Now().UTC()
		require.NoError(t, repo.UpdateLastUsed(ctx, key.ID, usedAt, "127.0.0.1"))
		found, err := repo.FindByID(ctx, key.ID, "user-lastused")
		require.NoError(t, err)
		assert.NotNil(t, found.LastUsedAt)
		assert.Equal(t, "127.0.0.1", found.LastUsedIP)
	})

	t.Run("UpdateLastUsed_NotFound", func(t *testing.T) {
		assert.Error(t, repo.UpdateLastUsed(ctx, "nonexistent", time.Now(), "0.0.0.0"))
	})
}
