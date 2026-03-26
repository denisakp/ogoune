package postgres

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	dbruntime "github.com/denisakp/pulseguard/internal/database"
	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func openNotificationRepo(t *testing.T) (*NotificationRepositoryImpl, context.Context, *dbruntime.Runtime) {
	t.Helper()

	cfg := dbruntime.Config{
		Driver:     dbruntime.DriverSQLite,
		SQLitePath: filepath.Join(t.TempDir(), "notification-repo.db"),
		LogLevel:   "silent",
	}

	runtime, err := dbruntime.Open(context.Background(), cfg)
	require.NoError(t, err)
	require.NotNil(t, runtime)

	repo := NewNotificationRepository(runtime.DB).(*NotificationRepositoryImpl)
	return repo, context.Background(), runtime
}

func seedIncident(t *testing.T, runtime *dbruntime.Runtime, suffix string) *domain.Incident {
	t.Helper()

	resource := &domain.Resource{
		Name:     "repo-resource-" + suffix,
		Type:     domain.ResourceHTTP,
		Target:   "https://example.com/" + suffix,
		Interval: 60,
		Timeout:  10,
		Status:   domain.StatusDown,
		IsActive: true,
	}
	require.NoError(t, runtime.DB.Create(resource).Error)

	incident := &domain.Incident{
		ResourceID: resource.ID,
		StartedAt:  time.Now().Add(-2 * time.Minute),
		Cause:      "connection_timeout",
	}
	require.NoError(t, runtime.DB.Create(incident).Error)

	return incident
}

func TestNotificationRepository_FindPendingOldestFirstAndTypeFiltered(t *testing.T) {
	repo, ctx, runtime := openNotificationRepo(t)
	incident := seedIncident(t, runtime, "pending-order")

	now := time.Now().UTC()
	events := []*domain.NotificationEvent{
		{IncidentID: incident.ID, Type: domain.NotificationEventTypeDown, Status: domain.NotificationEventStatusPending, Base: domain.Base{CreatedAt: now.Add(-3 * time.Minute), UpdatedAt: now.Add(-3 * time.Minute)}},
		{IncidentID: incident.ID, Type: domain.NotificationEventTypeReminder, Status: domain.NotificationEventStatusPending, Base: domain.Base{CreatedAt: now.Add(-2 * time.Minute), UpdatedAt: now.Add(-2 * time.Minute)}},
		{IncidentID: incident.ID, Type: domain.NotificationEventTypeUp, Status: domain.NotificationEventStatusPending, Base: domain.Base{CreatedAt: now.Add(-1 * time.Minute), UpdatedAt: now.Add(-1 * time.Minute)}},
		{IncidentID: incident.ID, Type: domain.NotificationEventTypeDown, Status: domain.NotificationEventStatusSent, Base: domain.Base{CreatedAt: now.Add(-30 * time.Second), UpdatedAt: now.Add(-30 * time.Second)}},
	}

	for _, event := range events {
		require.NoError(t, repo.Create(ctx, event))
	}

	pending, err := repo.FindPending(ctx, 10, 0)
	require.NoError(t, err)
	require.Len(t, pending, 2)

	assert.Equal(t, domain.NotificationEventTypeDown, pending[0].Type)
	assert.Equal(t, domain.NotificationEventTypeUp, pending[1].Type)
	assert.True(t, pending[0].CreatedAt.Before(pending[1].CreatedAt))
}

func TestNotificationRepository_ClaimPendingIsAtomic(t *testing.T) {
	repo, ctx, runtime := openNotificationRepo(t)
	incident := seedIncident(t, runtime, "claim")

	event := &domain.NotificationEvent{
		IncidentID: incident.ID,
		Type:       domain.NotificationEventTypeDown,
		Status:     domain.NotificationEventStatusPending,
	}
	require.NoError(t, repo.Create(ctx, event))

	claimed, err := repo.ClaimPending(ctx, event.ID, "worker-a", time.Now().UTC())
	require.NoError(t, err)
	assert.True(t, claimed)

	claimed, err = repo.ClaimPending(ctx, event.ID, "worker-b", time.Now().UTC())
	require.NoError(t, err)
	assert.False(t, claimed)

	stored, err := repo.FindByID(ctx, event.ID)
	require.NoError(t, err)
	require.NotNil(t, stored.ClaimOwner)
	assert.Equal(t, "worker-a", *stored.ClaimOwner)
}

func TestNotificationRepository_TerminalUpdatesPersistStatusAndCleanupClaim(t *testing.T) {
	repo, ctx, runtime := openNotificationRepo(t)
	incident := seedIncident(t, runtime, "terminal")

	createPending := func(eventType domain.NotificationEventType) *domain.NotificationEvent {
		event := &domain.NotificationEvent{
			IncidentID: incident.ID,
			Type:       eventType,
			Status:     domain.NotificationEventStatusPending,
		}
		require.NoError(t, repo.Create(ctx, event))
		claimed, err := repo.ClaimPending(ctx, event.ID, "worker-claim", time.Now().UTC())
		require.NoError(t, err)
		require.True(t, claimed)
		return event
	}

	sentEvent := createPending(domain.NotificationEventTypeDown)
	require.NoError(t, repo.MarkAsSent(ctx, sentEvent.ID, time.Now().UTC()))
	sentStored, err := repo.FindByID(ctx, sentEvent.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.NotificationEventStatusSent, sentStored.Status)
	assert.NotNil(t, sentStored.ProcessedAt)
	assert.Nil(t, sentStored.ClaimOwner)
	assert.Nil(t, sentStored.ClaimedAt)
	assert.Equal(t, "", sentStored.LastError)

	failedEvent := createPending(domain.NotificationEventTypeUp)
	require.NoError(t, repo.MarkAsFailed(ctx, failedEvent.ID, "dispatch failed", time.Now().UTC()))
	failedStored, err := repo.FindByID(ctx, failedEvent.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.NotificationEventStatusFailed, failedStored.Status)
	assert.NotNil(t, failedStored.ProcessedAt)
	assert.Nil(t, failedStored.ClaimOwner)
	assert.Nil(t, failedStored.ClaimedAt)
	assert.Contains(t, failedStored.LastError, "dispatch failed")

	expiredEvent := createPending(domain.NotificationEventTypeDown)
	require.NoError(t, repo.MarkAsExpired(ctx, expiredEvent.ID, "expired stale", time.Now().UTC()))
	expiredStored, err := repo.FindByID(ctx, expiredEvent.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.NotificationEventStatusExpired, expiredStored.Status)
	assert.NotNil(t, expiredStored.ProcessedAt)
	assert.Nil(t, expiredStored.ClaimOwner)
	assert.Nil(t, expiredStored.ClaimedAt)
	assert.Contains(t, expiredStored.LastError, "expired stale")
}
