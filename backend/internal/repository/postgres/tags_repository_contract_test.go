package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	domain "github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/repository/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTagsRepository_Contract(t *testing.T) {
	// Use fake implementation for contract tests
	repo := fake.NewTagsFake()

	t.Run("Create", func(t *testing.T) {
		tag := &domain.Tags{
			Base: domain.Base{
				ID:        "test-tag-1",
				CreatedAt: time.Now(),
			},
			Name: "Test Tag",
		}

		err := repo.Create(context.Background(), tag)
		require.NoError(t, err)

		// Test duplicate creation
		err = repo.Create(context.Background(), tag)
		assert.ErrorIs(t, err, fake.ErrDuplicate)

		// Test invalid input (empty ID)
		invalidTag := &domain.Tags{Name: "Invalid"}
		err = repo.Create(context.Background(), invalidTag)
		assert.ErrorIs(t, err, fake.ErrInvalidInput)
	})

	t.Run("FindByID", func(t *testing.T) {
		tag := &domain.Tags{
			Base: domain.Base{
				ID:        "test-tag-2",
				CreatedAt: time.Now(),
			},
			Name: "Test Tag 2",
		}

		err := repo.Create(context.Background(), tag)
		require.NoError(t, err)

		found, err := repo.FindByID(context.Background(), "test-tag-2")
		require.NoError(t, err)
		assert.Equal(t, tag.ID, found.ID)
		assert.Equal(t, tag.Name, found.Name)

		// Test not found
		_, err = repo.FindByID(context.Background(), "non-existent-id")
		assert.ErrorIs(t, err, fake.ErrNotFound)
	})

	t.Run("FindByName", func(t *testing.T) {
		tag := &domain.Tags{
			Base: domain.Base{
				ID:        "test-tag-3",
				CreatedAt: time.Now(),
			},
			Name: "Unique Tag Name",
		}

		err := repo.Create(context.Background(), tag)
		require.NoError(t, err)

		found, err := repo.FindByName(context.Background(), "Unique Tag Name")
		require.NoError(t, err)
		assert.Equal(t, tag.ID, found.ID)
		assert.Equal(t, tag.Name, found.Name)

		// Test not found
		_, err = repo.FindByName(context.Background(), "Non-Existent Name")
		assert.ErrorIs(t, err, fake.ErrNotFound)
	})

	t.Run("Update", func(t *testing.T) {
		tag := &domain.Tags{
			Base: domain.Base{
				ID:        "test-update",
				CreatedAt: time.Now(),
			},
			Name: "Initial Name",
		}

		err := repo.Create(context.Background(), tag)
		require.NoError(t, err)

		tag.Name = "Updated Name"
		err = repo.Update(context.Background(), tag)
		require.NoError(t, err)

		updated, err := repo.FindByID(context.Background(), "test-update")
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", updated.Name)

		// Test update non-existent
		nonExistent := &domain.Tags{
			Base: domain.Base{ID: "non-existent"},
			Name: "Doesn't Matter",
		}
		err = repo.Update(context.Background(), nonExistent)
		assert.ErrorIs(t, err, fake.ErrNotFound)
	})

	t.Run("Delete", func(t *testing.T) {
		tag := &domain.Tags{
			Base: domain.Base{
				ID:        "test-delete",
				CreatedAt: time.Now(),
			},
			Name: "To Be Deleted",
		}

		err := repo.Create(context.Background(), tag)
		require.NoError(t, err)

		err = repo.Delete(context.Background(), "test-delete")
		require.NoError(t, err)

		_, err = repo.FindByID(context.Background(), "test-delete")
		assert.ErrorIs(t, err, fake.ErrNotFound)

		// Test delete non-existent
		err = repo.Delete(context.Background(), "non-existent")
		assert.ErrorIs(t, err, fake.ErrNotFound)
	})

	t.Run("List", func(t *testing.T) {
		// Clear existing tags in fake
		repo = fake.NewTagsFake()

		// Create multiple tags
		for i := 1; i <= 5; i++ {
			tag := &domain.Tags{
				Base: domain.Base{
					ID:        fmt.Sprintf("tag-%d", i),
					CreatedAt: time.Now(),
				},
				Name: fmt.Sprintf("Tag %d", i),
			}
			err := repo.Create(context.Background(), tag)
			require.NoError(t, err)
		}

		tags, err := repo.List(context.Background(), 10, 0)
		require.NoError(t, err)
		assert.Len(t, tags, 5)

		tags, err = repo.List(context.Background(), 2, 1)
		require.NoError(t, err)
		assert.Len(t, tags, 2)
	})
}
