package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func seedPendingRetryFixture(t *testing.T, eventType domain.NotificationEventType, createdAt time.Time) (*PendingNotificationRetryService, *fake.NotificationFake, *domain.NotificationEvent, *fake.NotificationChannelFake) {
	t.Helper()

	notificationRepo := fake.NewNotificationFake()
	incidentRepo := fake.NewIncidentFake()
	channelRepo := fake.NewNotificationChannelFake()
	componentRepo := fake.NewComponentFake()

	resource := domain.Resource{
		Base:     domain.Base{ID: "res-pending-retry"},
		Name:     "pending-retry-resource",
		Type:     domain.ResourceHTTP,
		Target:   "https://example.com",
		Status:   domain.StatusDown,
		Interval: 60,
		Timeout:  10,
		IsActive: true,
	}

	incident := &domain.Incident{
		Base:       domain.Base{ID: "inc-pending-retry"},
		ResourceID: resource.ID,
		Resource:   resource,
		StartedAt:  createdAt.Add(-5 * time.Minute),
		Cause:      "connection_timeout",
	}

	_, err := incidentRepo.Create(context.Background(), incident)
	require.NoError(t, err)

	event := &domain.NotificationEvent{
		Base: domain.Base{
			CreatedAt: createdAt,
			UpdatedAt: createdAt,
		},
		IncidentID: incident.ID,
		Type:       eventType,
		Status:     domain.NotificationEventStatusPending,
	}
	require.NoError(t, notificationRepo.Create(context.Background(), event))

	svc := NewPendingNotificationRetryService(
		notificationRepo,
		incidentRepo,
		channelRepo,
		componentRepo,
		"test-worker",
		24*time.Hour,
	)
	svc.now = func() time.Time { return createdAt.Add(1 * time.Hour) }

	return svc, notificationRepo, event, channelRepo
}

func TestPendingNotificationRetryService_RetrySuccessMarksSent(t *testing.T) {
	now := time.Now().UTC()
	svc, notificationRepo, event, channelRepo := seedPendingRetryFixture(t, domain.NotificationEventTypeDown, now)

	var hits int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	channel := &domain.NotificationChannel{
		Base:   domain.Base{ID: "ch-webhook-success"},
		Name:   "webhook-success",
		Type:   domain.NotificationChannelType("webhook"),
		Config: []byte(`{"url":"` + server.URL + `"}`),
	}
	require.NoError(t, channelRepo.Create(context.Background(), channel))
	channelRepo.AssociateChannelWithResource("res-pending-retry", channel.ID)

	summary, err := svc.RetryPendingNotifications(context.Background(), 10)
	require.NoError(t, err)

	assert.Equal(t, 1, summary.ScannedCount)
	assert.Equal(t, 1, summary.RetriedCount)
	assert.Equal(t, int32(1), atomic.LoadInt32(&hits))

	stored, err := notificationRepo.FindByID(context.Background(), event.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.NotificationEventStatusSent, stored.Status)
	assert.NotNil(t, stored.ProcessedAt)
	assert.Nil(t, stored.ClaimOwner)
	assert.Nil(t, stored.ClaimedAt)
}

func TestPendingNotificationRetryService_RetryFailureMarksFailed(t *testing.T) {
	now := time.Now().UTC()
	svc, notificationRepo, event, channelRepo := seedPendingRetryFixture(t, domain.NotificationEventTypeUp, now)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	channel := &domain.NotificationChannel{
		Base:   domain.Base{ID: "ch-webhook-fail"},
		Name:   "webhook-fail",
		Type:   domain.NotificationChannelType("webhook"),
		Config: []byte(`{"url":"` + server.URL + `"}`),
	}
	require.NoError(t, channelRepo.Create(context.Background(), channel))
	channelRepo.AssociateChannelWithResource("res-pending-retry", channel.ID)

	summary, err := svc.RetryPendingNotifications(context.Background(), 10)
	require.NoError(t, err)

	assert.Equal(t, 1, summary.FailedCount)
	assert.Equal(t, 0, summary.RetriedCount)

	stored, err := notificationRepo.FindByID(context.Background(), event.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.NotificationEventStatusFailed, stored.Status)
	assert.NotEmpty(t, stored.LastError)
	assert.NotNil(t, stored.ProcessedAt)
}

