package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/denisakp/ogoune/internal/config"
	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/dto"
	icmppkg "github.com/denisakp/ogoune/internal/icmp"
	"github.com/denisakp/ogoune/internal/repository"
	"github.com/google/uuid"
)

const (
	defaultConfirmationChecks   = 2
	defaultConfirmationInterval = 30
)

// ResourceService orchestrates resource-related operations using repository interfaces.
// This service demonstrates the dependency injection pattern and serves as an example
// of how to compose repository operations while maintaining clean boundaries.
type ResourceService struct {
	resources          repository.ResourceRepository
	incidents          repository.IncidentRepository
	tags               repository.TagsRepository
	scheduler          repository.Scheduler
	monitoringActivity repository.MonitoringActivityRepository
	enrichment         *EnrichmentService
	components         *ComponentService
}

// NewResourceService creates a new ResourceService with the given repository dependencies.
func NewResourceService(
	resources repository.ResourceRepository,
	incidents repository.IncidentRepository,
	tags repository.TagsRepository,
	scheduler repository.Scheduler,
	monitoringActivity repository.MonitoringActivityRepository,
	enrichment *EnrichmentService,
	components *ComponentService,
) *ResourceService {
	return &ResourceService{
		resources:          resources,
		incidents:          incidents,
		tags:               tags,
		scheduler:          scheduler,
		monitoringActivity: monitoringActivity,
		enrichment:         enrichment,
		components:         components,
	}
}

// findOrCreateTags finds existing tags by name or creates new ones if they don't exist.
// It accepts tag names as strings and returns tag entities.
func (s *ResourceService) findOrCreateTags(ctx context.Context, tagNames []string) ([]*domain.Tags, error) {
	var tags []*domain.Tags

	for _, rawTag := range tagNames {
		tagName := strings.TrimSpace(rawTag)
		if tagName == "" {
			continue
		}

		// Backward compatibility: if client sends an existing tag ID, resolve it first.
		tagByID, err := s.tags.FindByID(ctx, tagName)
		if err == nil {
			tags = append(tags, tagByID)
			continue
		}
		if !errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("failed to find tag by id '%s': %w", tagName, err)
		}

		// Try to find the tag by name
		tag, err := s.tags.FindByName(ctx, tagName)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				// Tag doesn't exist, create it
				newTag := &domain.Tags{
					Name: tagName,
				}
				if err := s.tags.Create(ctx, newTag); err != nil {
					return nil, fmt.Errorf("failed to create tag '%s': %w", tagName, err)
				}
				tags = append(tags, newTag)
			} else {
				return nil, fmt.Errorf("failed to find tag '%s': %w", tagName, err)
			}
		} else {
			// Tag exists, use it
			tags = append(tags, tag)
		}
	}

	return tags, nil
}

