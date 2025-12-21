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

// TestStatusPage_90DayIntegration tests the complete 90-day status page functionality
func TestStatusPage_90DayIntegration(t *testing.T) {
	t.Run("returns no_data for days before resource creation", func(t *testing.T) {
		// Setup
		resourceRepo := fake.NewResourceFake()
		incidentRepo := fake.NewIncidentFake()
		activityRepo := fake.NewMonitoringActivityFake()
		ctx := context.Background()
		now := time.Now()

		// Create test resource that was created 10 days ago
		resourceCreatedAt := now.AddDate(0, 0, -10)
		resource := &domain.Resource{
			Base: domain.Base{
				ID:        "test-resource-new",
				CreatedAt: resourceCreatedAt,
			},
			Name:     "New API",
			Type:     domain.ResourceHTTP,
			Target:   "https://new-api.example.com",
			Status:   domain.StatusUp,
			IsActive: true,
			Interval: 300,
		}
		_, err := resourceRepo.Create(ctx, resource)
		require.NoError(t, err)

		// Create monitoring activities only for the last 10 days
		for day := 0; day < 10; day++ {
			for check := 0; check < 5; check++ {
				activityTime := now.AddDate(0, 0, -day).Add(time.Duration(-check) * time.Hour)
				activity := &domain.MonitoringActivity{
					Base: domain.Base{
						ID:        "activity-new-" + resource.ID + "-d" + string(rune(day+48)) + "-c" + string(rune(check+48)),
						CreatedAt: activityTime,
					},
					ResourceID:   resource.ID,
					Success:      true,
					ResponseTime: 150,
					Message:      "OK",
				}
				require.NoError(t, activityRepo.Create(ctx, activity))
			}
		}

		// Create service and get data
		service := NewStatusPageService(resourceRepo, incidentRepo, activityRepo, nil)
		data, err := service.GetData(ctx)

		// Assertions
		require.NoError(t, err)
		require.NotNil(t, data)
		require.Len(t, data.Resources, 1)

		resourceInfo := data.Resources[0]
		require.Len(t, resourceInfo.DailyStatusLast90Days, 90, "Must have exactly 90 days")

		// First 80 days (90 - 10) should be "no_data"
		for i := 0; i < 80; i++ {
			assert.Equal(t, "no_data", resourceInfo.DailyStatusLast90Days[i],
				"Day %d should be 'no_data' (resource didn't exist yet)", i)
		}

		// Last 10 days should be actual status (not "no_data")
		for i := 80; i < 90; i++ {
			assert.NotEqual(t, "no_data", resourceInfo.DailyStatusLast90Days[i],
				"Day %d should have real status (resource existed)", i)
			assert.Contains(t, []string{"up", "down", "degraded"}, resourceInfo.DailyStatusLast90Days[i],
				"Day %d should have valid status", i)
		}

		// Uptime should be 100% for the days it existed
		assert.Equal(t, 100.0, resourceInfo.UptimePercentageLast90Days)
	})

	t.Run("returns correct structure with 90-day data", func(t *testing.T) {
		// Setup
		resourceRepo := fake.NewResourceFake()
		incidentRepo := fake.NewIncidentFake()
		activityRepo := fake.NewMonitoringActivityFake()
		ctx := context.Background()
		now := time.Now()

		// Create test resource (created 100 days ago - older than 90-day window)
		resource := &domain.Resource{
			Base: domain.Base{
				ID:        "test-resource-1",
				CreatedAt: now.AddDate(0, 0, -100),
			},
			Name:     "Test API",
			Type:     domain.ResourceHTTP,
			Target:   "https://api.example.com",
			Status:   domain.StatusUp,
			IsActive: true,
			Interval: 300,
		}
		_, err := resourceRepo.Create(ctx, resource)
		require.NoError(t, err)

		// Create monitoring activities for the last 90 days
		// Creating a reasonable amount of data
		for day := 0; day < 90; day++ {
			for check := 0; check < 10; check++ { // 10 checks per day
				activityTime := now.AddDate(0, 0, -day).Add(time.Duration(-check) * time.Hour)
				activity := &domain.MonitoringActivity{
					Base: domain.Base{
						ID:        "activity-" + resource.ID + "-d" + string(rune(day+48)) + "-c" + string(rune(check+48)),
						CreatedAt: activityTime,
					},
					ResourceID:   resource.ID,
					Success:      true, // All successful for this test
					ResponseTime: 150,
					Message:      "OK",
				}
				require.NoError(t, activityRepo.Create(ctx, activity))
			}
		}

		// Create service and get data
		service := NewStatusPageService(resourceRepo, incidentRepo, activityRepo, nil)
		data, err := service.GetData(ctx)

		// Assertions
		require.NoError(t, err)
		require.NotNil(t, data)

		// Check global status
		assert.Equal(t, "all_systems_operational", data.GlobalStatus)
		assert.WithinDuration(t, time.Now(), data.GeneratedAt, 5*time.Second)

		// Check resources
		require.Len(t, data.Resources, 1)
		resourceInfo := data.Resources[0]

		assert.Equal(t, "test-resource-1", resourceInfo.ID)
		assert.Equal(t, "Test API", resourceInfo.Name)
		assert.Equal(t, "up", resourceInfo.CurrentStatus)

		// Check uptime percentage (should be 100%)
		assert.Equal(t, 100.0, resourceInfo.UptimePercentageLast90Days)

		// Check daily status array - CRITICAL: Must have exactly 90 entries
		require.Len(t, resourceInfo.DailyStatusLast90Days, 90, "Must have exactly 90 days")

		// Verify all days are valid status values
		for i, status := range resourceInfo.DailyStatusLast90Days {
			assert.Contains(t, []string{"up", "down", "degraded", "no_data"}, status,
				"Day %d has invalid status: %s", i, status)
		}
	})

	t.Run("global status reflects degraded resource", func(t *testing.T) {
		resourceRepo := fake.NewResourceFake()
		incidentRepo := fake.NewIncidentFake()
		activityRepo := fake.NewMonitoringActivityFake()
		ctx := context.Background()
		now := time.Now()

		// Create degraded resource (created 100 days ago - older than 90-day window)
		resource := &domain.Resource{
			Base: domain.Base{
				ID:        "degraded-resource",
				CreatedAt: now.AddDate(0, 0, -100),
			},
			Name:     "Degraded Service",
			Type:     domain.ResourceHTTP,
			Target:   "https://degraded.example.com",
			Status:   domain.StatusWarn, // This should map to "degraded"
			IsActive: true,
			Interval: 300,
		}
		_, err := resourceRepo.Create(ctx, resource)
		require.NoError(t, err)

		// Add some activities
		for i := 0; i < 10; i++ {
			activity := &domain.MonitoringActivity{
				Base: domain.Base{
					ID:        "activity-degraded-" + string(rune(i+48)),
					CreatedAt: time.Now().Add(time.Duration(-i) * time.Hour),
				},
				ResourceID:   resource.ID,
				Success:      true,
				ResponseTime: 150,
			}
			require.NoError(t, activityRepo.Create(ctx, activity))
		}

		service := NewStatusPageService(resourceRepo, incidentRepo, activityRepo, nil)
		data, err := service.GetData(ctx)

		require.NoError(t, err)
		assert.Equal(t, "some_systems_down", data.GlobalStatus,
			"Global status should be 'some_systems_down' when any resource is not 'up'")
		assert.Equal(t, "degraded", data.Resources[0].CurrentStatus)
		assert.Len(t, data.Resources[0].DailyStatusLast90Days, 90)

		// All days should be actual status values (resource is old)
		for _, status := range data.Resources[0].DailyStatusLast90Days {
			assert.NotEqual(t, "no_data", status, "Old resource should not have 'no_data' entries")
		}
	})

	t.Run("global status reflects down resource", func(t *testing.T) {
		resourceRepo := fake.NewResourceFake()
		incidentRepo := fake.NewIncidentFake()
		activityRepo := fake.NewMonitoringActivityFake()
		ctx := context.Background()
		now := time.Now()

		// Create down resource (created 100 days ago - older than 90-day window)
		resource := &domain.Resource{
			Base: domain.Base{
				ID:        "down-resource",
				CreatedAt: now.AddDate(0, 0, -100),
			},
			Name:     "Down Service",
			Type:     domain.ResourceHTTP,
			Target:   "https://down.example.com",
			Status:   domain.StatusDown,
			IsActive: true,
			Interval: 300,
		}
		_, err := resourceRepo.Create(ctx, resource)
		require.NoError(t, err)

		// Add some failed activities
		for i := 0; i < 10; i++ {
			activity := &domain.MonitoringActivity{
				Base: domain.Base{
					ID:        "activity-down-" + string(rune(i+48)),
					CreatedAt: time.Now().Add(time.Duration(-i) * time.Hour),
				},
				ResourceID:   resource.ID,
				Success:      false,
				ResponseTime: 0,
			}
			require.NoError(t, activityRepo.Create(ctx, activity))
		}

		service := NewStatusPageService(resourceRepo, incidentRepo, activityRepo, nil)
		data, err := service.GetData(ctx)

		require.NoError(t, err)
		assert.Equal(t, "some_systems_down", data.GlobalStatus)
		assert.Equal(t, "down", data.Resources[0].CurrentStatus)
		assert.Len(t, data.Resources[0].DailyStatusLast90Days, 90)

		// All days should be actual status values (resource is old)
		for _, status := range data.Resources[0].DailyStatusLast90Days {
			assert.NotEqual(t, "no_data", status, "Old resource should not have 'no_data' entries")
		}
	})

	t.Run("multiple resources aggregate correctly", func(t *testing.T) {
		resourceRepo := fake.NewResourceFake()
		incidentRepo := fake.NewIncidentFake()
		activityRepo := fake.NewMonitoringActivityFake()
		ctx := context.Background()
		now := time.Now()

		// Create 3 resources with different statuses (all created 100 days ago)
		resources := []*domain.Resource{
			{
				Base:     domain.Base{ID: "resource-1", CreatedAt: now.AddDate(0, 0, -100)},
				Name:     "Service A",
				Type:     domain.ResourceHTTP,
				Target:   "https://a.example.com",
				Status:   domain.StatusUp,
				IsActive: true,
				Interval: 300,
			},
			{
				Base:     domain.Base{ID: "resource-2", CreatedAt: now.AddDate(0, 0, -100)},
				Name:     "Service B",
				Type:     domain.ResourceHTTP,
				Target:   "https://b.example.com",
				Status:   domain.StatusDown,
				IsActive: true,
				Interval: 300,
			},
			{
				Base:     domain.Base{ID: "resource-3", CreatedAt: now.AddDate(0, 0, -100)},
				Name:     "Service C",
				Type:     domain.ResourceHTTP,
				Target:   "https://c.example.com",
				Status:   domain.StatusUp,
				IsActive: true,
				Interval: 300,
			},
		}

		for _, res := range resources {
			_, err := resourceRepo.Create(ctx, res)
			require.NoError(t, err)

			// Add minimal activities for each
			for i := 0; i < 5; i++ {
				activity := &domain.MonitoringActivity{
					Base: domain.Base{
						ID:        res.ID + "-activity-" + string(rune(i+48)),
						CreatedAt: now.Add(time.Duration(-i) * time.Hour),
					},
					ResourceID:   res.ID,
					Success:      res.Status == domain.StatusUp,
					ResponseTime: 150,
				}
				require.NoError(t, activityRepo.Create(ctx, activity))
			}
		}

		service := NewStatusPageService(resourceRepo, incidentRepo, activityRepo, nil)
		data, err := service.GetData(ctx)

		require.NoError(t, err)
		assert.Equal(t, "some_systems_down", data.GlobalStatus,
			"Global status should be 'some_systems_down' because Service B is down")
		assert.Len(t, data.Resources, 3)

		// Check each resource has 90 daily statuses
		for _, res := range data.Resources {
			assert.Len(t, res.DailyStatusLast90Days, 90,
				"Resource %s must have exactly 90 daily status entries", res.Name)
		}
	})

	t.Run("handles empty resources gracefully", func(t *testing.T) {
		resourceRepo := fake.NewResourceFake()
		incidentRepo := fake.NewIncidentFake()
		activityRepo := fake.NewMonitoringActivityFake()

		service := NewStatusPageService(resourceRepo, incidentRepo, activityRepo, nil)
		data, err := service.GetData(context.Background())

		require.NoError(t, err)
		assert.NotNil(t, data)
		assert.Empty(t, data.Resources)
		assert.Equal(t, "all_systems_operational", data.GlobalStatus)
	})

	t.Run("handles resource created exactly 90 days ago", func(t *testing.T) {
		resourceRepo := fake.NewResourceFake()
		incidentRepo := fake.NewIncidentFake()
		activityRepo := fake.NewMonitoringActivityFake()
		ctx := context.Background()
		now := time.Now()

		// Create resource exactly 90 days ago
		resourceCreatedAt := now.AddDate(0, 0, -90)
		resource := &domain.Resource{
			Base: domain.Base{
				ID:        "resource-90-days",
				CreatedAt: resourceCreatedAt,
			},
			Name:     "90-Day Old Service",
			Type:     domain.ResourceHTTP,
			Target:   "https://old.example.com",
			Status:   domain.StatusUp,
			IsActive: true,
			Interval: 300,
		}
		_, err := resourceRepo.Create(ctx, resource)
		require.NoError(t, err)

		// Add some activities
		for i := 0; i < 10; i++ {
			activity := &domain.MonitoringActivity{
				Base: domain.Base{
					ID:        "activity-90-" + string(rune(i+48)),
					CreatedAt: now.Add(time.Duration(-i) * time.Hour),
				},
				ResourceID:   resource.ID,
				Success:      true,
				ResponseTime: 150,
			}
			require.NoError(t, activityRepo.Create(ctx, activity))
		}

		service := NewStatusPageService(resourceRepo, incidentRepo, activityRepo, nil)
		data, err := service.GetData(ctx)

		require.NoError(t, err)
		require.Len(t, data.Resources, 1)
		require.Len(t, data.Resources[0].DailyStatusLast90Days, 90)

		// Should have no "no_data" entries (resource existed for entire 90-day window)
		for i, status := range data.Resources[0].DailyStatusLast90Days {
			assert.NotEqual(t, "no_data", status,
				"Day %d should not be 'no_data' for 90-day-old resource", i)
		}
	})

	t.Run("handles resource created more than 90 days ago", func(t *testing.T) {
		resourceRepo := fake.NewResourceFake()
		incidentRepo := fake.NewIncidentFake()
		activityRepo := fake.NewMonitoringActivityFake()
		ctx := context.Background()
		now := time.Now()

		// Create resource 120 days ago (older than window)
		resourceCreatedAt := now.AddDate(0, 0, -120)
		resource := &domain.Resource{
			Base: domain.Base{
				ID:        "resource-120-days",
				CreatedAt: resourceCreatedAt,
			},
			Name:     "Old Service",
			Type:     domain.ResourceHTTP,
			Target:   "https://veryold.example.com",
			Status:   domain.StatusUp,
			IsActive: true,
			Interval: 300,
		}
		_, err := resourceRepo.Create(ctx, resource)
		require.NoError(t, err)

		// Add some activities
		for i := 0; i < 10; i++ {
			activity := &domain.MonitoringActivity{
				Base: domain.Base{
					ID:        "activity-120-" + string(rune(i+48)),
					CreatedAt: now.Add(time.Duration(-i) * time.Hour),
				},
				ResourceID:   resource.ID,
				Success:      true,
				ResponseTime: 150,
			}
			require.NoError(t, activityRepo.Create(ctx, activity))
		}

		service := NewStatusPageService(resourceRepo, incidentRepo, activityRepo, nil)
		data, err := service.GetData(ctx)

		require.NoError(t, err)
		require.Len(t, data.Resources, 1)
		require.Len(t, data.Resources[0].DailyStatusLast90Days, 90)

		// Should have no "no_data" entries (resource is older than 90-day window)
		for i, status := range data.Resources[0].DailyStatusLast90Days {
			assert.NotEqual(t, "no_data", status,
				"Day %d should not be 'no_data' for old resource", i)
		}
	})
}

