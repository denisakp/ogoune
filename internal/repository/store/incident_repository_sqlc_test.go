package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
)

// TestIncidentRepository_SqlcContract drives the GORM contract suite against
// the sqlc-backed IncidentRepository on both dialects. US2 of spec 048.
func TestIncidentRepository_SqlcContract(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		seedResource(t, fx, "test-resource-1", "res-1")
		seedResource(t, fx, "resource-1", "res-1b")
		seedResource(t, fx, "resource-2", "res-2")

		repo := store.NewIncidentRepositorySQLC(fx.Runtime)
		runIncidentContract(t, repo)
	})
}

// TestIncidentRepository_GetIncidentStats_DialectParity verifies FR-007: the
// dialect-divergent SQL (CTE on PG, sub-queries on SQLite) returns identical
// integers for the same seeded fixture.
func TestIncidentRepository_GetIncidentStats_DialectParity(t *testing.T) {
	type result struct{ total, affected int }
	results := map[string]result{}

	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		seedResource(t, fx, "stats-res-a", "stats-res-a")
		seedResource(t, fx, "stats-res-b", "stats-res-b")

		repo := store.NewIncidentRepositorySQLC(fx.Runtime)
		ctx := context.Background()

		now := time.Now()
		seed := []struct {
			id, resourceID string
			startedAt      time.Time
		}{
			{"stats-inc-1", "stats-res-a", now.Add(-1 * time.Hour)},
			{"stats-inc-2", "stats-res-a", now.Add(-30 * time.Minute)},
			{"stats-inc-3", "stats-res-b", now.Add(-2 * time.Hour)},
			{"stats-inc-4", "stats-res-b", now.Add(-25 * time.Hour)},
		}
		for _, s := range seed {
			_, err := repo.Create(ctx, &domain.Incident{
				Base:       domain.Base{ID: s.id, CreatedAt: s.startedAt},
				ResourceID: s.resourceID,
				StartedAt:  s.startedAt,
			})
			require.NoError(t, err)
		}

		total, affected, err := repo.GetIncidentStats(ctx, 24)
		require.NoError(t, err)
		results[fx.Dialect] = result{total: total, affected: affected}

		assert.Equal(t, 3, total)
		assert.Equal(t, 2, affected)
	})

	if len(results) == 2 {
		assert.Equal(t, results["postgres"], results["sqlite"],
			"GetIncidentStats must be numerically identical across dialects")
	}
}
