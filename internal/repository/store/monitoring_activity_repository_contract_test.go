package store

import (
	"context"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMonitoringActivityRepository_Create(t *testing.T) {
	db := internaltest.GetTestDB(t)
	repo := NewMonitoringActivityRepository(db)
	resourceRepo := NewResourceRepository(db)

	// Create a test resource first
	resource := &domain.Resource{
		Name:     "Test Resource",
		Type:     domain.ResourceHTTP,
		Target:   "https://example.com",
		Interval: 60,
		Timeout:  30,
		IsActive: true,
	}
	_, err := resourceRepo.Create(context.Background(), resource)
	require.NoError(t, err)

	// Create a monitoring activity
	activity := &domain.MonitoringActivity{
		ResourceID:   resource.ID,
		Message:      "Check completed successfully",
		Success:      true,
		ResponseTime: 150,
		ResponseData: []byte("OK"),
	}

	err = repo.Create(context.Background(), activity)
	require.NoError(t, err)
	assert.NotEmpty(t, activity.ID)
	assert.NotZero(t, activity.CreatedAt)
}

func TestMonitoringActivityRepository_List(t *testing.T) {
	db := internaltest.GetTestDB(t)
	repo := NewMonitoringActivityRepository(db)
	resourceRepo := NewResourceRepository(db)

	// Create a test resource
	resource := &domain.Resource{
		Name:     "Test Resource",
		Type:     domain.ResourceHTTP,
		Target:   "https://example.com",
		Interval: 60,
		Timeout:  30,
		IsActive: true,
	}
	_, err := resourceRepo.Create(context.Background(), resource)
	require.NoError(t, err)

	// Create multiple activities
	activity1 := &domain.MonitoringActivity{
		ResourceID:   resource.ID,
		Message:      "First check",
		Success:      true,
		ResponseTime: 100,
	}
	activity2 := &domain.MonitoringActivity{
		ResourceID:   resource.ID,
		Message:      "Second check",
		Success:      false,
		ResponseTime: 500,
	}

	err = repo.Create(context.Background(), activity1)
	require.NoError(t, err)

	// Small delay to ensure different timestamps
	time.Sleep(10 * time.Millisecond)

	err = repo.Create(context.Background(), activity2)
	require.NoError(t, err)

	// List activities
	activities, err := repo.List(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(activities), 2)

	// Verify most recent is first (DESC order)
	if len(activities) >= 2 {
		assert.True(t, activities[0].CreatedAt.After(activities[1].CreatedAt) ||
			activities[0].CreatedAt.Equal(activities[1].CreatedAt))
	}
}

func TestMonitoringActivityRepository_FindByResourceID(t *testing.T) {
	db := internaltest.GetTestDB(t)
	repo := NewMonitoringActivityRepository(db)
	resourceRepo := NewResourceRepository(db)

	// Create two test resources
	resource1 := &domain.Resource{
		Name:     "Test Resource 1",
		Type:     domain.ResourceHTTP,
		Target:   "https://example1.com",
		Interval: 60,
		Timeout:  30,
		IsActive: true,
	}
	resource2 := &domain.Resource{
		Name:     "Test Resource 2",
		Type:     domain.ResourceHTTP,
		Target:   "https://example2.com",
		Interval: 60,
		Timeout:  30,
		IsActive: true,
	}

	_, err := resourceRepo.Create(context.Background(), resource1)
	require.NoError(t, err)
	_, err = resourceRepo.Create(context.Background(), resource2)
	require.NoError(t, err)

	// Create activities for resource1
	for i := 0; i < 3; i++ {
		activity := &domain.MonitoringActivity{
			ResourceID:   resource1.ID,
			Message:      "Check for resource 1",
			Success:      true,
			ResponseTime: 100 + i,
		}
		err = repo.Create(context.Background(), activity)
		require.NoError(t, err)
		time.Sleep(5 * time.Millisecond)
	}

	// Create activity for resource2
	activity := &domain.MonitoringActivity{
		ResourceID:   resource2.ID,
		Message:      "Check for resource 2",
		Success:      true,
		ResponseTime: 200,
	}
	err = repo.Create(context.Background(), activity)
	require.NoError(t, err)

	// Find activities for resource1
	activities, err := repo.FindByResourceID(context.Background(), resource1.ID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, activities, 3)

	// Verify all activities belong to resource1
	for _, act := range activities {
		assert.Equal(t, resource1.ID, act.ResourceID)
	}

	// Find activities for resource2
	activities, err = repo.FindByResourceID(context.Background(), resource2.ID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, activities, 1)
	assert.Equal(t, resource2.ID, activities[0].ResourceID)
}

func TestMonitoringActivityRepository_Pagination(t *testing.T) {
	db := internaltest.GetTestDB(t)
	repo := NewMonitoringActivityRepository(db)
	resourceRepo := NewResourceRepository(db)

	// Create a test resource
	resource := &domain.Resource{
		Name:     "Test Resource",
		Type:     domain.ResourceHTTP,
		Target:   "https://example.com",
		Interval: 60,
		Timeout:  30,
		IsActive: true,
	}
	_, err := resourceRepo.Create(context.Background(), resource)
	require.NoError(t, err)

	// Create 15 activities
	for i := 0; i < 15; i++ {
		activity := &domain.MonitoringActivity{
			ResourceID:   resource.ID,
			Message:      "Check",
			Success:      true,
			ResponseTime: i,
		}
		err = repo.Create(context.Background(), activity)
		require.NoError(t, err)
		time.Sleep(2 * time.Millisecond)
	}

	// Get first page (limit 10)
	page1, err := repo.List(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.Len(t, page1, 10)

	// Get second page (limit 10, offset 10)
	page2, err := repo.List(context.Background(), 10, 10)
	require.NoError(t, err)
	assert.Len(t, page2, 5)

	// Verify no overlap
	assert.NotEqual(t, page1[0].ID, page2[0].ID)
}
