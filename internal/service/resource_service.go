package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/denisakp/ogoune/internal/config"
	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/dto"
	icmppkg "github.com/denisakp/ogoune/internal/icmp"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository"
	"github.com/denisakp/ogoune/internal/repository/sqlc/dynquery"
	"github.com/google/uuid"
)

const (
	defaultConfirmationChecks   = 2
	defaultConfirmationInterval = 30
	errWrapFmt                  = "%w: %v"
)

// ResourceService orchestrates resource-related operations using repository interfaces.
// This service demonstrates the dependency injection pattern and serves as an example
// of how to compose repository operations while maintaining clean boundaries.
type ResourceService struct {
	resources          port.ResourceRepository
	incidents          port.IncidentRepository
	tags               port.TagsRepository
	channels           port.NotificationChannelRepository
	scheduler          port.ResourceScheduler
	monitoringActivity port.MonitoringActivityRepository
	enrichment         *EnrichmentService
	components         *ComponentService
	// resourceHealth serves the latest database health on the DETAIL path only
	// (spec 088). Optional: nil simply means no health is attached, so no existing
	// constructor call site changes.
	resourceHealth port.ResourceHealthRepository
}

// WithResourceHealth attaches the database-health store. Separate from the
// constructor to keep the wiring optional.
func (s *ResourceService) WithResourceHealth(repo port.ResourceHealthRepository) *ResourceService {
	s.resourceHealth = repo
	return s
}

// attachDatabaseHealth adds the monitor's latest database health, if any.
//
// Detail path only: the monitor list endpoints are read far more often and sit
// under a benchmark gate, so health is never loaded or joined there (FR-024a).
// A lookup failure leaves the field absent rather than failing the request --
// health is diagnostic context, not part of the monitor.
func (s *ResourceService) attachDatabaseHealth(ctx context.Context, rr *dto.ResourceResponse, resourceID string) {
	if s.resourceHealth == nil || rr == nil {
		return
	}
	rec, err := s.resourceHealth.FindByResourceID(ctx, resourceID)
	if err != nil {
		slog.Debug("db health: lookup failed", "resource_id", resourceID, "error", err)
		return
	}
	rr.DatabaseHealth = dto.MapDatabaseHealth(rec)
}

// NewResourceService creates a new ResourceService with the given repository dependencies.
func NewResourceService(
	resources port.ResourceRepository,
	incidents port.IncidentRepository,
	tags port.TagsRepository,
	channels port.NotificationChannelRepository,
	scheduler port.ResourceScheduler,
	monitoringActivity port.MonitoringActivityRepository,
	enrichment *EnrichmentService,
	components *ComponentService,
) *ResourceService {
	return &ResourceService{
		resources:          resources,
		incidents:          incidents,
		tags:               tags,
		channels:           channels,
		scheduler:          scheduler,
		monitoringActivity: monitoringActivity,
		enrichment:         enrichment,
		components:         components,
	}
}

// CreateResource creates a new resource using domain validation and persistence.
// After successful creation, it schedules monitoring for the resource and triggers
// asynchronous metadata enrichment so the HTTP request is not blocked by SSL/WHOIS lookups.
func (s *ResourceService) CreateResource(ctx context.Context, payload *dto.CreateResourcePayload) (*domain.Resource, error) {
	// The checks run in the order a caller sees them fail: capability, then
	// target, then confirmation window, then the type-specific fields.
	if err := checkICMPAvailable(payload.Type); err != nil {
		return nil, err
	}
	if payload.Type != domain.ResourceHeartbeat {
		if err := domain.ValidateResourceTarget(payload.Target, payload.Type); err != nil {
			return nil, fmt.Errorf(errWrapFmt, ErrValidationFailed, err)
		}
	}
	resolvedChecks, resolvedInterval, err := resolveConfirmation(payload)
	if err != nil {
		return nil, err
	}

	resource := &domain.Resource{
		Name:                  payload.Name,
		Type:                  payload.Type,
		Interval:              payload.Interval,
		Timeout:               payload.Timeout,
		Target:                payload.Target,
		IsActive:              true,
		Status:                domain.StatusPending,
		ConfirmationChecks:    resolvedChecks,
		ConfirmationInterval:  resolvedInterval,
		ExpiryAlertThresholds: payload.ExpiryAlertThresholds,
	}
	if err := applyTypeSpecificFields(resource, payload); err != nil {
		return nil, err
	}
	if err := applySmartAlerting(resource, payload); err != nil {
		return nil, err
	}
	if err := s.resolveComponent(ctx, resource, payload.ComponentID); err != nil {
		return nil, err
	}
	if err := s.resolveTagsAndChannels(ctx, resource, payload.Tags, payload.NotificationChannelNames); err != nil {
		return nil, err
	}

	// Create resource in database
	created, err := s.resources.Create(ctx, resource)
	if err != nil {
		return nil, err
	}

	// Schedule monitoring for the newly created resource
	if err := s.scheduler.Schedule(ctx, created); err != nil {
		// Log the error but don't fail the entire operation
		// The resource was created successfully, monitoring scheduling failed
		return created, fmt.Errorf(errWrapFmt, ErrSchedulerSync, err)
	}

	// Mark metadata as pending for API consumers until enrichment completes
	created.MetadataPending = true

	// Kick off async metadata enrichment (SSL/WHOIS) so the HTTP request returns quickly
	go s.asyncEnrichAndPersist(created)

	// Sync component state if applicable (best-effort)
	if created.ComponentID != nil && s.components != nil {
		_ = s.components.RecalculateAndNotify(ctx, *created.ComponentID)
	}

	return created, nil
}

