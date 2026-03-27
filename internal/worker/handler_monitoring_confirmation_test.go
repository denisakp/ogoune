package worker

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/monitoring"
	"github.com/denisakp/pulseguard/internal/repository/fake"
	"github.com/denisakp/pulseguard/internal/service"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mutableStrategy struct {
	mu     sync.RWMutex
	status domain.ResourceStatus
}

func (s *mutableStrategy) SetStatus(status domain.ResourceStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status = status
}

func (s *mutableStrategy) Execute(ctx context.Context, resource *domain.Resource) (domain.CheckResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return domain.CheckResult{Status: string(s.status), ResponseTime: 5 * time.Millisecond}, nil
}

type intervalCaptureScheduler struct {
	mu        sync.Mutex
	intervals []time.Duration
}

func (s *intervalCaptureScheduler) ScheduleWithInterval(ctx context.Context, resource *domain.Resource, interval time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.intervals = append(s.intervals, interval)
	return nil
}

func (s *intervalCaptureScheduler) Last() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.intervals) == 0 {
		return 0
	}
	return s.intervals[len(s.intervals)-1]
}

type noMaintenanceRepo struct{}

func (r *noMaintenanceRepo) Create(ctx context.Context, m *domain.Maintenance) (*domain.Maintenance, error) {
	return m, nil
}
func (r *noMaintenanceRepo) FindByID(ctx context.Context, id string) (*domain.Maintenance, error) {
	return nil, nil
}
func (r *noMaintenanceRepo) List(ctx context.Context, status string, limit, offset int) ([]*domain.Maintenance, error) {
	return nil, nil
}
func (r *noMaintenanceRepo) Update(ctx context.Context, m *domain.Maintenance) error { return nil }
func (r *noMaintenanceRepo) Delete(ctx context.Context, id string) error             { return nil }
func (r *noMaintenanceRepo) FindActiveForResource(ctx context.Context, resourceID string, now time.Time) ([]*domain.Maintenance, error) {
	return []*domain.Maintenance{}, nil
}

