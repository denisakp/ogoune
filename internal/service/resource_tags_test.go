package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/dto"
	"github.com/denisakp/ogoune/internal/repository/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceService_FindOrCreateTags_ResolvesExistingTagByName(t *testing.T) {
	resourceRepo := fake.NewResourceFake()
	incidentRepo := fake.NewIncidentFake()
	tagsRepo := fake.NewTagsFake()
	schedulerFake := fake.NewSchedulerFake()
	monitoringActivityRepo := fake.NewMonitoringActivityFake()
	channelRepo := fake.NewNotificationChannelFake()
	componentRepo := fake.NewComponentFake()
	enrichmentService := NewEnrichmentService(30 * time.Second)
	componentService := NewComponentService(componentRepo, resourceRepo, channelRepo)

	require.NoError(t, tagsRepo.Create(context.Background(), &domain.Tags{
		Base: domain.Base{ID: "tag-existing"},
		Name: "production",
	}))

	service := NewResourceService(resourceRepo, incidentRepo, tagsRepo, schedulerFake, monitoringActivityRepo, enrichmentService, componentService)

	created, err := service.CreateResource(context.Background(), &dto.CreateResourcePayload{
		Name:     "tagged-resource",
		Type:     domain.ResourceHTTP,
		Target:   "https://example.com",
		Interval: 60,
		Timeout:  5,
		Tags:     []string{"production"},
	})
	require.NoError(t, err)
	require.Len(t, created.Tags, 1)
	assert.Equal(t, "tag-existing", created.Tags[0].ID)
	assert.Equal(t, "production", created.Tags[0].Name)
}

func TestResourceService_AddTagsToResource_ResourceNotFound(t *testing.T) {
	service, _, _ := newResourceServiceForTest()

	err := service.AddTagsToResource(context.Background(), "missing-resource-id", []string{"tag-1"})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrResourceNotFound), "expected ErrResourceNotFound, got %v", err)
}
