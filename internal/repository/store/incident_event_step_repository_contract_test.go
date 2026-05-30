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

func TestIncidentEventStepRepository_Contract(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		// Seed parent rows: incident_event_steps.incident_id FK -> incidents.id
		// which itself FKs to resources.id.
		seedResource(t, fx, "res-ies", "res-ies")
		seen := map[string]bool{}
		seed := func(id string) {
			if seen[id] {
				return
			}
			seen[id] = true
			seedIncident(t, fx, id, "res-ies")
		}
		for _, id := range []string{"incident-1", "incident-456", "incident-789", "incident-delete"} {
			seed(id)
		}
		for i := 1; i <= 5; i++ {
			seed(fmt.Sprintf("incident-%d", i))
		}

		repo := store.NewIncidentEventStepRepository(fx.Runtime.GormDB())
		runIncidentEventStepContract(t, repo)
	})
}

func runIncidentEventStepContract(t *testing.T, repo port.IncidentEventStepRepository) {
	t.Helper()
	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		step := &domain.IncidentEventStep{
			IncidentID: "incident-1",
			Step:       domain.IncidentEventStepDetected,
		}
		created, err := repo.Create(ctx, step)
		require.NoError(t, err)
		assert.NotEmpty(t, created.ID)
	})

	t.Run("FindByID", func(t *testing.T) {
		msg := "Resource resolved"
		step := &domain.IncidentEventStep{
			Base:       domain.Base{ID: "test-step-2", CreatedAt: time.Now()},
			IncidentID: "incident-456",
			Step:       domain.IncidentEventStepResolved,
			Message:    &msg,
		}
		_, err := repo.Create(ctx, step)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, step.ID)
		require.NoError(t, err)
		assert.Equal(t, step.ID, found.ID)
		assert.Equal(t, step.IncidentID, found.IncidentID)
		assert.Equal(t, step.Step, found.Step)
		assert.Equal(t, msg, *found.Message)

		_, err = repo.FindByID(ctx, "non-existent-id")
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})

	t.Run("Update", func(t *testing.T) {
		initial := "Initial"
		step := &domain.IncidentEventStep{
			Base:       domain.Base{ID: "test-update-step", CreatedAt: time.Now()},
			IncidentID: "incident-789",
			Step:       domain.IncidentEventStepAlert,
			Message:    &initial,
		}
		_, err := repo.Create(ctx, step)
		require.NoError(t, err)

		updated := "Updated"
		step.Message = &updated
		step.Step = domain.IncidentEventStepDownAlert
		require.NoError(t, repo.Update(ctx, step))

		got, err := repo.FindByID(ctx, step.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated", *got.Message)
		assert.Equal(t, domain.IncidentEventStepDownAlert, got.Step)
	})

	t.Run("Delete", func(t *testing.T) {
		step := &domain.IncidentEventStep{
			Base:       domain.Base{ID: "test-delete-step", CreatedAt: time.Now()},
			IncidentID: "incident-delete",
			Step:       domain.IncidentEventStepUpAlert,
		}
		_, err := repo.Create(ctx, step)
		require.NoError(t, err)

		require.NoError(t, repo.Delete(ctx, step.ID))
		_, err = repo.FindByID(ctx, step.ID)
		assert.ErrorIs(t, err, repository.ErrNotFound)

		assert.ErrorIs(t, repo.Delete(ctx, "non-existent"), repository.ErrNotFound)
	})

	t.Run("List", func(t *testing.T) {
		before, err := repo.List(ctx, 1000, 0)
		require.NoError(t, err)
		baseline := len(before)

		for i := 1; i <= 5; i++ {
			msg := fmt.Sprintf("Step message %d", i)
			step := &domain.IncidentEventStep{
				Base:       domain.Base{ID: fmt.Sprintf("step-list-%d", i), CreatedAt: time.Now()},
				IncidentID: fmt.Sprintf("incident-%d", i),
				Step:       domain.IncidentEventStepDetected,
				Message:    &msg,
			}
			_, err := repo.Create(ctx, step)
			require.NoError(t, err)
		}

		all, err := repo.List(ctx, 1000, 0)
		require.NoError(t, err)
		assert.Equal(t, baseline+5, len(all))
	})
}
