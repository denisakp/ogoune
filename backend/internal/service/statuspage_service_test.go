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

func TestStatusPageService_GetData(t *testing.T) {
	t.Run("returns data with active resources and 90-day stats", func(t *testing.T) {
		// Setup fake repositories
		resourceRepo := fake.NewResourceFake()
		incidentRepo := fake.NewIncidentFake()
		activityRepo := fake.NewMonitoringActivityFake()

		// Create test data
		now := time.Now()
		resource := &domain.Resource{
			Base:        domain.Base{ID: "resource-1"},
			Name:        "Test Website",
			Type:        domain.ResourceHTTP,
			Target:      "https://example.com",
			Status:      domain.StatusUp,
			IsActive:    true,
			Interval:    60,
			LastChecked: &now,
		}

		// Store resource
		ctx := context.Background()
		_, err := resourceRepo.Create(ctx, resource)
		require.NoError(t, err)

		// Create monitoring activities (90 days of data)
		for i := 0; i < 200; i++ {
			activity := &domain.MonitoringActivity{
				Base:         domain.Base{ID: "activity-" + string(rune(i))},
				ResourceID:   resource.ID,
				Success:      i%10 != 0, // 90% success rate
				ResponseTime: 100 + i,
			}
			activity.CreatedAt = now.AddDate(0, 0, -i%90)
			require.NoError(t, activityRepo.Create(ctx, activity))
		}

		// Create service
		service := NewStatusPageService(resourceRepo, incidentRepo, activityRepo, nil)

		// Execute
		data, err := service.GetData(ctx)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, data)
		assert.Equal(t, "all_systems_operational", data.GlobalStatus)
		assert.Len(t, data.Resources, 1)

		// Check resource info
		resourceInfo := data.Resources[0]
		assert.Equal(t, "resource-1", resourceInfo.ID)
		assert.Equal(t, "Test Website", resourceInfo.Name)
		assert.Equal(t, "up", resourceInfo.CurrentStatus)
		assert.InDelta(t, 90.0, resourceInfo.UptimePercentageLast90Days, 5.0) // ~90% uptime
		assert.Len(t, resourceInfo.DailyStatusLast90Days, 90)                 // Must have 90 days
	})

	t.Run("sets global status to some_systems_down when resource is down", func(t *testing.T) {
		// Setup fake repositories
		resourceRepo := fake.NewResourceFake()
		incidentRepo := fake.NewIncidentFake()
		activityRepo := fake.NewMonitoringActivityFake()

		ctx := context.Background()
		now := time.Now()

		// Create down resource
		resource := &domain.Resource{
			Base:        domain.Base{ID: "resource-1"},
			Name:        "Down Website",
			Type:        domain.ResourceHTTP,
			Target:      "https://example.com",
			Status:      domain.StatusDown,
			IsActive:    true,
			Interval:    60,
			LastChecked: &now,
		}
		_, err := resourceRepo.Create(ctx, resource)
		require.NoError(t, err)

		// Create some monitoring activities
		for i := 0; i < 10; i++ {
			activity := &domain.MonitoringActivity{
				Base:         domain.Base{ID: "activity-" + string(rune(i))},
				ResourceID:   resource.ID,
				Success:      false,
				ResponseTime: 0,
			}
			activity.CreatedAt = now.AddDate(0, 0, -i)
			require.NoError(t, activityRepo.Create(ctx, activity))
		}

		service := NewStatusPageService(resourceRepo, incidentRepo, activityRepo, nil)

		data, err := service.GetData(ctx)

		require.NoError(t, err)
		assert.Equal(t, "some_systems_down", data.GlobalStatus)
		assert.Len(t, data.Resources, 1)
		assert.Equal(t, "down", data.Resources[0].CurrentStatus)
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
}

func TestStatusPageService_buildResourceStatusInfo(t *testing.T) {
	t.Run("calculates 90-day uptime correctly", func(t *testing.T) {
		resourceRepo := fake.NewResourceFake()
		incidentRepo := fake.NewIncidentFake()
		activityRepo := fake.NewMonitoringActivityFake()

		ctx := context.Background()
		now := time.Now()

		resource := &domain.Resource{
			Base:   domain.Base{ID: "resource-1"},
			Name:   "Test",
			Type:   domain.ResourceHTTP,
			Status: domain.StatusUp,
		}

		// 80 successful checks, 20 failures = 80% uptime
		for i := 0; i < 80; i++ {
			activity := &domain.MonitoringActivity{
				Base:         domain.Base{ID: "success-" + string(rune(i))},
				ResourceID:   resource.ID,
				Success:      true,
				ResponseTime: 100,
			}
			activity.CreatedAt = now.AddDate(0, 0, -i%90)
			require.NoError(t, activityRepo.Create(ctx, activity))
		}

		for i := 0; i < 20; i++ {
			activity := &domain.MonitoringActivity{
				Base:         domain.Base{ID: "failure-" + string(rune(i))},
				ResourceID:   resource.ID,
				Success:      false,
				ResponseTime: 0,
			}
			activity.CreatedAt = now.AddDate(0, 0, -(i+10)%90)
			require.NoError(t, activityRepo.Create(ctx, activity))
		}

		service := NewStatusPageService(resourceRepo, incidentRepo, activityRepo, nil)

		info, err := service.buildResourceStatusInfo(ctx, resource)

		require.NoError(t, err)
		assert.Equal(t, 80.0, info.UptimePercentageLast90Days)
		assert.Len(t, info.DailyStatusLast90Days, 90)
	})

	t.Run("handles resource with no activities", func(t *testing.T) {
		resourceRepo := fake.NewResourceFake()
		incidentRepo := fake.NewIncidentFake()
		activityRepo := fake.NewMonitoringActivityFake()

		resource := &domain.Resource{
			Base:   domain.Base{ID: "resource-1"},
			Name:   "Test",
			Type:   domain.ResourceHTTP,
			Status: domain.StatusUp,
		}

		service := NewStatusPageService(resourceRepo, incidentRepo, activityRepo, nil)

		info, err := service.buildResourceStatusInfo(context.Background(), resource)

		require.NoError(t, err)
		assert.Equal(t, 100.0, info.UptimePercentageLast90Days) // Default to 100% if no data
		assert.Len(t, info.DailyStatusLast90Days, 90)
	})
}

func TestStatusPageService_mapResourceStatus(t *testing.T) {
	service := NewStatusPageService(nil, nil, nil, nil)

	tests := []struct {
		name           string
		resourceStatus domain.ResourceStatus
		want           string
	}{
		{
			name:           "up status",
			resourceStatus: domain.StatusUp,
			want:           "up",
		},
		{
			name:           "down status",
			resourceStatus: domain.StatusDown,
			want:           "down",
		},
		{
			name:           "error status maps to down",
			resourceStatus: domain.StatusError,
			want:           "down",
		},
		{
			name:           "warn status maps to degraded",
			resourceStatus: domain.StatusWarn,
			want:           "degraded",
		},
		{
			name:           "pending status maps to degraded",
			resourceStatus: domain.StatusPending,
			want:           "degraded",
		},
		{
			name:           "unknown status maps to degraded",
			resourceStatus: domain.StatusUnknown,
			want:           "degraded",
		},
		{
			name:           "paused status maps to up",
			resourceStatus: domain.StatusPaused,
			want:           "up",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.mapResourceStatus(tt.resourceStatus)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStatusPageService_calculateDayStatus(t *testing.T) {
	service := NewStatusPageService(nil, nil, nil, nil)
	now := time.Now()
	dayStart := now.Truncate(24 * time.Hour)
	dayEnd := dayStart.Add(24 * time.Hour)

	t.Run("returns down for major incident", func(t *testing.T) {
		// Incident covering 18 hours (75% of day)
		incidents := []*domain.Incident{
			{
				StartedAt:  dayStart.Add(2 * time.Hour),
				ResolvedAt: ptrTime(dayStart.Add(20 * time.Hour)),
			},
		}

		status := service.calculateDayStatus(dayStart, dayEnd, nil, incidents)
		assert.Equal(t, "down", status)
	})

	t.Run("returns degraded for minor incident with good uptime", func(t *testing.T) {
		// Incident covering 2 hours (8% of day)
		incidents := []*domain.Incident{
			{
				StartedAt:  dayStart.Add(2 * time.Hour),
				ResolvedAt: ptrTime(dayStart.Add(4 * time.Hour)),
			},
		}

		// 95% success rate
		activities := make([]*domain.MonitoringActivity, 0)
		for i := 0; i < 95; i++ {
			activities = append(activities, &domain.MonitoringActivity{
				Base:    domain.Base{CreatedAt: dayStart.Add(time.Duration(i) * time.Minute)},
				Success: true,
			})
		}
		for i := 0; i < 5; i++ {
			activities = append(activities, &domain.MonitoringActivity{
				Base:    domain.Base{CreatedAt: dayStart.Add(time.Duration(i+95) * time.Minute)},
				Success: false,
			})
		}

		status := service.calculateDayStatus(dayStart, dayEnd, activities, incidents)
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

	t.Run("returns up when no data available", func(t *testing.T) {
		status := service.calculateDayStatus(dayStart, dayEnd, nil, nil)
		assert.Equal(t, "up", status)
	})
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{
			name:     "seconds only",
			duration: 45 * time.Second,
			want:     "45s",
		},
		{
			name:     "minutes only",
			duration: 15 * time.Minute,
			want:     "15m",
		},
		{
			name:     "hours and minutes",
			duration: 2*time.Hour + 30*time.Minute,
			want:     "2h 30m",
		},
		{
			name:     "days and hours",
			duration: 3*24*time.Hour + 5*time.Hour,
			want:     "3d 5h",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.duration)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Helper function to create time pointer
func ptrTime(t time.Time) *time.Time {
	return &t
}