// CreateResource creates a new resource using domain validation and persistence.
// After successful creation, it schedules monitoring for the resource and triggers
// asynchronous metadata enrichment so the HTTP request is not blocked by SSL/WHOIS lookups.
func (s *ResourceService) CreateResource(ctx context.Context, payload *dto.CreateResourcePayload) (*domain.Resource, error) {
	// Gate ICMP monitor creation: requires ENABLE_ICMP and runtime capability.
	if payload.Type == domain.ResourceICMP {
		cfg := config.Load()
		if !cfg.EnableICMP {
			return nil, ErrICMPUnavailable
		}
		if cap := icmppkg.Detect(); !cap.Available {
			return nil, ErrICMPUnavailable
		}
	}

	// Validate target format
	if payload.Type != domain.ResourceHeartbeat {
		if err := domain.ValidateResourceTarget(payload.Target, payload.Type); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrValidationFailed, err)
		}
	}

	defaultChecks, defaultInterval := confirmationDefaults()
	resolvedChecks, resolvedInterval := domain.ResolveConfirmationDefaults(
		payload.ConfirmationChecks,
		payload.ConfirmationInterval,
		defaultChecks,
		defaultInterval,
	)
	if payload.ConfirmationInterval == nil && payload.Interval > 1 && resolvedInterval >= payload.Interval {
		resolvedInterval = payload.Interval - 1
	}
	if err := domain.ValidateConfirmationSettings(payload.Interval, resolvedChecks, resolvedInterval); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrValidationFailed, err)
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

	if payload.Type == domain.ResourceHeartbeat {
		if payload.HeartbeatInterval == nil || payload.HeartbeatGrace == nil {
			return nil, fmt.Errorf("%w: heartbeat_interval and heartbeat_grace are required", ErrValidationFailed)
		}
		if err := domain.ValidateHeartbeatSettings(*payload.HeartbeatInterval, *payload.HeartbeatGrace); err != nil {
			return nil, err
		}
		slug := uuid.NewString()
		resource.HeartbeatSlug = &slug
		resource.HeartbeatInterval = payload.HeartbeatInterval
		resource.HeartbeatGrace = payload.HeartbeatGrace
		resource.Status = domain.StatusUp
		if resource.Target == "" {
			resource.Target = "heartbeat"
		}
	}

	if payload.Type == domain.ResourceKeyword {
		if err := validateKeywordFields(payload.Keyword, payload.KeywordMode); err != nil {
			return nil, err
		}
		resource.Keyword = payload.Keyword
		defaultMode := "contains"
		if payload.KeywordMode != nil {
			resource.KeywordMode = payload.KeywordMode
		} else {
			resource.KeywordMode = &defaultMode
		}
	}

	if payload.Type == domain.ResourceProtocol {
		if err := validateProtocolFields(payload.ProtocolType, payload.ProtocolPort); err != nil {
			return nil, err
		}
		resource.ProtocolType = payload.ProtocolType
		resource.ProtocolPort = payload.ProtocolPort
	}

	// Optional component assignment
	// Apply smart alerting config defaults and per-resource overrides
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
	if err := validateSmartAlertingFields(resource.FlapThreshold, resource.FlapWindowSeconds, resource.FlapMaxDurationMinutes, resource.ReminderIntervalMinutes); err != nil {
		return nil, err
	}

	// Optional component assignment
	if payload.ComponentID != nil && *payload.ComponentID != "" {
		if s.components == nil {
			return nil, fmt.Errorf("%w: component support is not configured", ErrValidationFailed)
		}
		if _, err := s.components.GetComponent(ctx, *payload.ComponentID); err != nil {
			return nil, fmt.Errorf("%w: invalid component reference", ErrValidationFailed)
		}
		resource.ComponentID = payload.ComponentID
	}

	// Find or create tags by name if provided
	if len(payload.Tags) > 0 {
		tags, err := s.findOrCreateTags(ctx, payload.Tags)
		if err != nil {
			return nil, fmt.Errorf("failed to process tags: %w", err)
		}
		resource.Tags = tags
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
		return created, fmt.Errorf("%w: %v", ErrSchedulerSync, err)
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

	return rr, nil
}

