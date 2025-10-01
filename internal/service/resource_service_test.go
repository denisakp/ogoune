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
