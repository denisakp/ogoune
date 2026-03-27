package store

import (
	"context"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T009 – APIKeyRepository contract tests using the fake implementation.
func TestAPIKeyRepository_Contract(t *testing.T) {
	repo := fake.NewAPIKeyRepository()

	t.Run("Create_Success", func(t *testing.T) {
		key := &domain.APIKey{
			UserID:    "user-1",
			Name:      "CI Pipeline",
			KeyHash:   "hash-abc",
			KeyPrefix: "pk_live_abc",
			Scope:     domain.APIKeyScopeReadWrite,
			IsActive:  true,
		}
		err := repo.Create(context.Background(), key)
		require.NoError(t, err)
		assert.NotEmpty(t, key.ID)
	})

	t.Run("Create_DuplicateHash_ReturnsError", func(t *testing.T) {
		key := &domain.APIKey{
			UserID:    "user-dup",
			Name:      "Duplicate",
			KeyHash:   "hash-duplicate",
			KeyPrefix: "pk_live_dup",
			Scope:     domain.APIKeyScopeRead,
			IsActive:  true,
		}
		err := repo.Create(context.Background(), key)
		require.NoError(t, err)

		key2 := &domain.APIKey{
			UserID:    "user-dup",
			Name:      "Duplicate2",
			KeyHash:   "hash-duplicate",
			KeyPrefix: "pk_live_dup",
			Scope:     domain.APIKeyScopeRead,
			IsActive:  true,
		}
		err = repo.Create(context.Background(), key2)
		assert.Error(t, err)
	})

	t.Run("FindByID_Success", func(t *testing.T) {
		key := &domain.APIKey{
			UserID:    "user-2",
			Name:      "Find Me",
			KeyHash:   "hash-find",
			KeyPrefix: "pk_live_fin",
			Scope:     domain.APIKeyScopeRead,
			IsActive:  true,
		}
		err := repo.Create(context.Background(), key)
		require.NoError(t, err)

		found, err := repo.FindByID(context.Background(), key.ID, "user-2")
		require.NoError(t, err)
		assert.Equal(t, key.ID, found.ID)
		assert.Equal(t, "Find Me", found.Name)
	})

	t.Run("FindByID_WrongUser_NotFound", func(t *testing.T) {
		key := &domain.APIKey{
			UserID:    "user-3",
			Name:      "Secret Key",
			KeyHash:   "hash-secret",
			KeyPrefix: "pk_live_sec",
			Scope:     domain.APIKeyScopeRead,
			IsActive:  true,
		}
		err := repo.Create(context.Background(), key)
		require.NoError(t, err)

		_, err = repo.FindByID(context.Background(), key.ID, "wrong-user")
		assert.Error(t, err)
	})

	t.Run("FindByKeyHash_Success", func(t *testing.T) {
		key := &domain.APIKey{
			UserID:    "user-4",
			Name:      "Hash Lookup",
			KeyHash:   "hash-lookup",
			KeyPrefix: "pk_live_loo",
			Scope:     domain.APIKeyScopeReadWrite,
			IsActive:  true,
		}
		err := repo.Create(context.Background(), key)
		require.NoError(t, err)

		found, err := repo.FindByKeyHash(context.Background(), "hash-lookup")
		require.NoError(t, err)
		assert.Equal(t, key.ID, found.ID)
	})

	t.Run("FindByKeyHash_NotFound", func(t *testing.T) {
		_, err := repo.FindByKeyHash(context.Background(), "nonexistent-hash")
		assert.Error(t, err)
	})

	t.Run("ListByUserID_ReturnsOwnedKeys", func(t *testing.T) {
		userID := "user-list"
		for i := 0; i < 3; i++ {
			key := &domain.APIKey{
				UserID:    userID,
				Name:      "Key",
				KeyHash:   "hash-list-" + string(rune('a'+i)),
				KeyPrefix: "pk_live_ls",
				Scope:     domain.APIKeyScopeRead,
				IsActive:  true,
			}
			require.NoError(t, repo.Create(context.Background(), key))
		}
		keys, err := repo.ListByUserID(context.Background(), userID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(keys), 3)
	})

	t.Run("CountByUserID", func(t *testing.T) {
		userID := "user-count"
		for i := 0; i < 2; i++ {
			key := &domain.APIKey{
				UserID:    userID,
				Name:      "Key",
				KeyHash:   "hash-count-" + string(rune('a'+i)),
				KeyPrefix: "pk_live_ct",
				Scope:     domain.APIKeyScopeRead,
				IsActive:  true,
			}
			require.NoError(t, repo.Create(context.Background(), key))
		}
		count, err := repo.CountByUserID(context.Background(), userID)
		require.NoError(t, err)
		assert.EqualValues(t, 2, count)
	})

	t.Run("Revoke_Success", func(t *testing.T) {
		key := &domain.APIKey{
			UserID:    "user-revoke",
			Name:      "To Revoke",
			KeyHash:   "hash-revoke",
			KeyPrefix: "pk_live_rev",
			Scope:     domain.APIKeyScopeRead,
			IsActive:  true,
		}
		err := repo.Create(context.Background(), key)
		require.NoError(t, err)

		err = repo.Revoke(context.Background(), key.ID, "user-revoke")
		require.NoError(t, err)

		found, err := repo.FindByID(context.Background(), key.ID, "user-revoke")
		require.NoError(t, err)
		assert.False(t, found.IsActive)
	})

	t.Run("Revoke_WrongUser_NotFound", func(t *testing.T) {
		key := &domain.APIKey{
			UserID:    "user-revoke2",
			Name:      "Protected Key",
			KeyHash:   "hash-protected",
			KeyPrefix: "pk_live_pro",
			Scope:     domain.APIKeyScopeRead,
			IsActive:  true,
		}
		err := repo.Create(context.Background(), key)
		require.NoError(t, err)

		err = repo.Revoke(context.Background(), key.ID, "attacker")
		assert.Error(t, err)
	})

	t.Run("UpdateLastUsed_Success", func(t *testing.T) {
		key := &domain.APIKey{
			UserID:    "user-lastused",
			Name:      "Track Me",
			KeyHash:   "hash-track",
			KeyPrefix: "pk_live_tra",
			Scope:     domain.APIKeyScopeReadWrite,
			IsActive:  true,
		}
		err := repo.Create(context.Background(), key)
		require.NoError(t, err)

		usedAt := time.Now().UTC()
		err = repo.UpdateLastUsed(context.Background(), key.ID, usedAt, "127.0.0.1")
		require.NoError(t, err)

		found, err := repo.FindByID(context.Background(), key.ID, "user-lastused")
		require.NoError(t, err)
		assert.NotNil(t, found.LastUsedAt)
		assert.Equal(t, "127.0.0.1", found.LastUsedIP)
	})

	t.Run("UpdateLastUsed_NotFound", func(t *testing.T) {
		err := repo.UpdateLastUsed(context.Background(), "nonexistent", time.Now(), "0.0.0.0")
		assert.Error(t, err)
	})
}