// UpdateResource updates an existing resource by ID with the provided payload.
// It fetches the resource, applies changes, validates, updates, and reschedules monitoring.
func (s *ResourceService) UpdateResource(ctx context.Context, id string, payload *dto.UpdateResourcePayload) (*domain.Resource, error) {
	// Fetch existing resource
	resource, err := s.resources.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrResourceNotFound
		}
		return nil, err
	}
	previousComponentID := resource.ComponentID

	// Apply updates from payload
	if payload.Name != nil {
		resource.Name = *payload.Name
	}
	if payload.Type != nil {
		resource.Type = *payload.Type
	}
	if payload.Target != nil {
		// Determine the resource type for validation (use new type if provided, else existing)
		resType := resource.Type
		if payload.Type != nil {
			resType = *payload.Type
		}
		if err := domain.ValidateResourceTarget(*payload.Target, resType); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrValidationFailed, err)
		}
		resource.Target = *payload.Target
	}
	if payload.Interval != nil {
		resource.Interval = *payload.Interval
	}
	if payload.Timeout != nil {
		resource.Timeout = *payload.Timeout
	}

	defaultChecks, defaultInterval := confirmationDefaults()
	if resource.ConfirmationChecks < 1 {
		resource.ConfirmationChecks = defaultChecks
	}
	if resource.ConfirmationInterval <= 0 {
		resource.ConfirmationInterval = defaultInterval
	}
	if payload.ConfirmationChecks != nil {
		resource.ConfirmationChecks = *payload.ConfirmationChecks
	}
	if payload.ConfirmationInterval != nil {
		resource.ConfirmationInterval = *payload.ConfirmationInterval
	}
	if err := domain.ValidateConfirmationSettings(resource.Interval, resource.ConfirmationChecks, resource.ConfirmationInterval); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrValidationFailed, err)
	}

	if payload.IsActive != nil {
		resource.IsActive = *payload.IsActive
	}

	if payload.ComponentID != nil {
		if s.components == nil {
			return nil, fmt.Errorf("%w: component support is not configured", ErrValidationFailed)
		}
		if *payload.ComponentID == "" {
			resource.ComponentID = nil
		} else {
			if _, err := s.components.GetComponent(ctx, *payload.ComponentID); err != nil {
				return nil, fmt.Errorf("%w: invalid component reference", ErrValidationFailed)
			}
			resource.ComponentID = payload.ComponentID
		}
	}

	// Handle tags update: payload.Tags accepts names (auto-create) and existing IDs.
	// Replace tag associations with the provided set.
	if payload.Tags != nil {
		if len(*payload.Tags) > 0 {
			tags, err := s.findOrCreateTags(ctx, *payload.Tags)
			if err != nil {
				return nil, fmt.Errorf("failed to process tags: %w", err)
			}
			resource.Tags = tags
		} else {
			// Clear all tags if empty slice provided
			resource.Tags = []*domain.Tags{}
		}
	}

	// Validate target format after updates
	if err := domain.ValidateResourceTarget(resource.Target, resource.Type); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrValidationFailed, err)
	}

	// Apply per-resource expiry alert threshold override (empty string → clear the field)
	if payload.ExpiryAlertThresholds != nil {
		if *payload.ExpiryAlertThresholds == "" {
			resource.ExpiryAlertThresholds = nil
		} else {
			resource.ExpiryAlertThresholds = payload.ExpiryAlertThresholds
		}
	}

	// Apply smart alerting overrides
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
	if err := validateSmartAlertingFields(resource.FlapThreshold, resource.FlapWindowSeconds, resource.FlapMaxDurationMinutes, resource.ReminderIntervalMinutes); err != nil {
		return nil, err
	}

	if resource.Type == domain.ResourceHeartbeat {
		if payload.HeartbeatInterval != nil {
			resource.HeartbeatInterval = payload.HeartbeatInterval
		}
		if payload.HeartbeatGrace != nil {
			resource.HeartbeatGrace = payload.HeartbeatGrace
		}
		if resource.HeartbeatInterval != nil && resource.HeartbeatGrace != nil {
			if err := domain.ValidateHeartbeatSettings(*resource.HeartbeatInterval, *resource.HeartbeatGrace); err != nil {
				return nil, fmt.Errorf("%w: %v", ErrValidationFailed, err)
			}
		}
	}

	if resource.Type == domain.ResourceKeyword {
		if payload.Keyword != nil {
			resource.Keyword = payload.Keyword
		}
		if payload.KeywordMode != nil {
			resource.KeywordMode = payload.KeywordMode
		}
		if err := validateKeywordFields(resource.Keyword, resource.KeywordMode); err != nil {
			return nil, err
		}
	}

	if resource.Type == domain.ResourceProtocol {
		if payload.ProtocolType != nil {
			resource.ProtocolType = payload.ProtocolType
		}
		if payload.ProtocolPort != nil {
			resource.ProtocolPort = payload.ProtocolPort
		}
		if err := validateProtocolFields(resource.ProtocolType, resource.ProtocolPort); err != nil {
			return nil, err
		}
	}

	// Defer SSL/WHOIS enrichment to avoid blocking the update request
	resource.MetadataPending = true
	go s.asyncEnrichAndPersist(resource)

	// Update resource in database
	if err := s.resources.Update(ctx, resource); err != nil {
		return nil, err
	}

	// Reschedule monitoring for the updated resource
	// This will unschedule the old task and create a new one with updated parameters
	if err := s.scheduler.Schedule(ctx, resource); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSchedulerSync, err)
	}

	// Re-evaluate impacted components (best-effort)
	if resource.ComponentID != nil && s.components != nil {
		_ = s.components.RecalculateAndNotify(ctx, *resource.ComponentID)
	}
	if previousComponentID != nil && s.components != nil {
		if resource.ComponentID == nil {
			// Resource removed from component - check if component is now empty
			count, err := s.resources.CountByComponentID(ctx, *previousComponentID)
			if err == nil && count == 0 {
				// Auto-cleanup empty component
				_ = s.components.DeleteComponent(ctx, *previousComponentID)
			} else if err == nil {
				// Recalculate status for previous component
				_ = s.components.RecalculateAndNotify(ctx, *previousComponentID)
			}
		} else if *previousComponentID != *resource.ComponentID {
			// Resource moved to different component - check if old component is now empty
			count, err := s.resources.CountByComponentID(ctx, *previousComponentID)
			if err == nil && count == 0 {
				_ = s.components.DeleteComponent(ctx, *previousComponentID)
			} else if err == nil {
				_ = s.components.RecalculateAndNotify(ctx, *previousComponentID)
			}
		}
	}

	return resource, nil
}