// checkICMPAvailable gates ICMP monitor creation on ENABLE_ICMP and on the
// runtime actually being able to send echo requests.
func checkICMPAvailable(t domain.ResourceType) error {
	if t != domain.ResourceICMP {
		return nil
	}
	if !config.Load().EnableICMP {
		return ErrICMPUnavailable
	}
	if cap := icmppkg.Detect(); !cap.Available {
		return ErrICMPUnavailable
	}
	return nil
}

// resolveConfirmation fills the confirmation window from the payload and the
// configured defaults, keeps the retry interval under the check interval when
// the caller left it to us, and validates the result.
func resolveConfirmation(payload *dto.CreateResourcePayload) (checks, interval int, err error) {
	defaultChecks, defaultInterval := confirmationDefaults()
	checks, interval = domain.ResolveConfirmationDefaults(
		payload.ConfirmationChecks,
		payload.ConfirmationInterval,
		defaultChecks,
		defaultInterval,
	)
	if payload.ConfirmationInterval == nil && payload.Interval > 1 && interval >= payload.Interval {
		interval = payload.Interval - 1
	}
	if err := domain.ValidateConfirmationSettings(payload.Interval, checks, interval); err != nil {
		return 0, 0, fmt.Errorf(errWrapFmt, ErrValidationFailed, err)
	}
	return checks, interval, nil
}

// applyTypeSpecificFields validates and copies the fields that exist for one
// monitor type only: heartbeat, keyword, protocol.
func applyTypeSpecificFields(resource *domain.Resource, payload *dto.CreateResourcePayload) error {
	switch payload.Type {
	case domain.ResourceHeartbeat:
		if payload.HeartbeatInterval == nil || payload.HeartbeatGrace == nil {
			return fmt.Errorf("%w: heartbeat_interval and heartbeat_grace are required", ErrValidationFailed)
		}
		if err := domain.ValidateHeartbeatSettings(*payload.HeartbeatInterval, *payload.HeartbeatGrace); err != nil {
			return err
		}
		slug := uuid.NewString()
		resource.HeartbeatSlug = &slug
		resource.HeartbeatInterval = payload.HeartbeatInterval
		resource.HeartbeatGrace = payload.HeartbeatGrace
		resource.Status = domain.StatusUp
		if resource.Target == "" {
			resource.Target = "heartbeat"
		}
	case domain.ResourceKeyword:
		if err := validateKeywordFields(payload.Keyword, payload.KeywordMode); err != nil {
			return err
		}
		resource.Keyword = payload.Keyword
		defaultMode := "contains"
		if payload.KeywordMode != nil {
			resource.KeywordMode = payload.KeywordMode
		} else {
			resource.KeywordMode = &defaultMode
		}
	case domain.ResourceProtocol:
		if err := validateProtocolFields(payload.ProtocolType, payload.ProtocolPort, payload.Target); err != nil {
			return err
		}
		resource.ProtocolType = payload.ProtocolType
		resource.ProtocolPort = payload.ProtocolPort
	}
	return nil
}

