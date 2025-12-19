package service

import (
	"context"
	"testing"
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/dto"
	"github.com/denisakp/pulseguard/internal/repository/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceService_CreateResource(t *testing.T) {
	resourceRepo := fake.NewResourceFake()
	incidentRepo := fake.NewIncidentFake()
	tagsRepo := fake.NewTagsFake()
	schedulerFake := fake.NewSchedulerFake()
	monitoringActivityRepo := fake.NewMonitoringActivityFake()
	enrichmentService := NewEnrichmentService(30 * time.Second)
	service := NewResourceService(resourceRepo, incidentRepo, tagsRepo, schedulerFake, monitoringActivityRepo, enrichmentService)

	payload := &dto.CreateResourcePayload{
		Name:     "Test Resource",
		Type:     domain.ResourceHTTP,
		Target:   "https://example.com",
		Interval: 60,
		Timeout:  5,
		Tags:     []string{},
	}

	_, err := service.CreateResource(context.Background(), payload)
	require.NoError(t, err)

	// Verify resource was created (find by name or filter by recent creation)
	resources, err := resourceRepo.List(context.Background(), 10, 0)
	require.NoError(t, err)
	require.Len(t, resources, 1)
	assert.Equal(t, "Test Resource", resources[0].Name)

	// Verify resource was scheduled
	assert.True(t, schedulerFake.IsScheduled(resources[0].ID))
}

func TestResourceService_ListAll(t *testing.T) {
	resourceRepo := fake.NewResourceFake()
	incidentRepo := fake.NewIncidentFake()
	tagsRepo := fake.NewTagsFake()
	schedulerFake := fake.NewSchedulerFake()
	monitoringActivityRepo := fake.NewMonitoringActivityFake()
	enrichmentService := NewEnrichmentService(30 * time.Second)
	service := NewResourceService(resourceRepo, incidentRepo, tagsRepo, schedulerFake, monitoringActivityRepo, enrichmentService)

	// Create some test resources
	resource1 := &domain.Resource{
		Base: domain.Base{
			ID:        "resource-1",
			CreatedAt: time.Now(),
		},
		Name:     "Test Resource 1",
		Type:     domain.ResourceHTTP,
		Target:   "https://example1.com",
		IsActive: true,
	}

	resource2 := &domain.Resource{
		Base: domain.Base{
			ID:        "resource-2",
			CreatedAt: time.Now(),
		},
		Name:     "Test Resource 2",
		Type:     domain.ResourceTCP,
		Target:   "localhost:8080",
		IsActive: false,
	}

	// Add resources to fake repository
	_, err := resourceRepo.Create(context.Background(), resource1)
	require.NoError(t, err)
	_, err = resourceRepo.Create(context.Background(), resource2)
	require.NoError(t, err)

	// List all resources
	resources, err := service.ListAll(context.Background())
	require.NoError(t, err)

	// Verify we got both resources
	assert.Len(t, resources, 2)

	// Verify resource details
	resourceMap := make(map[string]*domain.Resource)
	for _, r := range resources {
		resourceMap[r.ID] = r
	}

	assert.Equal(t, "Test Resource 1", resourceMap["resource-1"].Name)
	assert.Equal(t, "Test Resource 2", resourceMap["resource-2"].Name)
}

func TestResourceService_ListAll_EmptyRepository(t *testing.T) {
	resourceRepo := fake.NewResourceFake()
	incidentRepo := fake.NewIncidentFake()
	tagsRepo := fake.NewTagsFake()
	schedulerFake := fake.NewSchedulerFake()
	monitoringActivityRepo := fake.NewMonitoringActivityFake()
	enrichmentService := NewEnrichmentService(30 * time.Second)
	service := NewResourceService(resourceRepo, incidentRepo, tagsRepo, schedulerFake, monitoringActivityRepo, enrichmentService)

	// List all resources from empty repository
	resources, err := service.ListAll(context.Background())
	require.NoError(t, err)

	// Verify empty list
	assert.Empty(t, resources)
}
