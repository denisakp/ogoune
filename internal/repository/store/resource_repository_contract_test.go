package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
)

func TestResourceRepository_Contract(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		repo := store.NewResourceRepositorySQLC(fx.Runtime)
		runResourceContract(t, repo)
	})
}

func runResourceContract(t *testing.T, repo port.ResourceRepository) {
	t.Helper()
	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		res := &domain.Resource{
			Base:     domain.Base{ID: "test-resource-1", CreatedAt: time.Now()},
			Name:     "Test Resource",
			Type:     domain.ResourceHTTP,
			Target:   "https://example.com",
			Interval: 60, Timeout: 5,
			IsActive: true,
		}
		_, err := repo.Create(ctx, res)
		require.NoError(t, err)
	})

	t.Run("Create_HeartbeatFields", func(t *testing.T) {
		slug := "550e8400-e29b-41d4-a716-446655440000"
		interval, grace := 300, 60
		res := &domain.Resource{
			Base:              domain.Base{ID: "heartbeat-create-1", CreatedAt: time.Now()},
			Name:              "Heartbeat Resource",
			Type:              domain.ResourceHeartbeat,
			IsActive:          true,
			Status:            domain.StatusUp,
			HeartbeatSlug:     &slug,
			HeartbeatInterval: &interval,
			HeartbeatGrace:    &grace,
		}
		_, err := repo.Create(ctx, res)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, res.ID)
		require.NoError(t, err)
		require.NotNil(t, found.HeartbeatSlug)
		assert.Equal(t, slug, *found.HeartbeatSlug)
		assert.Equal(t, interval, *found.HeartbeatInterval)
		assert.Equal(t, grace, *found.HeartbeatGrace)
	})

	t.Run("FindByID", func(t *testing.T) {
		res := &domain.Resource{
			Base:     domain.Base{ID: "test-resource-2", CreatedAt: time.Now()},
			Name:     "Test Resource 2",
			Type:     domain.ResourceHTTP,
			Target:   "https://example2.com",
			IsActive: true,
		}
		_, err := repo.Create(ctx, res)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, res.ID)
		require.NoError(t, err)
		assert.Equal(t, res.Name, found.Name)
		assert.Equal(t, res.Target, found.Target)

		_, err = repo.FindByID(ctx, "nonexistent")
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})

	t.Run("FindByHeartbeatSlug", func(t *testing.T) {
		slug := "550e8400-e29b-41d4-a716-446655440001"
		interval, grace := 300, 60
		res := &domain.Resource{
			Base:              domain.Base{ID: "heartbeat-slug-1", CreatedAt: time.Now()},
			Name:              "Heartbeat Slug Lookup",
			Type:              domain.ResourceHeartbeat,
			IsActive:          true,
			Status:            domain.StatusUp,
			HeartbeatSlug:     &slug,
			HeartbeatInterval: &interval,
			HeartbeatGrace:    &grace,
		}
		_, err := repo.Create(ctx, res)
		require.NoError(t, err)

		found, err := repo.FindByHeartbeatSlug(ctx, slug)
		require.NoError(t, err)
		assert.Equal(t, res.ID, found.ID)

		_, err = repo.FindByHeartbeatSlug(ctx, "550e8400-e29b-41d4-a716-446655440999")
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})

	t.Run("Update", func(t *testing.T) {
		res := &domain.Resource{
			Base:     domain.Base{ID: "test-update-res", CreatedAt: time.Now()},
			Name:     "Original Name",
			IsActive: true,
		}
		_, err := repo.Create(ctx, res)
		require.NoError(t, err)

		res.Name = "Updated Name"
		require.NoError(t, repo.Update(ctx, res))

		found, err := repo.FindByID(ctx, res.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", found.Name)
	})

	t.Run("Delete_SoftDelete", func(t *testing.T) {
		res := &domain.Resource{
			Base:     domain.Base{ID: "test-delete-res", CreatedAt: time.Now()},
			Name:     "To Delete",
			IsActive: true,
		}
		_, err := repo.Create(ctx, res)
		require.NoError(t, err)

		require.NoError(t, repo.Delete(ctx, res.ID))

		// Soft delete: the GORM repo's FindByID filters by is_active=true so
		// the deleted row is invisible. (The fake's FindByID does not filter
		// and would still surface IsActive=false. That fake-specific behavior
		// is asserted in resource_repository_fake_test.go.)
		_, err = repo.FindByID(ctx, res.ID)
		assert.ErrorIs(t, err, repository.ErrNotFound)

		assert.ErrorIs(t, repo.Delete(ctx, "nonexistent"), repository.ErrNotFound)
	})

	t.Run("FindActive", func(t *testing.T) {
		activeRes := &domain.Resource{
			Base:     domain.Base{ID: "active-res", CreatedAt: time.Now()},
			Name:     "Active Resource",
			IsActive: true,
		}
		inactiveRes := &domain.Resource{
			Base:     domain.Base{ID: "inactive-res", CreatedAt: time.Now().Add(-1 * time.Minute)},
			Name:     "Inactive Resource",
			IsActive: false,
		}
		_, err := repo.Create(ctx, activeRes)
		require.NoError(t, err)
		_, err = repo.Create(ctx, inactiveRes)
		require.NoError(t, err)

		results, err := repo.FindActive(ctx, 100, 0)
		require.NoError(t, err)

		var found bool
		for _, r := range results {
			assert.True(t, r.IsActive)
			if r.ID == "active-res" {
				found = true
			}
		}
		assert.True(t, found)
	})

	t.Run("FindMissedHeartbeats_and_UpdateLastPingAt", func(t *testing.T) {
		now := time.Now().UTC()
		slugMissed := "550e8400-e29b-41d4-a716-446655440002"
		slugFresh := "550e8400-e29b-41d4-a716-446655440003"
		interval, grace := 300, 60
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
		_, err := repo.Create(ctx, missed)
		require.NoError(t, err)
		_, err = repo.Create(ctx, fresh)
		require.NoError(t, err)

		items, err := repo.FindMissedHeartbeats(ctx, now, 50)
		require.NoError(t, err)
		var missedFound bool
		for _, it := range items {
			if it.ID == "heartbeat-missed-1" {
				missedFound = true
			}
		}
		assert.True(t, missedFound)

		newPing := now
		require.NoError(t, repo.UpdateLastPingAt(ctx, "heartbeat-missed-1", newPing))

		updated, err := repo.FindByID(ctx, "heartbeat-missed-1")
		require.NoError(t, err)
		require.NotNil(t, updated.LastPingAt)
		assert.WithinDuration(t, newPing, *updated.LastPingAt, time.Second)

		err = repo.UpdateLastPingAt(ctx, "nonexistent", newPing)
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})
}