// applySmartAlerting starts from the configured flap/reminder defaults and
// lets the payload override each one, then validates the combination.
func applySmartAlerting(resource *domain.Resource, payload *dto.CreateResourcePayload) error {
	cfg := config.Load()
	resource.FlapDetectionEnabled = cfg.FlapDetectionEnabled
	resource.FlapThreshold = cfg.FlapThreshold
	resource.FlapWindowSeconds = cfg.FlapWindowSeconds
	resource.FlapMaxDurationMinutes = cfg.FlapMaxDurationMinutes
	resource.ReminderIntervalMinutes = cfg.ReminderIntervalMinutes
	if payload.FlapDetectionEnabled != nil {
		resource.FlapDetectionEnabled = *payload.FlapDetectionEnabled
	}
	if payload.FlapThreshold != nil {
		resource.FlapThreshold = *payload.FlapThreshold
	}
	if payload.FlapWindowSeconds != nil {
		resource.FlapWindowSeconds = *payload.FlapWindowSeconds
	}
	if payload.FlapMaxDurationMinutes != nil {
		resource.FlapMaxDurationMinutes = *payload.FlapMaxDurationMinutes
	}
	if payload.ReminderIntervalMinutes != nil {
		resource.ReminderIntervalMinutes = *payload.ReminderIntervalMinutes
	}
	return validateSmartAlertingFields(resource.FlapThreshold, resource.FlapWindowSeconds, resource.FlapMaxDurationMinutes, resource.ReminderIntervalMinutes)
}

// resolveComponent validates an optional component reference and assigns it.
func (s *ResourceService) resolveComponent(ctx context.Context, resource *domain.Resource, componentID *string) error {
	if componentID == nil || *componentID == "" {
		return nil
	}
	if s.components == nil {
		return fmt.Errorf("%w: component support is not configured", ErrValidationFailed)
	}
	if _, err := s.components.GetComponent(ctx, *componentID); err != nil {
		return fmt.Errorf("%w: invalid component reference", ErrValidationFailed)
	}
	resource.ComponentID = componentID
	return nil
}

// resolveTagsAndChannels finds or creates tags by name, and resolves
// notification channels by name (lookup only; a missing channel is a
// validation error -- the bulk import path relies on that).
func (s *ResourceService) resolveTagsAndChannels(ctx context.Context, resource *domain.Resource, tagNames, channelNames []string) error {
	if len(tagNames) > 0 {
		tags, err := s.findOrCreateTags(ctx, tagNames)
		if err != nil {
			return fmt.Errorf("failed to process tags: %w", err)
		}
		resource.Tags = tags
	}
	if len(channelNames) > 0 {
		channels, err := s.resolveChannelsByName(ctx, channelNames)
		if err != nil {
			return err
		}
		resource.NotificationChannels = channels
	}
	return nil
}

// GetResourceByID retrieves a resource by its ID.
func (s *ResourceService) GetResourceByID(ctx context.Context, id string) (*domain.Resource, error) {
	resource, err := s.resources.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrResourceNotFound
		}
		return nil, err
	}

	// load incidents
	// load incidents and convert to the expected value slice type
	incidentPtrs, err := s.incidents.FindByResource(ctx, id, 100, 0) // default limit 100
	if err != nil {
		return nil, fmt.Errorf("failed to load incidents for resource: %w", err)
	}
	incidents := make([]domain.Incident, 0, len(incidentPtrs))
	for _, inc := range incidentPtrs {
		if inc == nil {
			continue
		}
		incidents = append(incidents, *inc)
	}
	resource.Incidents = incidents

	return resource, nil

}

// GetResourceByIDWithResponseTimes retrieves a resource by its ID with recent response times.
func (s *ResourceService) GetResourceByIDWithResponseTimes(ctx context.Context, id string, limit int) (*dto.ResourceResponse, error) {
	// Get the base resource
	resource, err := s.GetResourceByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Normalize tags to include only id, name, and color to keep payload focused
	if len(resource.Tags) > 0 {
		trimmed := make([]*domain.Tags, 0, len(resource.Tags))
		for _, t := range resource.Tags {
			if t == nil {
				continue
			}
			trimmed = append(trimmed, &domain.Tags{
				Base:  domain.Base{ID: t.ID},
				Name:  t.Name,
				Color: t.Color,
			})
		}
		resource.Tags = trimmed
	}

	// Get recent response times
	responsePoints, err := s.monitoringActivity.GetRecentResponseTimes(ctx, id, limit)
	if err != nil {
		// Don't fail the entire request if we can't get response times
		// Just log and return resource without response times
		rr := &dto.ResourceResponse{
			Resource:      *resource,
			ResponseTimes: []dto.ResponseTimePoint{},
		}
		dto.EnrichResponseExpiry(rr)
		s.attachDatabaseHealth(ctx, rr, id)
		return rr, nil
	}

	// Map to DTO response times
	responseTimes := make([]dto.ResponseTimePoint, len(responsePoints))
	for i, point := range responsePoints {
		responseTimes[i] = dto.ResponseTimePoint{
			Timestamp:    point.Timestamp,
			ResponseTime: point.ResponseTime,
		}
	}

	rr := &dto.ResourceResponse{
		Resource:      *resource,
		ResponseTimes: responseTimes,
	}
	dto.EnrichResponseExpiry(rr)
	rr.Waiting = resource.IsHeartbeatWaiting()
	s.attachDatabaseHealth(ctx, rr, id)

	return rr, nil
}

