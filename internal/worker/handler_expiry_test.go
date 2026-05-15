package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Minimal mock implementations
// ---------------------------------------------------------------------------

// mockResourceRepo implements the activeResourceLister interface (only FindActive is needed).
type mockExpiryResourceRepo struct {
	resources []*domain.Resource
	findErr   error
}

func (m *mockExpiryResourceRepo) FindActive(_ context.Context, _, _ int) ([]*domain.Resource, error) {
	return m.resources, m.findErr
}

// mockExpiryChannelRepo returns an empty channel list.
type mockExpiryChannelRepo struct{}

func (m *mockExpiryChannelRepo) Create(_ context.Context, _ *domain.NotificationChannel) error {
	return nil
}
func (m *mockExpiryChannelRepo) FindByID(_ context.Context, _ string) (*domain.NotificationChannel, error) {
	return nil, nil
}
func (m *mockExpiryChannelRepo) List(_ context.Context, _, _ int) ([]*domain.NotificationChannel, error) {
	return nil, nil
}
func (m *mockExpiryChannelRepo) Update(_ context.Context, _ *domain.NotificationChannel) error {
	return nil
}
func (m *mockExpiryChannelRepo) Delete(_ context.Context, _ string) error { return nil }
func (m *mockExpiryChannelRepo) FindByType(_ context.Context, _ domain.NotificationChannelType) ([]*domain.NotificationChannel, error) {
	return nil, nil
}
func (m *mockExpiryChannelRepo) FindDefaultChannels(_ context.Context) ([]*domain.NotificationChannel, error) {
	return nil, nil
}
func (m *mockExpiryChannelRepo) FindByResourceID(_ context.Context, _ string) ([]*domain.NotificationChannel, error) {
	return nil, nil
}
func (m *mockExpiryChannelRepo) FindByComponentID(_ context.Context, _ string) ([]*domain.NotificationChannel, error) {
	return nil, nil
}

// mockEnricher is a simple enricher implementation.
type mockEnricher struct {
	metadata *domain.ResourceMetaData
	err      error
}

func (m *mockEnricher) Enrich(_ context.Context, _ *domain.Resource) (*domain.ResourceMetaData, error) {
	return m.metadata, m.err
}

// mockExpiryChecker records calls to CheckAndNotify, ResetLogs, and CleanupOldLogs.
type mockExpiryChecker struct {
	checkCalls   []string // resource IDs checked
	resetCalls   []string // resource IDs reset
	cleanupCalls int
	checkErr     error
}

