package store_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
)

func TestEscalationRepository_Contract(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		repo := store.NewEscalationRepositorySQLC(fx.Runtime)
		ctx := context.Background()

		mkPolicy := func(name string, priority int, active bool) *domain.EscalationPolicy {
			return &domain.EscalationPolicy{
				Name:     name,
				Scope:    domain.EscalationScope{Kind: domain.EscalationScopeTag, Value: "prod"},
				IsActive: active,
				Priority: priority,
				Steps: []domain.EscalationStep{
					{StepOrder: 1, DelayMinutes: 5, ChannelIDs: []string{"ch-1"}},
					{StepOrder: 2, DelayMinutes: 10, ChannelIDs: []string{"ch-2", "ch-3"}},
				},
			}
		}

		t.Run("Create_FindByID_persists_steps_and_channel_ids", func(t *testing.T) {
			p := mkPolicy("on-call", 1, true)
			require.NoError(t, repo.Create(ctx, p))
			require.NotEmpty(t, p.ID)
			got, err := repo.FindByID(ctx, p.ID)
			require.NoError(t, err)
			assert.Equal(t, "on-call", got.Name)
			assert.True(t, got.IsActive)
			assert.Equal(t, 1, got.Priority)
			require.Len(t, got.Steps, 2)
			assert.Equal(t, []string{"ch-2", "ch-3"}, got.Steps[1].ChannelIDs)
		})

		t.Run("List_orders_by_priority_asc", func(t *testing.T) {
			require.NoError(t, repo.Create(ctx, mkPolicy("pol-low", 50, true)))
			require.NoError(t, repo.Create(ctx, mkPolicy("pol-mid", 30, true)))
			list, err := repo.List(ctx)
			require.NoError(t, err)
			require.GreaterOrEqual(t, len(list), 2)
			for i := 1; i < len(list); i++ {
				assert.LessOrEqual(t, list[i-1].Priority, list[i].Priority)
			}
		})

		t.Run("Reorder_assigns_1_to_N", func(t *testing.T) {
			// Deactivate any existing active policies so the reorder partial-unique
			// constraint sees only our 3 fresh policies.
			existing, err := repo.List(ctx)
			require.NoError(t, err)
			for _, p := range existing {
				if p.IsActive {
					p.IsActive = false
					require.NoError(t, repo.Update(ctx, p))
				}
			}
			a := mkPolicy("reorder-A", 100, true)
			b := mkPolicy("reorder-B", 101, true)
			c := mkPolicy("reorder-C", 102, true)
			require.NoError(t, repo.Create(ctx, a))
			require.NoError(t, repo.Create(ctx, b))
			require.NoError(t, repo.Create(ctx, c))
			require.NoError(t, repo.Reorder(ctx, []string{c.ID, a.ID, b.ID}))
			gotC, _ := repo.FindByID(ctx, c.ID)
			gotA, _ := repo.FindByID(ctx, a.ID)
			gotB, _ := repo.FindByID(ctx, b.ID)
			assert.Equal(t, 1, gotC.Priority)
			assert.Equal(t, 2, gotA.Priority)
			assert.Equal(t, 3, gotB.Priority)
		})

		t.Run("Delete_cascades_steps", func(t *testing.T) {
			p := mkPolicy("delete-me", 200, true)
			require.NoError(t, repo.Create(ctx, p))
			require.NoError(t, repo.Delete(ctx, p.ID))
			_, err := repo.FindByID(ctx, p.ID)
			assert.ErrorIs(t, err, repository.ErrNotFound)
		})
	})
}
