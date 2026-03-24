package repository

import (
	"context"
	"errors"
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
)

// Common repository errors
var (
	ErrNotFound     = errors.New("repository: not found")
	ErrDuplicate    = errors.New("repository: duplicate")
	ErrInvalidInput = errors.New("repository: invalid input")
)

// PaginationParams holds common pagination parameters
type PaginationParams struct {
	Limit  int
	Offset int
}

// TagsRepository manages tag lifecycle
type TagsRepository interface {
	Create(ctx context.Context, t *domain.Tags) error
	FindByID(ctx context.Context, id string) (*domain.Tags, error)
	FindByIDs(ctx context.Context, ids []string) ([]*domain.Tags, error)
	FindByName(ctx context.Context, name string) (*domain.Tags, error)
	List(ctx context.Context, limit, offset int) ([]*domain.Tags, error)
	Update(ctx context.Context, t *domain.Tags) error
	Delete(ctx context.Context, id string) error
}

// ResourceRepository manages monitored resources
type ResourceRepository interface {
	Create(ctx context.Context, r *domain.Resource) (*domain.Resource, error)
	FindByID(ctx context.Context, id string) (*domain.Resource, error)
	List(ctx context.Context, limit, offset int) ([]*domain.Resource, error)
	Update(ctx context.Context, r *domain.Resource) error
	Delete(ctx context.Context, id string) error // soft delete (active=false)
	FindActive(ctx context.Context, limit, offset int) ([]*domain.Resource, error)
	FindByTag(ctx context.Context, tagName string, limit, offset int) ([]*domain.Resource, error)
	FindByComponentID(ctx context.Context, componentID string) ([]*domain.Resource, error)
	CountByComponentID(ctx context.Context, componentID string) (int64, error)
	// UpdateMetadata updates only the metadata fields for a resource, leaving associations intact
	UpdateMetadata(ctx context.Context, id string, metadata *domain.ResourceMetaData) error
	// FindScheduledResources returns all active resources with a non-nil schedule (used by scheduler startup)
	FindScheduledResources(ctx context.Context) ([]*domain.Resource, error)
}

// ComponentRepository manages logical component groups
type ComponentRepository interface {
	Create(ctx context.Context, c *domain.Component) (*domain.Component, error)
	FindByID(ctx context.Context, id string) (*domain.Component, error)
	List(ctx context.Context, limit, offset int) ([]*domain.Component, error)
	Update(ctx context.Context, c *domain.Component) error
	Delete(ctx context.Context, id string) error
	UpdateLastNotificationStatus(ctx context.Context, id string, status domain.ComponentStatus) error
}

// IncidentRepository manages incidents (unresolved vs resolved)
type IncidentRepository interface {
	Create(ctx context.Context, i *domain.Incident) (*domain.Incident, error)
	FindByID(ctx context.Context, id string) (*domain.Incident, error)
	List(ctx context.Context, limit, offset int) ([]*domain.Incident, error)
	Update(ctx context.Context, i *domain.Incident) error
	Delete(ctx context.Context, id string) error
	FindUnresolved(ctx context.Context, limit, offset int) ([]*domain.Incident, error)
	FindByResource(ctx context.Context, resourceID string, limit, offset int) ([]*domain.Incident, error)
	GetIncidentStats(ctx context.Context, hours int) (int, int, error)
}

// IncidentEventStepRepository manages lifecycle steps
type IncidentEventStepRepository interface {
	Create(ctx context.Context, s *domain.IncidentEventStep) (*domain.IncidentEventStep, error)
	FindByID(ctx context.Context, id string) (*domain.IncidentEventStep, error)
	List(ctx context.Context, limit, offset int) ([]*domain.IncidentEventStep, error)
	Update(ctx context.Context, s *domain.IncidentEventStep) error
	Delete(ctx context.Context, id string) error
}

// NotificationRepository handles notification events
type NotificationRepository interface {
	Create(ctx context.Context, n *domain.NotificationEvent) error
	FindByID(ctx context.Context, id string) (*domain.NotificationEvent, error)
	List(ctx context.Context, limit, offset int) ([]*domain.NotificationEvent, error)
	Update(ctx context.Context, n *domain.NotificationEvent) error
	Delete(ctx context.Context, id string) error
	FindPending(ctx context.Context, limit, offset int) ([]*domain.NotificationEvent, error)
	MarkAsSent(ctx context.Context, id string) error
}

