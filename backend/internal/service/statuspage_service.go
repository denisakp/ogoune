package service

import (
	"context"
	"fmt"
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/dto"
	"github.com/denisakp/pulseguard/internal/repository"
)

// StatusPageService provides data aggregation for public status pages.
type StatusPageService struct {
	resourceRepo           repository.ResourceRepository
	incidentRepo           repository.IncidentRepository
	monitoringActivityRepo repository.MonitoringActivityRepository
}

// NewStatusPageService creates a new StatusPageService instance.
func NewStatusPageService(
	resourceRepo repository.ResourceRepository,
	incidentRepo repository.IncidentRepository,
	monitoringActivityRepo repository.MonitoringActivityRepository,
) *StatusPageService {
	return &StatusPageService{
		resourceRepo:           resourceRepo,
		incidentRepo:           incidentRepo,
		monitoringActivityRepo: monitoringActivityRepo,
	}
}

// GetData fetches and aggregates all data needed for the status page.
func (s *StatusPageService) GetData(ctx context.Context) (*dto.StatusPageData, error) {
	// Fetch all active resources (limit 1000, offset 0 to get all)
	resources, err := s.resourceRepo.FindActive(ctx, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch resources: %w", err)
	}

	// Build resource status info
	resourceInfos := make([]dto.ResourceStatusInfo, 0, len(resources))
	for _, resource := range resources {
		info, err := s.buildResourceStatusInfo(ctx, resource)
		if err != nil {
			// Log error but continue processing other resources
			continue
		}
		resourceInfos = append(resourceInfos, info)
	}

	// Fetch recent incidents (limit 1000 for last 90 days)
	since := time.Now().AddDate(0, 0, -90)
	incidents, err := s.incidentRepo.List(ctx, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch incidents: %w", err)
	}

	// Filter and build incident summaries
	incidentSummaries := make([]dto.IncidentSummary, 0)
	for _, incident := range incidents {
		if incident.StartedAt.Before(since) {
			continue
		}
		summary := s.buildIncidentSummary(incident)
		incidentSummaries = append(incidentSummaries, summary)
	}

	return &dto.StatusPageData{
		Resources: resourceInfos,
		Incidents: incidentSummaries,
		Generated: time.Now(),
	}, nil
}

// buildResourceStatusInfo creates a ResourceStatusInfo from a Resource entity.
func (s *StatusPageService) buildResourceStatusInfo(ctx context.Context, resource *domain.Resource) (dto.ResourceStatusInfo, error) {
	// Calculate 30-day uptime
	since := time.Now().AddDate(0, 0, -30)
	activities, err := s.monitoringActivityRepo.FindByResourceID(ctx, resource.ID, 10000, 0)
	if err != nil {
		return dto.ResourceStatusInfo{}, fmt.Errorf("failed to fetch monitoring activities for resource %s: %w", resource.ID, err)
	}

	// Filter activities from last 30 days and calculate uptime
	var totalChecks, successfulChecks int
	var lastChecked time.Time
	var lastResponseTime int

	for _, activity := range activities {
		if activity.CreatedAt.Before(since) {
			continue
		}
		totalChecks++
		if activity.Success {
			successfulChecks++
		}
		if activity.CreatedAt.After(lastChecked) {
			lastChecked = activity.CreatedAt
			lastResponseTime = activity.ResponseTime
		}
	}

	uptime := 100.0
	if totalChecks > 0 {
		uptime = (float64(successfulChecks) / float64(totalChecks)) * 100
	}

	return dto.ResourceStatusInfo{
		ID:               resource.ID,
		Name:             resource.Name,
		Type:             string(resource.Type),
		CurrentStatus:    string(resource.Status),
		UptimeLast30Days: uptime,
		LastChecked:      lastChecked,
		ResponseTime:     lastResponseTime,
	}, nil
}

// buildIncidentSummary creates an IncidentSummary from an Incident entity.
func (s *StatusPageService) buildIncidentSummary(incident *domain.Incident) dto.IncidentSummary {
	var duration string
	var isOngoing bool

	if incident.ResolvedAt != nil {
		diff := incident.ResolvedAt.Sub(incident.StartedAt)
		duration = formatDuration(diff)
		isOngoing = false
	} else {
		diff := time.Since(incident.StartedAt)
		duration = formatDuration(diff)
		isOngoing = true
	}

	return dto.IncidentSummary{
		ID:         incident.ID,
		ResourceID: incident.ResourceID,
		Resource:   "", // Will be populated by handler if needed
		Reason:     incident.Reason,
		Cause:      incident.Cause,
		StartedAt:  incident.StartedAt,
		ResolvedAt: incident.ResolvedAt,
		Duration:   duration,
		IsOngoing:  isOngoing,
	}
}

// formatDuration formats a duration into a human-readable string.
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	return fmt.Sprintf("%dd %dh", days, hours)
}
