package worker

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/monitoring"
	"github.com/denisakp/pulseguard/internal/repository/fake"
	"github.com/denisakp/pulseguard/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type notificationRecorder struct {
	mu     sync.Mutex
	events []string
}

func (r *notificationRecorder) handler(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	var body map[string]any
	_ = json.NewDecoder(req.Body).Decode(&body)
	event, _ := body["event"].(string)
	if event == "" {
		event, _ = body["type"].(string)
	}
	r.mu.Lock()
	r.events = append(r.events, event)
	r.mu.Unlock()
	w.WriteHeader(http.StatusOK)
}

func (r *notificationRecorder) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.events)
}

func (r *notificationRecorder) last() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.events) == 0 {
		return ""
	}
	return r.events[len(r.events)-1]
}

func setupMonitoringHandlerForAlertingTests(
	t *testing.T,
	resource *domain.Resource,
	strategy *mutableStrategy,
) (*MonitoringTaskHandler, *fake.ResourceFake, *fake.IncidentFake, *fake.IncidentEventStepFake, *notificationRecorder, func()) {
	t.Helper()

	resourceRepo := fake.NewResourceFake()
	activityRepo := fake.NewMonitoringActivityFake()
	maintenanceRepo := &noMaintenanceRepo{}
	diagnosticsRepo := fake.NewIncidentDiagnosticsFake()
	incidentRepo := fake.NewIncidentFake()
	eventStepRepoInterface := fake.NewIncidentEventStepFake()
	eventStepRepo := eventStepRepoInterface.(*fake.IncidentEventStepFake)
	notificationRepo := fake.NewNotificationFake()
	notificationChannelRepo := fake.NewNotificationChannelFake()
	recorder := &notificationRecorder{}
	server := httptest.NewServer(http.HandlerFunc(recorder.handler))

	channelConfig, err := json.Marshal(map[string]string{"url": server.URL})
	require.NoError(t, err)

	channel := &domain.NotificationChannel{
		Base:   domain.Base{ID: "channel-1"},
		Name:   "webhook",
		Type:   domain.NotificationChannelType("webhook"),
		Config: channelConfig,
	}
	require.NoError(t, notificationChannelRepo.Create(context.Background(), channel))

	incidentService := monitoring.NewIncidentService(
		incidentRepo,
		eventStepRepoInterface,
		notificationRepo,
		notificationChannelRepo,
		diagnosticsRepo,
		nil,
	)

	if strategy == nil {
		strategy = &mutableStrategy{status: domain.StatusDown}
	}

	executor := domain.NewCheckExecutor(map[domain.ResourceType]domain.CheckStrategy{
		domain.ResourceHTTP: strategy,
	})

	created, err := resourceRepo.Create(context.Background(), resource)
	require.NoError(t, err)
	notificationChannelRepo.AssociateChannelWithResource(created.ID, channel.ID)

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

	cleanup := func() { server.Close() }
	return h, resourceRepo, incidentRepo, eventStepRepo, recorder, cleanup
}

func TestFlap_AlertSuppressedWhileFlapping(t *testing.T) {
	strategy := &mutableStrategy{status: domain.StatusDown}
	h, resourceRepo, incidentRepo, _, recorder, cleanup := setupMonitoringHandlerForAlertingTests(t, &domain.Resource{
		Name:                   "flap-suppressed",
		Type:                   domain.ResourceHTTP,
		Target:                 "https://example.com",
		Interval:               60,
		Timeout:                5,
		Status:                 domain.StatusUp,
		IsActive:               true,
		ConfirmationChecks:     2,
		FlapDetectionEnabled:   true,
		FlapThreshold:          2,
		FlapWindowSeconds:      600,
		FlapMaxDurationMinutes: 30,
	}, strategy)
	defer cleanup()

	resources, err := resourceRepo.List(context.Background(), 1, 0)
	require.NoError(t, err)
	resourceID := resources[0].ID

	processMonitoringTask(t, h, resourceID)
	strategy.SetStatus(domain.StatusUp)
	processMonitoringTask(t, h, resourceID)
	strategy.SetStatus(domain.StatusDown)
	processMonitoringTask(t, h, resourceID)
	strategy.SetStatus(domain.StatusUp)
	processMonitoringTask(t, h, resourceID)

	updated, err := resourceRepo.FindByID(context.Background(), resourceID)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusFlapping, updated.Status)
	incidents, err := incidentRepo.List(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.Len(t, incidents, 0)
	assert.Equal(t, 1, recorder.count())
	assert.Equal(t, "flapping", recorder.last())
}

func TestFlap_OneNotificationOnEntry(t *testing.T) {
	strategy := &mutableStrategy{status: domain.StatusDown}
	h, resourceRepo, _, _, recorder, cleanup := setupMonitoringHandlerForAlertingTests(t, &domain.Resource{
		Name:                   "flap-entry",
		Type:                   domain.ResourceHTTP,
		Target:                 "https://example.com",
		Interval:               60,
		Timeout:                5,
		Status:                 domain.StatusUp,
		IsActive:               true,
		ConfirmationChecks:     2,
		FlapDetectionEnabled:   true,
		FlapThreshold:          2,
		FlapWindowSeconds:      600,
		FlapMaxDurationMinutes: 30,
	}, strategy)
	defer cleanup()

	resources, err := resourceRepo.List(context.Background(), 1, 0)
	require.NoError(t, err)
	resourceID := resources[0].ID

	processMonitoringTask(t, h, resourceID)
	strategy.SetStatus(domain.StatusUp)
	processMonitoringTask(t, h, resourceID)
	strategy.SetStatus(domain.StatusDown)
	processMonitoringTask(t, h, resourceID)
	strategy.SetStatus(domain.StatusUp)
	processMonitoringTask(t, h, resourceID)
	strategy.SetStatus(domain.StatusDown)
	processMonitoringTask(t, h, resourceID)

	assert.Equal(t, 1, recorder.count())
}