func (m *mockExpiryChecker) CheckAndNotify(_ context.Context, r *domain.Resource, _ []*domain.NotificationChannel) error {
	m.checkCalls = append(m.checkCalls, r.ID)
	return m.checkErr
}
func (m *mockExpiryChecker) ResetLogs(_ context.Context, resourceID, _ string) error {
	m.resetCalls = append(m.resetCalls, resourceID)
	return nil
}
func (m *mockExpiryChecker) CleanupOldLogs(_ context.Context) error {
	m.cleanupCalls++
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func httpResource(id string) *domain.Resource {
	exp := time.Now().Add(6 * 24 * time.Hour)
	return &domain.Resource{
		Base: domain.Base{ID: id},
		Type: domain.ResourceHTTP,
		Metadata: &domain.ResourceMetaData{
			SSLExpirationDate: &exp,
		},
	}
}

func tcpResource(id string) *domain.Resource {
	return &domain.Resource{
		Base: domain.Base{ID: id},
		Type: domain.ResourceTCP,
	}
}

func emptyTask() *asynq.Task {
	return asynq.NewTask(TypeExpiryCheck, nil)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

// T015-1: only HTTP resources are processed; TCP resources are skipped.
func TestExpiryTaskHandler_SkipsNonHTTPResources(t *testing.T) {
	resources := &mockExpiryResourceRepo{
		resources: []*domain.Resource{
			httpResource("http-1"),
			tcpResource("tcp-1"),
		},
	}
	checker := &mockExpiryChecker{}

	h := NewExpiryTaskHandler(resources, &mockExpiryChannelRepo{}, &mockEnricher{metadata: &domain.ResourceMetaData{}}, checker)

	require.NoError(t, h.ProcessTask(context.Background(), emptyTask()))

	assert.Equal(t, []string{"http-1"}, checker.checkCalls, "only HTTP resource should be checked")
}

// T015-2: an error for one resource does not prevent other resources from being processed.
func TestExpiryTaskHandler_PerResourceErrorIsolation(t *testing.T) {
	resources := &mockExpiryResourceRepo{
		resources: []*domain.Resource{
			httpResource("http-1"),
			httpResource("http-2"),
		},
	}

	enricher := &mockEnricher{metadata: &domain.ResourceMetaData{}}

	// Override CheckAndNotify to error only on the first resource.
	customChecker := &conditionalErrorChecker{
		errOnID: "http-1",
	}

	h := NewExpiryTaskHandler(resources, &mockExpiryChannelRepo{}, enricher, customChecker)

	require.NoError(t, h.ProcessTask(context.Background(), emptyTask()))

	// Both resources should have been attempted.
	assert.Len(t, customChecker.checkCalls, 2)
}

// T015-3: CleanupOldLogs is always called after all resources are processed.
func TestExpiryTaskHandler_CleanupCalledAfterProcessing(t *testing.T) {
	resources := &mockExpiryResourceRepo{
		resources: []*domain.Resource{httpResource("http-1")},
	}
	checker := &mockExpiryChecker{}

	h := NewExpiryTaskHandler(resources, &mockExpiryChannelRepo{}, &mockEnricher{metadata: &domain.ResourceMetaData{}}, checker)

	require.NoError(t, h.ProcessTask(context.Background(), emptyTask()))

	assert.Equal(t, 1, checker.cleanupCalls, "CleanupOldLogs must be called once per run")
}

// T015-4: when enrichment fails for a resource, it is skipped but others continue.
func TestExpiryTaskHandler_EnrichmentFailureSkipsResource(t *testing.T) {
	resources := &mockExpiryResourceRepo{
		resources: []*domain.Resource{
			httpResource("http-1"),
			httpResource("http-2"),
		},
	}
	checker := &mockExpiryChecker{}

	// Enrich always fails
	h := NewExpiryTaskHandler(resources, &mockExpiryChannelRepo{}, &mockEnricher{err: errors.New("timeout")}, checker)

	require.NoError(t, h.ProcessTask(context.Background(), emptyTask()))

	// Enrichment failed for both — CheckAndNotify should not have been called.
	assert.Empty(t, checker.checkCalls)
	assert.Equal(t, 1, checker.cleanupCalls, "cleanup always runs")
}

// T015-5: renewal detection triggers ResetLogs when SSL date has advanced.
func TestExpiryTaskHandler_RenewalDetectionResetsSSLLogs(t *testing.T) {
	old := time.Now().Add(10 * 24 * time.Hour)
	fresh := time.Now().Add(365 * 24 * time.Hour)

	resource := httpResource("http-1")
	resource.Metadata.SSLExpirationDate = &old // stored date is soon

	enricher := &mockEnricher{
		metadata: &domain.ResourceMetaData{
			SSLExpirationDate: &fresh, // fresh date is far in the future → renewal
		},
	}
	checker := &mockExpiryChecker{}

	h := NewExpiryTaskHandler(&mockExpiryResourceRepo{resources: []*domain.Resource{resource}}, &mockExpiryChannelRepo{}, enricher, checker)
	require.NoError(t, h.ProcessTask(context.Background(), emptyTask()))

	assert.Contains(t, checker.resetCalls, "http-1", "ResetLogs should be called when SSL is renewed")
}

// ---------------------------------------------------------------------------
// helper: conditionalErrorChecker errors only on a specific resource ID
// ---------------------------------------------------------------------------

type conditionalErrorChecker struct {
	errOnID    string
	checkCalls []string
}

func (c *conditionalErrorChecker) CheckAndNotify(_ context.Context, r *domain.Resource, _ []*domain.NotificationChannel) error {
	c.checkCalls = append(c.checkCalls, r.ID)
	if r.ID == c.errOnID {
		return errors.New("forced error")
	}
	return nil
}
func (c *conditionalErrorChecker) ResetLogs(_ context.Context, _ string, _ string) error {
	return nil
}
func (c *conditionalErrorChecker) CleanupOldLogs(_ context.Context) error { return nil }
