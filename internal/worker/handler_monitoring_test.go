package worker

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/monitoring"
	"github.com/denisakp/ogoune/internal/port"
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
) (*MonitoringTaskHandler, *fake.ResourceFake, *fake.IncidentFake, port.IncidentDiagnosticsRepository) {
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
	}, &noopRecorder{})

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

// TestMissedHeartbeatCheckResult_CreatesIncidentWithCorrectCause verifies that
// a CheckResult synthesized by the heartbeat detector (Cause=MissedHeartbeat)
// results in an incident with cause "missed_heartbeat" when passed to
// IncidentService.CreateIncident. This validates the US4 incident pipeline (T037).
func TestMissedHeartbeatCheckResult_CreatesIncidentWithCorrectCause(t *testing.T) {
	incidentRepo := fake.NewIncidentFake()
	eventStepRepo := fake.NewIncidentEventStepFake()
	notificationRepo := fake.NewNotificationFake()
	notificationChannelRepo := fake.NewNotificationChannelFake()
	diagnosticsRepo := fake.NewIncidentDiagnosticsFake()

	incidentSvc := monitoring.NewIncidentService(
		incidentRepo,
		eventStepRepo,
		notificationRepo,
		notificationChannelRepo,
		diagnosticsRepo,
		nil,
	)

	resource := &domain.Resource{
		Base:     domain.Base{ID: "hb-missed-1"},
		Name:     "Backup Job",
		Type:     domain.ResourceHeartbeat,
		IsActive: true,
		Status:   domain.StatusUp,
	}

	cause := domain.MissedHeartbeat
	result := domain.CheckResult{
		Status:       string(domain.StatusDown),
		Cause:        &cause,
		ErrorMessage: "No ping received within the expected interval + grace period.",
	}

	err := incidentSvc.CreateIncident(context.Background(), resource, result)
	require.NoError(t, err)

	incidents, err := incidentRepo.List(context.Background(), 10, 0)
	require.NoError(t, err)
	require.Len(t, incidents, 1)
	assert.Equal(t, "missed_heartbeat", incidents[0].Cause)
}

// TestMissedHeartbeatIncident_DeduplicationSkipsSecondCall verifies that
// calling CreateIncident twice for the same down monitor only creates one incident (T036 dedup).
func TestMissedHeartbeatIncident_DeduplicationSkipsSecondCall(t *testing.T) {
	incidentRepo := fake.NewIncidentFake()
	eventStepRepo := fake.NewIncidentEventStepFake()
	notificationRepo := fake.NewNotificationFake()
	notificationChannelRepo := fake.NewNotificationChannelFake()
	diagnosticsRepo := fake.NewIncidentDiagnosticsFake()

	incidentSvc := monitoring.NewIncidentService(
		incidentRepo,
		eventStepRepo,
		notificationRepo,
		notificationChannelRepo,
		diagnosticsRepo,
		nil,
	)

	resource := &domain.Resource{
		Base:     domain.Base{ID: "hb-missed-dedup"},
		Name:     "Dedup Monitor",
		Type:     domain.ResourceHeartbeat,
		IsActive: true,
		Status:   domain.StatusUp,
	}

	cause := domain.MissedHeartbeat
	result := domain.CheckResult{
		Status: string(domain.StatusDown),
		Cause:  &cause,
	}

	// First call creates the incident
	err := incidentSvc.CreateIncident(context.Background(), resource, result)
	require.NoError(t, err)

	// Second call (detector re-runs) must be deduplicated
	err = incidentSvc.CreateIncident(context.Background(), resource, result)
	require.NoError(t, err)

	incidents, err := incidentRepo.List(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.Len(t, incidents, 1, "second CreateIncident call must be deduplicated")
}

// TestMissedHeartbeatRecovery_ResolveIncident verifies that calling ResolveIncident
// after a ping marks the incident as resolved with a recovery timestamp (T036 recovery).
func TestMissedHeartbeatRecovery_ResolveIncident(t *testing.T) {
	incidentRepo := fake.NewIncidentFake()
	eventStepRepo := fake.NewIncidentEventStepFake()
	notificationRepo := fake.NewNotificationFake()
	notificationChannelRepo := fake.NewNotificationChannelFake()
	diagnosticsRepo := fake.NewIncidentDiagnosticsFake()

	incidentSvc := monitoring.NewIncidentService(
		incidentRepo,
		eventStepRepo,
		notificationRepo,
		notificationChannelRepo,
		diagnosticsRepo,
		nil,
	)

	resource := &domain.Resource{
		Base:     domain.Base{ID: "hb-missed-recover"},
		Name:     "Recovery Monitor",
		Type:     domain.ResourceHeartbeat,
		IsActive: true,
		Status:   domain.StatusUp,
	}

	// Seed an unresolved incident
	started := time.Now().Add(-10 * time.Minute)
	_, err := incidentRepo.Create(context.Background(), &domain.Incident{
		ResourceID: resource.ID,
		Cause:      "missed_heartbeat",
		StartedAt:  started,
	})
	require.NoError(t, err)

	// Ping received → resolve
	recoveryResult := domain.CheckResult{Status: string(domain.StatusUp)}
	err = incidentSvc.ResolveIncident(context.Background(), resource, recoveryResult)
	require.NoError(t, err)

	incidents, err := incidentRepo.List(context.Background(), 10, 0)
	require.NoError(t, err)
	require.Len(t, incidents, 1)
	assert.NotNil(t, incidents[0].ResolvedAt, "incident must be resolved after recovery ping")
}