// asyncEnrichAndPersist performs metadata enrichment in the background and updates the resource
// without blocking the HTTP request lifecycle. It intentionally uses a background context with
// a bounded timeout to avoid leaking long-running WHOIS/SSL lookups.
func (s *ResourceService) asyncEnrichAndPersist(r *domain.Resource) {
	if r == nil || s.enrichment == nil {
		return
	}

	// Use a background context with a soft timeout to keep enrichment bounded
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	// Copy the minimal data needed for enrichment to avoid accidental mutation
	resourceCopy := &domain.Resource{Target: r.Target, Type: r.Type, Timeout: r.Timeout}

	metadata, err := s.enrichment.Enrich(ctx, resourceCopy)
	if err != nil {
		// Best-effort enrichment; log and exit without impacting the created resource
		// (actual logging handled upstream if needed)
		return
	}

	if metadata == nil {
		return
	}

	// Persist the metadata without touching tags/associations
	_ = s.resources.UpdateMetadata(context.Background(), r.ID, metadata)
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
			log.Printf("[resource-service] failed to resolve incident for deleted resource %s: %v", resourceID, err)
		}
	}
	// ErrNotFound from FindActiveByResourceID is expected (no active incident) — proceed silently.

	if err := s.resources.Delete(ctx, resourceID); err != nil {
		return err
	}

	// Unschedule monitoring for the deleted resource
	if err := s.scheduler.Unschedule(ctx, resourceID); err != nil {
		log.Printf("[resource-service] failed to unschedule resource %s: %v", resourceID, err)
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

func confirmationDefaults() (int, int) {
	cfg := config.Load()
	checks := cfg.ConfirmationChecks
	if checks < 1 {
		checks = defaultConfirmationChecks
	}
	interval := cfg.ConfirmationInterval
	if interval <= 0 {
		interval = defaultConfirmationInterval
	}
	return checks, interval
}

// validateKeywordFields validates keyword-specific fields when resource type is keyword.
func validateKeywordFields(keyword *string, keywordMode *string) error {
	if keyword == nil || *keyword == "" {
		return fmt.Errorf("%w: keyword is required for keyword monitor type", ErrValidationFailed)
	}
	if len(*keyword) > 500 {
		return fmt.Errorf("%w: keyword must not exceed 500 characters", ErrValidationFailed)
	}
	if keywordMode != nil && *keywordMode != "contains" && *keywordMode != "not_contains" {
		return fmt.Errorf("%w: keyword_mode must be 'contains' or 'not_contains'", ErrValidationFailed)
	}
	return nil
}

// validateProtocolFields validates protocol-specific fields when resource type is protocol.
func validateProtocolFields(protocolType *string, protocolPort *int) error {
	if protocolType == nil || *protocolType == "" {
		return fmt.Errorf("%w: protocol_type is required when resource type is 'protocol'", ErrValidationFailed)
	}
	validTypes := map[string]bool{"redis": true, "mongodb": true, "ftp": true, "ssh": true}
	if !validTypes[*protocolType] {
		return fmt.Errorf("%w: protocol_type must be one of: redis, mongodb, ftp, ssh", ErrValidationFailed)
	}
	if protocolPort != nil && (*protocolPort < 1 || *protocolPort > 65535) {
		return fmt.Errorf("%w: protocol_port must be between 1 and 65535", ErrValidationFailed)
	}
	return nil
}

// validateSmartAlertingFields validates the smart alerting configuration fields.
func validateSmartAlertingFields(flapThreshold, flapWindowSeconds, flapMaxDurationMinutes, reminderIntervalMinutes int) error {
	if flapThreshold < 2 {
		return fmt.Errorf("%w: flap_threshold must be >= 2 (got %d)", ErrValidationFailed, flapThreshold)
	}
	if flapWindowSeconds < 60 || flapWindowSeconds > 3600 {
		return fmt.Errorf("%w: flap_window_seconds must be between 60 and 3600 (got %d)", ErrValidationFailed, flapWindowSeconds)
	}
	if flapMaxDurationMinutes < 0 {
		return fmt.Errorf("%w: flap_max_duration_minutes must be >= 0 (got %d)", ErrValidationFailed, flapMaxDurationMinutes)
	}
	if reminderIntervalMinutes < 0 {
		return fmt.Errorf("%w: reminder_interval_minutes must be >= 0 (got %d)", ErrValidationFailed, reminderIntervalMinutes)
	}
	return nil
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

// PauseMonitoring pauses monitoring for a specific resource by setting IsActive to false
// and unscheduling its monitoring tasks.
func (s *ResourceService) PauseMonitoring(ctx context.Context, resourceID string) error {
	// Retrieve the resource
	resource, err := s.resources.FindByID(ctx, resourceID)
	if err != nil {
		return err
	}

	// Check if already paused
	if !resource.IsActive {
		return nil // Already paused, nothing to do
	}

	// Set IsActive to false
	resource.IsActive = false

	// Update the resource in the database
	if err := s.resources.Update(ctx, resource); err != nil {
		return err
	}

	// Unschedule monitoring tasks for this resource
	if err := s.scheduler.Unschedule(ctx, resourceID); err != nil {
		// Log the error but consider the pause operation successful
		// since the database state has been updated
		return err
	}

	return nil
}

// ResumeMonitoring resumes monitoring for a specific resource by setting IsActive to true
// and rescheduling its monitoring tasks.
func (s *ResourceService) ResumeMonitoring(ctx context.Context, resourceID string) error {
	// Retrieve the resource
	resource, err := s.resources.FindByID(ctx, resourceID)
	if err != nil {
		return err
	}

	// Check if already active
	if resource.IsActive {
		return nil // Already active, nothing to do
	}

	// Set IsActive to true
	resource.IsActive = true

	// Update the resource in the database
	if err := s.resources.Update(ctx, resource); err != nil {
		return err
	}

	// Schedule monitoring tasks for this resource
	if err := s.scheduler.Schedule(ctx, resource); err != nil {
		// Log the error but consider the resume operation successful
		// since the database state has been updated
		return err
	}

	return nil
}

// AddTagsToResource adds multiple tags to a resource using GORM's Association mode.
func (s *ResourceService) AddTagsToResource(ctx context.Context, resourceID string, tagIDs []string) error {
	// Fetch the resource
	resource, err := s.resources.FindByID(ctx, resourceID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("%w: resource not found", ErrResourceNotFound)
		}
		return err
	}

	// Fetch all tags
	var tags []*domain.Tags
	for _, tagID := range tagIDs {
		tag, err := s.tags.FindByID(ctx, tagID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return fmt.Errorf("%w: tag with ID '%s' not found", ErrValidationFailed, tagID)
			}
			return err
		}
		tags = append(tags, tag)
	}

	// Use GORM association to append tags
	// This requires database access, so we need to get the DB instance
	// For now, we'll append tags to the resource and update
	resource.Tags = append(resource.Tags, tags...)

	if err := s.resources.Update(ctx, resource); err != nil {
		return fmt.Errorf("failed to add tags to resource: %w", err)
	}

	return nil
}

