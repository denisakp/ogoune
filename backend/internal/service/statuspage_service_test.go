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
	t.Run("returns data with active resources and recent incidents", func(t *testing.T) {
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

		// Create monitoring activities (30 days of data)
		for i := 0; i < 100; i++ {
			activity := &domain.MonitoringActivity{
				Base:         domain.Base{ID: "activity-" + string(rune(i))},
				ResourceID:   resource.ID,
				Success:      i%10 != 0, // 90% success rate
				ResponseTime: 100 + i,
			}
			activity.CreatedAt = now.AddDate(0, 0, -i%30)
			require.NoError(t, activityRepo.Create(ctx, activity))
		}

		// Create recent incident
		incident := &domain.Incident{
			Base:       domain.Base{ID: "incident-1"},
			ResourceID: resource.ID,
			Cause:      "http_error",
			StartedAt:  now.Add(-2 * time.Hour),
			ResolvedAt: nil, // Ongoing
		}
		_, err = incidentRepo.Create(ctx, incident)
		require.NoError(t, err)

		// Create service
		service := NewStatusPageService(resourceRepo, incidentRepo, activityRepo)

		// Execute
		data, err := service.GetData(ctx)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, data)
		assert.Len(t, data.Resources, 1)
		assert.Len(t, data.Incidents, 1)

		// Check resource info
		resourceInfo := data.Resources[0]
		assert.Equal(t, "resource-1", resourceInfo.ID)
		assert.Equal(t, "Test Website", resourceInfo.Name)
		assert.Equal(t, "http", resourceInfo.Type)
		assert.Equal(t, "up", resourceInfo.CurrentStatus)
		assert.InDelta(t, 90.0, resourceInfo.UptimeLast30Days, 5.0) // ~90% uptime

		// Check incident info
		incidentInfo := data.Incidents[0]
		assert.Equal(t, "incident-1", incidentInfo.ID)
		assert.True(t, incidentInfo.IsOngoing)
		assert.Nil(t, incidentInfo.ResolvedAt)
	})

	t.Run("handles empty resources gracefully", func(t *testing.T) {
		resourceRepo := fake.NewResourceFake()
		incidentRepo := fake.NewIncidentFake()
		activityRepo := fake.NewMonitoringActivityFake()

		service := NewStatusPageService(resourceRepo, incidentRepo, activityRepo)

		data, err := service.GetData(context.Background())

		require.NoError(t, err)
		assert.NotNil(t, data)
		assert.Empty(t, data.Resources)
		assert.Empty(t, data.Incidents)
	})

	t.Run("filters incidents older than 90 days", func(t *testing.T) {
		resourceRepo := fake.NewResourceFake()
		incidentRepo := fake.NewIncidentFake()
		activityRepo := fake.NewMonitoringActivityFake()

		ctx := context.Background()

		// Create old incident (100 days ago)
		oldIncident := &domain.Incident{
			Base:       domain.Base{ID: "old-incident"},
			ResourceID: "resource-1",
			Cause:      "test",
			StartedAt:  time.Now().AddDate(0, 0, -100),
		}
		_, err := incidentRepo.Create(ctx, oldIncident)
		require.NoError(t, err)

		// Create recent incident (10 days ago)
		recentIncident := &domain.Incident{
			Base:       domain.Base{ID: "recent-incident"},
			ResourceID: "resource-1",
			Cause:      "test",
			StartedAt:  time.Now().AddDate(0, 0, -10),
		}
		_, err = incidentRepo.Create(ctx, recentIncident)
		require.NoError(t, err)

		service := NewStatusPageService(resourceRepo, incidentRepo, activityRepo)

		data, err := service.GetData(ctx)

		require.NoError(t, err)
		assert.Len(t, data.Incidents, 1)
		assert.Equal(t, "recent-incident", data.Incidents[0].ID)
	})
}

