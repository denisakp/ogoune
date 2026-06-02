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

// TestTagsRepository_Contract validates the GORM-backed TagsRepository
// against the dual-dialect helper. Contract = behaviors that MUST hold for
// any implementation of port.TagsRepository, regardless of backend.
//
// Fake-only assertions (e.g. ErrDuplicate / ErrInvalidInput sentinels that
// the in-memory fake emits but the GORM impl does not map to) live in
// tags_repository_fake_test.go.
func TestTagsRepository_Contract(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		repo := store.NewTagsRepositorySQLC(fx.Runtime)
		runTagsContract(t, repo)
	})
}

func runTagsContract(t *testing.T, repo port.TagsRepository) {
	t.Helper()
	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		tag := &domain.Tags{
			Base: domain.Base{ID: "01TAG00000000000000000001A", CreatedAt: time.Now()},
			Name: "Test Tag",
		}
		require.NoError(t, repo.Create(ctx, tag))
	})

	t.Run("FindByID", func(t *testing.T) {
		tag := &domain.Tags{
			Base: domain.Base{ID: "01TAG00000000000000000002A", CreatedAt: time.Now()},
			Name: "Test Tag 2",
		}
		require.NoError(t, repo.Create(ctx, tag))

		found, err := repo.FindByID(ctx, tag.ID)
		require.NoError(t, err)
		assert.Equal(t, tag.ID, found.ID)
		assert.Equal(t, tag.Name, found.Name)

		_, err = repo.FindByID(ctx, "non-existent-id")
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})

	t.Run("FindByName", func(t *testing.T) {
		tag := &domain.Tags{
			Base: domain.Base{ID: "01TAG00000000000000000003A", CreatedAt: time.Now()},
			Name: "Unique Tag Name",
		}
		require.NoError(t, repo.Create(ctx, tag))

		found, err := repo.FindByName(ctx, tag.Name)
		require.NoError(t, err)
		assert.Equal(t, tag.ID, found.ID)
		assert.Equal(t, tag.Name, found.Name)

		_, err = repo.FindByName(ctx, "Non-Existent Name")
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})

	t.Run("Update", func(t *testing.T) {
		tag := &domain.Tags{
			Base: domain.Base{ID: "01TAG00000000000000000004A", CreatedAt: time.Now()},
			Name: "Initial Name",
		}
		require.NoError(t, repo.Create(ctx, tag))

		tag.Name = "Updated Name"
		require.NoError(t, repo.Update(ctx, tag))

		updated, err := repo.FindByID(ctx, tag.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", updated.Name)

		// NOTE: GORM's Save() upserts on non-existent rows — it does NOT
		// return ErrNotFound. The fake implementation does. That assertion
		// is therefore fake-specific and lives in tags_repository_fake_test.go.
	})

	t.Run("Delete", func(t *testing.T) {
		tag := &domain.Tags{
			Base: domain.Base{ID: "01TAG00000000000000000005A", CreatedAt: time.Now()},
			Name: "To Be Deleted",
		}
		require.NoError(t, repo.Create(ctx, tag))
		require.NoError(t, repo.Delete(ctx, tag.ID))

		_, err := repo.FindByID(ctx, tag.ID)
		assert.ErrorIs(t, err, repository.ErrNotFound)

		err = repo.Delete(ctx, "non-existent")
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})

	t.Run("List", func(t *testing.T) {
		// Insert N distinct rows. Other sub-tests may have created tags in
		// this dialect's DB earlier; we count what THIS sub-test added.
		const n = 5
		before, err := repo.List(ctx, 1000, 0)
		require.NoError(t, err)
		baseline := len(before)

		for i := 1; i <= n; i++ {
			tag := &domain.Tags{
				Base: domain.Base{
					ID:        fmt.Sprintf("01TAGLIST%016d", i),
					CreatedAt: time.Now(),
				},
				Name: fmt.Sprintf("Tag List %d", i),
			}
			require.NoError(t, repo.Create(ctx, tag))
		}

		got, err := repo.List(ctx, 1000, 0)
		require.NoError(t, err)
		assert.Equal(t, baseline+n, len(got))

		page, err := repo.List(ctx, 2, 1)
		require.NoError(t, err)
		assert.Len(t, page, 2)
	})
}
