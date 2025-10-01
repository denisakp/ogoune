package service

import (
	"context"
	"testing"
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/repository/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceService_CreateResource(t *testing.T) {
	resourceRepo := fake.NewResourceFake()
	incidentRepo := fake.NewIncidentFake()
	schedulerFake := fake.NewSchedulerFake()
	service := NewResourceService(resourceRepo, incidentRepo, schedulerFake)

	resource := &domain.Resource{
		Base: domain.Base{
			ID:        "test-resource",
			CreatedAt: time.Now(),
		},
		Name:     "Test Resource",
		Type:     domain.ResourceHTTP,
		Target:   "https://example.com",
		IsActive: true,
	}

	err := service.CreateResource(context.Background(), resource)
	require.NoError(t, err)

	// Verify resource was created
	found, err := resourceRepo.FindByID(context.Background(), "test-resource")
	require.NoError(t, err)
	assert.Equal(t, "Test Resource", found.Name)

	// Verify resource was scheduled
	assert.True(t, schedulerFake.IsScheduled("test-resource"))
}

func TestResourceService_ListAll(t *testing.T) {
	resourceRepo := fake.NewResourceFake()
	incidentRepo := fake.NewIncidentFake()
	schedulerFake := fake.NewSchedulerFake()
	service := NewResourceService(resourceRepo, incidentRepo, schedulerFake)

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
	err := resourceRepo.Create(context.Background(), resource1)
	require.NoError(t, err)
	err = resourceRepo.Create(context.Background(), resource2)
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
	schedulerFake := fake.NewSchedulerFake()
	service := NewResourceService(resourceRepo, incidentRepo, schedulerFake)

	// List all resources from empty repository
	resources, err := service.ListAll(context.Background())
	require.NoError(t, err)

	// Verify empty list
	assert.Empty(t, resources)
}
