package monitoring

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func findNotificationByType(t *testing.T, notifications []*domain.NotificationEvent, eventType domain.NotificationEventType) *domain.NotificationEvent {
	t.Helper()
	for _, n := range notifications {
		if n.Type == eventType {
			return n
		}
	}
	t.Fatalf("notification type %s not found", eventType)
	return nil
}

func TestIncidentService_CreateIncident_PersistsPendingThenMarksSent(t *testing.T) {
	service, incidentRepo, _, notificationRepo, channelRepo, asynqClient := setupTestService()
	defer asynqClient.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	channel := &domain.NotificationChannel{
		Base:   domain.Base{ID: "notif-down-sent"},
		Name:   "down-sent",
		Type:   domain.NotificationChannelType("webhook"),
		Config: []byte(`{"url":"` + server.URL + `"}`),
	}
	require.NoError(t, channelRepo.Create(context.Background(), channel))

	resource := &domain.Resource{
		Base:     domain.Base{ID: "res-notif-down"},
		Name:     "resource-down",
		Target:   "https://example.com/down",
		Type:     domain.ResourceHTTP,
		Status:   domain.StatusDown,
		Interval: 60,
		Timeout:  10,
		IsActive: true,
	}
	channelRepo.AssociateChannelWithResource(resource.ID, channel.ID)

	err := service.CreateIncident(context.Background(), resource, domain.CheckResult{Status: "down", ResponseData: "timeout"})
	require.NoError(t, err)

	incidents, err := incidentRepo.FindByResource(context.Background(), resource.ID, 10, 0)
	require.NoError(t, err)
	require.Len(t, incidents, 1)

	notifications, err := notificationRepo.FindByIncident(context.Background(), incidents[0].ID, 20, 0)
	require.NoError(t, err)
	require.NotEmpty(t, notifications)

	down := findNotificationByType(t, notifications, domain.NotificationEventTypeDown)
	assert.Equal(t, domain.NotificationEventStatusSent, down.Status)
	assert.NotNil(t, down.ProcessedAt)
}

func TestIncidentService_CreateIncident_PersistsPendingThenMarksFailed(t *testing.T) {
	service, incidentRepo, _, notificationRepo, channelRepo, asynqClient := setupTestService()
	defer asynqClient.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	channel := &domain.NotificationChannel{
		Base:   domain.Base{ID: "notif-down-failed"},
		Name:   "down-failed",
		Type:   domain.NotificationChannelType("webhook"),
		Config: []byte(`{"url":"` + server.URL + `"}`),
	}
	require.NoError(t, channelRepo.Create(context.Background(), channel))

	resource := &domain.Resource{
		Base:     domain.Base{ID: "res-notif-failed"},
		Name:     "resource-failed",
		Target:   "https://example.com/fail",
		Type:     domain.ResourceHTTP,
		Status:   domain.StatusDown,
		Interval: 60,
		Timeout:  10,
		IsActive: true,
	}
	channelRepo.AssociateChannelWithResource(resource.ID, channel.ID)

	err := service.CreateIncident(context.Background(), resource, domain.CheckResult{Status: "down", ResponseData: "timeout"})
	require.NoError(t, err)

	incidents, err := incidentRepo.FindByResource(context.Background(), resource.ID, 10, 0)
	require.NoError(t, err)
	require.Len(t, incidents, 1)

	notifications, err := notificationRepo.FindByIncident(context.Background(), incidents[0].ID, 20, 0)
	require.NoError(t, err)
	require.NotEmpty(t, notifications)

	down := findNotificationByType(t, notifications, domain.NotificationEventTypeDown)
	assert.Equal(t, domain.NotificationEventStatusFailed, down.Status)
	assert.NotEmpty(t, down.LastError)
	assert.NotNil(t, down.ProcessedAt)
}

func TestIncidentService_ResolveIncident_CreatesUpNotificationAndMarksSent(t *testing.T) {
	service, incidentRepo, _, notificationRepo, channelRepo, asynqClient := setupTestService()
	defer asynqClient.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	channel := &domain.NotificationChannel{
		Base:   domain.Base{ID: "notif-up-sent"},
		Name:   "up-sent",
		Type:   domain.NotificationChannelType("webhook"),
		Config: []byte(`{"url":"` + server.URL + `"}`),
	}
	require.NoError(t, channelRepo.Create(context.Background(), channel))

	resource := &domain.Resource{
		Base:     domain.Base{ID: "res-up"},
		Name:     "resource-up",
		Target:   "https://example.com/up",
		Type:     domain.ResourceHTTP,
		Status:   domain.StatusDown,
		Interval: 60,
		Timeout:  10,
		IsActive: true,
	}
	channelRepo.AssociateChannelWithResource(resource.ID, channel.ID)

	require.NoError(t, service.CreateIncident(context.Background(), resource, domain.CheckResult{Status: "down", ResponseData: "timeout"}))
	require.NoError(t, service.ResolveIncident(context.Background(), resource, domain.CheckResult{Status: "up", ResponseData: "ok"}))

	incidents, err := incidentRepo.FindByResource(context.Background(), resource.ID, 10, 0)
	require.NoError(t, err)
	require.Len(t, incidents, 1)

	notifications, err := notificationRepo.FindByIncident(context.Background(), incidents[0].ID, 20, 0)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(notifications), 2)

	up := findNotificationByType(t, notifications, domain.NotificationEventTypeUp)
	assert.Equal(t, domain.NotificationEventStatusSent, up.Status)
	assert.NotNil(t, up.ProcessedAt)
}
