package monitoring

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// chanEmitter records emitted notifications on a buffered channel.
type chanEmitter struct {
	ch  chan domain.EmittedNotification
	err error
}

func (e *chanEmitter) Emit(_ context.Context, n domain.EmittedNotification) error {
	e.ch <- n
	return e.err
}

func downResource() *domain.Resource {
	return &domain.Resource{
		Base:     domain.Base{ID: "res-emit"},
		Name:     "Example API",
		Target:   "https://example.com",
		Type:     domain.ResourceHTTP,
		Timeout:  30,
		IsActive: true,
		Status:   domain.StatusUp,
	}
}

func awaitEmit(t *testing.T, ch chan domain.EmittedNotification) domain.EmittedNotification {
	t.Helper()
	select {
	case n := <-ch:
		return n
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for emitted notification")
		return domain.EmittedNotification{}
	}
}

func TestIncident_EmitsFeedNotification_OnDetectAndResolve(t *testing.T) {
	service, _, _, _, _, asynqClient := setupTestService()
	defer asynqClient.Close()
	emitter := &chanEmitter{ch: make(chan domain.EmittedNotification, 4)}
	service.SetNotificationEmitter(emitter)
	ctx := context.Background()
	resource := downResource()

	require.NoError(t, service.CreateIncident(ctx, resource, domain.CheckResult{Status: "down", ResponseData: "connection timeout"}))
	detected := awaitEmit(t, emitter.ch)
	assert.Equal(t, domain.NotificationCategoryIncident, detected.Category)
	assert.Equal(t, domain.NotificationSeverityError, detected.Severity)
	assert.Contains(t, detected.Title, "Example API")
	require.NotNil(t, detected.DeepLink)
	assert.Contains(t, *detected.DeepLink, "/incidents/")

	require.NoError(t, service.ResolveIncident(ctx, resource, domain.CheckResult{Status: "up"}))
	resolved := awaitEmit(t, emitter.ch)
	assert.Equal(t, domain.NotificationSeveritySuccess, resolved.Severity)
}

func TestIncident_EmitFailure_DoesNotBreakIncidentHandling(t *testing.T) {
	service, incidentRepo, _, _, _, asynqClient := setupTestService()
	defer asynqClient.Close()
	emitter := &chanEmitter{ch: make(chan domain.EmittedNotification, 4), err: errors.New("feed down")}
	service.SetNotificationEmitter(emitter)
	ctx := context.Background()
	resource := downResource()

	// Even though the emitter returns an error, incident creation MUST succeed.
	require.NoError(t, service.CreateIncident(ctx, resource, domain.CheckResult{Status: "down", ResponseData: "connection timeout"}))
	_ = awaitEmit(t, emitter.ch) // drain the emitted notification

	incidents, err := incidentRepo.FindByResource(ctx, resource.ID, 10, 0)
	require.NoError(t, err)
	require.Len(t, incidents, 1, "incident must be persisted despite emitter error")
}
