package database

import (
	"context"
	"testing"
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
	postgresrepo "github.com/denisakp/pulseguard/internal/repository/postgres"
	"github.com/stretchr/testify/require"
)

func TestSQLiteCRUDForResourcesAndIncidents(t *testing.T) {
	runtime := openSQLiteTestRuntime(t)
	ctx := context.Background()

	resourceRepo := postgresrepo.NewResourceRepository(runtime.DB)
	incidentRepo := postgresrepo.NewIncidentRepository(runtime.DB)

	resource, err := resourceRepo.Create(ctx, &domain.Resource{
		Name:     "Website",
		Type:     domain.ResourceHTTP,
		Target:   "https://example.com",
		Interval: 60,
		Timeout:  5,
	})
	require.NoError(t, err)
	require.NotEmpty(t, resource.ID)

	loadedResource, err := resourceRepo.FindByID(ctx, resource.ID)
	require.NoError(t, err)
	require.Equal(t, resource.Name, loadedResource.Name)

	now := time.Now().UTC().Round(time.Second)
	incident, err := incidentRepo.Create(ctx, &domain.Incident{
		ResourceID: resource.ID,
		StartedAt:  now,
		Cause:      "timeout",
		Details:    []byte("timeout reached"),
	})
	require.NoError(t, err)

	loadedIncident, err := incidentRepo.FindByID(ctx, incident.ID)
	require.NoError(t, err)
	require.Equal(t, resource.ID, loadedIncident.ResourceID)

	resolvedAt := now.Add(5 * time.Minute)
	loadedIncident.ResolvedAt = &resolvedAt
	require.NoError(t, incidentRepo.Update(ctx, loadedIncident))

	updatedIncident, err := incidentRepo.FindByID(ctx, incident.ID)
	require.NoError(t, err)
	require.NotNil(t, updatedIncident.ResolvedAt)
	ctxList := context.Background()
	incidents, err := incidentRepo.FindByResource(ctxList, resource.ID, 10, 0)
	require.NoError(t, err)
	require.Len(t, incidents, 1)
}