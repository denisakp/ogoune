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

func TestIntegrationRepository_Contract(t *testing.T) {
	// Use fake implementation for contract tests
	repo := fake.NewIntegrationFake()

	t.Run("Create", func(t *testing.T) {
		integration := &domain.Integration{
			Base: domain.Base{
				ID:        "test-integration-1",
				CreatedAt: time.Now(),
			},
			Name:     "Test Slack Integration",
			Target:   "https://hooks.slack.com/test",
			Type:     domain.IntegrationSlack,
			IsActive: true,
		}

		err := repo.Create(context.Background(), integration)
		require.NoError(t, err)

		// Test duplicate creation
		err = repo.Create(context.Background(), integration)
		assert.ErrorIs(t, err, fake.ErrDuplicate)

		// Test invalid input (empty ID)
		invalidIntegration := &domain.Integration{Name: "Invalid"}
		err = repo.Create(context.Background(), invalidIntegration)
		assert.ErrorIs(t, err, fake.ErrInvalidInput)
	})

	t.Run("FindByID", func(t *testing.T) {
		integration := &domain.Integration{
			Base: domain.Base{
				ID:        "test-integration-2",
				CreatedAt: time.Now(),
			},
			Name:     "Test SMTP Integration",
			Target:   "smtp://mail.example.com:587",
			Type:     domain.IntegrationSMTP,
			IsActive: true,
		}

		err := repo.Create(context.Background(), integration)
		require.NoError(t, err)

		found, err := repo.FindByID(context.Background(), "test-integration-2")
		require.NoError(t, err)
		assert.Equal(t, integration.ID, found.ID)
		assert.Equal(t, integration.Name, found.Name)
		assert.Equal(t, integration.Type, found.Type)

		// Test not found
		_, err = repo.FindByID(context.Background(), "non-existent-id")
		assert.ErrorIs(t, err, fake.ErrNotFound)
	})

	t.Run("Update", func(t *testing.T) {
		integration := &domain.Integration{
			Base: domain.Base{
				ID:        "test-update",
				CreatedAt: time.Now(),
			},
			Name:     "Initial Name",
			Target:   "initial-target",
			Type:     domain.IntegrationSlack,
			IsActive: true,
		}

		err := repo.Create(context.Background(), integration)
		require.NoError(t, err)

		integration.Name = "Updated Name"
		integration.IsActive = false
		err = repo.Update(context.Background(), integration)
		require.NoError(t, err)

		updated, err := repo.FindByID(context.Background(), "test-update")
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", updated.Name)
		assert.False(t, updated.IsActive)

		// Test update non-existent
		nonExistent := &domain.Integration{
			Base: domain.Base{ID: "non-existent"},
			Name: "Doesn't Matter",
		}
		err = repo.Update(context.Background(), nonExistent)
		assert.ErrorIs(t, err, fake.ErrNotFound)
	})

	t.Run("Delete", func(t *testing.T) {
		integration := &domain.Integration{
			Base: domain.Base{
				ID:        "test-delete",
				CreatedAt: time.Now(),
			},
			Name:     "To Be Deleted",
			Target:   "delete-target",
			Type:     domain.IntegrationGoogleChat,
			IsActive: true,
		}

		err := repo.Create(context.Background(), integration)
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
		// Clear existing integrations by creating a new fake
		repo = fake.NewIntegrationFake()

		// Create multiple integrations
		for i := 1; i <= 5; i++ {
			integration := &domain.Integration{
				Base: domain.Base{
					ID:        fmt.Sprintf("integration-%d", i),
					CreatedAt: time.Now(),
				},
				Name:     fmt.Sprintf("Integration %d", i),
				Target:   fmt.Sprintf("target-%d", i),
				Type:     domain.IntegrationSlack,
				IsActive: i%2 == 1, // Alternate active/inactive
			}
			err := repo.Create(context.Background(), integration)
			require.NoError(t, err)
		}

		integrations, err := repo.List(context.Background(), 10, 0)
		require.NoError(t, err)
		assert.Len(t, integrations, 5)

		integrations, err = repo.List(context.Background(), 2, 1)
		require.NoError(t, err)
		assert.Len(t, integrations, 2)
	})

	t.Run("FindActiveByType", func(t *testing.T) {
		// Clear existing integrations by creating a new fake
		repo = fake.NewIntegrationFake()

		// Create integrations of different types and active states
		integrations := []*domain.Integration{
			{
				Base:     domain.Base{ID: "slack-1", CreatedAt: time.Now()},
				Name:     "Active Slack 1",
				Type:     domain.IntegrationSlack,
				IsActive: true,
			},
			{
				Base:     domain.Base{ID: "slack-2", CreatedAt: time.Now()},
				Name:     "Inactive Slack",
				Type:     domain.IntegrationSlack,
				IsActive: false,
			},
			{
				Base:     domain.Base{ID: "slack-3", CreatedAt: time.Now()},
				Name:     "Active Slack 2",
				Type:     domain.IntegrationSlack,
				IsActive: true,
			},
			{
				Base:     domain.Base{ID: "smtp-1", CreatedAt: time.Now()},
				Name:     "Active SMTP",
				Type:     domain.IntegrationSMTP,
				IsActive: true,
			},
		}

		for _, integration := range integrations {
			err := repo.Create(context.Background(), integration)
			require.NoError(t, err)
		}

		// Find active Slack integrations
		activeSlack, err := repo.FindActiveByType(context.Background(), domain.IntegrationSlack, 10, 0)
		require.NoError(t, err)
		assert.Len(t, activeSlack, 2)
		for _, integration := range activeSlack {
			assert.Equal(t, domain.IntegrationSlack, integration.Type)
			assert.True(t, integration.IsActive)
		}

		// Find active SMTP integrations
		activeSMTP, err := repo.FindActiveByType(context.Background(), domain.IntegrationSMTP, 10, 0)
		require.NoError(t, err)
		assert.Len(t, activeSMTP, 1)
		assert.Equal(t, "Active SMTP", activeSMTP[0].Name)

		// Find active Google Chat integrations (should be empty)
		activeGoogleChat, err := repo.FindActiveByType(context.Background(), domain.IntegrationGoogleChat, 10, 0)
		require.NoError(t, err)
		assert.Len(t, activeGoogleChat, 0)
	})
}
