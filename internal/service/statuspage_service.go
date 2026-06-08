package service

import (
	"context"
	"fmt"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/dto"
	"github.com/denisakp/ogoune/internal/port"
)

// StatusPageService provides data aggregation for public status pages.
type StatusPageService struct {
	resourceRepo           port.ResourceRepository
	incidentRepo           port.IncidentRepository
	monitoringActivityRepo port.MonitoringActivityRepository
	maintenanceRepo        port.MaintenanceRepository
	settingsRepo           port.StatusPageSettingsRepository
	componentRepo          port.ComponentRepository
}

// NewStatusPageService creates a new StatusPageService instance.
func NewStatusPageService(
	resourceRepo port.ResourceRepository,
	incidentRepo port.IncidentRepository,
	monitoringActivityRepo port.MonitoringActivityRepository,
	maintenanceRepo port.MaintenanceRepository,
	settingsRepo port.StatusPageSettingsRepository,
	componentRepo port.ComponentRepository,
) *StatusPageService {
	return &StatusPageService{
		resourceRepo:           resourceRepo,
		incidentRepo:           incidentRepo,
		monitoringActivityRepo: monitoringActivityRepo,
		maintenanceRepo:        maintenanceRepo,
		settingsRepo:           settingsRepo,
		componentRepo:          componentRepo,
	}
}

// GetData fetches and aggregates all data needed for the status page.
func (s *StatusPageService) GetData(ctx context.Context) (*dto.StatusPageData, error) {
	// Get settings to check if paused monitors should be hidden
	var settings *domain.StatusPageSettings
	if s.settingsRepo != nil {
		var err error
		settings, err = s.settingsRepo.Get(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch settings: %w", err)
		}
	}

	// Fetch all active resources
	resources, err := s.resourceRepo.FindActive(ctx, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch resources: %w", err)
	}

	components := []*domain.Component{}
	if s.componentRepo != nil {
		components, _ = s.componentRepo.List(ctx, 1000, 0)
	}

	// Filter paused resources if setting is enabled (default: hide paused)
	hidePaused := settings == nil || settings.HidePausedMonitors
	if hidePaused {
		filteredResources := make([]*domain.Resource, 0, len(resources))
		for _, r := range resources {
			if r.Status != domain.StatusPaused {
				filteredResources = append(filteredResources, r)
			}
		}
		resources = filteredResources
	}

	// Build resource status info with 90-day data
	resourceInfos := make([]dto.ResourceStatusInfo, 0, len(resources))
	resourcesByComponent := make(map[string][]dto.ResourceStatusInfo)
	hasNonOperationalResource := false

	for _, resource := range resources {
		info, err := s.buildResourceStatusInfo(ctx, resource)
		if err != nil {
			// Log error but continue processing other resources
			continue
		}

		// If resource belongs to a component, add it to the component's resource list
		// but do NOT add it to the standalone resourceInfos list
		if resource.ComponentID != nil {
			resourcesByComponent[*resource.ComponentID] = append(resourcesByComponent[*resource.ComponentID], info)
		} else {
			// Only standalone resources (no component) go to the main resources list
			resourceInfos = append(resourceInfos, info)
		}

		// Check if any resource is not "up", excluding pending/unknown resources
		// Newly added or unchecked resources should not count as an outage
		if info.CurrentStatus != "up" && info.CurrentStatus != "degraded" {
			hasNonOperationalResource = true
		} else if info.CurrentStatus == "degraded" {
			// Only count degraded if it's not from pending/unknown status
			if resource.Status != domain.StatusPending && resource.Status != domain.StatusUnknown {
				hasNonOperationalResource = true
			}
		}
	}

	// Calculate global status
	globalStatus := "all_systems_operational"
	if hasNonOperationalResource {
		globalStatus = "some_systems_down"
	}

	// Build settings DTO (only include public-facing fields)
	var settingsDTO *dto.StatusPageSettings
	if settings != nil {
		settingsDTO = &dto.StatusPageSettings{
			Name:                 settings.Name,
			HomepageURL:          settings.HomepageURL,
			UmamiWebsiteID:       settings.UmamiWebsiteID,
			UmamiScriptURL:       settings.UmamiScriptURL,
			EnableDetailsPage:    settings.EnableDetailsPage,
			ShowUptimePercentage: settings.ShowUptimePercentage,
		}
	}

	componentInfos := make([]dto.ComponentStatusInfo, 0, len(components))
	for _, component := range components {
		infos := resourcesByComponent[component.ID]
		componentStatus := s.mapComponentStatus(infos)
		componentInfos = append(componentInfos, dto.ComponentStatusInfo{
			ID:            component.ID,
			Name:          component.Name,
			Description:   component.Description,
			CurrentStatus: componentStatus,
			Resources:     infos,
		})
	}

	return &dto.StatusPageData{
		GlobalStatus: globalStatus,
		GeneratedAt:  time.Now(),
		Resources:    resourceInfos,
		Components:   componentInfos,
		Settings:     settingsDTO,
	}, nil
}

