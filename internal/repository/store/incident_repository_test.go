package store

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	dbruntime "github.com/denisakp/ogoune/internal/database"
	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func openIncidentRepo(t *testing.T) (*IncidentRepositoryImpl, *dbruntime.Runtime) {
	t.Helper()

	cfg := dbruntime.Config{
		Driver:     dbruntime.DriverSQLite,
		SQLitePath: filepath.Join(t.TempDir(), "incident-repo.db"),
		LogLevel:   "silent",
	}

	runtime, err := dbruntime.Open(context.Background(), cfg)
	require.NoError(t, err)
	require.NotNil(t, runtime)

	repo := NewIncidentRepository(runtime.DB).(*IncidentRepositoryImpl)
	return repo, runtime
}

func seedResource(t *testing.T, runtime *dbruntime.Runtime, suffix string) *domain.Resource {
	t.Helper()

	resource := &domain.Resource{
		Name:     "test-resource-" + suffix,
		Type:     domain.ResourceHTTP,
		Target:   "https://example.com/" + suffix,
		Interval: 60,
		Timeout:  10,
		Status:   domain.StatusDown,
		IsActive: true,
	}
	require.NoError(t, runtime.DB.Create(resource).Error)
	return resource
}

func TestFindUnresolved(t *testing.T) {
	repo, runtime := openIncidentRepo(t)
	ctx := context.Background()
	resource := seedResource(t, runtime, "find-unresolved")

	now := time.Now().UTC()
	resolvedTime := now.Add(-1 * time.Minute)

	t.Run("returns only unresolved incidents", func(t *testing.T) {
		// Create a resolved incident
		resolved := &domain.Incident{
			ResourceID: resource.ID,
			StartedAt:  now.Add(-10 * time.Minute),
			ResolvedAt: &resolvedTime,
			Cause:      "timeout",
		}
		require.NoError(t, runtime.DB.Create(resolved).Error)

		// Create an unresolved incident
		unresolved := &domain.Incident{
			ResourceID: resource.ID,
			StartedAt:  now.Add(-5 * time.Minute),
			ResolvedAt: nil,
			Cause:      "connection_refused",
		}
		require.NoError(t, runtime.DB.Create(unresolved).Error)

		results, err := repo.FindUnresolved(ctx, 10, 0)
		require.NoError(t, err)

		// Must return only the unresolved incident
		require.Len(t, results, 1)
		assert.Equal(t, unresolved.ID, results[0].ID)
		assert.Nil(t, results[0].ResolvedAt)
	})

	t.Run("returns empty when all incidents resolved", func(t *testing.T) {
		// Clean up and create only resolved incidents
		runtime.DB.Where("1=1").Delete(&domain.Incident{})

		resolved := &domain.Incident{
			ResourceID: resource.ID,
			StartedAt:  now.Add(-10 * time.Minute),
			ResolvedAt: &resolvedTime,
			Cause:      "timeout",
		}
		require.NoError(t, runtime.DB.Create(resolved).Error)

		results, err := repo.FindUnresolved(ctx, 10, 0)
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("returns all when none resolved", func(t *testing.T) {
		// Clean up and create only unresolved incidents
		runtime.DB.Where("1=1").Delete(&domain.Incident{})

		for i := 0; i < 3; i++ {
			inc := &domain.Incident{
				ResourceID: resource.ID,
				StartedAt:  now.Add(-time.Duration(i+1) * time.Minute),
				ResolvedAt: nil,
				Cause:      "connection_refused",
			}
			require.NoError(t, runtime.DB.Create(inc).Error)
		}

		results, err := repo.FindUnresolved(ctx, 10, 0)
		require.NoError(t, err)
		assert.Len(t, results, 3)
		for _, r := range results {
			assert.Nil(t, r.ResolvedAt)
		}
	})
}

func TestFindActiveByResourceID(t *testing.T) {
	repo, runtime := openIncidentRepo(t)
	ctx := context.Background()
	resource := seedResource(t, runtime, "find-active")

	now := time.Now().UTC()
	resolvedTime := now.Add(-1 * time.Minute)

	t.Run("returns only unresolved incident for resource", func(t *testing.T) {
		// Create a resolved incident
		resolved := &domain.Incident{
			ResourceID: resource.ID,
			StartedAt:  now.Add(-10 * time.Minute),
			ResolvedAt: &resolvedTime,
			Cause:      "timeout",
		}
		require.NoError(t, runtime.DB.Create(resolved).Error)

		// Create an unresolved incident
		unresolved := &domain.Incident{
			ResourceID: resource.ID,
			StartedAt:  now.Add(-5 * time.Minute),
			ResolvedAt: nil,
			Cause:      "connection_refused",
		}
		require.NoError(t, runtime.DB.Create(unresolved).Error)

		result, err := repo.FindActiveByResourceID(ctx, resource.ID)
		require.NoError(t, err)
		assert.Equal(t, unresolved.ID, result.ID)
		assert.Nil(t, result.ResolvedAt)
	})

	t.Run("returns ErrNotFound when all resolved", func(t *testing.T) {
		runtime.DB.Where("1=1").Delete(&domain.Incident{})

		resolved := &domain.Incident{
			ResourceID: resource.ID,
			StartedAt:  now.Add(-10 * time.Minute),
			ResolvedAt: &resolvedTime,
			Cause:      "timeout",
		}
		require.NoError(t, runtime.DB.Create(resolved).Error)

		_, err := repo.FindActiveByResourceID(ctx, resource.ID)
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})
}