// RemoveTagFromResource removes a specific tag from a resource.
func (s *ResourceService) RemoveTagFromResource(ctx context.Context, resourceID string, tagID string) error {
	// Fetch the resource with its tags
	resource, err := s.resources.FindByID(ctx, resourceID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("%w: resource not found", ErrResourceNotFound)
		}
		return err
	}

	// Verify tag exists
	_, err = s.tags.FindByID(ctx, tagID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("%w: tag not found", ErrResourceNotFound)
		}
		return err
	}

	// Filter out the tag to be removed
	var filteredTags []*domain.Tags
	tagFound := false
	for _, tag := range resource.Tags {
		if tag.ID != tagID {
			filteredTags = append(filteredTags, tag)
		} else {
			tagFound = true
		}
	}

	if !tagFound {
		return fmt.Errorf("%w: tag not associated with resource", ErrValidationFailed)
	}

	resource.Tags = filteredTags

	if err := s.resources.Update(ctx, resource); err != nil {
		return fmt.Errorf("failed to remove tag from resource: %w", err)
	}

	return nil
}

// GetResourceByHeartbeatSlug retrieves a resource by its heartbeat slug.
func (s *ResourceService) GetResourceByHeartbeatSlug(ctx context.Context, slug string) (*domain.Resource, error) {
	resource, err := s.resources.FindByHeartbeatSlug(ctx, slug)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrResourceNotFound
		}
		return nil, err
	}
	return resource, nil
}

