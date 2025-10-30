package monitoring

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/repository/fake"
	"github.com/denisakp/pulseguard/pkg/notifier"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helpers
func setupTestService(smtpEnabled bool, smtpRecipient, smtpSender string) (*IncidentService, *fake.IncidentFake, *fake.IncidentEventStepFake, *fake.NotificationFake, *fake.IntegrationFake, *asynq.Client) {
	incidentRepo := fake.NewIncidentFake()
	eventStepRepoInterface := fake.NewIncidentEventStepFake()
	notificationRepo := fake.NewNotificationFake()
	integrationRepoInterface := fake.NewIntegrationFake()
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: "localhost:6379"})
	factory := notifier.NewNotifierFactory()

	service := NewIncidentService(
		incidentRepo,
		eventStepRepoInterface,
		notificationRepo,
		integrationRepoInterface,
		asynqClient,
		factory,
		smtpEnabled,
		smtpRecipient,
		smtpSender,
		"smtp.example.com", // SMTP host for testing
		"587",              // SMTP port
		"testuser",         // SMTP user
		"testpass",         // SMTP password
	)

	// Type assert to concrete types for test access
	eventStepRepo := eventStepRepoInterface.(*fake.IncidentEventStepFake)
	integrationRepo := integrationRepoInterface.(*fake.IntegrationFake)

	return service, incidentRepo, eventStepRepo, notificationRepo, integrationRepo, asynqClient
}

// ============================================================
// UNIT TESTS - Testing internal behavior and edge cases
// ============================================================

func TestExtractCause(t *testing.T) {
	tests := []struct {
		name     string
		result   domain.CheckResult
		expected string
	}{
		{
			name: "connection timeout",
			result: domain.CheckResult{
				Status:       "down",
				ResponseData: "connection timeout exceeded",
			},
			expected: "connection_timeout",
		},
		{
			name: "connection refused",
			result: domain.CheckResult{
				Status:       "down",
				ResponseData: "connection refused by server",
			},
			expected: "connection_refused",
		},
		{
			name: "invalid status code",
			result: domain.CheckResult{
				Status:       "down",
				ResponseData: "received status code 500",
			},
			expected: "invalid_status_code",
		},
		{
			name: "dns failure",
			result: domain.CheckResult{
				Status:       "down",
				ResponseData: "failed to resolve dns name",
			},
			expected: "dns_resolution_failure",
		},
		{
			name: "ssl error",
			result: domain.CheckResult{
				Status:       "down",
				ResponseData: "ssl certificate verification failed",
			},
			expected: "ssl_certificate_error",
		},
		{
			name: "generic down",
			result: domain.CheckResult{
				Status:       "down",
				ResponseData: "service unavailable",
			},
			expected: "health_check_failed",
		},
		{
			name: "execution error",
			result: domain.CheckResult{
				Status:       "error",
				ResponseData: "failed to execute check",
			},
			expected: "check_execution_error",
		},
		{
			name: "unknown",
			result: domain.CheckResult{
				Status:       "unknown",
				ResponseData: "",
			},
			expected: "unknown_failure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cause := extractCause(tt.result)
			assert.Equal(t, tt.expected, cause)
		})
	}
}

// ============================================================
// INTEGRATION TESTS - Testing complete workflows
// ============================================================

func TestIncidentService_CreateIncident_ThreeFailureRule(t *testing.T) {
	service, incidentRepo, _, _, _, asynqClient := setupTestService(false, "test@example.com", "sender@example.com")
	defer asynqClient.Close()

	resource := &domain.Resource{
		Base: domain.Base{
			ID:        "resource-1",
			CreatedAt: time.Now(),
		},
		Name:         "Test Resource",
		Type:         domain.ResourceHTTP,
		Target:       "https://example.com",
		Status:       domain.StatusDown,
		FailureCount: 3, // At 3rd failure
		IsActive:     true,
	}

	result := domain.CheckResult{
		Status:       "down",
		ResponseData: "connection timeout",
		ResponseTime: 5 * time.Second,
	}

	ctx := context.Background()

	err := service.CreateIncident(ctx, resource, result)
	require.NoError(t, err)

	// Verify incident was created with correct cause
	incidents, err := incidentRepo.FindByResource(ctx, resource.ID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, incidents, 1)
	assert.Equal(t, "connection_timeout", incidents[0].Cause)
	assert.Nil(t, incidents[0].ResolvedAt, "New incident should be unresolved")
}

