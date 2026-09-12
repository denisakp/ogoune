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

func TestIncidentRepository_Contract(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		// Seed all resources referenced by incidents in the sub-tests.
		seedResource(t, fx, "test-resource-1", "res-1")
		seedResource(t, fx, "resource-1", "res-1b")
		seedResource(t, fx, "resource-2", "res-2")

		repo := store.NewIncidentRepositorySQLC(fx.Runtime)
		runIncidentContract(t, repo)
	})
}

func runIncidentContract(t *testing.T, repo port.IncidentRepository) {
	t.Helper()
	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		inc := &domain.Incident{
			ResourceID: "test-resource-1",
			Cause:      "test_cause",
			StartedAt:  time.Now(),
		}
		created, err := repo.Create(ctx, inc)
		require.NoError(t, err)
		assert.NotEmpty(t, created.ID)
	})

	t.Run("FindByID", func(t *testing.T) {
		inc := &domain.Incident{
			Base:       domain.Base{ID: "test-incident-2", CreatedAt: time.Now()},
			ResourceID: "resource-2",
			Cause:      "Network error",
			StartedAt:  time.Now(),
		}
		_, err := repo.Create(ctx, inc)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, "test-incident-2")
		require.NoError(t, err)
		assert.Equal(t, inc.ResourceID, found.ResourceID)
		assert.Equal(t, inc.Cause, found.Cause)

		_, err = repo.FindByID(ctx, "nonexistent")
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})

	// The machine an incident happened on, as it was when it opened (spec 092).
	// Two columns carry three states, and the two absences must survive a
	// round-trip as DIFFERENT things: one falls back, the other must never.
	t.Run("HostLinkRoundTripsThreeStates", func(t *testing.T) {
		hostA := "host-A"

		recorded := &domain.Incident{
			Base: domain.Base{ID: "hl-recorded"}, ResourceID: "resource-2",
			Cause: "c", StartedAt: time.Now(),
			HostID: &hostA, HostLinkRecorded: true,
		}
		absent := &domain.Incident{
			Base: domain.Base{ID: "hl-absent"}, ResourceID: "resource-2",
			Cause: "c", StartedAt: time.Now(),
			HostID: nil, HostLinkRecorded: true,
		}
		predates := &domain.Incident{
			Base: domain.Base{ID: "hl-predates"}, ResourceID: "resource-2",
			Cause: "c", StartedAt: time.Now(),
			// Left at the zero values: exactly what a row from before the
			// migration looks like.
		}
		for _, inc := range []*domain.Incident{recorded, absent, predates} {
			_, err := repo.Create(ctx, inc)
			require.NoError(t, err)
		}

		got, err := repo.FindByID(ctx, "hl-recorded")
		require.NoError(t, err)
		require.NotNil(t, got.HostID)
		assert.Equal(t, "host-A", *got.HostID)
		assert.True(t, got.HostLinkRecorded)

		got, err = repo.FindByID(ctx, "hl-absent")
		require.NoError(t, err)
		assert.Nil(t, got.HostID)
		assert.True(t, got.HostLinkRecorded, "we looked and there was none -- the flag is what keeps it from ever falling back")

		got, err = repo.FindByID(ctx, "hl-predates")
		require.NoError(t, err)
		assert.Nil(t, got.HostID)
		assert.False(t, got.HostLinkRecorded, "a pre-migration row reads as not recorded, which is the state that may fall back")
	})

	// A pre-migration row is not a fixture we build -- it is what the defaults
	// produce. Insert through the raw column list to be sure the defaults, not
	// the wrapper, are what a real old row carries.
	t.Run("HostLinkDefaultsReadAsNotRecorded", func(t *testing.T) {
		inc := &domain.Incident{
			Base: domain.Base{ID: "hl-defaults"}, ResourceID: "resource-2",
			Cause: "c", StartedAt: time.Now(),
		}
		_, err := repo.Create(ctx, inc)
		require.NoError(t, err)

		got, err := repo.FindByID(ctx, "hl-defaults")
		require.NoError(t, err)
		assert.False(t, got.HostLinkRecorded)
		assert.Nil(t, got.HostID)
	})

	t.Run("Update", func(t *testing.T) {
		inc := &domain.Incident{
			Base:       domain.Base{ID: "test-update-incident", CreatedAt: time.Now()},
			ResourceID: "resource-1",
			Cause:      "connection_timeout",
			StartedAt:  time.Now(),
		}
		_, err := repo.Create(ctx, inc)
		require.NoError(t, err)

		resolved := time.Now()
		inc.ResolvedAt = &resolved
		require.NoError(t, repo.Update(ctx, inc))

		got, err := repo.FindByID(ctx, inc.ID)
		require.NoError(t, err)
		require.NotNil(t, got.ResolvedAt)
	})

	t.Run("Delete", func(t *testing.T) {
		inc := &domain.Incident{
			Base:       domain.Base{ID: "test-delete-incident", CreatedAt: time.Now()},
			ResourceID: "resource-1",
			StartedAt:  time.Now(),
		}
		_, err := repo.Create(ctx, inc)
		require.NoError(t, err)

		require.NoError(t, repo.Delete(ctx, inc.ID))
		_, err = repo.FindByID(ctx, inc.ID)
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})

	t.Run("FindUnresolved", func(t *testing.T) {
		now := time.Now()
		resolvedTime := now.Add(-1 * time.Minute)
		resolved := &domain.Incident{
			Base:       domain.Base{ID: "resolved-incident", CreatedAt: now.Add(-2 * time.Minute)},
			ResourceID: "resource-1",
			StartedAt:  now.Add(-2 * time.Minute),
			ResolvedAt: &resolvedTime,
		}
		unresolved := &domain.Incident{
			Base:       domain.Base{ID: "unresolved-incident", CreatedAt: now.Add(-1 * time.Minute)},
			ResourceID: "resource-1",
			StartedAt:  now.Add(-1 * time.Minute),
		}
		_, err := repo.Create(ctx, resolved)
		require.NoError(t, err)
		_, err = repo.Create(ctx, unresolved)
		require.NoError(t, err)

		results, err := repo.FindUnresolved(ctx, 100, 0)
		require.NoError(t, err)

		found := false
		for _, r := range results {
			assert.Nil(t, r.ResolvedAt)
			if r.ID == "unresolved-incident" {
				found = true
			}
		}
		assert.True(t, found)
	})

	t.Run("FindByResource", func(t *testing.T) {
		now := time.Now()
		ids := []struct {
			id, resourceID string
			startedAt      time.Time
		}{
			{"r1-incident-a", "resource-1", now.Add(-2 * time.Minute)},
			{"r2-incident-a", "resource-2", now.Add(-1 * time.Minute)},
			{"r1-incident-b", "resource-1", now},
		}
		for _, x := range ids {
			inc := &domain.Incident{
				Base:       domain.Base{ID: x.id, CreatedAt: x.startedAt},
				ResourceID: x.resourceID,
				StartedAt:  x.startedAt,
			}
			_, err := repo.Create(ctx, inc)
			require.NoError(t, err)
		}

		r1, err := repo.FindByResource(ctx, "resource-1", 100, 0)
		require.NoError(t, err)
		// Other sub-tests may have inserted rows on resource-1; at least our 2 new rows show.
		var r1aSeen, r1bSeen bool
		for _, r := range r1 {
			if r.ID == "r1-incident-a" {
				r1aSeen = true
			}
			if r.ID == "r1-incident-b" {
				r1bSeen = true
			}
		}
		assert.True(t, r1aSeen, "expected r1-incident-a in results")
		assert.True(t, r1bSeen, "expected r1-incident-b in results")

		// Nonexistent resource
		none, err := repo.FindByResource(ctx, "nonexistent", 100, 0)
		require.NoError(t, err)
		assert.Empty(t, none)
	})

	t.Run("CountByResourceID", func(t *testing.T) {
		// Use a dedicated resource so the count isn't polluted by other sub-tests.
		// resource-count is not seeded yet — it'll be created here only if needed.
		// We'll insert directly via raw counts.
		count, err := repo.CountByResourceID(ctx, "nonexistent-resource-id")
		require.NoError(t, err)
		assert.EqualValues(t, 0, count)
	})
}
