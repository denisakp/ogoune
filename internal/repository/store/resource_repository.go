package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository"
	"github.com/denisakp/ogoune/internal/repository/sqlc/dynquery"
	"gorm.io/gorm"
)

// ResourceRepositoryImpl provides GORM-based implementation of ResourceRepository
type ResourceRepositoryImpl struct {
	db *gorm.DB
}

// NewResourceRepository creates a new ResourceRepository using GORM
func NewResourceRepository(db *gorm.DB) port.ResourceRepository {
	return &ResourceRepositoryImpl{db: db}
}

// Create persists a new resource record to the database.
func (r *ResourceRepositoryImpl) Create(ctx context.Context, resource *domain.Resource) (*domain.Resource, error) {
	if err := r.db.WithContext(ctx).Create(resource).Error; err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}
	return resource, nil
}

// FindByID retrieves a resource by its ID.
func (r *ResourceRepositoryImpl) FindByID(ctx context.Context, id string) (*domain.Resource, error) {
	var resource domain.Resource
	err := r.db.WithContext(ctx).
		Preload("Tags").
		Preload("Component").
		Preload("Credential").
		First(&resource, "id = ? AND is_active = ?", id, true).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find resource by ID: %w", err)
	}
	return &resource, nil
}

// FindByHeartbeatSlug retrieves an active heartbeat resource by slug.
func (r *ResourceRepositoryImpl) FindByHeartbeatSlug(ctx context.Context, slug string) (*domain.Resource, error) {
	var resource domain.Resource
	err := r.db.WithContext(ctx).
		Preload("Tags").
		Preload("Component").
		First(&resource, "heartbeat_slug = ? AND type = ? AND is_active = ?", slug, domain.ResourceHeartbeat, true).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find resource by heartbeat slug: %w", err)
	}
	return &resource, nil
}

// List retrieves all resources with pagination, ordered by creation time descending.
func (r *ResourceRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*domain.Resource, error) {
	var resources []*domain.Resource
	err := r.db.WithContext(ctx).
		Preload("Tags").
		Preload("Component").
		Where("is_active = ?", true).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&resources).Error

	if err != nil {
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}
	return resources, nil
}

