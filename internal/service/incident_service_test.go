package service

import (
	"context"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildRecoveryTestService creates a minimal ResourceService suitable for recovery tests.
func buildRecoveryTestService(resources *fake.ResourceFake, incidents *fake.IncidentFake) *ResourceService {
	return &ResourceService{
		resources: resources,
		incidents: incidents,
	}
}

func TestHandleHeartbeatRecovery_UpdatesStatusToUp(t *testing.T) {
	resourceRepo := fake.NewResourceFake()
	incidentRepo := fake.NewIncidentFake()

	resource := &domain.Resource{
		Base:     domain.Base{ID: "hb-r1"},
		Name:     "Backup",
		Type:     domain.ResourceHeartbeat,
		IsActive: true,
		Status:   domain.StatusDown,
	}
	_, err := resourceRepo.Create(context.Background(), resource)
	require.NoError(t, err)

	svc := buildRecoveryTestService(resourceRepo, incidentRepo)
	err = svc.HandleHeartbeatRecovery(context.Background(), resource)
	require.NoError(t, err)

	updated, err := resourceRepo.FindByID(context.Background(), resource.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusUp, updated.Status)
}

func TestHandleHeartbeatRecovery_ResolvesActiveIncident(t *testing.T) {
	resourceRepo := fake.NewResourceFake()
	incidentRepo := fake.NewIncidentFake()

	resource := &domain.Resource{
		Base:     domain.Base{ID: "hb-r2"},
		Name:     "Cron",
		Type:     domain.ResourceHeartbeat,
		IsActive: true,
		Status:   domain.StatusDown,
	}
	_, err := resourceRepo.Create(context.Background(), resource)
	require.NoError(t, err)

	started := time.Now().Add(-5 * time.Minute)
	incident := &domain.Incident{
		ResourceID: resource.ID,
		Cause:      "missed_heartbeat",
		StartedAt:  started,
	}
	_, err = incidentRepo.Create(context.Background(), incident)
	require.NoError(t, err)

	svc := buildRecoveryTestService(resourceRepo, incidentRepo)
	err = svc.HandleHeartbeatRecovery(context.Background(), resource)
	require.NoError(t, err)

	// The incident should now be resolved (ResolvedAt set)
	incidents, err := incidentRepo.FindByResource(context.Background(), resource.ID, 10, 0)
	require.NoError(t, err)
	require.Len(t, incidents, 1)
	assert.NotNil(t, incidents[0].ResolvedAt, "incident must be resolved after heartbeat recovery")
}

func TestHandleHeartbeatRecovery_NoActiveIncident_IsNoop(t *testing.T) {
	resourceRepo := fake.NewResourceFake()
	incidentRepo := fake.NewIncidentFake()

	resource := &domain.Resource{
		Base:     domain.Base{ID: "hb-r3"},
		Name:     "NoIncident",
		Type:     domain.ResourceHeartbeat,
		IsActive: true,
		Status:   domain.StatusDown,
	}
	_, err := resourceRepo.Create(context.Background(), resource)
	require.NoError(t, err)

	svc := buildRecoveryTestService(resourceRepo, incidentRepo)
	err = svc.HandleHeartbeatRecovery(context.Background(), resource)
	// No active incident: recovery still succeeds (idempotent)
	require.NoError(t, err)

	updated, err := resourceRepo.FindByID(context.Background(), resource.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusUp, updated.Status)
}