func TestIncidentService_CreateIncident_IdempotentCheck(t *testing.T) {
	service, incidentRepo, _, _, _, asynqClient := setupTestService(false, "test@example.com", "sender@example.com")
	defer asynqClient.Close()

	resource := &domain.Resource{
		Base: domain.Base{
			ID:        "resource-1",
			CreatedAt: time.Now(),
		},
		Name:         "Test Resource",
		Type:         domain.ResourceHTTP,
		Target:       "https://example.com",
		Status:       domain.StatusDown,
		FailureCount: 3,
		IsActive:     true,
	}

	result := domain.CheckResult{
		Status:       "down",
		ResponseData: "connection timeout",
		ResponseTime: 5 * time.Second,
	}

	ctx := context.Background()

	// Create incident first time
	err := service.CreateIncident(ctx, resource, result)
	require.NoError(t, err)

	// Try to create incident again (should be idempotent)
	err = service.CreateIncident(ctx, resource, result)
	require.NoError(t, err)

	// Should still only have one incident
	incidents, err := incidentRepo.FindByResource(ctx, resource.ID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, incidents, 1, "Should not create duplicate incident")
}

func TestIncidentService_SMTPDisabled_NoNotifications(t *testing.T) {
	service, incidentRepo, _, notificationRepo, _, asynqClient := setupTestService(false, "", "")
	defer asynqClient.Close()

	resource := &domain.Resource{
		Base: domain.Base{
			ID:        "resource-1",
			CreatedAt: time.Now(),
		},
		Name:         "Test Resource",
		Type:         domain.ResourceHTTP,
		Target:       "https://example.com",
		Status:       domain.StatusDown,
		FailureCount: 3,
		IsActive:     true,
	}

	result := domain.CheckResult{
		Status:       "down",
		ResponseData: "connection timeout",
		ResponseTime: 5 * time.Second,
	}

	ctx := context.Background()

	err := service.CreateIncident(ctx, resource, result)
	require.NoError(t, err)

	// Verify incident was created
	incidents, err := incidentRepo.List(ctx, 10, 0)
	require.NoError(t, err)
	assert.Len(t, incidents, 1)

	// With SMTP disabled and no integrations, no notifications should be created
	notifications, err := notificationRepo.List(ctx, 10, 0)
	require.NoError(t, err)
	assert.Len(t, notifications, 0)
}

func TestIncidentService_SMTPEnabled_CreatesNotification(t *testing.T) {
	service, incidentRepo, _, notificationRepo, _, asynqClient := setupTestService(true, "admin@example.com", "noreply@example.com")
	defer asynqClient.Close()

	resource := &domain.Resource{
		Base:         domain.Base{ID: "resource-1"},
		Name:         "Test Resource",
		Type:         domain.ResourceHTTP,
		Target:       "https://example.com",
		Status:       domain.StatusDown,
		FailureCount: 3,
	}

	result := domain.CheckResult{
		Status:       "down",
		ResponseTime: 0,
		ResponseData: "Connection timeout",
	}

	ctx := context.Background()

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
}

func TestIncidentService_ResolveIncident(t *testing.T) {
	service, incidentRepo, _, _, _, asynqClient := setupTestService(false, "test@example.com", "sender@example.com")
	defer asynqClient.Close()

	resource := &domain.Resource{
		Base: domain.Base{
			ID:        "resource-1",
			CreatedAt: time.Now(),
		},
		Name:         "Test Resource",
		Type:         domain.ResourceHTTP,
		Target:       "https://example.com",
		Status:       domain.StatusUp,
		FailureCount: 0,
		IsActive:     true,
	}

	// Create an active incident first
	activeIncident := &domain.Incident{
		Base: domain.Base{
			ID:        "incident-1",
			CreatedAt: time.Now(),
		},
		ResourceID: resource.ID,
		Resource:   *resource,
		Cause:      "connection_timeout",
		ResolvedAt: nil, // Active
		StartedAt:  time.Now().Add(-10 * time.Minute),
	}

	ctx := context.Background()
	_, err := incidentRepo.Create(ctx, activeIncident)
	require.NoError(t, err)

	result := domain.CheckResult{
		Status:       "up",
		ResponseData: "success",
		ResponseTime: 100 * time.Millisecond,
	}

	err = service.ResolveIncident(ctx, resource, result)
	require.NoError(t, err)

	// Verify incident was resolved
	resolvedIncident, err := incidentRepo.FindByID(ctx, activeIncident.ID)
	require.NoError(t, err)
	assert.NotNil(t, resolvedIncident.ResolvedAt, "Incident should be resolved")
}