// MarkHeartbeatPing records a ping timestamp for the given heartbeat resource.
func (s *ResourceService) MarkHeartbeatPing(ctx context.Context, resourceID string, at time.Time) error {
	if err := s.resources.UpdateLastPingAt(ctx, resourceID, at); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrResourceNotFound
		}
		return err
	}
	return nil
}

// HandleHeartbeatRecovery resolves the active incident for a previously-down heartbeat monitor.
// It updates the resource status to 'up' and marks the most recent unresolved incident as resolved.
// Errors are non-fatal from the caller's perspective — the ping has already been recorded.
func (s *ResourceService) HandleHeartbeatRecovery(ctx context.Context, resource *domain.Resource) error {
	// Persist the status transition to 'up' using a targeted column update.
	// The generic Update() uses a map that intentionally excludes 'status' (monitoring-controlled),
	// so we use UpdateStatus here to avoid that exclusion.
	if err := s.resources.UpdateStatus(ctx, resource.ID, domain.StatusUp); err != nil {
		return fmt.Errorf("failed to update heartbeat monitor status on recovery: %w", err)
	}

	// Resolve the latest unresolved incident for this resource
	recentIncidents, err := s.incidents.FindByResource(ctx, resource.ID, 10, 0)
	if err != nil {
		return fmt.Errorf("failed to find incidents for heartbeat recovery: %w", err)
	}

	for _, incident := range recentIncidents {
		if incident.ResolvedAt == nil {
			now := time.Now()
			incident.ResolvedAt = &now
			if err := s.incidents.Update(ctx, incident); err != nil {
				return fmt.Errorf("failed to resolve heartbeat incident: %w", err)
			}
			break
		}
	}

	return nil
}
