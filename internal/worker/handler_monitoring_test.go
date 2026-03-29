package worker

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/monitoring"
	"github.com/denisakp/ogoune/internal/repository"
	"github.com/denisakp/ogoune/internal/repository/fake"
	"github.com/denisakp/ogoune/internal/service"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMonitoringHandlerForEnrichmentTests(
	t *testing.T,
	resource *domain.Resource,
	strategy domain.CheckStrategy,
) (*MonitoringTaskHandler, *fake.ResourceFake, *fake.IncidentFake, repository.IncidentDiagnosticsRepository) {
	t.Helper()

	resourceRepo := fake.NewResourceFake()
	activityRepo := fake.NewMonitoringActivityFake()
	maintenanceRepo := &noMaintenanceRepo{}
	diagnosticsRepo := fake.NewIncidentDiagnosticsFake()
	incidentRepo := fake.NewIncidentFake()
	eventStepRepo := fake.NewIncidentEventStepFake()
	notificationRepo := fake.NewNotificationFake()
	notificationChannelRepo := fake.NewNotificationChannelFake()
	incidentService := monitoring.NewIncidentService(
		incidentRepo,
		eventStepRepo,
		notificationRepo,
		notificationChannelRepo,
		diagnosticsRepo,
		nil,
	)

	executor := domain.NewCheckExecutor(map[domain.ResourceType]domain.CheckStrategy{
		domain.ResourceHTTP: strategy,
		domain.ResourceICMP: strategy,
	})

	created, err := resourceRepo.Create(context.Background(), resource)
	require.NoError(t, err)
	require.NotNil(t, created)

	h := NewMonitoringTaskHandler(
		resourceRepo,
		activityRepo,
		maintenanceRepo,
		diagnosticsRepo,
		executor,
		incidentService,
		(*service.ComponentService)(nil),
		&intervalCaptureScheduler{},
	)

	return h, resourceRepo, incidentRepo, diagnosticsRepo
}

func TestMonitoringHandler_EnrichesDiagnosticsForDownNonICMP(t *testing.T) {
	t.Setenv("ENABLE_ICMP", "true")

	strategy := &mutableStrategy{status: domain.StatusDown}
	h, resourceRepo, incidentRepo, diagnosticsRepo := setupMonitoringHandlerForEnrichmentTests(t, &domain.Resource{
		Name:               "http-down-enrichment",
		Type:               domain.ResourceHTTP,
		Target:             "https://example.com",
		Interval:           60,
		Timeout:            5,
		Status:             domain.StatusUp,
		IsActive:           true,
		ConfirmationChecks: 1,
	}, strategy)

	resources, err := resourceRepo.List(context.Background(), 1, 0)
	require.NoError(t, err)
	require.Len(t, resources, 1)

	payload, err := json.Marshal(map[string]string{"resource_id": resources[0].ID})
	require.NoError(t, err)
	require.NoError(t, h.ProcessTask(context.Background(), asynq.NewTask("monitoring:check", payload)))

	incidents, err := incidentRepo.List(context.Background(), 10, 0)
	require.NoError(t, err)
	require.Len(t, incidents, 1)

	diag, err := diagnosticsRepo.FindByIncidentID(context.Background(), incidents[0].ID)
	require.NoError(t, err)
	require.NotNil(t, diag)
	assert.NotEmpty(t, diag.RootCauseHint)
	assert.NotNil(t, diag.ICMPAvailable)
}

func TestMonitoringHandler_EnrichesDiagnosticsForDownICMP_NoSecondProbePolicy(t *testing.T) {
	t.Setenv("ENABLE_ICMP", "true")

	strategy := &mutableStrategy{status: domain.StatusDown}
	h, resourceRepo, incidentRepo, diagnosticsRepo := setupMonitoringHandlerForEnrichmentTests(t, &domain.Resource{
		Name:               "icmp-down-enrichment",
		Type:               domain.ResourceICMP,
		Target:             "1.1.1.1",
		Interval:           60,
		Timeout:            5,
		Status:             domain.StatusUp,
		IsActive:           true,
		ConfirmationChecks: 1,
	}, strategy)

	resources, err := resourceRepo.List(context.Background(), 1, 0)
	require.NoError(t, err)
	require.Len(t, resources, 1)

	payload, err := json.Marshal(map[string]string{"resource_id": resources[0].ID})
	require.NoError(t, err)
	require.NoError(t, h.ProcessTask(context.Background(), asynq.NewTask("monitoring:check", payload)))

	incidents, err := incidentRepo.List(context.Background(), 10, 0)
	require.NoError(t, err)
	require.Len(t, incidents, 1)

	diag, err := diagnosticsRepo.FindByIncidentID(context.Background(), incidents[0].ID)
	require.NoError(t, err)
	require.NotNil(t, diag)
	assert.Equal(t, "host_unreachable", diag.RootCauseHint)
	require.NotNil(t, diag.ICMPAvailable)
	require.NotNil(t, diag.ICMPReachable)
	assert.Equal(t, false, *diag.ICMPReachable)
	assert.Nil(t, diag.ICMPRttMs)
}

func TestMonitoringHandler_EnrichmentSkippedWhenICMPDisabled(t *testing.T) {
	t.Setenv("ENABLE_ICMP", "false")

	strategy := &mutableStrategy{status: domain.StatusDown}
	h, resourceRepo, incidentRepo, diagnosticsRepo := setupMonitoringHandlerForEnrichmentTests(t, &domain.Resource{
		Name:               "http-down-no-enrichment",
		Type:               domain.ResourceHTTP,
		Target:             "https://example.com",
		Interval:           60,
		Timeout:            5,
		Status:             domain.StatusUp,
		IsActive:           true,
		ConfirmationChecks: 1,
	}, strategy)

	resources, err := resourceRepo.List(context.Background(), 1, 0)
	require.NoError(t, err)
	require.Len(t, resources, 1)

	payload, err := json.Marshal(map[string]string{"resource_id": resources[0].ID})
	require.NoError(t, err)
	require.NoError(t, h.ProcessTask(context.Background(), asynq.NewTask("monitoring:check", payload)))

	incidents, err := incidentRepo.List(context.Background(), 10, 0)
	require.NoError(t, err)
	require.Len(t, incidents, 1)

	diag, err := diagnosticsRepo.FindByIncidentID(context.Background(), incidents[0].ID)
	require.NoError(t, err)
	require.NotNil(t, diag)
	assert.Empty(t, diag.RootCauseHint)
	assert.Nil(t, diag.ICMPAvailable)
	assert.Nil(t, diag.ICMPReachable)
	assert.Nil(t, diag.ICMPRttMs)
}
