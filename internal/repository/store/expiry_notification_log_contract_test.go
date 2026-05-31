package store_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
)

func TestExpiryNotificationLogRepository_Contract(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		seedResources(t, fx, "res-enl-1", "res-enl-2", "res-enl-old")
		repo := store.NewExpiryNotificationLogRepository(fx.Runtime.GormDB())
		runExpiryNotificationLogContract(t, repo)
	})
}

// seedResources inserts minimal Resource rows so the expiry_notification_logs.resource_id FK resolves.
func seedResources(t *testing.T, fx *internaltest.DialectFixture, ids ...string) {
	t.Helper()
	for _, id := range ids {
		res := &domain.Resource{
			Base:   domain.Base{ID: id},
			Name:   "seed-" + id,
			Type:   domain.ResourceHTTP,
			Target: "https://example.invalid/" + id,
		}
		if err := fx.Runtime.GormDB().Create(res).Error; err != nil {
			t.Fatalf("seed resource %q: %v", id, err)
		}
	}
}

func runExpiryNotificationLogContract(t *testing.T, repo port.ExpiryNotificationLogRepository) {
	t.Helper()
	ctx := context.Background()
	tag := fmt.Sprintf("%d", time.Now().UnixNano())
	now := time.Now()

	t.Run("CountByKey_Empty", func(t *testing.T) {
		n, err := repo.CountByKey(ctx, "res-enl-1", "ssl", 7)
		require.NoError(t, err)
		assert.Equal(t, int64(0), n)
	})

	t.Run("Create_and_CountByKey", func(t *testing.T) {
		log := &domain.ExpiryNotificationLog{
			Base:       domain.Base{ID: "01ENL" + tag + "0001"},
			ResourceID: "res-enl-1",
			ExpiryType: "ssl",
			Threshold:  30,
			SentAt:     now,
		}
		require.NoError(t, repo.Create(ctx, log))

		n, err := repo.CountByKey(ctx, "res-enl-1", "ssl", 30)
		require.NoError(t, err)
		assert.Equal(t, int64(1), n)
	})

	t.Run("DeleteByResourceIDAndType", func(t *testing.T) {
		for i, threshold := range []int{7, 14, 30} {
			log := &domain.ExpiryNotificationLog{
				Base:       domain.Base{ID: fmt.Sprintf("01ENL%sDEL%02d", tag, i)},
				ResourceID: "res-enl-2",
				ExpiryType: "ssl",
				Threshold:  threshold,
				SentAt:     now,
			}
			require.NoError(t, repo.Create(ctx, log))
		}

		// Delete-by-no-match must not error.
		require.NoError(t, repo.DeleteByResourceIDAndType(ctx, "res-enl-2", "domain"))
		n, err := repo.CountByKey(ctx, "res-enl-2", "ssl", 7)
		require.NoError(t, err)
		assert.Equal(t, int64(1), n)

		// Delete the actual rows.
		require.NoError(t, repo.DeleteByResourceIDAndType(ctx, "res-enl-2", "ssl"))
		for _, threshold := range []int{7, 14, 30} {
			n, err := repo.CountByKey(ctx, "res-enl-2", "ssl", threshold)
			require.NoError(t, err)
			assert.Equal(t, int64(0), n)
		}
	})

	t.Run("DeleteOlderThan", func(t *testing.T) {
		oldLog := &domain.ExpiryNotificationLog{
			Base:       domain.Base{ID: "01ENL" + tag + "OLD"},
			ResourceID: "res-enl-old",
			ExpiryType: "ssl",
			Threshold:  1,
			SentAt:     now.Add(-48 * time.Hour),
		}
		recentLog := &domain.ExpiryNotificationLog{
			Base:       domain.Base{ID: "01ENL" + tag + "NEW"},
			ResourceID: "res-enl-old",
			ExpiryType: "ssl",
			Threshold:  2,
			SentAt:     now,
		}
		require.NoError(t, repo.Create(ctx, oldLog))
		require.NoError(t, repo.Create(ctx, recentLog))

		cutoff := now.Add(-24 * time.Hour)
		require.NoError(t, repo.DeleteOlderThan(ctx, cutoff))

		nOld, err := repo.CountByKey(ctx, "res-enl-old", "ssl", 1)
		require.NoError(t, err)
		assert.Equal(t, int64(0), nOld, "old row must be deleted")
		nNew, err := repo.CountByKey(ctx, "res-enl-old", "ssl", 2)
		require.NoError(t, err)
		assert.Equal(t, int64(1), nNew, "recent row must remain")
	})
}
