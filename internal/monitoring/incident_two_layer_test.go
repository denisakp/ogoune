package monitoring_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/monitoring"
	"github.com/denisakp/pulseguard/internal/repository/fake"
	"github.com/denisakp/pulseguard/pkg/notifier"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIncidentService_TwoLayeredNotifications tests the two-layered notification architecture
func TestIncidentService_TwoLayeredNotifications(t *testing.T) {
	// Setup test dependencies
	ctx := context.Background()
	incidentRepo := fake.NewIncidentFake()
	eventStepRepo := fake.NewIncidentEventStepFake()
	notificationRepo := fake.NewNotificationFake()
	integrationRepo := fake.NewIntegrationFake()

	// Mock Asynq client (not used in this test)
	redisOpt := asynq.RedisClientOpt{Addr: "localhost:6379"}
	asynqClient := asynq.NewClient(redisOpt)
	defer asynqClient.Close()

	factory := notifier.NewNotifierFactory()

	t.Run("CreateIncident with SMTP disabled and no integrations", func(t *testing.T) {
		// Create incident service with SMTP disabled
		service := monitoring.NewIncidentService(
			incidentRepo,
			eventStepRepo,
			notificationRepo,
			integrationRepo,
			asynqClient,
			factory,
			false, // SMTP disabled
			"",
			"",
		)

		// Create test resource
		resource := &domain.Resource{
			Base:         domain.Base{ID: "resource-1"},
			Name:         "Test Resource",
			Type:         domain.ResourceHTTP,
			Target:       "https://example.com",
			Status:       domain.StatusDown,
			FailureCount: 3,
		}

		// Create incident
		result := domain.CheckResult{
			Status:       "down",
			ResponseTime: 0,
			ResponseData: "Connection timeout",
		}

		err := service.CreateIncident(ctx, resource, result)
		require.NoError(t, err)

		// Verify incident was created
		incidents, err := incidentRepo.List(ctx, 10, 0)
		require.NoError(t, err)
		assert.Len(t, incidents, 1)
		assert.Nil(t, incidents[0].ResolvedAt) // Should be active

		// Verify no notifications were sent (SMTP disabled, no integrations)
		notifications, err := notificationRepo.List(ctx, 10, 0)
		require.NoError(t, err)
		assert.Len(t, notifications, 0)
	})

	t.Run("CreateIncident with SMTP enabled", func(t *testing.T) {
		// Reset repositories
		incidentRepo = fake.NewIncidentFake()
		eventStepRepo = fake.NewIncidentEventStepFake()
		notificationRepo = fake.NewNotificationFake()
		integrationRepo = fake.NewIntegrationFake()

		// Create incident service with SMTP enabled
		service := monitoring.NewIncidentService(
			incidentRepo,
			eventStepRepo,
			notificationRepo,
			integrationRepo,
			asynqClient,
			factory,
			true, // SMTP enabled
			"admin@example.com",
			"noreply@example.com",
		)

		// Create test resource
		resource := &domain.Resource{
			Base:         domain.Base{ID: "resource-2"},
			Name:         "Test Resource 2",
			Type:         domain.ResourceHTTP,
			Target:       "https://example.com",
			Status:       domain.StatusDown,
			FailureCount: 3,
		}

		// Create incident
		result := domain.CheckResult{
			Status:       "down",
			ResponseTime: 0,
			ResponseData: "Connection timeout",
		}

		err := service.CreateIncident(ctx, resource, result)
		require.NoError(t, err)

		// Verify incident was created
		incidents, err := incidentRepo.List(ctx, 10, 0)
		require.NoError(t, err)
		assert.Len(t, incidents, 1)

		// Verify SMTP notification was logged (even if send failed)
		notifications, err := notificationRepo.List(ctx, 10, 0)
		require.NoError(t, err)
		assert.Len(t, notifications, 1)
		assert.Equal(t, domain.NotificationEventTypeDown, notifications[0].Type)
	})

	t.Run("CreateIncident with user-configured integrations filtered by event type", func(t *testing.T) {
		// Reset repositories
		incidentRepo = fake.NewIncidentFake()
		eventStepRepo = fake.NewIncidentEventStepFake()
		notificationRepo = fake.NewNotificationFake()
		integrationRepo = fake.NewIntegrationFake()

		// Create Google Chat integration subscribed to "down" events
		googleChatConfig, _ := json.Marshal(map[string]string{
			"type":        "google_chat",
			"webhook_url": "https://chat.googleapis.com/v1/spaces/AAAA/messages?key=123",
		})
		googleChatEventTypes, _ := json.Marshal([]string{"down", "up"})

		googleChatIntegration := &domain.Integration{
			Base:       domain.Base{ID: "integration-1"},
			Name:       "Google Chat Alerts",
			IsActive:   true,
			Config:     googleChatConfig,
			EventTypes: googleChatEventTypes,
		}
		err := integrationRepo.Create(ctx, googleChatIntegration)
		require.NoError(t, err)

		// Create Slack integration subscribed only to "up" events (should be filtered out)
		slackConfig, _ := json.Marshal(map[string]string{
			"type":        "slack",
			"webhook_url": "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX",
		})
		slackEventTypes, _ := json.Marshal([]string{"up"})
		slackIntegration := &domain.Integration{
			Base:       domain.Base{ID: "integration-2"},
			Name:       "Slack Recovery Alerts",
			IsActive:   true,
			Config:     slackConfig,
			EventTypes: slackEventTypes,
		}
		err = integrationRepo.Create(ctx, slackIntegration)
		require.NoError(t, err)

		// Create incident service with SMTP disabled
		service := monitoring.NewIncidentService(
			incidentRepo,
			eventStepRepo,
			notificationRepo,
			integrationRepo,
			asynqClient,
			factory,
			false, // SMTP disabled
			"",
			"",
		)

		// Create test resource
		resource := &domain.Resource{
			Base:         domain.Base{ID: "resource-3"},
			Name:         "Test Resource 3",
			Type:         domain.ResourceHTTP,
			Target:       "https://example.com",
			Status:       domain.StatusDown,
			FailureCount: 3,
		}

		// Create incident (DOWN event)
		result := domain.CheckResult{
			Status:       "down",
			ResponseTime: 0,
			ResponseData: "Connection timeout",
		}

		err = service.CreateIncident(ctx, resource, result)
		require.NoError(t, err)

		// Verify incident was created
		incidents, err := incidentRepo.List(ctx, 10, 0)
		require.NoError(t, err)
		assert.Len(t, incidents, 1)

		// Verify notification was logged for Google Chat only (Slack should be filtered out)
		// Note: The actual send will fail since it's a placeholder, but the notification should be logged
		notifications, err := notificationRepo.List(ctx, 10, 0)
		require.NoError(t, err)
		// Should have 1 notification for Google Chat (Slack filtered out because not subscribed to "down")
		assert.Len(t, notifications, 1)
		assert.Equal(t, domain.NotificationEventTypeDown, notifications[0].Type)
	})

	t.Run("ResolveIncident with user-configured integrations filtered by event type", func(t *testing.T) {
		// Reset repositories
		incidentRepo = fake.NewIncidentFake()
		eventStepRepo = fake.NewIncidentEventStepFake()
		notificationRepo = fake.NewNotificationFake()
		integrationRepo = fake.NewIntegrationFake()

		// Create Slack integration subscribed to "up" events
		slackConfig, _ := json.Marshal(map[string]string{
			"type":        "slack",
			"webhook_url": "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX",
		})
		slackEventTypes, _ := json.Marshal([]string{"up"})
		slackIntegration := &domain.Integration{
			Base:       domain.Base{ID: "integration-3"},
			Name:       "Slack Recovery Alerts",
			IsActive:   true,
			Config:     slackConfig,
			EventTypes: slackEventTypes,
		}
		err := integrationRepo.Create(ctx, slackIntegration)
		require.NoError(t, err)

		// Create incident service with SMTP disabled
		service := monitoring.NewIncidentService(
			incidentRepo,
			eventStepRepo,
			notificationRepo,
			integrationRepo,
			asynqClient,
			factory,
			false, // SMTP disabled
			"",
			"",
		)

		// Create test resource
		resource := &domain.Resource{
			Base:         domain.Base{ID: "resource-4"},
			Name:         "Test Resource 4",
			Type:         domain.ResourceHTTP,
			Target:       "https://example.com",
			Status:       domain.StatusUp,
			FailureCount: 0,
		}

		// Create an active incident first
		incident := &domain.Incident{
			Base:       domain.Base{ID: "incident-1"},
			ResourceID: resource.ID,
			Resource:   *resource,
			Reason:     "Test incident",
			Cause:      "connection_timeout",
			ResolvedAt: nil,
			StartedAt:  time.Now().Add(-5 * time.Minute),
		}
		err = incidentRepo.Create(ctx, incident)
		require.NoError(t, err)

		// Resolve the incident (UP event)
		result := domain.CheckResult{
			Status:       "up",
			ResponseTime: 100,
			ResponseData: "OK",
		}

		err = service.ResolveIncident(ctx, resource, result)
		require.NoError(t, err)

		// Verify incident was resolved
		incidents, err := incidentRepo.List(ctx, 10, 0)
		require.NoError(t, err)
		assert.Len(t, incidents, 1)
		assert.NotNil(t, incidents[0].ResolvedAt) // Should be resolved

		// Verify notification was logged for Slack (subscribed to "up" events)
		notifications, err := notificationRepo.List(ctx, 10, 0)
		require.NoError(t, err)
		assert.Len(t, notifications, 1)
		assert.Equal(t, domain.NotificationEventTypeUp, notifications[0].Type)
	})
}