func TestStatusPageService_MapResourceStatus(t *testing.T) {
	service := NewStatusPageService(nil, nil, nil, nil)

	tests := []struct {
		name           string
		resourceStatus domain.ResourceStatus
		expectedStatus string
	}{
		{"StatusUp maps to up", domain.StatusUp, "up"},
		{"StatusDown maps to down", domain.StatusDown, "down"},
		{"StatusError maps to down", domain.StatusError, "down"},
		{"StatusWarn maps to degraded", domain.StatusWarn, "degraded"},
		{"StatusPending maps to degraded", domain.StatusPending, "degraded"},
		{"StatusUnknown maps to degraded", domain.StatusUnknown, "degraded"},
		{"StatusPaused maps to up", domain.StatusPaused, "up"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.mapResourceStatus(tt.resourceStatus)
			assert.Equal(t, tt.expectedStatus, result)
		})
	}
}

func TestStatusPageService_CalculateDayStatus(t *testing.T) {
	service := NewStatusPageService(nil, nil, nil, nil)
	now := time.Now()
	dayStart := now.Truncate(24 * time.Hour)
	dayEnd := dayStart.Add(24 * time.Hour)

	t.Run("returns up for high success rate", func(t *testing.T) {
		// 99% success rate
		activities := make([]*domain.MonitoringActivity, 0)
		for i := 0; i < 99; i++ {
			activities = append(activities, &domain.MonitoringActivity{
				Base:    domain.Base{CreatedAt: dayStart.Add(time.Duration(i) * time.Minute)},
				Success: true,
			})
		}
		activities = append(activities, &domain.MonitoringActivity{
			Base:    domain.Base{CreatedAt: dayStart.Add(99 * time.Minute)},
			Success: false,
		})

		status := service.calculateDayStatus(dayStart, dayEnd, activities, nil)
		assert.Equal(t, "up", status)
	})

	t.Run("returns degraded for moderate success rate", func(t *testing.T) {
		// 85% success rate
		activities := make([]*domain.MonitoringActivity, 0)
		for i := 0; i < 85; i++ {
			activities = append(activities, &domain.MonitoringActivity{
				Base:    domain.Base{CreatedAt: dayStart.Add(time.Duration(i) * time.Minute)},
				Success: true,
			})
		}
		for i := 0; i < 15; i++ {
			activities = append(activities, &domain.MonitoringActivity{
				Base:    domain.Base{CreatedAt: dayStart.Add(time.Duration(i+85) * time.Minute)},
				Success: false,
			})
		}

		status := service.calculateDayStatus(dayStart, dayEnd, activities, nil)
		assert.Equal(t, "degraded", status)
	})

	t.Run("returns down for low success rate", func(t *testing.T) {
		// 30% success rate
		activities := make([]*domain.MonitoringActivity, 0)
		for i := 0; i < 30; i++ {
			activities = append(activities, &domain.MonitoringActivity{
				Base:    domain.Base{CreatedAt: dayStart.Add(time.Duration(i) * time.Minute)},
				Success: true,
			})
		}
		for i := 0; i < 70; i++ {
			activities = append(activities, &domain.MonitoringActivity{
				Base:    domain.Base{CreatedAt: dayStart.Add(time.Duration(i+30) * time.Minute)},
				Success: false,
			})
		}

		status := service.calculateDayStatus(dayStart, dayEnd, activities, nil)
		assert.Equal(t, "down", status)
	})

	t.Run("returns up when no data available for existing resource", func(t *testing.T) {
		// Note: "no_data" is handled at a higher level based on resource creation date
		// This test verifies that when a resource exists but has no monitoring data for a day,
		// we assume "up" rather than returning "no_data"
		status := service.calculateDayStatus(dayStart, dayEnd, nil, nil)
		assert.Equal(t, "up", status)
	})

	t.Run("returns down for major incident covering >50% of day", func(t *testing.T) {
		// Incident covering 18 hours (75% of day)
		incidentStart := dayStart.Add(2 * time.Hour)
		incidentEnd := dayStart.Add(20 * time.Hour)
		incidents := []*domain.Incident{
			{
				StartedAt:  incidentStart,
				ResolvedAt: &incidentEnd,
			},
		}

		status := service.calculateDayStatus(dayStart, dayEnd, nil, incidents)
		assert.Equal(t, "down", status)
	})

	t.Run("returns degraded for minor incident", func(t *testing.T) {
		// Incident covering 2 hours (8% of day)
		incidentStart := dayStart.Add(2 * time.Hour)
		incidentEnd := dayStart.Add(4 * time.Hour)
		incidents := []*domain.Incident{
			{
				StartedAt:  incidentStart,
				ResolvedAt: &incidentEnd,
			},
		}

		// Good activities otherwise
		activities := make([]*domain.MonitoringActivity, 0)
		for i := 0; i < 100; i++ {
			activities = append(activities, &domain.MonitoringActivity{
				Base:    domain.Base{CreatedAt: dayStart.Add(time.Duration(i) * time.Minute)},
				Success: true,
			})
		}

		status := service.calculateDayStatus(dayStart, dayEnd, activities, incidents)
		assert.Equal(t, "degraded", status)
	})
}
