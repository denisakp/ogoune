package store_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
)

// TestNotificationFeedRepository_Contract — spec 072. Dual-dialect (PG + SQLite).
func TestNotificationFeedRepository_Contract(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		seedUsers(t, fx, "user-A", "user-B")
		repo := store.NewNotificationFeedRepositorySQLC(fx.Runtime)
		ctx := context.Background()
		base := time.Now().UTC().Truncate(time.Second)

		strptr := func(s string) *string { return &s }
		mk := func(userID *string, cat string, ago time.Duration) *domain.FeedNotification {
			return &domain.FeedNotification{
				UserID:     userID,
				Category:   cat,
				Severity:   domain.NotificationSeverityError,
				Title:      "t",
				OccurredAt: base.Add(-ago),
			}
		}

		t.Run("Create_returns_row_with_id", func(t *testing.T) {
			got, err := repo.Create(ctx, mk(nil, domain.NotificationCategoryIncident, time.Minute))
			require.NoError(t, err)
			require.NotEmpty(t, got.ID)
			assert.True(t, got.Unread())
		})

		t.Run("List_scopes_instance_wide_plus_user_and_orders_desc", func(t *testing.T) {
			_, _ = repo.Create(ctx, mk(nil, domain.NotificationCategorySystem, 3*time.Minute))    // instance-wide
			_, _ = repo.Create(ctx, mk(strptr("user-A"), domain.NotificationCategoryGeneral, 1*time.Minute)) // user-A
			_, _ = repo.Create(ctx, mk(strptr("user-B"), domain.NotificationCategoryGeneral, 1*time.Minute)) // user-B (not visible to A)

			list, err := repo.ListForUser(ctx, "user-A", nil, 50, 0)
			require.NoError(t, err)
			require.NotEmpty(t, list)
			for i := 1; i < len(list); i++ {
				assert.False(t, list[i-1].OccurredAt.Before(list[i].OccurredAt), "must be newest-first")
			}
			for _, n := range list {
				if n.UserID != nil {
					assert.Equal(t, "user-A", *n.UserID, "must not see other users' targeted notifications")
				}
			}
		})

		t.Run("Category_filter", func(t *testing.T) {
			cat := domain.NotificationCategorySystem
			list, err := repo.ListForUser(ctx, "user-A", &cat, 50, 0)
			require.NoError(t, err)
			for _, n := range list {
				assert.Equal(t, domain.NotificationCategorySystem, n.Category)
			}
		})

		t.Run("MarkRead_idempotent_missing_returns_zero", func(t *testing.T) {
			n, _ := repo.Create(ctx, mk(nil, domain.NotificationCategoryIncident, time.Minute))
			rows, err := repo.MarkRead(ctx, n.ID, time.Now())
			require.NoError(t, err)
			assert.EqualValues(t, 1, rows)
			// second mark still affects the existing row (idempotent), preserves read_at
			rows2, err := repo.MarkRead(ctx, n.ID, time.Now())
			require.NoError(t, err)
			assert.EqualValues(t, 1, rows2)
			// missing id → 0 rows
			missing, err := repo.MarkRead(ctx, "01HZZZNONEXISTENT00000000", time.Now())
			require.NoError(t, err)
			assert.EqualValues(t, 0, missing)
		})

		t.Run("MarkAllRead_respects_before_boundary", func(t *testing.T) {
			u := strptr("user-mark")
			seedUsers(t, fx, "user-mark")
			old1, _ := repo.Create(ctx, mk(u, domain.NotificationCategoryIncident, 10*time.Minute))
			recent, _ := repo.Create(ctx, mk(u, domain.NotificationCategoryIncident, 0)) // occurred now (base)
			boundary := base.Add(-5 * time.Minute)
			marked, err := repo.MarkAllRead(ctx, "user-mark", boundary, time.Now())
			require.NoError(t, err)
			assert.EqualValues(t, 1, marked, "only the one before the boundary is marked")

			list, _ := repo.ListForUser(ctx, "user-mark", nil, 50, 0)
			byID := map[string]*domain.FeedNotification{}
			for _, n := range list {
				byID[n.ID] = n
			}
			assert.False(t, byID[old1.ID].Unread(), "old should be read")
			assert.True(t, byID[recent.ID].Unread(), "recent (after boundary) stays unread")
		})

		t.Run("DeleteOlderThan", func(t *testing.T) {
			u := strptr("user-del")
			seedUsers(t, fx, "user-del")
			_, _ = repo.Create(ctx, mk(u, domain.NotificationCategoryIncident, 100*24*time.Hour)) // 100d old
			keep, _ := repo.Create(ctx, mk(u, domain.NotificationCategoryIncident, time.Hour))
			cutoff := base.Add(-90 * 24 * time.Hour)
			deleted, err := repo.DeleteOlderThan(ctx, cutoff)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, deleted, int64(1))
			list, _ := repo.ListForUser(ctx, "user-del", nil, 200, 0)
			// Filter to this user's own targeted notifications (instance-wide rows
			// from earlier sub-tests are also visible and share the DB).
			var mine []*domain.FeedNotification
			for _, n := range list {
				if n.UserID != nil && *n.UserID == "user-del" {
					mine = append(mine, n)
				}
			}
			require.Len(t, mine, 1, "the 100d-old targeted notification must be pruned")
			assert.Equal(t, keep.ID, mine[0].ID)
		})

		// SC-005: volume — default page returns newest 50 in order, count correct.
		t.Run("Volume_page_and_count", func(t *testing.T) {
			u := strptr("user-vol")
			seedUsers(t, fx, "user-vol")
			const total = 200
			for i := 0; i < total; i++ {
				_, err := repo.Create(ctx, &domain.FeedNotification{
					UserID:     u,
					Category:   domain.NotificationCategoryGeneral,
					Severity:   domain.NotificationSeverityInfo,
					Title:      fmt.Sprintf("n-%d", i),
					OccurredAt: base.Add(-time.Duration(i) * time.Second),
				})
				require.NoError(t, err)
			}
			page, err := repo.ListForUser(ctx, "user-vol", strptr(domain.NotificationCategoryGeneral), 50, 0)
			require.NoError(t, err)
			require.Len(t, page, 50)
			for i := 1; i < len(page); i++ {
				assert.False(t, page[i-1].OccurredAt.Before(page[i].OccurredAt))
			}
			count, err := repo.CountForUser(ctx, "user-vol", strptr(domain.NotificationCategoryGeneral))
			require.NoError(t, err)
			assert.EqualValues(t, total, count)
		})
	})
}