// UpdateResource updates an existing resource by ID with the provided payload.
func (s *ResourceService) UpdateResource(ctx context.Context, id string, payload *dto.UpdateResourcePayload) (*domain.Resource, error) {
	resource, err := s.resources.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrResourceNotFound
		}
		return nil, err
	}
	previousComponentID := resource.ComponentID

	if err := s.applyResourcePayload(ctx, resource, payload); err != nil {
		return nil, err
	}

	if err := s.validateTypeSpecificUpdate(resource, payload); err != nil {
		return nil, err
	}

	resource.MetadataPending = true
	go s.asyncEnrichAndPersist(resource)

	if err := s.resources.Update(ctx, resource); err != nil {
		return nil, err
	}

	if err := s.scheduler.Schedule(ctx, resource); err != nil {
		return nil, fmt.Errorf(errWrapFmt, ErrSchedulerSync, err)
	}

	s.reconcileComponentChange(ctx, resource, previousComponentID)

	return resource, nil
}

// DeleteResource soft deletes a resource and unschedules its monitoring.
func (s *ResourceService) DeleteResource(ctx context.Context, resourceID string) error {
	var componentID *string
	if res, err := s.resources.FindByID(ctx, resourceID); err == nil {
		componentID = res.ComponentID
	}

	// Resolve any active incident before soft-deleting to prevent orphan incidents.
	if incident, err := s.incidents.FindActiveByResourceID(ctx, resourceID); err == nil {
		now := time.Now()
		incident.ResolvedAt = &now
		if err := s.incidents.Update(ctx, incident); err != nil {
			slog.Error("failed to resolve incident for deleted resource", "resource_id", resourceID, "error", err)
		}
	}
	// ErrNotFound from FindActiveByResourceID is expected (no active incident) — proceed silently.

	if err := s.resources.Delete(ctx, resourceID); err != nil {
		return err
	}

	// Unschedule monitoring for the deleted resource
	if err := s.scheduler.Unschedule(ctx, resourceID); err != nil {
		slog.Error("failed to unschedule resource", "resource_id", resourceID, "error", err)
	}

	// Auto-cleanup component if it becomes empty after resource deletion
	if componentID != nil && s.components != nil {
		count, err := s.resources.CountByComponentID(ctx, *componentID)
		if err == nil && count == 0 {
			// Component is now empty - auto-delete it
			_ = s.components.DeleteComponent(ctx, *componentID)
		} else if err == nil {
			// Component still has resources - recalculate status
			_ = s.components.RecalculateAndNotify(ctx, *componentID)
		}
	}

	return nil
}

// ListActiveResources returns all active resources with pagination.
func (s *ResourceService) ListActiveResources(ctx context.Context, limit, offset int) ([]*domain.Resource, error) {
	return s.resources.FindActive(ctx, limit, offset)
}

// ListResourcesByTag returns resources filtered by a specific tag.
func (s *ResourceService) ListResourcesByTag(ctx context.Context, tagName string, limit, offset int) ([]*domain.Resource, error) {
	return s.resources.FindByTag(ctx, tagName, limit, offset)
}

// ListUnresolvedIncidents returns unresolved incidents for a specific resource.
func (s *ResourceService) ListUnresolvedIncidents(ctx context.Context, resourceID string) ([]*domain.Incident, error) {
	// First verify resource exists
	_, err := s.resources.FindByID(ctx, resourceID)
	if err != nil {
		return nil, err
	}

	// Get unresolved incidents for this resource
	return s.incidents.FindByResource(ctx, resourceID, 50, 0) // Default limit of 50
}

// ListAll retrieves all monitoring resources from the repository.
// This method supports listing all resources without pagination for simple use cases.
func (s *ResourceService) ListAll(ctx context.Context) ([]*domain.Resource, error) {
	// Use a large limit to get all resources (can be optimized with proper pagination later)
	return s.resources.List(ctx, 1000, 0)
}

// ListByFilter passes the dynamic filter through to the repo.
func (s *ResourceService) ListByFilter(ctx context.Context, f dynquery.MonitorFilter, page, perPage int) ([]*domain.Resource, int, error) {
	return s.resources.ListResourcesByFilter(ctx, f, page, perPage)
}
