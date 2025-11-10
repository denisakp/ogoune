package monitoring

import (
	"context"
	"testing"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/repository/fake"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helpers
func setupTestService(smtpEnabled bool, smtpRecipient, smtpSender string) (*IncidentService, *fake.IncidentFake, *fake.IncidentEventStepFake, *fake.NotificationFake, *asynq.Client) {
	incidentRepo := fake.NewIncidentFake()
	eventStepRepoInterface := fake.NewIncidentEventStepFake()
	notificationRepo := fake.NewNotificationFake()
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: "localhost:6379"})

	service := NewIncidentService(
		incidentRepo,
		eventStepRepoInterface,
		notificationRepo,
		asynqClient,
		smtpEnabled,
		smtpRecipient,
		smtpSender,
		"smtp.example.com", // SMTP host for testing
		"587",              // SMTP port
		"testuser",         // SMTP user
		"testpass",         // SMTP password
		"",                 // webhook URL (disabled for tests)
		nil,                // webhook secret (disabled for tests)
	)

	// Type assert to concrete types for test access
	eventStepRepo := eventStepRepoInterface.(*fake.IncidentEventStepFake)

	return service, incidentRepo, eventStepRepo, notificationRepo, asynqClient
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
				ResponseData: "",
			},
			expected: "health_check_failed",
		},
		{
			name: "execution error",
			result: domain.CheckResult{
				Status:       "error",
				ResponseData: "some error",
			},
			expected: "check_execution_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractCause(tt.result)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ============================================================
// INTEGRATION TESTS - Testing incident lifecycle
// ============================================================

func TestIncidentService_CreateIncident_Success(t *testing.T) {
	service, incidentRepo, eventStepRepo, notificationRepo, asynqClient := setupTestService(false, "admin@example.com", "noreply@example.com")
	defer asynqClient.Close()

	ctx := context.Background()

	// Create test resource
	resource := &domain.Resource{
		Base:         domain.Base{ID: "res-123"},
		Name:         "Example API",
		Target:       "https://example.com",
		Type:         domain.ResourceHTTP,
		Timeout:      30,
		IsActive:     true,
		FailureCount: 0,
		Status:       domain.StatusUp,
	}

	// Create test result
	result := domain.CheckResult{
		Status:       "down",
		ResponseData: "connection timeout",
	}

	// Create incident
	err := service.CreateIncident(ctx, resource, result)
	require.NoError(t, err)

	// Verify incident was created
	incidents, err := incidentRepo.FindByResource(ctx, resource.ID, 10, 0)
	require.NoError(t, err)
	require.Equal(t, 1, len(incidents))
	assert.Equal(t, resource.ID, incidents[0].ResourceID)
	assert.Nil(t, incidents[0].ResolvedAt)
	assert.Equal(t, "connection_timeout", incidents[0].Cause)

	// Verify event step was created
	steps, err := eventStepRepo.List(ctx, 10, 0)
	require.NoError(t, err)
	assert.Greater(t, len(steps), 0)
	assert.Equal(t, domain.IncidentEventStepDetected, steps[0].Step)
}

func TestIncidentService_CreateIncident_DuplicatePrevention(t *testing.T) {
	service, incidentRepo, _, _, asynqClient := setupTestService(false, "", "")
	defer asynqClient.Close()

	ctx := context.Background()

	resource := &domain.Resource{
		Base:         domain.Base{ID: "res-456"},
		Name:         "API Server",
		Target:       "https://api.example.com",
		Type:         domain.ResourceHTTP,
		IsActive:     true,
		FailureCount: 0,
		Status:       domain.StatusDown,
	}

	result := domain.CheckResult{
		Status:       "down",
		ResponseData: "connection refused",
	}

	// Create first incident
	err := service.CreateIncident(ctx, resource, result)
	require.NoError(t, err)

	// Try to create second incident - should be skipped
	err = service.CreateIncident(ctx, resource, result)
	require.NoError(t, err)

	// Verify only one incident exists
	incidents, err := incidentRepo.FindByResource(ctx, resource.ID, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, 1, len(incidents))
}

func TestIncidentService_ResolveIncident_Success(t *testing.T) {
	service, incidentRepo, eventStepRepo, _, asynqClient := setupTestService(false, "", "")
	defer asynqClient.Close()

	ctx := context.Background()

	resource := &domain.Resource{
		Base:         domain.Base{ID: "res-789"},
		Name:         "Database",
		Target:       "db.example.com:5432",
		Type:         domain.ResourceTCP,
		IsActive:     true,
		FailureCount: 3,
		Status:       domain.StatusDown,
	}

	downResult := domain.CheckResult{
		Status:       "down",
		ResponseData: "connection timeout",
	}

	// Create incident
	err := service.CreateIncident(ctx, resource, downResult)
	require.NoError(t, err)

	// Now resolve it
	upResult := domain.CheckResult{
		Status:       "up",
		ResponseData: "OK",
	}

	err = service.ResolveIncident(ctx, resource, upResult)
	require.NoError(t, err)

	// Verify incident is resolved
	incidents, err := incidentRepo.FindByResource(ctx, resource.ID, 10, 0)
	require.NoError(t, err)
	require.Equal(t, 1, len(incidents))
	assert.NotNil(t, incidents[0].ResolvedAt)

	// Verify resolved event step was created
	steps, err := eventStepRepo.List(ctx, 10, 0)
	require.NoError(t, err)
	foundResolved := false
	for _, step := range steps {
		if step.Step == domain.IncidentEventStepResolved {
			foundResolved = true
			break
		}
	}
	assert.True(t, foundResolved, "Resolved event step should have been created")
}

func TestIncidentService_ResolveIncident_NoActiveIncident(t *testing.T) {
	service, _, _, _, asynqClient := setupTestService(false, "", "")
	defer asynqClient.Close()

	ctx := context.Background()

	resource := &domain.Resource{
		Base:     domain.Base{ID: "res-999"},
		Name:     "Orphaned Resource",
		Target:   "orphan.example.com",
		Type:     domain.ResourceHTTP,
		IsActive: true,
		Status:   domain.StatusUp,
	}

	upResult := domain.CheckResult{
		Status:       "up",
		ResponseData: "OK",
	}

	// Try to resolve without creating incident - should return nil (no error)
	err := service.ResolveIncident(ctx, resource, upResult)
	require.NoError(t, err)
}

func TestIncidentService_CreateIncident_WithSMTPEnabled(t *testing.T) {
	service, incidentRepo, eventStepRepo, notificationRepo, asynqClient := setupTestService(true, "admin@example.com", "noreply@example.com")
	defer asynqClient.Close()

	ctx := context.Background()

	resource := &domain.Resource{
		Base:         domain.Base{ID: "res-smtp"},
		Name:         "SMTP Test Resource",
		Target:       "smtp-test.example.com",
		Type:         domain.ResourceHTTP,
		IsActive:     true,
		FailureCount: 0,
		Status:       domain.StatusUp,
	}

	result := domain.CheckResult{
		Status:       "down",
		ResponseData: "connection refused",
	}

	// Create incident with SMTP enabled
	err := service.CreateIncident(ctx, resource, result)
	require.NoError(t, err)

	// Verify incident was created
	incidents, err := incidentRepo.FindByResource(ctx, resource.ID, 10, 0)
	require.NoError(t, err)
	require.Equal(t, 1, len(incidents))

	// Verify event steps include down alert
	steps, err := eventStepRepo.List(ctx, 10, 0)
	require.NoError(t, err)
	foundDownAlert := false
	for _, step := range steps {
		if step.Step == domain.IncidentEventStepDownAlert {
			foundDownAlert = true
			break
		}
	}
	assert.True(t, foundDownAlert, "Down alert event step should have been created when SMTP is enabled")

	// Verify notification event was created
	notifications, err := notificationRepo.List(ctx, 10, 0)
	require.NoError(t, err)
	assert.Greater(t, len(notifications), 0)
}

func TestIncidentService_ResolveIncident_WithSMTPEnabled(t *testing.T) {
	service, incidentRepo, eventStepRepo, notificationRepo, asynqClient := setupTestService(true, "admin@example.com", "noreply@example.com")
	defer asynqClient.Close()

	ctx := context.Background()

	resource := &domain.Resource{
		Base:         domain.Base{ID: "res-resolve-smtp"},
		Name:         "Resolve SMTP Test",
		Target:       "resolve-test.example.com",
		Type:         domain.ResourceHTTP,
		IsActive:     true,
		FailureCount: 3,
		Status:       domain.StatusDown,
	}

	// Create incident
	downResult := domain.CheckResult{
		Status:       "down",
		ResponseData: "timeout",
	}
	err := service.CreateIncident(ctx, resource, downResult)
	require.NoError(t, err)

	// Resolve incident
	upResult := domain.CheckResult{
		Status:       "up",
		ResponseData: "OK",
	}
	err = service.ResolveIncident(ctx, resource, upResult)
	require.NoError(t, err)

	// Verify event steps include up alert
	steps, err := eventStepRepo.List(ctx, 10, 0)
	require.NoError(t, err)
	foundUpAlert := false
	for _, step := range steps {
		if step.Step == domain.IncidentEventStepUpAlert {
			foundUpAlert = true
			break
		}
	}
	assert.True(t, foundUpAlert, "Up alert event step should have been created when SMTP is enabled")

	// Verify notification events were created (both down and up)
	notifications, err := notificationRepo.List(ctx, 100, 0)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(notifications), 2, "Should have at least down and up notifications")
}

func TestStringPtr(t *testing.T) {
	testStr := "test"
	result := stringPtr(testStr)
	require.NotNil(t, result)
	assert.Equal(t, testStr, *result)
}
