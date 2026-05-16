package service

import (
	"context"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
)

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
		return
	}

	if metadata == nil {
		return
	}

	// Persist the metadata without touching tags/associations
	_ = s.resources.UpdateMetadata(context.Background(), r.ID, metadata)
}
