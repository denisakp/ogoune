package service

import (
	"context"
	"errors"
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/dto"
)

// LiveSnapshotServiceInterface provides the live snapshot aggregation operation.
type LiveSnapshotServiceInterface interface {
	GetLiveSnapshot(ctx context.Context, resourceID string) (*dto.LiveSnapshotResponse, error)
}

// ResourceServiceInterface defines methods used by LiveSnapshotService.
type ResourceServiceInterface interface {
	GetResourceByID(ctx context.Context, id string) (*domain.Resource, error)
}

// MonitoringActivityServiceInterface defines methods used by LiveSnapshotService.
type MonitoringActivityServiceInterface interface {
	GetResourceUptimeStats(ctx context.Context, resourceID string) (*dto.LiveStats, error)
	ListByResourceID(ctx context.Context, resourceID string, limit, offset int) ([]*domain.MonitoringActivity, error)
}

// IncidentServiceInterface defines methods used by LiveSnapshotService.
type IncidentServiceInterface interface {
	GetActiveIncident(ctx context.Context, resourceID string) (*dto.LiveActiveIncident, error)
}

// LiveSnapshotService aggregates live monitor data from multiple services.
type LiveSnapshotService struct {
	resourceService ResourceServiceInterface
	activityService MonitoringActivityServiceInterface
	incidentService IncidentServiceInterface
}

// NewLiveSnapshotService creates a new LiveSnapshotService.
func NewLiveSnapshotService(
	resourceService ResourceServiceInterface,
	activityService MonitoringActivityServiceInterface,
	incidentService IncidentServiceInterface,
) *LiveSnapshotService {
	return &LiveSnapshotService{
		resourceService: resourceService,
		activityService: activityService,
		incidentService: incidentService,
	}
}

// GetLiveSnapshot returns an aggregated snapshot of current resource state,
// uptime stats, active incident, and latest activities.
func (s *LiveSnapshotService) GetLiveSnapshot(ctx context.Context, resourceID string) (*dto.LiveSnapshotResponse, error) {
	resource, err := s.resourceService.GetResourceByID(ctx, resourceID)
	if err != nil {
		if errors.Is(err, ErrResourceNotFound) {
			return nil, err
		}
		return nil, err
	}

	stats := dto.LiveStats{}
	if liveStats, statsErr := s.activityService.GetResourceUptimeStats(ctx, resourceID); statsErr == nil && liveStats != nil {
		stats = *liveStats
	}

	activeIncident, incidentErr := s.incidentService.GetActiveIncident(ctx, resourceID)
	if incidentErr != nil {
		activeIncident = nil
	}

	recentActivities, activitiesErr := s.activityService.ListByResourceID(ctx, resourceID, 20, 0)
	if activitiesErr != nil {
		recentActivities = []*domain.MonitoringActivity{}
	}

	if len(recentActivities) > 0 {
		lastResponseTime := recentActivities[0].ResponseTime
		stats.LastResponseTime = &lastResponseTime
	}

	return &dto.LiveSnapshotResponse{
		Resource:         resource,
		Stats:            stats,
		ActiveIncident:   activeIncident,
		RecentActivities: recentActivities,
		FetchedAt:        time.Now().UTC(),
	}, nil
}