func setupMonitoringHandlerForConfirmationTests(
	t *testing.T,
	resource *domain.Resource,
	strategy *mutableStrategy,
) (*MonitoringTaskHandler, *fake.ResourceFake, *fake.IncidentFake, *fake.IncidentEventStepFake, *intervalCaptureScheduler) {
	t.Helper()

	resourceRepo := fake.NewResourceFake()
	activityRepo := fake.NewMonitoringActivityFake()
	maintenanceRepo := &noMaintenanceRepo{}
	diagnosticsRepo := fake.NewIncidentDiagnosticsFake()
	incidentRepo := fake.NewIncidentFake()
	eventStepRepo := fake.NewIncidentEventStepFake()
	eventStepFake, ok := eventStepRepo.(*fake.IncidentEventStepFake)
	require.True(t, ok)
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
	intervalSpy := &intervalCaptureScheduler{}

	if strategy == nil {
		strategy = &mutableStrategy{status: domain.StatusDown}
	}

	executor := domain.NewCheckExecutor(map[domain.ResourceType]domain.CheckStrategy{
		domain.ResourceHTTP: strategy,
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
		intervalSpy,
	)

	return h, resourceRepo, incidentRepo, eventStepFake, intervalSpy
}

func processMonitoringTask(t *testing.T, h *MonitoringTaskHandler, resourceID string) {
	t.Helper()
	payload, err := json.Marshal(map[string]string{"resource_id": resourceID})
	require.NoError(t, err)
	require.NoError(t, h.ProcessTask(context.Background(), asynq.NewTask("monitoring:check", payload)))
}

func TestMonitoringConfirmation_NoIncidentOnFirstFailure(t *testing.T) {
	strategy := &mutableStrategy{status: domain.StatusDown}
	h, resourceRepo, incidentMgr, _, intervalSpy := setupMonitoringHandlerForConfirmationTests(t, &domain.Resource{
		Name:                 "first-failure",
		Type:                 domain.ResourceHTTP,
		Target:               "https://example.com",
		Interval:             60,
		Timeout:              5,
		Status:               domain.StatusUp,
		IsActive:             true,
		ConfirmationChecks:   3,
		ConfirmationInterval: 10,
	}, strategy)

	resources, err := resourceRepo.List(context.Background(), 1, 0)
	require.NoError(t, err)
	require.Len(t, resources, 1)

	processMonitoringTask(t, h, resources[0].ID)

	updated, err := resourceRepo.FindByID(context.Background(), resources[0].ID)
	require.NoError(t, err)
	assert.Equal(t, 1, updated.FailureCount)
	incidents, err := incidentMgr.List(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.Len(t, incidents, 0)
	assert.Equal(t, 10*time.Second, intervalSpy.Last())
}

func TestMonitoringConfirmation_IncidentCreatedAtExactThreshold(t *testing.T) {
	strategy := &mutableStrategy{status: domain.StatusDown}
	h, resourceRepo, incidentMgr, _, _ := setupMonitoringHandlerForConfirmationTests(t, &domain.Resource{
		Name:                 "threshold",
		Type:                 domain.ResourceHTTP,
		Target:               "https://example.com",
		Interval:             60,
		Timeout:              5,
		Status:               domain.StatusUp,
		IsActive:             true,
		ConfirmationChecks:   2,
		ConfirmationInterval: 10,
	}, strategy)

	resources, err := resourceRepo.List(context.Background(), 1, 0)
	require.NoError(t, err)

	processMonitoringTask(t, h, resources[0].ID)
	processMonitoringTask(t, h, resources[0].ID)

	incidents, err := incidentMgr.List(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.Len(t, incidents, 1)
}

func TestMonitoringConfirmation_NoDuplicateIncidentsAfterThreshold(t *testing.T) {
	strategy := &mutableStrategy{status: domain.StatusDown}
	h, resourceRepo, incidentMgr, _, _ := setupMonitoringHandlerForConfirmationTests(t, &domain.Resource{
		Name:                 "no-duplicate",
		Type:                 domain.ResourceHTTP,
		Target:               "https://example.com",
		Interval:             60,
		Timeout:              5,
		Status:               domain.StatusUp,
		IsActive:             true,
		ConfirmationChecks:   2,
		ConfirmationInterval: 10,
	}, strategy)

	resources, err := resourceRepo.List(context.Background(), 1, 0)
	require.NoError(t, err)

	for i := 0; i < 4; i++ {
		processMonitoringTask(t, h, resources[0].ID)
	}

	incidents, err := incidentMgr.List(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.Len(t, incidents, 1)
}

func TestMonitoringConfirmation_FalsePositiveRecoveryCreatesNoIncident(t *testing.T) {
	strategy := &mutableStrategy{status: domain.StatusDown}
	h, resourceRepo, incidentMgr, _, intervalSpy := setupMonitoringHandlerForConfirmationTests(t, &domain.Resource{
		Name:                 "false-positive",
		Type:                 domain.ResourceHTTP,
		Target:               "https://example.com",
		Interval:             45,
		Timeout:              5,
		Status:               domain.StatusUp,
		IsActive:             true,
		ConfirmationChecks:   3,
		ConfirmationInterval: 9,
	}, strategy)

	resources, err := resourceRepo.List(context.Background(), 1, 0)
	require.NoError(t, err)
	resourceID := resources[0].ID

	processMonitoringTask(t, h, resourceID)
	processMonitoringTask(t, h, resourceID)

	strategy.SetStatus(domain.StatusUp)
	processMonitoringTask(t, h, resourceID)

	updated, err := resourceRepo.FindByID(context.Background(), resourceID)
	require.NoError(t, err)
	assert.Equal(t, 0, updated.FailureCount)
	assert.Equal(t, domain.StatusUp, updated.Status)
	incidents, err := incidentMgr.List(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.Len(t, incidents, 0)
	assert.Equal(t, 45*time.Second, intervalSpy.Last())
}

func TestMonitoringConfirmation_PreThresholdRecoveryResetsFailureCount(t *testing.T) {
	strategy := &mutableStrategy{status: domain.StatusDown}
	h, resourceRepo, _, _, _ := setupMonitoringHandlerForConfirmationTests(t, &domain.Resource{
		Name:                 "reset-counter",
		Type:                 domain.ResourceHTTP,
		Target:               "https://example.com",
		Interval:             30,
		Timeout:              5,
		Status:               domain.StatusUp,
		IsActive:             true,
		ConfirmationChecks:   3,
		ConfirmationInterval: 8,
	}, strategy)

	resources, err := resourceRepo.List(context.Background(), 1, 0)
	require.NoError(t, err)
	resourceID := resources[0].ID

	processMonitoringTask(t, h, resourceID)
	updated, err := resourceRepo.FindByID(context.Background(), resourceID)
	require.NoError(t, err)
	assert.Equal(t, 1, updated.FailureCount)

	strategy.SetStatus(domain.StatusUp)
	processMonitoringTask(t, h, resourceID)

	updated, err = resourceRepo.FindByID(context.Background(), resourceID)
	require.NoError(t, err)
	assert.Equal(t, 0, updated.FailureCount)
}

func TestMonitoringConfirmation_PerResourceSerializedProcessing(t *testing.T) {
	strategy := &mutableStrategy{status: domain.StatusDown}
	h, resourceRepo, incidentMgr, _, _ := setupMonitoringHandlerForConfirmationTests(t, &domain.Resource{
		Name:                 "serialized",
		Type:                 domain.ResourceHTTP,
		Target:               "https://example.com",
		Interval:             30,
		Timeout:              5,
		Status:               domain.StatusUp,
		IsActive:             true,
		ConfirmationChecks:   2,
		ConfirmationInterval: 6,
	}, strategy)

	resources, err := resourceRepo.List(context.Background(), 1, 0)
	require.NoError(t, err)
	resourceID := resources[0].ID

	var wg sync.WaitGroup
	errCh := make(chan error, 5)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			payload, err := json.Marshal(map[string]string{"resource_id": resourceID})
			if err != nil {
				errCh <- err
				return
			}
			errCh <- h.ProcessTask(context.Background(), asynq.NewTask("monitoring:check", payload))
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		require.NoError(t, err)
	}

	updated, err := resourceRepo.FindByID(context.Background(), resourceID)
	require.NoError(t, err)
	assert.Equal(t, 5, updated.FailureCount)
	incidents, err := incidentMgr.List(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.Len(t, incidents, 1)
}

func TestMonitoringConfirmation_NoIncidentOnSecondFailureBelowThreshold(t *testing.T) {
	strategy := &mutableStrategy{status: domain.StatusDown}
	h, resourceRepo, incidentMgr, _, _ := setupMonitoringHandlerForConfirmationTests(t, &domain.Resource{
		Name:                 "second-failure",
		Type:                 domain.ResourceHTTP,
		Target:               "https://example.com",
		Interval:             60,
		Timeout:              5,
		Status:               domain.StatusUp,
		IsActive:             true,
		ConfirmationChecks:   3,
		ConfirmationInterval: 10,
	}, strategy)

	resources, err := resourceRepo.List(context.Background(), 1, 0)
	require.NoError(t, err)

	processMonitoringTask(t, h, resources[0].ID)
	processMonitoringTask(t, h, resources[0].ID)

	updated, err := resourceRepo.FindByID(context.Background(), resources[0].ID)
	require.NoError(t, err)
	assert.Equal(t, 2, updated.FailureCount)

	incidents, err := incidentMgr.List(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.Len(t, incidents, 0)
}

func TestMonitoringConfirmation_EventTimestampsAfterThresholdTiming(t *testing.T) {
	strategy := &mutableStrategy{status: domain.StatusDown}
	h, resourceRepo, incidentMgr, eventSteps, _ := setupMonitoringHandlerForConfirmationTests(t, &domain.Resource{
		Name:                 "timestamp-traceability",
		Type:                 domain.ResourceHTTP,
		Target:               "https://example.com",
		Interval:             60,
		Timeout:              5,
		Status:               domain.StatusUp,
		IsActive:             true,
		ConfirmationChecks:   2,
		ConfirmationInterval: 15,
	}, strategy)

	resources, err := resourceRepo.List(context.Background(), 1, 0)
	require.NoError(t, err)
	resourceID := resources[0].ID

	firstFailureAt := time.Now()
	processMonitoringTask(t, h, resourceID)
	assert.Eventually(t, func() bool {
		incidents, err := incidentMgr.List(context.Background(), 10, 0)
		return err == nil && len(incidents) == 0
	}, time.Second, 10*time.Millisecond)

	thresholdMetAt := time.Now()
	processMonitoringTask(t, h, resourceID)

	incidents, err := incidentMgr.List(context.Background(), 10, 0)
	require.NoError(t, err)
	require.Len(t, incidents, 1)

	incidentID := incidents[0].ID
	detectedSteps := eventSteps.CountByIncidentAndStep(incidentID, domain.IncidentEventStepDetected)
	assert.Equal(t, 1, detectedSteps)

	steps := eventSteps.FindByIncidentID(incidentID)
	require.NotEmpty(t, steps)
	for _, step := range steps {
		assert.True(t, step.CreatedAt.After(firstFailureAt) || step.CreatedAt.Equal(firstFailureAt))
		assert.True(t, step.CreatedAt.After(thresholdMetAt) || step.CreatedAt.Equal(thresholdMetAt))
	}
}

func TestMonitoringConfirmation_PersistenceFailureSkipsIncidentAndRetriesNextCycle(t *testing.T) {
	strategy := &mutableStrategy{status: domain.StatusDown}
	h, resourceRepo, incidentMgr, _, intervalSpy := setupMonitoringHandlerForConfirmationTests(t, &domain.Resource{
		Name:                 "persist-fail",
		Type:                 domain.ResourceHTTP,
		Target:               "https://example.com",
		Interval:             60,
		Timeout:              5,
		Status:               domain.StatusUp,
		IsActive:             true,
		ConfirmationChecks:   1,
		ConfirmationInterval: 7,
	}, strategy)

	resources, err := resourceRepo.List(context.Background(), 1, 0)
	require.NoError(t, err)
	resourceID := resources[0].ID

	resourceRepo.FailNextUpdate(errors.New("simulated persistence error"))
	processMonitoringTask(t, h, resourceID)

	incidents, err := incidentMgr.List(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.Len(t, incidents, 0)
	assert.Equal(t, 7*time.Second, intervalSpy.Last())

	processMonitoringTask(t, h, resourceID)
	incidents, err = incidentMgr.List(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.Len(t, incidents, 1)
}

func TestMonitoringConfirmation_ImmediateIncidentWhenAlreadyAboveThresholdNoActiveIncident(t *testing.T) {
	strategy := &mutableStrategy{status: domain.StatusDown}
	h, resourceRepo, incidentMgr, _, _ := setupMonitoringHandlerForConfirmationTests(t, &domain.Resource{
		Name:                 "reconcile-above-threshold",
		Type:                 domain.ResourceHTTP,
		Target:               "https://example.com",
		Interval:             60,
		Timeout:              5,
		Status:               domain.StatusDown,
		IsActive:             true,
		FailureCount:         5,
		ConfirmationChecks:   3,
		ConfirmationInterval: 10,
	}, strategy)

	resources, err := resourceRepo.List(context.Background(), 1, 0)
	require.NoError(t, err)
	resourceID := resources[0].ID

	processMonitoringTask(t, h, resourceID)

	incidents, err := incidentMgr.List(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.Len(t, incidents, 1)
}

func TestMonitoringConfirmation_ImmediateIncidentWhenConfirmationChecksOne(t *testing.T) {
	strategy := &mutableStrategy{status: domain.StatusDown}
	h, resourceRepo, incidentMgr, _, _ := setupMonitoringHandlerForConfirmationTests(t, &domain.Resource{
		Name:                 "checks-one",
		Type:                 domain.ResourceHTTP,
		Target:               "https://example.com",
		Interval:             60,
		Timeout:              5,
		Status:               domain.StatusUp,
		IsActive:             true,
		ConfirmationChecks:   1,
		ConfirmationInterval: 10,
	}, strategy)

	resources, err := resourceRepo.List(context.Background(), 1, 0)
	require.NoError(t, err)

	processMonitoringTask(t, h, resources[0].ID)

	incidents, err := incidentMgr.List(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.Len(t, incidents, 1)
}

func TestMonitoringConfirmation_NonPositiveConfirmationChecksTreatedAsImmediate(t *testing.T) {
	strategy := &mutableStrategy{status: domain.StatusDown}
	h, resourceRepo, incidentMgr, _, _ := setupMonitoringHandlerForConfirmationTests(t, &domain.Resource{
		Name:                 "checks-zero",
		Type:                 domain.ResourceHTTP,
		Target:               "https://example.com",
		Interval:             60,
		Timeout:              5,
		Status:               domain.StatusUp,
		IsActive:             true,
		ConfirmationChecks:   0,
		ConfirmationInterval: 10,
	}, strategy)

	resources, err := resourceRepo.List(context.Background(), 1, 0)
	require.NoError(t, err)

	processMonitoringTask(t, h, resources[0].ID)

	incidents, err := incidentMgr.List(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.Len(t, incidents, 1)
}

func TestMonitoringConfirmation_InactiveResourceSkipsProcessing(t *testing.T) {
	strategy := &mutableStrategy{status: domain.StatusDown}
	h, resourceRepo, incidentMgr, _, intervalSpy := setupMonitoringHandlerForConfirmationTests(t, &domain.Resource{
		Name:                 "inactive",
		Type:                 domain.ResourceHTTP,
		Target:               "https://example.com",
		Interval:             60,
		Timeout:              5,
		Status:               domain.StatusUp,
		IsActive:             false,
		ConfirmationChecks:   1,
		ConfirmationInterval: 10,
	}, strategy)

	resources, err := resourceRepo.List(context.Background(), 1, 0)
	require.NoError(t, err)

	processMonitoringTask(t, h, resources[0].ID)

	updated, err := resourceRepo.FindByID(context.Background(), resources[0].ID)
	require.NoError(t, err)
	assert.Equal(t, 0, updated.FailureCount)
	assert.Equal(t, time.Duration(0), intervalSpy.Last())

	incidents, err := incidentMgr.List(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.Len(t, incidents, 0)
}

func TestMonitoringConfirmation_PersistenceFailureLogsAndDoesNotPanic(t *testing.T) {
	strategy := &mutableStrategy{status: domain.StatusDown}
	h, resourceRepo, incidentMgr, _, _ := setupMonitoringHandlerForConfirmationTests(t, &domain.Resource{
		Name:                 "log-on-fail",
		Type:                 domain.ResourceHTTP,
		Target:               "https://example.com",
		Interval:             60,
		Timeout:              5,
		Status:               domain.StatusUp,
		IsActive:             true,
		ConfirmationChecks:   1,
		ConfirmationInterval: 5,
	}, strategy)

	resources, err := resourceRepo.List(context.Background(), 1, 0)
	require.NoError(t, err)
	resourceID := resources[0].ID

	resourceRepo.FailNextUpdate(errors.New("forced failure"))

	originalStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	processMonitoringTask(t, h, resourceID)

	require.NoError(t, w.Close())
	os.Stdout = originalStdout
	output, err := io.ReadAll(r)
	require.NoError(t, err)
	_ = r.Close()

	assert.Contains(t, strings.ToLower(string(output)), "failed to persist failure progression")

	incidents, err := incidentMgr.List(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.Len(t, incidents, 0)
}
