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

func TestIncidentDiagnosticsRepository_Contract(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		seedIncidents(t, fx, "inc-diag-1", "inc-diag-2", "inc-diag-3")
		repo := store.NewIncidentDiagnosticsRepository(fx.Runtime.GormDB())
		runIncidentDiagnosticsContract(t, repo)
	})
}

// seedIncidents inserts minimal Incident + Resource rows so the FK chain resolves.
func seedIncidents(t *testing.T, fx *internaltest.DialectFixture, incidentIDs ...string) {
	t.Helper()
	resourceID := "res-for-diag-" + fmt.Sprintf("%d", time.Now().UnixNano())
	res := &domain.Resource{
		Base:   domain.Base{ID: resourceID},
		Name:   "diag-resource",
		Type:   domain.ResourceHTTP,
		Target: "https://example.invalid/diag",
	}
	require.NoError(t, fx.Runtime.GormDB().Create(res).Error)
	for _, id := range incidentIDs {
		inc := &domain.Incident{
			Base:       domain.Base{ID: id},
			ResourceID: resourceID,
			Cause:      "test",
			StartedAt:  time.Now(),
		}
		require.NoError(t, fx.Runtime.GormDB().Create(inc).Error)
	}
}

func runIncidentDiagnosticsContract(t *testing.T, repo port.IncidentDiagnosticsRepository) {
	t.Helper()
	ctx := context.Background()

	t.Run("Create_and_FindByIncidentID", func(t *testing.T) {
		d := &domain.IncidentDiagnostics{
			IncidentID:      "inc-diag-1",
			RequestMethod:   "GET",
			RequestURL:      "https://example.invalid/",
			RequestHeaders:  map[string]string{"User-Agent": "test"},
			ResponseHeaders: map[string]string{},
			HTTPStatusCode:  500,
			ResponseBody:    "internal error",
			FailureType:     "invalid_status_code",
			ErrorMessage:    "5xx",
			TotalDuration:   123,
		}
		created, err := repo.Create(ctx, d)
		require.NoError(t, err)
		require.NotNil(t, created)
		assert.NotEmpty(t, created.ID)

		got, err := repo.FindByIncidentID(ctx, "inc-diag-1")
		require.NoError(t, err)
		assert.Equal(t, "GET", got.RequestMethod)
		assert.Equal(t, 500, got.HTTPStatusCode)
		assert.Equal(t, "test", got.RequestHeaders["User-Agent"])
	})

	t.Run("FindByIncidentID_NotFound", func(t *testing.T) {
		_, err := repo.FindByIncidentID(ctx, "nope-incident")
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})

	t.Run("Update_with_ICMP_enrichment", func(t *testing.T) {
		d := &domain.IncidentDiagnostics{
			IncidentID:      "inc-diag-2",
			RequestMethod:   "HEAD",
			RequestHeaders:  map[string]string{},
			ResponseHeaders: map[string]string{},
			FailureType:     "connection_timeout",
		}
		created, err := repo.Create(ctx, d)
		require.NoError(t, err)

		yes := true
		rtt := 42
		created.ICMPAvailable = &yes
		created.ICMPReachable = &yes
		created.ICMPRttMs = &rtt
		created.RootCauseHint = "service_down"
		require.NoError(t, repo.Update(ctx, created))

		got, err := repo.FindByIncidentID(ctx, "inc-diag-2")
		require.NoError(t, err)
		require.NotNil(t, got.ICMPAvailable)
		assert.True(t, *got.ICMPAvailable)
		require.NotNil(t, got.ICMPRttMs)
		assert.Equal(t, 42, *got.ICMPRttMs)
		assert.Equal(t, "service_down", got.RootCauseHint)
	})

	// NOTE: Update on a non-existent row is intentionally NOT asserted as
	// ErrNotFound at the contract level. GORM's Save() upserts (and triggers
	// an FK violation here because incident_id has no parent row), while the
	// sqlc impl returns ErrNotFound via :execrows on zero-rows-affected.
	// The behavior divergence is GORM-specific and not required by the port
	// contract (see FR-006: "when the port's contract requires it").

	t.Run("Delete_and_Delete_NotFound", func(t *testing.T) {
		d := &domain.IncidentDiagnostics{
			IncidentID:      "inc-diag-3",
			RequestMethod:   "GET",
			RequestHeaders:  map[string]string{},
			ResponseHeaders: map[string]string{},
		}
		created, err := repo.Create(ctx, d)
		require.NoError(t, err)

		require.NoError(t, repo.Delete(ctx, created.ID))

		_, err = repo.FindByIncidentID(ctx, "inc-diag-3")
		assert.ErrorIs(t, err, repository.ErrNotFound)

		err = repo.Delete(ctx, created.ID)
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})
}