func TestFlap_ExitAfterStable(t *testing.T) {
	strategy := &mutableStrategy{status: domain.StatusDown}
	h, resourceRepo, _, _, recorder, cleanup := setupMonitoringHandlerForAlertingTests(t, &domain.Resource{
		Name:                   "flap-stable",
		Type:                   domain.ResourceHTTP,
		Target:                 "https://example.com",
		Interval:               60,
		Timeout:                5,
		Status:                 domain.StatusUp,
		IsActive:               true,
		ConfirmationChecks:     2,
		FlapDetectionEnabled:   true,
		FlapThreshold:          2,
		FlapWindowSeconds:      1,
		FlapMaxDurationMinutes: 30,
	}, strategy)
	defer cleanup()

	resources, err := resourceRepo.List(context.Background(), 1, 0)
	require.NoError(t, err)
	resourceID := resources[0].ID

	processMonitoringTask(t, h, resourceID)
	strategy.SetStatus(domain.StatusUp)
	processMonitoringTask(t, h, resourceID)
	strategy.SetStatus(domain.StatusDown)
	processMonitoringTask(t, h, resourceID)

	time.Sleep(1100 * time.Millisecond)
	strategy.SetStatus(domain.StatusUp)
	processMonitoringTask(t, h, resourceID)

	updated, err := resourceRepo.FindByID(context.Background(), resourceID)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusUp, updated.Status)
	assert.Nil(t, updated.FlapStartedAt)
	assert.Equal(t, 2, recorder.count())
	assert.Equal(t, "flapping_stabilized", recorder.last())
}

func TestFlap_DisabledPerResource(t *testing.T) {
	strategy := &mutableStrategy{status: domain.StatusDown}
	h, resourceRepo, _, _, recorder, cleanup := setupMonitoringHandlerForAlertingTests(t, &domain.Resource{
		Name:                 "flap-disabled",
		Type:                 domain.ResourceHTTP,
		Target:               "https://example.com",
		Interval:             60,
		Timeout:              5,
		Status:               domain.StatusUp,
		IsActive:             true,
		ConfirmationChecks:   3,
		FlapDetectionEnabled: false,
		FlapThreshold:        2,
		FlapWindowSeconds:    600,
	}, strategy)
	defer cleanup()

	resources, err := resourceRepo.List(context.Background(), 1, 0)
	require.NoError(t, err)
	resourceID := resources[0].ID

	processMonitoringTask(t, h, resourceID)
	strategy.SetStatus(domain.StatusUp)
	processMonitoringTask(t, h, resourceID)
	strategy.SetStatus(domain.StatusDown)
	processMonitoringTask(t, h, resourceID)

	updated, err := resourceRepo.FindByID(context.Background(), resourceID)
	require.NoError(t, err)
	assert.NotEqual(t, domain.StatusFlapping, updated.Status)
	assert.Equal(t, 0, recorder.count())
}

func TestFlap_ForcedIncidentAfterMaxDuration(t *testing.T) {
	strategy := &mutableStrategy{status: domain.StatusDown}
	h, resourceRepo, incidentRepo, _, recorder, cleanup := setupMonitoringHandlerForAlertingTests(t, &domain.Resource{
		Name:                   "flap-force-incident",
		Type:                   domain.ResourceHTTP,
		Target:                 "https://example.com",
		Interval:               60,
		Timeout:                5,
		Status:                 domain.StatusUp,
		IsActive:               true,
		ConfirmationChecks:     2,
		FlapDetectionEnabled:   true,
		FlapThreshold:          2,
		FlapWindowSeconds:      600,
		FlapMaxDurationMinutes: 1,
	}, strategy)
	defer cleanup()

	resources, err := resourceRepo.List(context.Background(), 1, 0)
	require.NoError(t, err)
	resourceID := resources[0].ID

	processMonitoringTask(t, h, resourceID)
	strategy.SetStatus(domain.StatusUp)
	processMonitoringTask(t, h, resourceID)
	strategy.SetStatus(domain.StatusDown)
	processMonitoringTask(t, h, resourceID)

	updated, err := resourceRepo.FindByID(context.Background(), resourceID)
	require.NoError(t, err)
	startedAt := time.Now().Add(-2 * time.Minute)
	updated.FlapStartedAt = &startedAt
	require.NoError(t, resourceRepo.Update(context.Background(), updated))

	processMonitoringTask(t, h, resourceID)

	updated, err = resourceRepo.FindByID(context.Background(), resourceID)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusDown, updated.Status)
	incidents, err := incidentRepo.List(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.Len(t, incidents, 1)
	assert.GreaterOrEqual(t, recorder.count(), 1)
}