// buildResourceStatusInfo creates a ResourceStatusInfo from a Resource entity.
func (s *StatusPageService) buildResourceStatusInfo(ctx context.Context, resource *domain.Resource) (dto.ResourceStatusInfo, error) {
	// Calculate current status (map domain status to simplified enum)
	currentStatus := s.mapResourceStatus(resource.Status)

	// Calculate 90-day uptime and daily status
	uptimePercentage, dailyStatus, err := s.calculate90DayData(ctx, resource)
	if err != nil {
		return dto.ResourceStatusInfo{}, fmt.Errorf("failed to calculate 90-day data for resource %s: %w", resource.ID, err)
	}

	return dto.ResourceStatusInfo{
		ID:                         resource.ID,
		Name:                       resource.Name,
		CurrentStatus:              currentStatus,
		UptimePercentageLast90Days: uptimePercentage,
		DailyStatusLast90Days:      dailyStatus,
	}, nil
}

// mapResourceStatus maps domain.ResourceStatus to simplified status enum
func (s *StatusPageService) mapResourceStatus(status domain.ResourceStatus) string {
	switch status {
	case domain.StatusUp:
		return "up"
	case domain.StatusDown, domain.StatusError:
		return "down"
	case domain.StatusWarn, domain.StatusPending, domain.StatusUnknown:
		return "degraded"
	case domain.StatusPaused:
		return "up" // Paused resources are considered operational
	default:
		return "degraded"
	}
}

func (s *StatusPageService) mapComponentStatus(resources []dto.ResourceStatusInfo) string {
	hasDown := false
	hasDegraded := false

	for _, r := range resources {
		switch r.CurrentStatus {
		case "down":
			hasDown = true
		case "degraded":
			hasDegraded = true
		}
	}

	switch {
	case hasDown:
		return "down"
	case hasDegraded:
		return "degraded"
	default:
		return "up"
	}
}

// GetResourceDetailStatus fetches detailed status information for a single resource
func (s *StatusPageService) GetResourceDetailStatus(ctx context.Context, resourceID string) (*dto.ResourceDetailStatusData, error) {
	// Fetch the resource
	resource, err := s.resourceRepo.FindByID(ctx, resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch resource: %w", err)
	}

	// Get current status
	currentStatus := s.mapResourceStatus(resource.Status)

	// Get 90-day uptime history (reuse existing logic)
	uptimePercentage90Days, dailyStatus90Days, err := s.calculate90DayData(ctx, resource)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate 90-day data: %w", err)
	}

	// Calculate uptime summary for different time windows
	uptimeSummary, err := s.calculateUptimeSummary(ctx, resource)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate uptime summary: %w", err)
	}
	uptimeSummary.Last90Days = uptimePercentage90Days

	// Calculate response time summary for last 7 days
	responseTimeSummary, err := s.calculateResponseTimeSummary7Days(ctx, resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate response time summary: %w", err)
	}

	// Get recent events (incidents and resolutions)
	recentEvents, err := s.buildRecentEvents(ctx, resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to build recent events: %w", err)
	}

	// Determine last updated time
	lastUpdated := time.Now()
	if resource.LastChecked != nil {
		lastUpdated = *resource.LastChecked
	}

	detail := &dto.ResourceDetailStatusData{
		ID:                    resource.ID,
		Name:                  resource.Name,
		CurrentStatus:         currentStatus,
		LastUpdated:           lastUpdated,
		UptimeHistory90Days:   dailyStatus90Days,
		UptimeSummary:         uptimeSummary,
		ResponseTimeSummary7D: responseTimeSummary,
		RecentEvents:          recentEvents,
	}

	// Attach maintenance banner information: active has priority, else upcoming scheduled
	// Active or currently within window
	// Maintenance repo may be nil in some test contexts
	now := time.Now()
	if s.maintenanceRepo != nil {
		activeMaintenances, err := s.maintenanceRepo.FindActiveForResource(ctx, resourceID, now)
		if err == nil && len(activeMaintenances) > 0 {
			m := activeMaintenances[0]
			detail.Maintenance = &dto.MaintenanceBanner{
				Status:   "active",
				Title:    m.Title,
				StartAt:  m.StartAt,
				EndAt:    m.EndAt,
				Timezone: m.Timezone,
			}
			return detail, nil
		}

		// Upcoming scheduled: pick the nearest future StartAt
		scheduled, err := s.maintenanceRepo.List(ctx, "scheduled", 100, 0)
		if err == nil && len(scheduled) > 0 {
			var candidate *domain.Maintenance
			for _, m := range scheduled {
				if m.StartAt == nil || m.StartAt.Before(now) {
					continue
				}
				// ensure the maintenance applies to this resource
				applies := false
				for _, r := range m.Resources {
					if r.ID == resourceID {
						applies = true
						break
					}
				}
				if !applies {
					continue
				}
				if candidate == nil || m.StartAt.Before(*candidate.StartAt) {
					candidate = m
				}
			}
			if candidate != nil {
				detail.Maintenance = &dto.MaintenanceBanner{
					Status:   "scheduled",
					Title:    candidate.Title,
					StartAt:  candidate.StartAt,
					EndAt:    candidate.EndAt,
					Timezone: candidate.Timezone,
				}
			}
		}
	}

	return detail, nil
}

