package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
)

func TestUptimeDailyAggRepository_Contract(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		repo := store.NewUptimeDailyAggRepositorySQLC(fx.Runtime)
		ctx := context.Background()

		mk := func(rid string, day time.Time, ratio float64) *domain.UptimeDailyAgg {
			return &domain.UptimeDailyAgg{
				ResourceID:  rid,
				Day:         day,
				Samples:     288,
				Up:          280,
				Degraded:    5,
				Down:        3,
				UptimeRatio: ratio,
				ComputedAt:  time.Now().UTC(),
			}
		}

		d0 := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
		d1 := d0.AddDate(0, 0, 1)
		d2 := d0.AddDate(0, 0, 2)

		t.Run("Upsert_inserts_then_updates_same_day", func(t *testing.T) {
			a := mk("res-A", d0, 0.9722)
			require.NoError(t, repo.Upsert(ctx, a))
			a2 := mk("res-A", d0, 0.9999)
			a2.Samples = 290
			require.NoError(t, repo.Upsert(ctx, a2))

			rows, err := repo.FindForResource(ctx, "res-A", d0, d0)
			require.NoError(t, err)
			require.Len(t, rows, 1)
			assert.Equal(t, 290, rows[0].Samples)
			assert.InDelta(t, 0.9999, rows[0].UptimeRatio, 0.0001)
		})

		t.Run("FindRange_filters_by_resourceIDs_and_window", func(t *testing.T) {
			require.NoError(t, repo.Upsert(ctx, mk("res-B", d0, 0.95)))
			require.NoError(t, repo.Upsert(ctx, mk("res-B", d1, 0.96)))
			require.NoError(t, repo.Upsert(ctx, mk("res-B", d2, 0.97)))
			require.NoError(t, repo.Upsert(ctx, mk("res-C", d1, 0.85)))

			rows, err := repo.FindRange(ctx, []string{"res-B", "res-C"}, d0, d1)
			require.NoError(t, err)
			assert.Len(t, rows, 3)
		})

		t.Run("FindRange_empty_resourceIDs_returns_empty", func(t *testing.T) {
			rows, err := repo.FindRange(ctx, nil, d0, d2)
			require.NoError(t, err)
			assert.Empty(t, rows)
		})

		t.Run("FindForResource_orders_by_day_asc", func(t *testing.T) {
			require.NoError(t, repo.Upsert(ctx, mk("res-D", d2, 0.99)))
			require.NoError(t, repo.Upsert(ctx, mk("res-D", d0, 0.97)))
			require.NoError(t, repo.Upsert(ctx, mk("res-D", d1, 0.98)))

			rows, err := repo.FindForResource(ctx, "res-D", d0, d2)
			require.NoError(t, err)
			require.Len(t, rows, 3)
			for i := 1; i < len(rows); i++ {
				assert.True(t, !rows[i-1].Day.After(rows[i].Day))
			}
		})
	})
}
