package store_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
)

func TestIncidentEventStepRepository_SqlcContract(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		seedResource(t, fx, "res-ies", "res-ies")
		seen := map[string]bool{}
		seedI := func(id string) {
			if seen[id] {
				return
			}
			seen[id] = true
			seedIncident(t, fx, id, "res-ies")
		}
		for _, id := range []string{"incident-1", "incident-456", "incident-789", "incident-delete"} {
			seedI(id)
		}
		for i := 1; i <= 5; i++ {
			seedI(fmt.Sprintf("incident-%d", i))
		}
		repo := store.NewIncidentEventStepRepositorySQLC(fx.Runtime)
		runIncidentEventStepContract(t, repo)
	})
}

// TestIncidentEventStepRepository_SqlcJoinPreload verifies that FindByID
// populates the embedded Incident via single JOIN (Clarification Q1).
func TestIncidentEventStepRepository_SqlcJoinPreload(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		seedResource(t, fx, "res-ies-join", "res-ies-join")
		seedIncident(t, fx, "incident-join", "res-ies-join")
		ctx := context.Background()
		repo := store.NewIncidentEventStepRepositorySQLC(fx.Runtime)

		step := &domain.IncidentEventStep{
			Base:       domain.Base{ID: "01STEPJOIN" + fmt.Sprintf("%d", time.Now().UnixNano())},
			IncidentID: "incident-join",
			Step:       domain.IncidentEventStepType("detected"),
		}
		_, err := repo.Create(ctx, step)
		require.NoError(t, err)

		got, err := repo.FindByID(ctx, step.ID)
		require.NoError(t, err)
		assert.Equal(t, step.IncidentID, got.IncidentID)
		assert.Equal(t, "incident-join", got.Incident.ID, "JOIN must populate embedded Incident.ID")
		assert.Equal(t, "res-ies-join", got.Incident.ResourceID, "JOIN must populate Incident.ResourceID")
		assert.Equal(t, "seed", got.Incident.Cause, "JOIN must populate Incident.Cause")
	})
}
