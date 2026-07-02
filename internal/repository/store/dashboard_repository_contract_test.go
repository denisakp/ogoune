package store_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
)

// seedNamedUser creates a user with an explicit display name (seedUsers leaves it empty).
func seedNamedUser(t *testing.T, fx *internaltest.DialectFixture, id, name string) {
	t.Helper()
	repo := store.NewUserRepositorySQLC(fx.Runtime)
	_, err := repo.Create(context.Background(), &domain.User{
		Base:           domain.Base{ID: id},
		Email:          id + "@example.invalid",
		Name:           name,
		HashedPassword: "hash",
	})
	require.NoError(t, err)
}

func mkDashboard(owner, name string) *domain.Dashboard {
	return &domain.Dashboard{
		OwnerID: owner,
		Name:    name,
		Scope:   domain.DashboardScope{Mode: domain.DashboardScopeModeTag, Payload: domain.DashboardScopePayload{TagIDs: []string{"t1"}}},
		Widgets: []domain.WidgetInstance{
			{ID: "w1", WidgetTypeID: domain.WidgetTypeUptimeStat, Position: 0},
			{ID: "w2", WidgetTypeID: domain.WidgetTypeIncidentsList, Position: 1},
		},
		DefaultTimeRange: "24h",
		RefreshInterval:  "1m",
		Visibility:       "team",
	}
}

func TestDashboardRepository_Contract(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		seedNamedUser(t, fx, "owner-A", "Alice")
		seedNamedUser(t, fx, "owner-B", "Bob")
		repo := store.NewDashboardRepositorySQLC(fx.Runtime)
		ctx := context.Background()

		t.Run("Create_then_FindByID_with_owner_name_from_JOIN", func(t *testing.T) {
			created, err := repo.Create(ctx, mkDashboard("owner-A", "Prod"))
			require.NoError(t, err)
			require.NotEmpty(t, created.ID)
			assert.Equal(t, "Alice", created.OwnerName, "owner_name must come from the users JOIN")

			got, err := repo.FindByID(ctx, created.ID)
			require.NoError(t, err)
			assert.Equal(t, "Prod", got.Name)
			assert.Equal(t, "owner-A", got.OwnerID)
			assert.Equal(t, "Alice", got.OwnerName)
			assert.Equal(t, domain.DashboardScopeModeTag, got.Scope.Mode)
			require.Len(t, got.Widgets, 2)
			assert.Equal(t, "w1", got.Widgets[0].ID)
			assert.Equal(t, 0, got.Widgets[0].Position)
		})

		t.Run("FindByID_missing_returns_ErrNotFound", func(t *testing.T) {
			_, err := repo.FindByID(ctx, "does-not-exist")
			assert.ErrorIs(t, err, repository.ErrNotFound)
		})

		t.Run("List_newest_updated_first_with_owner_name", func(t *testing.T) {
			list, err := repo.List(ctx, 50, 0)
			require.NoError(t, err)
			require.NotEmpty(t, list)
			for i := 1; i < len(list); i++ {
				assert.False(t, list[i-1].UpdatedAt.Before(list[i].UpdatedAt), "newest-updated first")
			}
			for _, d := range list {
				assert.NotEmpty(t, d.OwnerName, "each row carries owner_name from JOIN")
			}
		})

		t.Run("Update_and_UpdateWidgets_round_trip_order", func(t *testing.T) {
			d, err := repo.Create(ctx, mkDashboard("owner-A", "ToEdit"))
			require.NoError(t, err)

			d.Name = "Edited"
			require.NoError(t, repo.Update(ctx, d))
			got, _ := repo.FindByID(ctx, d.ID)
			assert.Equal(t, "Edited", got.Name)

			reordered := []domain.WidgetInstance{
				{ID: "w2", WidgetTypeID: domain.WidgetTypeIncidentsList, Position: 0},
				{ID: "w1", WidgetTypeID: domain.WidgetTypeUptimeStat, Position: 1},
			}
			require.NoError(t, repo.UpdateWidgets(ctx, d.ID, reordered, got.UpdatedAt.Add(1)))
			got2, _ := repo.FindByID(ctx, d.ID)
			require.Len(t, got2.Widgets, 2)
			assert.Equal(t, "w2", got2.Widgets[0].ID, "widget order round-trips exactly")
			assert.Equal(t, "w1", got2.Widgets[1].ID)
		})

		t.Run("Delete_then_NotFound", func(t *testing.T) {
			d, _ := repo.Create(ctx, mkDashboard("owner-B", "ToDelete"))
			require.NoError(t, repo.Delete(ctx, d.ID))
			_, err := repo.FindByID(ctx, d.ID)
			assert.ErrorIs(t, err, repository.ErrNotFound)
			assert.ErrorIs(t, repo.Delete(ctx, d.ID), repository.ErrNotFound)
		})

		t.Run("Cascade_delete_on_owner_removal", func(t *testing.T) {
			seedNamedUser(t, fx, "owner-cascade", "Casey")
			d, _ := repo.Create(ctx, mkDashboard("owner-cascade", "Doomed"))
			userRepo := store.NewUserRepositorySQLC(fx.Runtime)
			require.NoError(t, userRepo.Delete(ctx, "owner-cascade"))
			_, err := repo.FindByID(ctx, d.ID)
			assert.ErrorIs(t, err, repository.ErrNotFound, "owner deletion cascades to dashboards")
		})

		// SC-006: volume — default page returns newest in order.
		t.Run("Volume_page", func(t *testing.T) {
			seedNamedUser(t, fx, "owner-vol", "Val")
			for i := 0; i < 200; i++ {
				_, err := repo.Create(ctx, mkDashboard("owner-vol", fmt.Sprintf("d-%d", i)))
				require.NoError(t, err)
			}
			page, err := repo.List(ctx, 50, 0)
			require.NoError(t, err)
			require.Len(t, page, 50)
			for i := 1; i < len(page); i++ {
				assert.False(t, page[i-1].UpdatedAt.Before(page[i].UpdatedAt))
			}
		})
	})
}
