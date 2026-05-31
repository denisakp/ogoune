package store_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
)

func TestMaintenanceRepository_SqlcContract(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		seedResources(t, fx, "res-mtc-1", "res-mtc-2", "res-mtc-3")
		repo := store.NewMaintenanceRepositorySQLC(fx.Runtime)
		runMaintenanceContract(t, repo)
	})
}

// TestMaintenanceRepository_SqlcMM_DiffOnUpdate asserts the sqlc-specific
// M2M update semantics: DELETE removed + INSERT added inside one tx (FR-007).
// GORM's Save() appends instead, so this assertion is sqlc-only.
func TestMaintenanceRepository_SqlcMM_DiffOnUpdate(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		seedResources(t, fx, "res-mtc-diff-1", "res-mtc-diff-2", "res-mtc-diff-3")
		ctx := context.Background()
		repo := store.NewMaintenanceRepositorySQLC(fx.Runtime)
		now := time.Now()
		start := now.Add(-time.Hour)
		end := now.Add(time.Hour)

		m := &domain.Maintenance{
			Base:     domain.Base{ID: "01MTCDIFF" + fmt.Sprintf("%d", time.Now().UnixNano())},
			Title:    "Diff test",
			Strategy: domain.OneTime,
			Status:   "scheduled",
			StartAt:  &start,
			EndAt:    &end,
			Resources: []*domain.Resource{
				{Base: domain.Base{ID: "res-mtc-diff-1"}},
			},
		}
		_, err := repo.Create(ctx, m)
		require.NoError(t, err)

		// Target: remove res-mtc-diff-1, add res-mtc-diff-2 + 3.
		m.Resources = []*domain.Resource{
			{Base: domain.Base{ID: "res-mtc-diff-2"}},
			{Base: domain.Base{ID: "res-mtc-diff-3"}},
		}
		require.NoError(t, repo.Update(ctx, m))

		got, err := repo.FindByID(ctx, m.ID)
		require.NoError(t, err)
		require.Len(t, got.Resources, 2, "sqlc Update must diff: 1 removed, 2 added → exactly 2 remain")
		gotIDs := []string{got.Resources[0].ID, got.Resources[1].ID}
		assert.NotContains(t, gotIDs, "res-mtc-diff-1")
		assert.Contains(t, gotIDs, "res-mtc-diff-2")
		assert.Contains(t, gotIDs, "res-mtc-diff-3")
	})
}

// TestMaintenanceRepository_SqlcMM_TxRollback (SC-008) — forces FK violation on
// junction INSERT, asserts principal row is NOT persisted (tx rollback).
func TestMaintenanceRepository_SqlcMM_TxRollback(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		seedResources(t, fx, "res-mtc-rb-valid")
		ctx := context.Background()
		repo := store.NewMaintenanceRepositorySQLC(fx.Runtime)
		tag := fmt.Sprintf("%d", time.Now().UnixNano())
		now := time.Now()
		start := now.Add(-time.Hour)
		end := now.Add(time.Hour)

		m := &domain.Maintenance{
			Base:     domain.Base{ID: "01MTCRB" + tag},
			Title:    "Rollback test",
			Strategy: domain.OneTime,
			Status:   "scheduled",
			StartAt:  &start,
			EndAt:    &end,
			Resources: []*domain.Resource{
				{Base: domain.Base{ID: "res-mtc-rb-valid"}},
				{Base: domain.Base{ID: "res-does-not-exist-" + tag}}, // FK violation
			},
		}
		_, err := repo.Create(ctx, m)
		require.Error(t, err, "Create with invalid junction must fail")

		// Principal MUST NOT exist after rollback.
		_, findErr := repo.FindByID(ctx, m.ID)
		assert.ErrorIs(t, findErr, repository.ErrNotFound, "tx rollback must remove principal row")
	})
}