// Update modifies an existing resource record in the database.
// It properly handles the many-to-many relationship with tags by replacing them.
func (r *ResourceRepositoryImpl) Update(ctx context.Context, resource *domain.Resource) error {
	// Use a transaction to ensure atomicity
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Use a map to avoid GORM skipping zero-value fields (e.g. bool=false, int=0, *int=nil).
		// Only include user-modifiable columns; monitoring-controlled fields (status, last_checked,
		// failure_count, last_ping_at, heartbeat_slug) are intentionally excluded.
		updates := map[string]interface{}{
			"name":                      resource.Name,
			"type":                      resource.Type,
			"target":                    resource.Target,
			"interval":                  resource.Interval,
			"timeout":                   resource.Timeout,
			"is_active":                 resource.IsActive,
			"confirmation_checks":       resource.ConfirmationChecks,
			"confirmation_interval":     resource.ConfirmationInterval,
			"component_id":              resource.ComponentID,
			"expiry_alert_thresholds":   resource.ExpiryAlertThresholds,
			"flap_detection_enabled":    resource.FlapDetectionEnabled,
			"flap_threshold":            resource.FlapThreshold,
			"flap_window_seconds":       resource.FlapWindowSeconds,
			"flap_max_duration_minutes": resource.FlapMaxDurationMinutes,
			"reminder_interval_minutes": resource.ReminderIntervalMinutes,
			"heartbeat_interval":        resource.HeartbeatInterval,
			"heartbeat_grace":           resource.HeartbeatGrace,
		}
		if err := tx.Model(resource).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update resource fields: %w", err)
		}

		// Replace tags using Association API to properly handle many-to-many relationship
		if err := tx.Model(resource).Association("Tags").Replace(resource.Tags); err != nil {
			return fmt.Errorf("failed to update resource tags: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// Delete performs a soft delete by setting IsActive to false.
func (r *ResourceRepositoryImpl) Delete(ctx context.Context, id string) error {
	// Soft delete: set IsActive to false
	result := r.db.WithContext(ctx).
		Model(&domain.Resource{}).
		Where("id = ? AND is_active = ?", id, true).
		Update("is_active", false)

	if result.Error != nil {
		return fmt.Errorf("failed to delete resource: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

// FindActive retrieves all active resources with pagination, ordered by creation time descending.
func (r *ResourceRepositoryImpl) FindActive(ctx context.Context, limit, offset int) ([]*domain.Resource, error) {
	var resources []*domain.Resource
	err := r.db.WithContext(ctx).
		Preload("Tags").
		Preload("Component").
		Where("is_active = ?", true).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&resources).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find active resources: %w", err)
	}
	return resources, nil
}

// FindScheduledResources retrieves all active resources (for scheduler startup loading).
// All active resources are assumed to be schedulable.
func (r *ResourceRepositoryImpl) FindScheduledResources(ctx context.Context) ([]*domain.Resource, error) {
	var resources []*domain.Resource
	err := r.db.WithContext(ctx).
		Preload("Tags").
		Preload("Component").
		Where("is_active = ?", true).
		Order("id ASC").
		Find(&resources).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find scheduled resources: %w", err)
	}
	return resources, nil
}

// FindByTag retrieves all resources associated with a specific tag name with pagination.
func (r *ResourceRepositoryImpl) FindByTag(ctx context.Context, tagName string, limit, offset int) ([]*domain.Resource, error) {
	var resources []*domain.Resource
	err := r.db.WithContext(ctx).
		Preload("Tags").
		Preload("Component").
		Joins("JOIN resource_tags ON resources.id = resource_tags.resource_id").
		Joins("JOIN tags ON resource_tags.tag_id = tags.id").
		Where("tags.name = ? AND resources.is_active = ?", tagName, true).
		Order("resources.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&resources).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find resources by tag: %w", err)
	}
	return resources, nil
}

// FindByComponentID returns resources assigned to a component.
func (r *ResourceRepositoryImpl) FindByComponentID(ctx context.Context, componentID string) ([]*domain.Resource, error) {
	var resources []*domain.Resource
	err := r.db.WithContext(ctx).
		Preload("Tags").
		Preload("Component").
		Where("component_id = ? AND is_active = ?", componentID, true).
		Order("created_at DESC").
		Find(&resources).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find resources by component: %w", err)
	}
	return resources, nil
}

// CountByComponentID returns how many resources are assigned to a component.
func (r *ResourceRepositoryImpl) CountByComponentID(ctx context.Context, componentID string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&domain.Resource{}).
		Where("component_id = ? AND is_active = ?", componentID, true).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count resources for component: %w", err)
	}
	return count, nil
}

// FindMissedHeartbeats returns heartbeat resources that exceeded interval+grace and are still up.
func (r *ResourceRepositoryImpl) FindMissedHeartbeats(ctx context.Context, now time.Time, limit int) ([]*domain.Resource, error) {
	if limit <= 0 {
		limit = 1000
	}

	var resources []*domain.Resource
	err := r.db.WithContext(ctx).
		Where("type = ?", domain.ResourceHeartbeat).
		Where("status = ?", domain.StatusUp).
		Where("is_active = ?", true).
		Where("last_ping_at IS NOT NULL").
		Where("(strftime('%s', last_ping_at) + heartbeat_interval + heartbeat_grace) < ?", now.Unix()).
		Order("last_ping_at ASC").
		Limit(limit).
		Find(&resources).Error
	if err != nil {
		// Postgres fallback expression when SQLite strftime is not supported.
		err = r.db.WithContext(ctx).
			Where("type = ?", domain.ResourceHeartbeat).
			Where("status = ?", domain.StatusUp).
			Where("is_active = ?", true).
			Where("last_ping_at IS NOT NULL").
			Where("EXTRACT(EPOCH FROM last_ping_at) + heartbeat_interval + heartbeat_grace < ?", now.Unix()).
			Order("last_ping_at ASC").
			Limit(limit).
			Find(&resources).Error
		if err != nil {
			return nil, fmt.Errorf("failed to find missed heartbeats: %w", err)
		}
	}

	return resources, nil
}

