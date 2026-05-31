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
	"github.com/denisakp/ogoune/internal/repository"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
)

func TestMaintenanceRepository_Contract(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		seedResources(t, fx, "res-mtc-1", "res-mtc-2", "res-mtc-3")
		repo := store.NewMaintenanceRepository(fx.Runtime.GormDB())
		runMaintenanceContract(t, repo)
	})
}

func runMaintenanceContract(t *testing.T, repo port.MaintenanceRepository) {
	t.Helper()
	ctx := context.Background()
	tag := fmt.Sprintf("%d", time.Now().UnixNano())
	now := time.Now()
	start := now.Add(-1 * time.Hour)
	end := now.Add(2 * time.Hour)

	t.Run("Create_with_resources_and_FindByID_preload", func(t *testing.T) {
		m := &domain.Maintenance{
			Base:     domain.Base{ID: "01MTC" + tag + "001"},
			Title:    "Maintenance " + tag,
			Strategy: domain.OneTime,
			Status:   "scheduled",
			StartAt:  &start,
			EndAt:    &end,
			Resources: []*domain.Resource{
				{Base: domain.Base{ID: "res-mtc-1"}},
				{Base: domain.Base{ID: "res-mtc-2"}},
			},
		}
		created, err := repo.Create(ctx, m)
		require.NoError(t, err)
		require.NotNil(t, created)

		got, err := repo.FindByID(ctx, m.ID)
		require.NoError(t, err)
		assert.Equal(t, m.Title, got.Title)
		assert.Equal(t, "scheduled", got.Status)
		require.Len(t, got.Resources, 2)
		gotIDs := []string{got.Resources[0].ID, got.Resources[1].ID}
		assert.Contains(t, gotIDs, "res-mtc-1")
		assert.Contains(t, gotIDs, "res-mtc-2")
	})

	t.Run("FindByID_NotFound", func(t *testing.T) {
		_, err := repo.FindByID(ctx, "no-such-maintenance-"+tag)
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})

	t.Run("Update_principal_fields", func(t *testing.T) {
		// Note: M2M update behavior diverges between GORM (Save() appends to
		// junction) and sqlc impl (computes DELETE+INSERT diff inside WithTx).
		// The diff-precision assertion lives in maintenance_repository_sqlc_test.go
		// (TestMaintenanceRepository_SqlcMM_DiffOnUpdate). The contract only
		// asserts that principal fields update correctly across both impls.
		m := &domain.Maintenance{
			Base:     domain.Base{ID: "01MTC" + tag + "UPD"},
			Title:    "Update " + tag,
			Strategy: domain.OneTime,
			Status:   "scheduled",
			StartAt:  &start,
			EndAt:    &end,
			Resources: []*domain.Resource{
				{Base: domain.Base{ID: "res-mtc-1"}},
			},
		}
		_, err := repo.Create(ctx, m)
		require.NoError(t, err)

		m.Title = "Updated"
		require.NoError(t, repo.Update(ctx, m))

		got, err := repo.FindByID(ctx, m.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated", got.Title)
	})

	t.Run("Delete_cascades_junction", func(t *testing.T) {
		m := &domain.Maintenance{
			Base:     domain.Base{ID: "01MTC" + tag + "DEL"},
			Title:    "Delete me",
			Strategy: domain.OneTime,
			Status:   "scheduled",
			StartAt:  &start,
			EndAt:    &end,
			Resources: []*domain.Resource{
				{Base: domain.Base{ID: "res-mtc-1"}},
			},
		}
		_, err := repo.Create(ctx, m)
		require.NoError(t, err)
		require.NoError(t, repo.Delete(ctx, m.ID))

		_, err = repo.FindByID(ctx, m.ID)
		assert.ErrorIs(t, err, repository.ErrNotFound)
		// FK CASCADE removes junction; we don't directly assert here (would need raw SQL),
		// but a subsequent Create with same resources must succeed (no orphan junction conflicts).
	})

	t.Run("FindActiveForResource", func(t *testing.T) {
		// Active maintenance.
		mActive := &domain.Maintenance{
			Base:     domain.Base{ID: "01MTC" + tag + "ACT"},
			Title:    "Active",
			Strategy: domain.OneTime,
			Status:   "active",
			Resources: []*domain.Resource{
				{Base: domain.Base{ID: "res-mtc-1"}},
			},
		}
		_, err := repo.Create(ctx, mActive)
		require.NoError(t, err)

		// Scheduled but in window.
		mScheduled := &domain.Maintenance{
			Base:     domain.Base{ID: "01MTC" + tag + "SCH"},
			Title:    "Scheduled-in-window",
			Strategy: domain.OneTime,
			Status:   "scheduled",
			StartAt:  &start,
			EndAt:    &end,
			Resources: []*domain.Resource{
				{Base: domain.Base{ID: "res-mtc-1"}},
			},
		}
		_, err = repo.Create(ctx, mScheduled)
		require.NoError(t, err)

		// Expired (outside window).
		expiredStart := now.Add(-48 * time.Hour)
		expiredEnd := now.Add(-24 * time.Hour)
		mExpired := &domain.Maintenance{
			Base:     domain.Base{ID: "01MTC" + tag + "EXP"},
			Title:    "Expired",
			Strategy: domain.OneTime,
			Status:   "scheduled",
			StartAt:  &expiredStart,
			EndAt:    &expiredEnd,
			Resources: []*domain.Resource{
				{Base: domain.Base{ID: "res-mtc-1"}},
			},
		}
		_, err = repo.Create(ctx, mExpired)
		require.NoError(t, err)

		got, err := repo.FindActiveForResource(ctx, "res-mtc-1", now)
		require.NoError(t, err)
		ids := make(map[string]bool)
		for _, m := range got {
			ids[m.ID] = true
		}
		assert.True(t, ids[mActive.ID], "active maintenance must be returned")
		assert.True(t, ids[mScheduled.ID], "scheduled-in-window must be returned")
		assert.False(t, ids[mExpired.ID], "expired must NOT be returned")
	})
}
