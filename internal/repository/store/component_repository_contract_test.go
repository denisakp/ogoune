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

func TestComponentRepository_Contract(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		repo := store.NewComponentRepository(fx.Runtime.GormDB())
		runComponentContract(t, repo)
	})
}

func runComponentContract(t *testing.T, repo port.ComponentRepository) {
	t.Helper()
	ctx := context.Background()
	tag := fmt.Sprintf("%d", time.Now().UnixNano())

	t.Run("Create_and_FindByID", func(t *testing.T) {
		desc := "test component"
		c := &domain.Component{
			Base:        domain.Base{ID: "01CMP" + tag + "001"},
			Name:        "Component " + tag,
			Description: &desc,
		}
		created, err := repo.Create(ctx, c)
		require.NoError(t, err)
		require.NotNil(t, created)

		got, err := repo.FindByID(ctx, c.ID)
		require.NoError(t, err)
		assert.Equal(t, c.Name, got.Name)
		require.NotNil(t, got.Description)
		assert.Equal(t, desc, *got.Description)
		assert.Equal(t, domain.ComponentStatus("up"), got.LastNotificationStatus)
	})

	t.Run("FindByID_NotFound", func(t *testing.T) {
		_, err := repo.FindByID(ctx, "no-such-component-"+tag)
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})

	t.Run("Update", func(t *testing.T) {
		c := &domain.Component{
			Base: domain.Base{ID: "01CMP" + tag + "UPD"},
			Name: "Original " + tag,
		}
		_, err := repo.Create(ctx, c)
		require.NoError(t, err)

		c.Name = "Updated " + tag
		newDesc := "now described"
		c.Description = &newDesc
		require.NoError(t, repo.Update(ctx, c))

		got, err := repo.FindByID(ctx, c.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated "+tag, got.Name)
		require.NotNil(t, got.Description)
		assert.Equal(t, "now described", *got.Description)
	})

	t.Run("UpdateLastNotificationStatus", func(t *testing.T) {
		c := &domain.Component{
			Base: domain.Base{ID: "01CMP" + tag + "STS"},
			Name: "Status " + tag,
		}
		_, err := repo.Create(ctx, c)
		require.NoError(t, err)

		require.NoError(t, repo.UpdateLastNotificationStatus(ctx, c.ID, domain.ComponentStatus("down")))

		got, err := repo.FindByID(ctx, c.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.ComponentStatus("down"), got.LastNotificationStatus)

		err = repo.UpdateLastNotificationStatus(ctx, "no-such-id-"+tag, "down")
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})

	t.Run("Delete", func(t *testing.T) {
		c := &domain.Component{
			Base: domain.Base{ID: "01CMP" + tag + "DEL"},
			Name: "ToDelete " + tag,
		}
		_, err := repo.Create(ctx, c)
		require.NoError(t, err)
		require.NoError(t, repo.Delete(ctx, c.ID))

		_, err = repo.FindByID(ctx, c.ID)
		assert.ErrorIs(t, err, repository.ErrNotFound)

		err = repo.Delete(ctx, c.ID)
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})

	t.Run("List_pagination", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			_, err := repo.Create(ctx, &domain.Component{
				Base: domain.Base{ID: fmt.Sprintf("01CMP%sLST%02d", tag, i)},
				Name: fmt.Sprintf("list-%s-%d", tag, i),
			})
			require.NoError(t, err)
		}
		page, err := repo.List(ctx, 2, 0)
		require.NoError(t, err)
		assert.Len(t, page, 2)
	})
}