func TestPendingNotificationRetryService_StaleMarksExpiredAndSkipsDispatch(t *testing.T) {
	createdAt := time.Now().UTC().Add(-26 * time.Hour)
	svc, notificationRepo, event, channelRepo := seedPendingRetryFixture(t, domain.NotificationEventTypeDown, createdAt)
	svc.now = func() time.Time { return createdAt.Add(27 * time.Hour) }

	var hits int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	channel := &domain.NotificationChannel{
		Base:   domain.Base{ID: "ch-webhook-stale"},
		Name:   "webhook-stale",
		Type:   domain.NotificationChannelType("webhook"),
		Config: []byte(`{"url":"` + server.URL + `"}`),
	}
	require.NoError(t, channelRepo.Create(context.Background(), channel))
	channelRepo.AssociateChannelWithResource("res-pending-retry", channel.ID)

	summary, err := svc.RetryPendingNotifications(context.Background(), 10)
	require.NoError(t, err)

	assert.Equal(t, 1, summary.ExpiredCount)
	assert.Equal(t, 0, summary.RetriedCount)
	assert.Equal(t, int32(0), atomic.LoadInt32(&hits))

	stored, err := notificationRepo.FindByID(context.Background(), event.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.NotificationEventStatusExpired, stored.Status)
	assert.Contains(t, stored.LastError, "expired")
	assert.NotNil(t, stored.ProcessedAt)
}

func TestPendingNotificationRetryService_ConcurrentRunsDoNotDuplicateDispatch(t *testing.T) {
	now := time.Now().UTC()
	notificationRepo := fake.NewNotificationFake()
	incidentRepo := fake.NewIncidentFake()
	channelRepo := fake.NewNotificationChannelFake()
	componentRepo := fake.NewComponentFake()

	resource := domain.Resource{
		Base:     domain.Base{ID: "res-concurrent"},
		Name:     "resource-concurrent",
		Type:     domain.ResourceHTTP,
		Target:   "https://example.com/concurrent",
		Status:   domain.StatusDown,
		Interval: 60,
		Timeout:  10,
		IsActive: true,
	}
	incident := &domain.Incident{
		Base:       domain.Base{ID: "inc-concurrent"},
		ResourceID: resource.ID,
		Resource:   resource,
		StartedAt:  now.Add(-5 * time.Minute),
		Cause:      "connection_timeout",
	}
	_, err := incidentRepo.Create(context.Background(), incident)
	require.NoError(t, err)

	event := &domain.NotificationEvent{
		Base:       domain.Base{CreatedAt: now, UpdatedAt: now},
		IncidentID: incident.ID,
		Type:       domain.NotificationEventTypeDown,
		Status:     domain.NotificationEventStatusPending,
	}
	require.NoError(t, notificationRepo.Create(context.Background(), event))

	var hits int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		atomic.AddInt32(&hits, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	channel := &domain.NotificationChannel{
		Base:   domain.Base{ID: "ch-concurrent"},
		Name:   "webhook-concurrent",
		Type:   domain.NotificationChannelType("webhook"),
		Config: []byte(`{"url":"` + server.URL + `"}`),
	}
	require.NoError(t, channelRepo.Create(context.Background(), channel))
	channelRepo.AssociateChannelWithResource(resource.ID, channel.ID)

	svcA := NewPendingNotificationRetryService(notificationRepo, incidentRepo, channelRepo, componentRepo, "worker-a", 24*time.Hour)
	svcB := NewPendingNotificationRetryService(notificationRepo, incidentRepo, channelRepo, componentRepo, "worker-b", 24*time.Hour)
	svcA.now = func() time.Time { return now.Add(1 * time.Minute) }
	svcB.now = func() time.Time { return now.Add(1 * time.Minute) }

	var wg sync.WaitGroup
	var summaryA, summaryB PendingNotificationRetrySummary
	wg.Add(2)
	go func() {
		defer wg.Done()
		summaryA, _ = svcA.RetryPendingNotifications(context.Background(), 10)
	}()
	go func() {
		defer wg.Done()
		summaryB, _ = svcB.RetryPendingNotifications(context.Background(), 10)
	}()
	wg.Wait()

	assert.Equal(t, int32(1), atomic.LoadInt32(&hits), "notification should be dispatched once")
	assert.Equal(t, 1, summaryA.RetriedCount+summaryB.RetriedCount, "only one retry pass can process the claimed event")
	assert.GreaterOrEqual(t, summaryA.SkippedClaimedCount+summaryB.SkippedClaimedCount, 1, "second pass should skip claimed event")

	stored, err := notificationRepo.FindByID(context.Background(), event.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.NotificationEventStatusSent, stored.Status)
}