// ============================================================
// TWO-LAYER NOTIFICATION ARCHITECTURE TESTS
// ============================================================

func TestIncidentService_IntegrationFiltering_DownEvent(t *testing.T) {
	service, incidentRepo, _, notificationRepo, integrationRepo, asynqClient := setupTestService(false, "", "")
	defer asynqClient.Close()

	ctx := context.Background()

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
		"webhook_url": "https://hooks.slack.com/services/T00/B00/XXX",
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

	resource := &domain.Resource{
		Base:         domain.Base{ID: "resource-1"},
		Name:         "Test Resource",
		Type:         domain.ResourceHTTP,
		Target:       "https://example.com",
		Status:       domain.StatusDown,
		FailureCount: 3,
	}

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

	// Verify only Google Chat notification was logged (Slack filtered out)
	notifications, err := notificationRepo.List(ctx, 10, 0)
	require.NoError(t, err)
	assert.Len(t, notifications, 1, "Only Google Chat should send notification for 'down' event")
	assert.Equal(t, domain.NotificationEventTypeDown, notifications[0].Type)
}

func TestIncidentService_IntegrationFiltering_UpEvent(t *testing.T) {
	service, incidentRepo, _, notificationRepo, integrationRepo, asynqClient := setupTestService(false, "", "")
	defer asynqClient.Close()

	ctx := context.Background()

	// Create Slack integration subscribed to "up" events
	slackConfig, _ := json.Marshal(map[string]string{
		"type":        "slack",
		"webhook_url": "https://hooks.slack.com/services/T00/B00/XXX",
	})
	slackEventTypes, _ := json.Marshal([]string{"up"})
	slackIntegration := &domain.Integration{
		Base:       domain.Base{ID: "integration-1"},
		Name:       "Slack Recovery Alerts",
		IsActive:   true,
		Config:     slackConfig,
		EventTypes: slackEventTypes,
	}
	err := integrationRepo.Create(ctx, slackIntegration)
	require.NoError(t, err)

	resource := &domain.Resource{
		Base:         domain.Base{ID: "resource-1"},
		Name:         "Test Resource",
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
		Cause:      "connection_timeout",
		ResolvedAt: nil,
		StartedAt:  time.Now().Add(-5 * time.Minute),
	}
	_, err = incidentRepo.Create(ctx, incident)
	require.NoError(t, err)

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
	assert.NotNil(t, incidents[0].ResolvedAt, "Incident should be resolved")

	// Verify Slack notification was logged (subscribed to "up" events)
	notifications, err := notificationRepo.List(ctx, 10, 0)
	require.NoError(t, err)
	assert.Len(t, notifications, 1)
	assert.Equal(t, domain.NotificationEventTypeUp, notifications[0].Type)
}

func TestIncidentService_TwoLayerArchitecture_SMTPAndIntegrations(t *testing.T) {
	service, incidentRepo, _, notificationRepo, integrationRepo, asynqClient := setupTestService(true, "admin@example.com", "noreply@example.com")
	defer asynqClient.Close()

	ctx := context.Background()

	// Create Google Chat integration
	googleChatConfig, _ := json.Marshal(map[string]string{
		"type":        "google_chat",
		"webhook_url": "https://chat.googleapis.com/v1/spaces/AAAA/messages?key=123",
	})
	googleChatEventTypes, _ := json.Marshal([]string{"down"})

	googleChatIntegration := &domain.Integration{
		Base:       domain.Base{ID: "integration-1"},
		Name:       "Google Chat Alerts",
		IsActive:   true,
		Config:     googleChatConfig,
		EventTypes: googleChatEventTypes,
	}
	err := integrationRepo.Create(ctx, googleChatIntegration)
	require.NoError(t, err)

	resource := &domain.Resource{
		Base:         domain.Base{ID: "resource-1"},
		Name:         "Test Resource",
		Type:         domain.ResourceHTTP,
		Target:       "https://example.com",
		Status:       domain.StatusDown,
		FailureCount: 3,
	}

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

	// Verify both SMTP and integration notifications were logged
	notifications, err := notificationRepo.List(ctx, 10, 0)
	require.NoError(t, err)
	assert.Len(t, notifications, 2, "Should have notifications from both SMTP (stage 1) and integration (stage 2)")

	// Verify notification types
	notifTypes := make(map[domain.NotificationEventType]int)
	for _, notif := range notifications {
		notifTypes[notif.Type]++
	}
	assert.Equal(t, 2, notifTypes[domain.NotificationEventTypeDown], "Both SMTP and Google Chat should send down notifications")
}
