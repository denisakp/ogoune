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

func TestAnnouncementRepository_Contract(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		repo := store.NewAnnouncementRepositorySQLC(fx.Runtime)
		ctx := context.Background()

		t.Run("empty", func(t *testing.T) {
			rows, err := repo.ListActive(ctx)
			require.NoError(t, err)
			assert.Empty(t, rows)
		})

		var activeID string
		t.Run("Create_active_and_inactive_then_ListActive_filters", func(t *testing.T) {
			a, err := repo.Create(ctx, &domain.Announcement{Severity: domain.AnnouncementWarning, Title: "Maintenance", Description: "soon", Dismissible: true, Active: true})
			require.NoError(t, err)
			require.NotEmpty(t, a.ID)
			activeID = a.ID
			_, err = repo.Create(ctx, &domain.Announcement{Severity: domain.AnnouncementInfo, Title: "Archived", Dismissible: false, Active: false})
			require.NoError(t, err)

			rows, err := repo.ListActive(ctx)
			require.NoError(t, err)
			require.Len(t, rows, 1)
			assert.Equal(t, "Maintenance", rows[0].Title)
			assert.Equal(t, domain.AnnouncementWarning, rows[0].Severity)
			assert.True(t, rows[0].Dismissible)
		})

		t.Run("Delete_removes_from_active", func(t *testing.T) {
			require.NoError(t, repo.Delete(ctx, activeID))
			rows, err := repo.ListActive(ctx)
			require.NoError(t, err)
			assert.Empty(t, rows)
		})

		t.Run("Delete_missing_is_ErrNotFound", func(t *testing.T) {
			require.ErrorIs(t, repo.Delete(ctx, "nope"), repository.ErrNotFound)
		})
	})
}
