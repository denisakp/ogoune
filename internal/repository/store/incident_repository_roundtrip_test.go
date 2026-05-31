package store_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	pgsqlc "github.com/denisakp/ogoune/internal/repository/sqlc/pg"
	sqlitesqlc "github.com/denisakp/ogoune/internal/repository/sqlc/sqlite"
	"github.com/denisakp/ogoune/internal/repository/store"
)

// incidentRoundtripFactory wires a sqlc incident repo with a counter.
func incidentRoundtripFactory(t *testing.T, fx *internaltest.DialectFixture) (port.IncidentRepository, *internaltest.Counter) {
	t.Helper()
	rt := fx.Runtime
	switch fx.Dialect {
	case "postgres":
		c, dbtx := internaltest.NewPGCounter(rt.PgxPool())
		q := pgsqlc.New(dbtx)
		return store.NewIncidentRepositorySQLCForTest(q, nil), c
	case "sqlite":
		c, dbtx := internaltest.NewSQLiteCounter(rt.SQLiteDB())
		q := sqlitesqlc.New(dbtx)
		return store.NewIncidentRepositorySQLCForTest(nil, q), c
	default:
		t.Fatalf("unknown dialect %q", fx.Dialect)
		return nil, nil
	}
}

// TestIncidentRepository_FindByID_RoundTripBound verifies the bound for
// the incident single-row read path (FR-004 of spec 049): 1 principal +
// 1 per preloaded relation (Resource, IncidentDiagnostics).
func TestIncidentRepository_FindByID_RoundTripBound(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		ctx := context.Background()
		seedResource(t, fx, "irt-res-1", "irt-res-1")

		// Seed one incident.
		seedRepo := store.NewIncidentRepository(fx.Runtime.GormDB())
		_, err := seedRepo.Create(ctx, &domain.Incident{
			Base:       domain.Base{ID: "irt-inc-1", CreatedAt: time.Now()},
			ResourceID: "irt-res-1",
			StartedAt:  time.Now(),
		})
		require.NoError(t, err)

		repo, counter := incidentRoundtripFactory(t, fx)
		counter.Reset()
		loaded, err := repo.FindByID(ctx, "irt-inc-1")
		require.NoError(t, err)
		require.NotNil(t, loaded)

		// 1 principal SELECT + Resource IN-query + IncidentDiagnostics IN-query = 3.
		assert.EqualValues(t, 3, counter.Snapshot(),
			"FindByID: expected 3 round-trips (1 principal + Resource + IncidentDiagnostics)")
	})
}

// TestIncidentRepository_List_RoundTripBound verifies the bound for the
// list path (FR-004 of spec 049). N invariant in round-trip count.
func TestIncidentRepository_List_RoundTripBound(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		ctx := context.Background()
		seedResource(t, fx, "ilrt-res", "ilrt-res")

		seedRepo := store.NewIncidentRepository(fx.Runtime.GormDB())
		for _, n := range []int{10, 100} {
			require.NoError(t, fx.Runtime.GormDB().Exec("DELETE FROM incident_diagnostics").Error)
			require.NoError(t, fx.Runtime.GormDB().Exec("DELETE FROM incidents").Error)
			for i := 0; i < n; i++ {
				_, err := seedRepo.Create(ctx, &domain.Incident{
					Base:       domain.Base{ID: fmt.Sprintf("ilrt-%04d", i), CreatedAt: time.Now()},
					ResourceID: "ilrt-res",
					StartedAt:  time.Now(),
				})
				require.NoError(t, err)
			}

			repo, counter := incidentRoundtripFactory(t, fx)
			counter.Reset()
			out, err := repo.List(ctx, n, 0)
			require.NoError(t, err)
			assert.Len(t, out, n)

			// 1 principal + Resource + IncidentDiagnostics = 3.
			assert.EqualValuesf(t, 3, counter.Snapshot(),
				"List(N=%d): expected 3 round-trips invariant in N", n)
		}
	})
}
