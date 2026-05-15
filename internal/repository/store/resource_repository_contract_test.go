package store

import (
	"context"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceRepository_Contract(t *testing.T) {
	// Use fake implementation for contract tests
	repo := fake.NewResourceFake()

	t.Run("Create", func(t *testing.T) {
		resource := &domain.Resource{
			Base: domain.Base{
				ID:        "test-resource-1",
				CreatedAt: time.Now(),
			},
			Name:     "Test Resource",
			Type:     domain.ResourceHTTP,
			Target:   "https://example.com",
			Interval: 60,
			Timeout:  5,
			IsActive: true,
		}

		_, err := repo.Create(context.Background(), resource)
		require.NoError(t, err)

		// Test duplicate creation
		_, err = repo.Create(context.Background(), resource)
		assert.ErrorIs(t, err, fake.ErrDuplicate)

		// Test invalid input (empty ID)
		invalidResource := &domain.Resource{Name: "Invalid"}
		_, err = repo.Create(context.Background(), invalidResource)
		// This test is no longer valid as BeforeCreate hook now generates the ID
		// assert.ErrorIs(t, err, fake.ErrInvalidInput)
	})

	t.Run("Create_HeartbeatFields", func(t *testing.T) {
		slug := "550e8400-e29b-41d4-a716-446655440000"
		interval := 300
		grace := 60

		resource := &domain.Resource{
			Base: domain.Base{
				ID:        "heartbeat-create-1",
				CreatedAt: time.Now(),
			},
			Name:              "Heartbeat Resource",
			Type:              domain.ResourceHeartbeat,
			IsActive:          true,
			Status:            domain.StatusUp,
			HeartbeatSlug:     &slug,
			HeartbeatInterval: &interval,
			HeartbeatGrace:    &grace,
		}

		_, err := repo.Create(context.Background(), resource)
		require.NoError(t, err)

		found, err := repo.FindByID(context.Background(), "heartbeat-create-1")
		require.NoError(t, err)
		require.NotNil(t, found.HeartbeatSlug)
		require.NotNil(t, found.HeartbeatInterval)
		require.NotNil(t, found.HeartbeatGrace)
		assert.Equal(t, slug, *found.HeartbeatSlug)
		assert.Equal(t, interval, *found.HeartbeatInterval)
		assert.Equal(t, grace, *found.HeartbeatGrace)
	})

	t.Run("FindByID", func(t *testing.T) {
		resource := &domain.Resource{
			Base: domain.Base{
				ID:        "test-resource-2",
				CreatedAt: time.Now(),
			},
			Name:     "Test Resource 2",
			Type:     domain.ResourceHTTP,
			Target:   "https://example2.com",
			IsActive: true,
		}

		_, err := repo.Create(context.Background(), resource)
		require.NoError(t, err)

		found, err := repo.FindByID(context.Background(), "test-resource-2")
		require.NoError(t, err)
		assert.Equal(t, resource.Name, found.Name)
		assert.Equal(t, resource.Target, found.Target)

		// Test not found
		_, err = repo.FindByID(context.Background(), "nonexistent")
		assert.ErrorIs(t, err, fake.ErrNotFound)
	})

	t.Run("FindByHeartbeatSlug", func(t *testing.T) {
		slug := "550e8400-e29b-41d4-a716-446655440001"
		interval := 300
		grace := 60

		resource := &domain.Resource{
			Base:              domain.Base{ID: "heartbeat-slug-1", CreatedAt: time.Now()},
			Name:              "Heartbeat Slug Lookup",
			Type:              domain.ResourceHeartbeat,
			IsActive:          true,
			Status:            domain.StatusUp,
			HeartbeatSlug:     &slug,
			HeartbeatInterval: &interval,
			HeartbeatGrace:    &grace,
		}

		_, err := repo.Create(context.Background(), resource)
		require.NoError(t, err)

		found, err := repo.FindByHeartbeatSlug(context.Background(), slug)
		require.NoError(t, err)
		assert.Equal(t, "heartbeat-slug-1", found.ID)

		_, err = repo.FindByHeartbeatSlug(context.Background(), "550e8400-e29b-41d4-a716-446655440999")
		assert.ErrorIs(t, err, fake.ErrNotFound)
	})

	t.Run("Update", func(t *testing.T) {
		resource := &domain.Resource{
			Base: domain.Base{
				ID:        "test-update",
				CreatedAt: time.Now(),
			},
			Name:     "Original Name",
			IsActive: true,
		}

		_, err := repo.Create(context.Background(), resource)
		require.NoError(t, err)

		// Update the resource
		resource.Name = "Updated Name"
		err = repo.Update(context.Background(), resource)
		require.NoError(t, err)

		// Verify update
		found, err := repo.FindByID(context.Background(), "test-update")
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", found.Name)

		// Test update nonexistent
		nonExistent := &domain.Resource{
			Base: domain.Base{ID: "nonexistent"},
		}
		err = repo.Update(context.Background(), nonExistent)
		assert.ErrorIs(t, err, fake.ErrNotFound)
	})

	t.Run("Delete_SoftDelete", func(t *testing.T) {
		resource := &domain.Resource{
			Base: domain.Base{
				ID:        "test-delete",
				CreatedAt: time.Now(),
			},
			Name:     "To Delete",
			IsActive: true,
		}

		_, err := repo.Create(context.Background(), resource)
		require.NoError(t, err)

		// Delete (soft delete)
		err = repo.Delete(context.Background(), "test-delete")
		require.NoError(t, err)

		// Resource should still exist but be inactive
		found, err := repo.FindByID(context.Background(), "test-delete")
		require.NoError(t, err)
		assert.False(t, found.IsActive)

		// Test delete nonexistent
		err = repo.Delete(context.Background(), "nonexistent")
		assert.ErrorIs(t, err, fake.ErrNotFound)
	})

	t.Run("FindActive", func(t *testing.T) {
		// Create active and inactive resources
		activeRes := &domain.Resource{
			Base: domain.Base{
				ID:        "active-res",
				CreatedAt: time.Now(),
			},
			Name:     "Active Resource",
			IsActive: true,
		}

		inactiveRes := &domain.Resource{
			Base: domain.Base{
				ID:        "inactive-res",
				CreatedAt: time.Now().Add(-1 * time.Minute),
			},
			Name:     "Inactive Resource",
			IsActive: false,
		}

		_, err := repo.Create(context.Background(), activeRes)
		require.NoError(t, err)
		_, err = repo.Create(context.Background(), inactiveRes)
		require.NoError(t, err)

		// Find active resources
		active, err := repo.FindActive(context.Background(), 10, 0)
		require.NoError(t, err)

		// Should only return active resources
		found := false
		for _, res := range active {
			assert.True(t, res.IsActive)
			if res.ID == "active-res" {
				found = true
			}
		}
		assert.True(t, found, "Should find the active resource")
	})

	t.Run("FindMissedHeartbeats_and_UpdateLastPingAt", func(t *testing.T) {
		now := time.Now().UTC()
		slugMissed := "550e8400-e29b-41d4-a716-446655440002"
		slugFresh := "550e8400-e29b-41d4-a716-446655440003"
		interval := 300
		grace := 60

		missedAt := now.Add(-10 * time.Minute)
		freshAt := now.Add(-1 * time.Minute)

		missed := &domain.Resource{
			Base:              domain.Base{ID: "heartbeat-missed-1", CreatedAt: now.Add(-2 * time.Minute)},
			Name:              "Missed",
			Type:              domain.ResourceHeartbeat,
			Status:            domain.StatusUp,
			IsActive:          true,
			HeartbeatSlug:     &slugMissed,
			HeartbeatInterval: &interval,
			HeartbeatGrace:    &grace,
			LastPingAt:        &missedAt,
		}
		fresh := &domain.Resource{
			Base:              domain.Base{ID: "heartbeat-fresh-1", CreatedAt: now.Add(-1 * time.Minute)},
			Name:              "Fresh",
			Type:              domain.ResourceHeartbeat,
			Status:            domain.StatusUp,
			IsActive:          true,
			HeartbeatSlug:     &slugFresh,
			HeartbeatInterval: &interval,
			HeartbeatGrace:    &grace,
			LastPingAt:        &freshAt,
		}

		_, err := repo.Create(context.Background(), missed)
		require.NoError(t, err)
		_, err = repo.Create(context.Background(), fresh)
		require.NoError(t, err)

		items, err := repo.FindMissedHeartbeats(context.Background(), now, 50)
		require.NoError(t, err)
		require.Len(t, items, 1)
		assert.Equal(t, "heartbeat-missed-1", items[0].ID)

		newPing := now
		err = repo.UpdateLastPingAt(context.Background(), "heartbeat-missed-1", newPing)
		require.NoError(t, err)

		updated, err := repo.FindByID(context.Background(), "heartbeat-missed-1")
		require.NoError(t, err)
		require.NotNil(t, updated.LastPingAt)
		assert.WithinDuration(t, newPing, *updated.LastPingAt, time.Second)

		err = repo.UpdateLastPingAt(context.Background(), "nonexistent", newPing)
		assert.ErrorIs(t, err, fake.ErrNotFound)
	})
}
