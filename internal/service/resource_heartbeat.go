package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository"
)

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