// UpdateLastPingAt updates last_ping_at for a heartbeat resource.
func (r *ResourceRepositoryImpl) UpdateLastPingAt(ctx context.Context, id string, at time.Time) error {
	result := r.db.WithContext(ctx).
		Model(&domain.Resource{}).
		Where("id = ? AND type = ? AND is_active = ?", id, domain.ResourceHeartbeat, true).
		Update("last_ping_at", at)

	if result.Error != nil {
		return fmt.Errorf("failed to update last_ping_at: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

// UpdateMonitoringState persists the monitoring-controlled fields after a check cycle.
// Pointer fields on the request distinguish "preserve" (nil) from "write zero" (&zero).
func (r *ResourceRepositoryImpl) UpdateMonitoringState(ctx context.Context, id string, req port.UpdateMonitoringStateRequest) error {
	updates := map[string]interface{}{}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.FailureCount != nil {
		updates["failure_count"] = *req.FailureCount
	}
	if req.LastChecked != nil {
		updates["last_checked"] = *req.LastChecked
	}
	if req.LastStatusTransition != nil {
		updates["last_status_transition"] = *req.LastStatusTransition
	}
	if req.FlapStartedAt != nil {
		updates["flap_started_at"] = *req.FlapStartedAt
	}
	if len(updates) == 0 {
		return nil
	}
	result := r.db.WithContext(ctx).
		Model(&domain.Resource{}).
		Where("id = ?", id).
		Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update monitoring state: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

// UpdateStatus sets only the status column for a resource.
// Used by heartbeat recovery to persist the transition to 'up' without touching other fields.
func (r *ResourceRepositoryImpl) UpdateStatus(ctx context.Context, id string, status domain.ResourceStatus) error {
	result := r.db.WithContext(ctx).
		Model(&domain.Resource{}).
		Where("id = ?", id).
		Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("failed to update resource status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

// UpdateMetadata updates only the metadata fields of a resource to avoid touching associations.
// Pointer fields on the request distinguish "preserve" (nil) from "write zero" (&zero).
func (r *ResourceRepositoryImpl) UpdateMetadata(ctx context.Context, id string, req port.UpdateMetadataRequest) error {
	updates := map[string]any{}
	if req.SSLExpirationDate != nil {
		updates["ssl_expiration_date"] = *req.SSLExpirationDate
	}
	if req.SSLIssuer != nil {
		updates["ssl_issuer"] = *req.SSLIssuer
	}
	if req.DomainExpirationDate != nil {
		updates["domain_expiration_date"] = *req.DomainExpirationDate
	}
	if req.DomainRegistrar != nil {
		updates["domain_registrar"] = *req.DomainRegistrar
	}
	if len(updates) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).
		Model(&domain.Resource{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update resource metadata: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

// ListResourcesByFilter implements dynamic filtering for the v1 monitors
// list endpoint (spec 051). Uses squirrel to build the ID-only and COUNT
// queries, then re-fetches full rows via GORM with preloads so the response
// mapper still has Tags / Component available.
func (r *ResourceRepositoryImpl) ListResourcesByFilter(ctx context.Context, f dynquery.MonitorFilter, page, perPage int) ([]*domain.Resource, int, error) {
	idSQL, idArgs, err := dynquery.BuildMonitorIDsQuery(f, page, perPage, sq.Question)
	if err != nil {
		return nil, 0, fmt.Errorf("dynquery: build monitor ids: %w", err)
	}
	var ids []string
	if err := r.db.WithContext(ctx).Raw(idSQL, idArgs...).Scan(&ids).Error; err != nil {
		return nil, 0, fmt.Errorf("list monitors by filter (ids): %w", err)
	}

	countSQL, countArgs, err := dynquery.BuildMonitorCountQuery(f, sq.Question)
	if err != nil {
		return nil, 0, fmt.Errorf("dynquery: build monitor count: %w", err)
	}
	var total int64
	if err := r.db.WithContext(ctx).Raw(countSQL, countArgs...).Scan(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("list monitors by filter (count): %w", err)
	}

	if len(ids) == 0 {
		return []*domain.Resource{}, int(total), nil
	}

	var resources []*domain.Resource
	if err := r.db.WithContext(ctx).
		Preload("Tags").
		Preload("Component").
		Where("id IN ?", ids).
		Order("created_at DESC").
		Find(&resources).Error; err != nil {
		return nil, 0, fmt.Errorf("list monitors by filter (rows): %w", err)
	}
	return resources, int(total), nil
}
