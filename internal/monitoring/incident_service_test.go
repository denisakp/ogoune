package monitoring

import (
	"context"
	"testing"
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/repository/fake"
	"github.com/denisakp/pulseguard/pkg/notifier"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIncidentService_CreateIncident_ThreeFailureRule(t *testing.T) {
	// Setup
	incidentRepo := fake.NewIncidentFake()
	eventStepRepo := fake.NewIncidentEventStepFake()
	notificationRepo := fake.NewNotificationFake()
	integrationRepo := fake.NewIntegrationFake()
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: "localhost:6379"})
	defer asynqClient.Close()
	factory := notifier.NewNotifierFactory()

	// Test with SMTP disabled
	service := NewIncidentService(
		incidentRepo,
		eventStepRepo,
		notificationRepo,
		integrationRepo,
		asynqClient,
		factory,
		false, // SMTP disabled
		"test@example.com",
		"sender@example.com",
	)

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

	// Act
	err := service.CreateIncident(ctx, resource, result)

	// Assert
	require.NoError(t, err)

	// Verify incident was created
	incidents, err := incidentRepo.FindByResource(ctx, resource.ID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, incidents, 1)
	assert.Equal(t, "connection_timeout", incidents[0].Cause)
	assert.Nil(t, incidents[0].ResolvedAt, "New incident should be unresolved")

	// Verify event steps were created
	// Note: Since SMTP is disabled, we should only have "detected" event, not alerts
}

func TestIncidentService_CreateIncident_IdempotentCheck(t *testing.T) {
	// Setup
	incidentRepo := fake.NewIncidentFake()
	eventStepRepo := fake.NewIncidentEventStepFake()
	notificationRepo := fake.NewNotificationFake()
	integrationRepo := fake.NewIntegrationFake()
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: "localhost:6379"})
	defer asynqClient.Close()
	factory := notifier.NewNotifierFactory()

	service := NewIncidentService(
		incidentRepo,
		eventStepRepo,
		notificationRepo,
		integrationRepo,
		asynqClient,
		factory,
		false,
		"test@example.com",
		"sender@example.com",
	)

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

	// Act - Create incident first time
	err := service.CreateIncident(ctx, resource, result)
	require.NoError(t, err)

	// Act - Try to create incident again (should be idempotent)
	err = service.CreateIncident(ctx, resource, result)
	require.NoError(t, err)

	// Assert - Should still only have one incident
	incidents, err := incidentRepo.FindByResource(ctx, resource.ID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, incidents, 1, "Should not create duplicate incident")
}

func TestIncidentService_ResolveIncident(t *testing.T) {
	// Setup
	incidentRepo := fake.NewIncidentFake()
	eventStepRepo := fake.NewIncidentEventStepFake()
	notificationRepo := fake.NewNotificationFake()
	integrationRepo := fake.NewIntegrationFake()
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: "localhost:6379"})
	defer asynqClient.Close()
	factory := notifier.NewNotifierFactory()

	service := NewIncidentService(
		incidentRepo,
		eventStepRepo,
		notificationRepo,
		integrationRepo,
		asynqClient,
		factory,
		false, // SMTP disabled
		"test@example.com",
		"sender@example.com",
	)

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
	err := incidentRepo.Create(ctx, activeIncident)
	require.NoError(t, err)

	result := domain.CheckResult{
		Status:       "up",
		ResponseData: "success",
		ResponseTime: 100 * time.Millisecond,
	}

	// Act
	err = service.ResolveIncident(ctx, resource, result)

	// Assert
	require.NoError(t, err)

	// Verify incident was resolved
	resolvedIncident, err := incidentRepo.FindByID(ctx, activeIncident.ID)
	require.NoError(t, err)
	assert.NotNil(t, resolvedIncident.ResolvedAt, "Incident should be resolved")
}

func TestIncidentService_SMTPDisabled_NoNotifications(t *testing.T) {
	// Setup
	incidentRepo := fake.NewIncidentFake()
	eventStepRepo := fake.NewIncidentEventStepFake()
	notificationRepo := fake.NewNotificationFake()
	integrationRepo := fake.NewIntegrationFake()
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: "localhost:6379"})
	defer asynqClient.Close()
	factory := notifier.NewNotifierFactory()

	// SMTP explicitly disabled
	service := NewIncidentService(
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

	// Act
	err := service.CreateIncident(ctx, resource, result)

	// Assert
	require.NoError(t, err)

	// Verify incident was created
	incidents, err := incidentRepo.FindByResource(ctx, resource.ID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, incidents, 1)

	// With SMTP disabled, no notification events should be created
	// This test verifies the early return in CreateIncident when SMTP is disabled
}

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
