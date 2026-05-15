package service

import (
	"context"
	"testing"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTagService_CreateTag(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := fake.NewTagsFake()
		service := NewTagService(repo)

		tag := &domain.Tags{
			Base: domain.Base{ID: "tag-1"},
			Name: "production",
		}

		err := service.CreateTag(ctx, tag)
		assert.NoError(t, err)

		// Verify tag was created
		created, err := repo.FindByID(ctx, "tag-1")
		require.NoError(t, err)
		assert.Equal(t, "production", created.Name)
	})

	t.Run("nil tag", func(t *testing.T) {
		repo := fake.NewTagsFake()
		service := NewTagService(repo)

		err := service.CreateTag(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tag cannot be nil")
	})

	t.Run("empty name", func(t *testing.T) {
		repo := fake.NewTagsFake()
		service := NewTagService(repo)

		tag := &domain.Tags{
			Base: domain.Base{ID: "tag-1"},
			Name: "",
		}

		err := service.CreateTag(ctx, tag)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tag name is required")
	})

	t.Run("duplicate name", func(t *testing.T) {
		repo := fake.NewTagsFake()
		service := NewTagService(repo)

		tag1 := &domain.Tags{
			Base: domain.Base{ID: "tag-1"},
			Name: "production",
		}
		err := service.CreateTag(ctx, tag1)
		require.NoError(t, err)

		tag2 := &domain.Tags{
			Base: domain.Base{ID: "tag-2"},
			Name: "production",
		}
		err = service.CreateTag(ctx, tag2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})
}

func TestTagService_ListTags(t *testing.T) {
	ctx := context.Background()
	repo := fake.NewTagsFake()
	service := NewTagService(repo)

	// Create test tags
	tags := []*domain.Tags{
		{Base: domain.Base{ID: "tag-1"}, Name: "production"},
		{Base: domain.Base{ID: "tag-2"}, Name: "staging"},
		{Base: domain.Base{ID: "tag-3"}, Name: "development"},
	}

	for _, tag := range tags {
		err := service.CreateTag(ctx, tag)
		require.NoError(t, err)
	}

	// List all tags
	result, err := service.ListTags(ctx, 10, 0)
	assert.NoError(t, err)
	assert.Len(t, result, 3)
}

func TestTagService_GetTagByID(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := fake.NewTagsFake()
		service := NewTagService(repo)

		tag := &domain.Tags{
			Base: domain.Base{ID: "tag-1"},
			Name: "production",
		}
		err := service.CreateTag(ctx, tag)
		require.NoError(t, err)

		result, err := service.GetTagByID(ctx, "tag-1")
		assert.NoError(t, err)
		assert.Equal(t, "production", result.Name)
	})

	t.Run("not found", func(t *testing.T) {
		repo := fake.NewTagsFake()
		service := NewTagService(repo)

		_, err := service.GetTagByID(ctx, "non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestTagService_UpdateTag(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := fake.NewTagsFake()
		service := NewTagService(repo)

		tag := &domain.Tags{
			Base: domain.Base{ID: "tag-1"},
			Name: "production",
		}
		err := service.CreateTag(ctx, tag)
		require.NoError(t, err)

		updated, err := service.UpdateTag(ctx, "tag-1", "prod", nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, "prod", updated.Name)
	})

	t.Run("empty name", func(t *testing.T) {
		repo := fake.NewTagsFake()
		service := NewTagService(repo)

		tag := &domain.Tags{
			Base: domain.Base{ID: "tag-1"},
			Name: "production",
		}
		err := service.CreateTag(ctx, tag)
		require.NoError(t, err)

		_, err = service.UpdateTag(ctx, "tag-1", "", nil, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tag name is required")
	})

	t.Run("not found", func(t *testing.T) {
		repo := fake.NewTagsFake()
		service := NewTagService(repo)

		_, err := service.UpdateTag(ctx, "non-existent", "newname", nil, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("duplicate name conflict", func(t *testing.T) {
		repo := fake.NewTagsFake()
		service := NewTagService(repo)

		tag1 := &domain.Tags{
			Base: domain.Base{ID: "tag-1"},
			Name: "production",
		}
		tag2 := &domain.Tags{
			Base: domain.Base{ID: "tag-2"},
			Name: "staging",
		}
		err := service.CreateTag(ctx, tag1)
		require.NoError(t, err)
		err = service.CreateTag(ctx, tag2)
		require.NoError(t, err)

		_, err = service.UpdateTag(ctx, "tag-2", "production", nil, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})
}

func TestTagService_DeleteTag(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := fake.NewTagsFake()
		service := NewTagService(repo)

		tag := &domain.Tags{
			Base: domain.Base{ID: "tag-1"},
			Name: "production",
		}
		err := service.CreateTag(ctx, tag)
		require.NoError(t, err)

		err = service.DeleteTag(ctx, "tag-1")
		assert.NoError(t, err)

		// Verify tag was deleted
		_, err = repo.FindByID(ctx, "tag-1")
		assert.Error(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		repo := fake.NewTagsFake()
		service := NewTagService(repo)

		err := service.DeleteTag(ctx, "non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}