// MonitoringActivityRepository manages monitoring activity records
type MonitoringActivityRepository interface {
	Create(ctx context.Context, activity *domain.MonitoringActivity) error
	List(ctx context.Context, limit, offset int) ([]*domain.MonitoringActivity, error)
	FindByResourceID(ctx context.Context, resourceID string, limit, offset int) ([]*domain.MonitoringActivity, error)
	GetUptimeStats(ctx context.Context, resourceID string) ([]domain.UptimeStat, error)
	GetRecentResponseTimes(ctx context.Context, resourceID string, limit int) ([]domain.ResponseTimePoint, error)
	GetGlobalUptimeStats(ctx context.Context, hours int) (float64, error)
}

// Scheduler defines the interface for scheduling monitoring tasks
type Scheduler interface {
	Schedule(ctx context.Context, r *domain.Resource) error
	Unschedule(ctx context.Context, resourceID string) error
}

// NotificationChannelRepository manages notification channels
type NotificationChannelRepository interface {
	Create(ctx context.Context, channel *domain.NotificationChannel) error
	FindByID(ctx context.Context, id string) (*domain.NotificationChannel, error)
	List(ctx context.Context, limit, offset int) ([]*domain.NotificationChannel, error)
	Update(ctx context.Context, channel *domain.NotificationChannel) error
	Delete(ctx context.Context, id string) error
	FindByType(ctx context.Context, channelType domain.NotificationChannelType) ([]*domain.NotificationChannel, error)
	FindDefaultChannels(ctx context.Context) ([]*domain.NotificationChannel, error)
	// FindByResourceID returns all notification channels associated with a resource
	FindByResourceID(ctx context.Context, resourceID string) ([]*domain.NotificationChannel, error)
	// FindByComponentID returns all notification channels associated with a component
	FindByComponentID(ctx context.Context, componentID string) ([]*domain.NotificationChannel, error)
}

// MaintenanceRepository manages maintenance windows
type MaintenanceRepository interface {
	Create(ctx context.Context, m *domain.Maintenance) (*domain.Maintenance, error)
	FindByID(ctx context.Context, id string) (*domain.Maintenance, error)
	List(ctx context.Context, status string, limit, offset int) ([]*domain.Maintenance, error)
	Update(ctx context.Context, m *domain.Maintenance) error
	Delete(ctx context.Context, id string) error
	// Returns maintenances currently active for a resource at the provided time
	FindActiveForResource(ctx context.Context, resourceID string, now time.Time) ([]*domain.Maintenance, error)
}

// StatusPageSettingsRepository manages status page configuration
type StatusPageSettingsRepository interface {
	Get(ctx context.Context) (*domain.StatusPageSettings, error)
	Upsert(ctx context.Context, settings *domain.StatusPageSettings) error
}

// UserRepository manages user accounts and authentication
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) (*domain.User, error)
	FindByID(ctx context.Context, id string) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error
	UpdatePassword(ctx context.Context, userID string, hashedPassword string) error
	UpdateLastLogin(ctx context.Context, userID string) error
	UpdateTwoFactorSecret(ctx context.Context, userID string, secret string, enabled bool) error
}

// IncidentDiagnosticsRepository manages detailed diagnostic information for incidents
type IncidentDiagnosticsRepository interface {
	Create(ctx context.Context, d *domain.IncidentDiagnostics) (*domain.IncidentDiagnostics, error)
	FindByIncidentID(ctx context.Context, incidentID string) (*domain.IncidentDiagnostics, error)
	Update(ctx context.Context, d *domain.IncidentDiagnostics) error
	Delete(ctx context.Context, id string) error
}

// ExpiryNotificationLogRepository manages deduplication records for expiry alerts.
type ExpiryNotificationLogRepository interface {
	// CountByKey returns how many log entries exist for the given resource/type/threshold combination.
	CountByKey(ctx context.Context, resourceID, expiryType string, threshold int) (int64, error)

	// Create persists a new log entry after a successful notification dispatch.
	Create(ctx context.Context, log *domain.ExpiryNotificationLog) error

	// DeleteByResourceIDAndType removes all logs for a resource+type pair (called on renewal reset).
	DeleteByResourceIDAndType(ctx context.Context, resourceID, expiryType string) error

	// DeleteOlderThan removes log entries with sent_at older than the given cutoff time.
	DeleteOlderThan(ctx context.Context, cutoff time.Time) error
}