func TestStatusPageService_buildResourceStatusInfo(t *testing.T) {
	t.Run("calculates uptime correctly", func(t *testing.T) {
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

		// 8 successful checks, 2 failures = 80% uptime
		for i := 0; i < 8; i++ {
			activity := &domain.MonitoringActivity{
				Base:         domain.Base{ID: "success-" + string(rune(i))},
				ResourceID:   resource.ID,
				Success:      true,
				ResponseTime: 100,
			}
			activity.CreatedAt = now.AddDate(0, 0, -i)
			require.NoError(t, activityRepo.Create(ctx, activity))
		}

		for i := 0; i < 2; i++ {
			activity := &domain.MonitoringActivity{
				Base:         domain.Base{ID: "failure-" + string(rune(i))},
				ResourceID:   resource.ID,
				Success:      false,
				ResponseTime: 0,
			}
			activity.CreatedAt = now.AddDate(0, 0, -i-8)
			require.NoError(t, activityRepo.Create(ctx, activity))
		}

		service := NewStatusPageService(resourceRepo, incidentRepo, activityRepo)

		info, err := service.buildResourceStatusInfo(ctx, resource)

		require.NoError(t, err)
		assert.Equal(t, 80.0, info.UptimeLast30Days)
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

		service := NewStatusPageService(resourceRepo, incidentRepo, activityRepo)

		info, err := service.buildResourceStatusInfo(context.Background(), resource)

		require.NoError(t, err)
		assert.Equal(t, 100.0, info.UptimeLast30Days) // Default to 100% if no data
	})

	t.Run("filters activities older than 30 days", func(t *testing.T) {
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

		// Old activity (40 days ago) - should be ignored
		oldActivity := &domain.MonitoringActivity{
			Base:       domain.Base{ID: "old-activity"},
			ResourceID: resource.ID,
			Success:    false,
		}
		oldActivity.CreatedAt = now.AddDate(0, 0, -40)
		require.NoError(t, activityRepo.Create(ctx, oldActivity))

		// Recent activity (10 days ago) - should be counted
		recentActivity := &domain.MonitoringActivity{
			Base:       domain.Base{ID: "recent-activity"},
			ResourceID: resource.ID,
			Success:    true,
		}
		recentActivity.CreatedAt = now.AddDate(0, 0, -10)
		require.NoError(t, activityRepo.Create(ctx, recentActivity))

		service := NewStatusPageService(resourceRepo, incidentRepo, activityRepo)

		info, err := service.buildResourceStatusInfo(ctx, resource)

		require.NoError(t, err)
		assert.Equal(t, 100.0, info.UptimeLast30Days) // Only recent success counted
	})
}

func TestStatusPageService_buildIncidentSummary(t *testing.T) {
	t.Run("handles ongoing incident", func(t *testing.T) {
		now := time.Now()
		incident := &domain.Incident{
			Base:       domain.Base{ID: "incident-1"},
			ResourceID: "resource-1",
			Cause:      "http_error",
			StartedAt:  now.Add(-2 * time.Hour),
			ResolvedAt: nil,
		}

		service := NewStatusPageService(nil, nil, nil)
		summary := service.buildIncidentSummary(incident)

		assert.True(t, summary.IsOngoing)
		assert.Nil(t, summary.ResolvedAt)
		assert.Contains(t, summary.Duration, "h") // Duration in hours
	})

	t.Run("handles resolved incident", func(t *testing.T) {
		now := time.Now()
		resolvedAt := now.Add(-1 * time.Hour)
		incident := &domain.Incident{
			Base:       domain.Base{ID: "incident-1"},
			ResourceID: "resource-1",
			Cause:      "http_error",
			StartedAt:  now.Add(-3 * time.Hour),
			ResolvedAt: &resolvedAt,
		}

		service := NewStatusPageService(nil, nil, nil)
		summary := service.buildIncidentSummary(incident)

		assert.False(t, summary.IsOngoing)
		assert.NotNil(t, summary.ResolvedAt)
		assert.Contains(t, summary.Duration, "h") // ~2 hours duration
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
