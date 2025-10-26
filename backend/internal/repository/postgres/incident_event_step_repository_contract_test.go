package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	domain "github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/repository/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIncidentEventStepRepository_Contract(t *testing.T) {
	// Use fake implementation for contract tests
	repo := fake.NewIncidentEventStepFake()

	t.Run("Create", func(t *testing.T) {
		message := "Incident detected"
		step := &domain.IncidentEventStep{
			Base: domain.Base{
				ID:        "test-step-1",
				CreatedAt: time.Now(),
			},
			IncidentID: "incident-123",
			Step:       domain.IncidentEventStepDetected,
			Message:    &message,
		}

		_, err := repo.Create(context.Background(), step)
		require.NoError(t, err)

		// Test duplicate creation
		_, err = repo.Create(context.Background(), step)
		assert.ErrorIs(t, err, fake.ErrDuplicate)

		// Test invalid input (empty ID)
		invalidStep := &domain.IncidentEventStep{IncidentID: "incident-123"}
		_, err = repo.Create(context.Background(), invalidStep)
		assert.ErrorIs(t, err, fake.ErrInvalidInput)
	})

	t.Run("FindByID", func(t *testing.T) {
		message := "Resource resolved"
		step := &domain.IncidentEventStep{
			Base: domain.Base{
				ID:        "test-step-2",
				CreatedAt: time.Now(),
			},
			IncidentID: "incident-456",
			Step:       domain.IncidentEventStepResolved,
			Message:    &message,
		}

		_, err := repo.Create(context.Background(), step)
		require.NoError(t, err)

		found, err := repo.FindByID(context.Background(), "test-step-2")
		require.NoError(t, err)
		assert.Equal(t, step.ID, found.ID)
		assert.Equal(t, step.IncidentID, found.IncidentID)
		assert.Equal(t, step.Step, found.Step)
		assert.Equal(t, *step.Message, *found.Message)

		// Test not found
		_, err = repo.FindByID(context.Background(), "non-existent-id")
		assert.ErrorIs(t, err, fake.ErrNotFound)
	})

	t.Run("Update", func(t *testing.T) {
		initialMessage := "Initial message"
		step := &domain.IncidentEventStep{
			Base: domain.Base{
				ID:        "test-update",
				CreatedAt: time.Now(),
			},
			IncidentID: "incident-789",
			Step:       domain.IncidentEventStepAlert,
			Message:    &initialMessage,
		}

		_, err := repo.Create(context.Background(), step)
		require.NoError(t, err)

		updatedMessage := "Updated message"
		step.Message = &updatedMessage
		step.Step = domain.IncidentEventStepDownAlert
		err = repo.Update(context.Background(), step)
		require.NoError(t, err)

		updated, err := repo.FindByID(context.Background(), "test-update")
		require.NoError(t, err)
		assert.Equal(t, "Updated message", *updated.Message)
		assert.Equal(t, domain.IncidentEventStepDownAlert, updated.Step)

		// Test update non-existent
		nonExistent := &domain.IncidentEventStep{
			Base: domain.Base{ID: "non-existent"},
			Step: domain.IncidentEventStepDetected,
		}
		err = repo.Update(context.Background(), nonExistent)
		assert.ErrorIs(t, err, fake.ErrNotFound)
	})

	t.Run("Delete", func(t *testing.T) {
		step := &domain.IncidentEventStep{
			Base: domain.Base{
				ID:        "test-delete",
				CreatedAt: time.Now(),
			},
			IncidentID: "incident-delete",
			Step:       domain.IncidentEventStepUpAlert,
		}

		_, err := repo.Create(context.Background(), step)
		require.NoError(t, err)

		err = repo.Delete(context.Background(), "test-delete")
		require.NoError(t, err)

		_, err = repo.FindByID(context.Background(), "test-delete")
		assert.ErrorIs(t, err, fake.ErrNotFound)

		// Test delete non-existent
		err = repo.Delete(context.Background(), "non-existent")
		assert.ErrorIs(t, err, fake.ErrNotFound)
	})

	t.Run("List", func(t *testing.T) {
		// Clear existing steps by creating a new fake
		repo = fake.NewIncidentEventStepFake()

		// Create multiple steps
		for i := 1; i <= 5; i++ {
			message := fmt.Sprintf("Step message %d", i)
			step := &domain.IncidentEventStep{
				Base: domain.Base{
					ID:        fmt.Sprintf("step-%d", i),
					CreatedAt: time.Now(),
				},
				IncidentID: fmt.Sprintf("incident-%d", i),
				Step:       domain.IncidentEventStepDetected,
				Message:    &message,
			}
			_, err := repo.Create(context.Background(), step)
			require.NoError(t, err)
		}

		steps, err := repo.List(context.Background(), 10, 0)
		require.NoError(t, err)
		assert.Len(t, steps, 5)

		steps, err = repo.List(context.Background(), 2, 1)
		require.NoError(t, err)
		assert.Len(t, steps, 2)
	})
}
