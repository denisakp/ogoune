package postgres

import (
	"context"
	"encoding/json"
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
		slackConfig, _ := json.Marshal(map[string]string{
			"type":        "slack",
			"webhook_url": "https://hooks.slack.com/test",
		})

		integration := &domain.Integration{
			Base: domain.Base{
				ID:        "test-integration-1",
				CreatedAt: time.Now(),
			},
			Name:     "Test Slack Integration",
			IsActive: true,
			Config:   slackConfig,
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
		smtpConfig, _ := json.Marshal(map[string]string{
			"type":      "smtp",
			"recipient": "test@example.com",
			"sender":    "alerts@example.com",
		})

		integration := &domain.Integration{
			Base: domain.Base{
				ID:        "test-integration-2",
				CreatedAt: time.Now(),
			},
			Name:     "Test SMTP Integration",
			IsActive: true,
			Config:   smtpConfig,
		}

		err := repo.Create(context.Background(), integration)
		require.NoError(t, err)

		found, err := repo.FindByID(context.Background(), "test-integration-2")
		require.NoError(t, err)
		assert.Equal(t, integration.ID, found.ID)
		assert.Equal(t, integration.Name, found.Name)
		assert.Equal(t, domain.IntegrationSMTP, found.GetType())

		// Test not found
		_, err = repo.FindByID(context.Background(), "non-existent-id")
		assert.ErrorIs(t, err, fake.ErrNotFound)
	})

	t.Run("Update", func(t *testing.T) {
		slackConfig, _ := json.Marshal(map[string]string{
			"type":        "slack",
			"webhook_url": "https://hooks.slack.com/initial",
		})

		integration := &domain.Integration{
			Base: domain.Base{
				ID:        "test-update",
				CreatedAt: time.Now(),
			},
			Name:     "Initial Name",
			IsActive: true,
			Config:   slackConfig,
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
		googleChatConfig, _ := json.Marshal(map[string]string{
			"type":        "google_chat",
			"webhook_url": "https://chat.googleapis.com/test",
		})

		integration := &domain.Integration{
			Base: domain.Base{
				ID:        "test-delete",
				CreatedAt: time.Now(),
			},
			Name:     "To Be Deleted",
			IsActive: true,
			Config:   googleChatConfig,
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

		slackConfig, _ := json.Marshal(map[string]string{
			"type":        "slack",
			"webhook_url": "https://hooks.slack.com/test",
		})

		// Create multiple integrations
		for i := 1; i <= 5; i++ {
			integration := &domain.Integration{
				Base: domain.Base{
					ID:        fmt.Sprintf("integration-%d", i),
					CreatedAt: time.Now(),
				},
				Name:     fmt.Sprintf("Integration %d", i),
				IsActive: i%2 == 1, // Alternate active/inactive
				Config:   slackConfig,
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

		slackConfig, _ := json.Marshal(map[string]string{
			"type":        "slack",
			"webhook_url": "https://hooks.slack.com/test",
		})

		smtpConfig, _ := json.Marshal(map[string]string{
			"type":      "smtp",
			"recipient": "test@example.com",
		})

		// Create integrations of different types and active states
		integrations := []*domain.Integration{
			{
				Base:     domain.Base{ID: "slack-1", CreatedAt: time.Now()},
				Name:     "Active Slack 1",
				IsActive: true,
				Config:   slackConfig,
			},
			{
				Base:     domain.Base{ID: "slack-2", CreatedAt: time.Now()},
				Name:     "Inactive Slack",
				IsActive: false,
				Config:   slackConfig,
			},
			{
				Base:     domain.Base{ID: "slack-3", CreatedAt: time.Now()},
				Name:     "Active Slack 2",
				IsActive: true,
				Config:   slackConfig,
			},
			{
				Base:     domain.Base{ID: "smtp-1", CreatedAt: time.Now()},
				Name:     "Active SMTP",
				IsActive: true,
				Config:   smtpConfig,
			},
		}

		for _, integration := range integrations {
			err := repo.Create(context.Background(), integration)
			require.NoError(t, err)
		}

		// Find active Slack integrations
		slackIntegrations, err := repo.FindActiveByType(context.Background(), domain.IntegrationSlack, 10, 0)
		require.NoError(t, err)
		assert.Len(t, slackIntegrations, 2) // Should only return active ones

		for _, integration := range slackIntegrations {
			assert.Equal(t, domain.IntegrationSlack, integration.GetType())
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
