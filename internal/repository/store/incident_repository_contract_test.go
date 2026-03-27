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

func TestIncidentRepository_Contract(t *testing.T) {
	t.Run("Create", func(t *testing.T) {
		repo := fake.NewIncidentFake()

		incident := &domain.Incident{
			ResourceID: "test-resource-1",
			Cause:      "test_cause",
			StartedAt:  time.Now(),
		}

		created, err := repo.Create(context.Background(), incident)
		require.NoError(t, err)
		assert.NotEmpty(t, created.ID)

		// Test duplicate creation
		_, err = repo.Create(context.Background(), created)
		assert.ErrorIs(t, err, fake.ErrDuplicate)
	})

	t.Run("FindByID", func(t *testing.T) {
		repo := fake.NewIncidentFake()

		incident := &domain.Incident{
			Base: domain.Base{
				ID:        "test-incident-2",
				CreatedAt: time.Now(),
			},
			ResourceID: "resource-2",
			Cause:      "Network error",
			StartedAt:  time.Now(),
			ResolvedAt: nil, // Active incident
		}

		_, err := repo.Create(context.Background(), incident)
		require.NoError(t, err)

		found, err := repo.FindByID(context.Background(), "test-incident-2")
		require.NoError(t, err)
		assert.Equal(t, incident.ResourceID, found.ResourceID)
		assert.Equal(t, incident.Cause, found.Cause)

		// Test not found
		_, err = repo.FindByID(context.Background(), "nonexistent")
		assert.ErrorIs(t, err, fake.ErrNotFound)
	})

	t.Run("Update", func(t *testing.T) {
		repo := fake.NewIncidentFake()

		incident := &domain.Incident{
			Base: domain.Base{
				ID:        "test-update-incident",
				CreatedAt: time.Now(),
			},
			ResourceID: "resource-1",
			Cause:      "connection_timeout",
			StartedAt:  time.Now(),
			ResolvedAt: nil, // Active incident
		}

		_, err := repo.Create(context.Background(), incident)
		require.NoError(t, err)

		// Update the incident (resolve it)
		incident.Cause = "connection_timeout" // Keep cause as is
		resolvedAt := time.Now()
		incident.ResolvedAt = &resolvedAt

		err = repo.Update(context.Background(), incident)
		require.NoError(t, err)

		// Verify update
		found, err := repo.FindByID(context.Background(), "test-update-incident")
		require.NoError(t, err)
		assert.Equal(t, "connection_timeout", found.Cause)
		assert.NotNil(t, found.ResolvedAt, "Incident should be resolved")
	})

	t.Run("Delete", func(t *testing.T) {
		repo := fake.NewIncidentFake()

		incident := &domain.Incident{
			Base: domain.Base{
				ID:        "test-delete-incident",
				CreatedAt: time.Now(),
			},
			ResourceID: "resource-1",
			StartedAt:  time.Now(),
		}

		_, err := repo.Create(context.Background(), incident)
		require.NoError(t, err)

		// Delete
		err = repo.Delete(context.Background(), "test-delete-incident")
		require.NoError(t, err)

		// Should not exist
		_, err = repo.FindByID(context.Background(), "test-delete-incident")
		assert.ErrorIs(t, err, fake.ErrNotFound)
	})

	t.Run("FindUnresolved", func(t *testing.T) {
		repo := fake.NewIncidentFake()

		now := time.Now()
		resolvedTime := now.Add(-1 * time.Minute)

		// Create resolved and unresolved incidents
		resolvedIncident := &domain.Incident{
			Base: domain.Base{
				ID:        "resolved-incident",
				CreatedAt: now.Add(-2 * time.Minute),
			},
			ResourceID: "resource-1",
			StartedAt:  now.Add(-2 * time.Minute),
			ResolvedAt: &resolvedTime, // Resolved incident
		}

		unresolvedIncident := &domain.Incident{
			Base: domain.Base{
				ID:        "unresolved-incident",
				CreatedAt: now.Add(-1 * time.Minute),
			},
			ResourceID: "resource-1",
			StartedAt:  now.Add(-1 * time.Minute),
			ResolvedAt: nil, // Active incident
		}

		_, err := repo.Create(context.Background(), resolvedIncident)
		require.NoError(t, err)
		_, err = repo.Create(context.Background(), unresolvedIncident)
		require.NoError(t, err)

		// Find unresolved incidents
		unresolved, err := repo.FindUnresolved(context.Background(), 10, 0)
		require.NoError(t, err)

		// Should only return unresolved incidents
		found := false
		for _, inc := range unresolved {
			assert.Nil(t, inc.ResolvedAt, "Should only return unresolved incidents (ResolvedAt must be nil)")
			if inc.ID == "unresolved-incident" {
				found = true
			}
		}
		assert.True(t, found, "Should find the unresolved incident")
	})

	t.Run("FindByResource", func(t *testing.T) {
		repo := fake.NewIncidentFake()

		now := time.Now()

		// Create incidents for different resources
		incident1 := &domain.Incident{
			Base: domain.Base{
				ID:        "resource1-incident",
				CreatedAt: now.Add(-2 * time.Minute),
			},
			ResourceID: "resource-1",
			StartedAt:  now.Add(-2 * time.Minute),
		}

		incident2 := &domain.Incident{
			Base: domain.Base{
				ID:        "resource2-incident",
				CreatedAt: now.Add(-1 * time.Minute),
			},
			ResourceID: "resource-2",
			StartedAt:  now.Add(-1 * time.Minute),
		}

		incident3 := &domain.Incident{
			Base: domain.Base{
				ID:        "resource1-incident2",
				CreatedAt: now,
			},
			ResourceID: "resource-1",
			StartedAt:  now,
		}

		_, err := repo.Create(context.Background(), incident1)
		require.NoError(t, err)
		_, err = repo.Create(context.Background(), incident2)
		require.NoError(t, err)
		_, err = repo.Create(context.Background(), incident3)
		require.NoError(t, err)

		// Find incidents for resource-1
		incidents, err := repo.FindByResource(context.Background(), "resource-1", 10, 0)
		require.NoError(t, err)
		assert.Len(t, incidents, 2)

		// Should be ordered by started_at DESC, so newest first
		assert.Equal(t, "resource1-incident2", incidents[0].ID)
		assert.Equal(t, "resource1-incident", incidents[1].ID)

		// Find incidents for resource-2
		incidents, err = repo.FindByResource(context.Background(), "resource-2", 10, 0)
		require.NoError(t, err)
		assert.Len(t, incidents, 1)
		assert.Equal(t, "resource2-incident", incidents[0].ID)

		// Find incidents for nonexistent resource
		incidents, err = repo.FindByResource(context.Background(), "nonexistent", 10, 0)
		require.NoError(t, err)
		assert.Empty(t, incidents)
	})
}
