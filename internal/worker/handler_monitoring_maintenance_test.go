package worker

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/monitoring"
	"github.com/denisakp/ogoune/internal/repository/fake"
	"github.com/denisakp/ogoune/internal/service"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// activeMaintenanceRepo returns active maintenances for FindActiveForResource.
type activeMaintenanceRepo struct {
	maintenances []*domain.Maintenance
}

func (r *activeMaintenanceRepo) Create(_ context.Context, m *domain.Maintenance) (*domain.Maintenance, error) {
	return m, nil
}
func (r *activeMaintenanceRepo) FindByID(_ context.Context, _ string) (*domain.Maintenance, error) {
	return nil, nil
}
func (r *activeMaintenanceRepo) List(_ context.Context, _ string, _, _ int) ([]*domain.Maintenance, error) {
	return nil, nil
}
func (r *activeMaintenanceRepo) Update(_ context.Context, _ *domain.Maintenance) error { return nil }
func (r *activeMaintenanceRepo) Delete(_ context.Context, _ string) error              { return nil }
func (r *activeMaintenanceRepo) FindActiveForResource(_ context.Context, _ string, _ time.Time) ([]*domain.Maintenance, error) {
	return r.maintenances, nil
}

func TestMonitoringMaintenance_SkipsIncidentDuringActiveMaintenance(t *testing.T) {
	strategy := &mutableStrategy{status: domain.StatusDown}

	resourceRepo := fake.NewResourceFake()
	activityRepo := fake.NewMonitoringActivityFake()
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
	}, &noopRecorder{})

	resource := &domain.Resource{
		Name:               "maintenance-resource",
		Type:               domain.ResourceHTTP,
		Target:             "https://example.com",
		Interval:           60,
		Timeout:            5,
		Status:             domain.StatusUp,
		IsActive:           true,
		ConfirmationChecks: 1,
	}

	created, err := resourceRepo.Create(context.Background(), resource)
	require.NoError(t, err)
	require.NotNil(t, created)

	now := time.Now()
	maintenanceRepo := &activeMaintenanceRepo{
		maintenances: []*domain.Maintenance{
			{
				Base:      domain.Base{ID: "maint-active-1"},
				Title:     "Scheduled Downtime",
				Strategy:  domain.OneTime,
				Status:    "active",
				StartedAt: &now,
			},
		},
	}

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

	// Process the monitoring task while maintenance is active
	payload, err := json.Marshal(map[string]string{"resource_id": created.ID})
	require.NoError(t, err)
	err = h.ProcessTask(context.Background(), asynq.NewTask("monitoring:check", payload))
	require.NoError(t, err)

	// Verify: monitoring activity was created with IsMaintenance=true
	activities, err := activityRepo.FindByResourceID(context.Background(), created.ID, 10, 0)
	require.NoError(t, err)
	require.Len(t, activities, 1)
	assert.True(t, activities[0].IsMaintenance, "activity should be marked as maintenance")

	// Verify: no incidents were created
	incidents, err := incidentRepo.List(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.Len(t, incidents, 0, "no incidents should be created during maintenance")

	// Verify: resource status was NOT changed (still UP, not DOWN)
	updated, err := resourceRepo.FindByID(context.Background(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusUp, updated.Status, "resource status should not change during maintenance")
	assert.Equal(t, 0, updated.FailureCount, "failure count should not increment during maintenance")
}
