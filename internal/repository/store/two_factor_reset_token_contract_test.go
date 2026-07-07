package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
)

func TestTwoFactorResetTokenRepository_Contract(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		seedUsers(t, fx, "user-2fa-A")
		repo := store.NewTwoFactorResetTokenRepositorySQLC(fx.Runtime)
		ctx := context.Background()
		now := time.Now()

		t.Run("Create_then_Consume_single_use", func(t *testing.T) {
			tok := &domain.TwoFactorResetToken{
				TokenHash: "hash-A",
				UserID:    "user-2fa-A",
				ExpiresAt: now.Add(15 * time.Minute),
			}
			require.NoError(t, repo.Create(ctx, tok))
			got, err := repo.ConsumeByHash(ctx, "hash-A", now)
			require.NoError(t, err)
			assert.Equal(t, "user-2fa-A", got.UserID)
			// second consume must fail
			_, err = repo.ConsumeByHash(ctx, "hash-A", now)
			assert.ErrorIs(t, err, repository.ErrNotFound)
		})

		t.Run("Expired_token_not_consumable", func(t *testing.T) {
			tok := &domain.TwoFactorResetToken{
				TokenHash: "hash-expired",
				UserID:    "user-2fa-A",
				ExpiresAt: now.Add(-1 * time.Minute),
			}
			require.NoError(t, repo.Create(ctx, tok))
			_, err := repo.ConsumeByHash(ctx, "hash-expired", now)
			assert.ErrorIs(t, err, repository.ErrNotFound)
		})

		t.Run("CountRecentByUser_for_rate_limit", func(t *testing.T) {
			for i := 0; i < 3; i++ {
				require.NoError(t, repo.Create(ctx, &domain.TwoFactorResetToken{
					TokenHash: "h-rate-" + time.Now().Format("150405.000000") + string(rune('a'+i)),
					UserID:    "user-2fa-A",
					ExpiresAt: now.Add(15 * time.Minute),
				}))
			}
			n, err := repo.CountRecentByUser(ctx, "user-2fa-A", now.Add(-1*time.Hour))
			require.NoError(t, err)
			assert.GreaterOrEqual(t, n, int64(3))
		})

		t.Run("DeleteExpired_purges", func(t *testing.T) {
			require.NoError(t, repo.Create(ctx, &domain.TwoFactorResetToken{
				TokenHash: "h-purge",
				UserID:    "user-2fa-A",
				ExpiresAt: now.Add(-10 * time.Minute),
			}))
			require.NoError(t, repo.DeleteExpired(ctx, now))
			_, err := repo.ConsumeByHash(ctx, "h-purge", now)
			assert.ErrorIs(t, err, repository.ErrNotFound)
		})
	})
}
