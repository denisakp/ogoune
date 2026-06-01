package port

import (
	"context"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/sqlc/dynquery"
)

// TagsRepository manages tag lifecycle.
type TagsRepository interface {
	Create(ctx context.Context, t *domain.Tags) error
	FindByID(ctx context.Context, id string) (*domain.Tags, error)
	FindByIDs(ctx context.Context, ids []string) ([]*domain.Tags, error)
	FindByName(ctx context.Context, name string) (*domain.Tags, error)
	List(ctx context.Context, limit, offset int) ([]*domain.Tags, error)
	Update(ctx context.Context, t *domain.Tags) error
	Delete(ctx context.Context, id string) error
}

// UpdateMonitoringStateRequest carries the columns mutated by the monitoring
// worker after a check cycle. Pointer semantics: nil preserves the existing
// column value; a non-nil pointer writes it. For nullable timestamp columns
// the outer pointer is double-indirected so callers can distinguish
// "preserve" (outer nil) from "set to NULL" (outer non-nil, inner nil).
type UpdateMonitoringStateRequest struct {
	Status               *domain.ResourceStatus
	FailureCount         *int
	LastChecked          **time.Time
	LastStatusTransition **time.Time
	FlapStartedAt        **time.Time
}

// UpdateMetadataRequest carries the SSL/domain expiry fields populated by the
// metadata-enrichment path. Same nil-vs-non-nil semantics as
// UpdateMonitoringStateRequest; nullable timestamps use **time.Time.
type UpdateMetadataRequest struct {
	SSLExpirationDate    **time.Time
	SSLIssuer            *string
	DomainExpirationDate **time.Time
	DomainRegistrar      *string
}

// ResourceRepository manages monitored resources.
type ResourceRepository interface {
	Create(ctx context.Context, r *domain.Resource) (*domain.Resource, error)
	FindByID(ctx context.Context, id string) (*domain.Resource, error)
	FindByHeartbeatSlug(ctx context.Context, slug string) (*domain.Resource, error)
	List(ctx context.Context, limit, offset int) ([]*domain.Resource, error)
	Update(ctx context.Context, r *domain.Resource) error
	Delete(ctx context.Context, id string) error
	FindActive(ctx context.Context, limit, offset int) ([]*domain.Resource, error)
	FindByTag(ctx context.Context, tagName string, limit, offset int) ([]*domain.Resource, error)
	FindByComponentID(ctx context.Context, componentID string) ([]*domain.Resource, error)
	CountByComponentID(ctx context.Context, componentID string) (int64, error)
	FindMissedHeartbeats(ctx context.Context, now time.Time, limit int) ([]*domain.Resource, error)
	UpdateLastPingAt(ctx context.Context, id string, at time.Time) error
	UpdateStatus(ctx context.Context, id string, status domain.ResourceStatus) error
	UpdateMonitoringState(ctx context.Context, id string, req UpdateMonitoringStateRequest) error
	UpdateMetadata(ctx context.Context, id string, req UpdateMetadataRequest) error
	FindScheduledResources(ctx context.Context) ([]*domain.Resource, error)
	ListResourcesByFilter(ctx context.Context, f dynquery.MonitorFilter, page, perPage int) ([]*domain.Resource, int, error)
}

// ComponentRepository manages logical component groups.
type ComponentRepository interface {
	Create(ctx context.Context, c *domain.Component) (*domain.Component, error)
	FindByID(ctx context.Context, id string) (*domain.Component, error)
	List(ctx context.Context, limit, offset int) ([]*domain.Component, error)
	Update(ctx context.Context, c *domain.Component) error
	Delete(ctx context.Context, id string) error
	UpdateLastNotificationStatus(ctx context.Context, id string, status domain.ComponentStatus) error
}

// IncidentRepository manages incidents (unresolved vs resolved).
type IncidentRepository interface {
	Create(ctx context.Context, i *domain.Incident) (*domain.Incident, error)
	FindByID(ctx context.Context, id string) (*domain.Incident, error)
	List(ctx context.Context, limit, offset int) ([]*domain.Incident, error)
	Update(ctx context.Context, i *domain.Incident) error
	Delete(ctx context.Context, id string) error
	FindUnresolved(ctx context.Context, limit, offset int) ([]*domain.Incident, error)
	FindByResource(ctx context.Context, resourceID string, limit, offset int) ([]*domain.Incident, error)
	GetIncidentStats(ctx context.Context, hours int) (int, int, error)
	FindActiveByResourceID(ctx context.Context, resourceID string) (*domain.Incident, error)
	HasActiveIncident(ctx context.Context) (bool, error)
	FindLastResolved(ctx context.Context) (*domain.Incident, error)
	CountByResourceID(ctx context.Context, resourceID string) (int64, error)
	ListIncidentsByFilter(ctx context.Context, f dynquery.IncidentFilter, page, perPage int) ([]*domain.Incident, int, error)
}

// IncidentEventStepRepository manages lifecycle steps.
type IncidentEventStepRepository interface {
	Create(ctx context.Context, s *domain.IncidentEventStep) (*domain.IncidentEventStep, error)
	FindByID(ctx context.Context, id string) (*domain.IncidentEventStep, error)
	FindLastByIncidentAndStep(ctx context.Context, incidentID string, step domain.IncidentEventStepType) (*domain.IncidentEventStep, error)
	List(ctx context.Context, limit, offset int) ([]*domain.IncidentEventStep, error)
	Update(ctx context.Context, s *domain.IncidentEventStep) error
	Delete(ctx context.Context, id string) error
}

// NotificationRepository handles notification events.
type NotificationRepository interface {
	Create(ctx context.Context, n *domain.NotificationEvent) error
	FindByID(ctx context.Context, id string) (*domain.NotificationEvent, error)
	List(ctx context.Context, limit, offset int) ([]*domain.NotificationEvent, error)
	Update(ctx context.Context, n *domain.NotificationEvent) error
	Delete(ctx context.Context, id string) error
	FindPending(ctx context.Context, limit, offset int) ([]*domain.NotificationEvent, error)
	ClaimPending(ctx context.Context, id, claimOwner string, claimedAt time.Time) (bool, error)
	MarkAsSent(ctx context.Context, id string, processedAt time.Time) error
	MarkAsFailed(ctx context.Context, id, lastError string, processedAt time.Time) error
	MarkAsExpired(ctx context.Context, id, lastError string, processedAt time.Time) error
}

// MonitoringActivityRepository manages monitoring activity records.
type MonitoringActivityRepository interface {
	Create(ctx context.Context, activity *domain.MonitoringActivity) error
	List(ctx context.Context, limit, offset int) ([]*domain.MonitoringActivity, error)
	FindByResourceID(ctx context.Context, resourceID string, limit, offset int) ([]*domain.MonitoringActivity, error)
	CountTransitionsInWindow(ctx context.Context, resourceID string, windowStart time.Time) (int, error)
	GetUptimeStats(ctx context.Context, resourceID string) ([]domain.UptimeStat, error)
	GetRecentResponseTimes(ctx context.Context, resourceID string, limit int) ([]domain.ResponseTimePoint, error)
	GetGlobalUptimeStats(ctx context.Context, hours int) (float64, error)
	GetUptimeByWindow(ctx context.Context, resourceID string, hours int) (*float64, error)
	GetAvgResponseTimeByWindow(ctx context.Context, resourceID string, hours int) (*int, error)
}

// ResourceScheduler defines the interface for scheduling monitoring tasks
// at the service layer (schedule/unschedule with domain.Resource).
// Named ResourceScheduler to avoid collision with the full Scheduler interface.
type ResourceScheduler interface {
	Schedule(ctx context.Context, r *domain.Resource) error
	Unschedule(ctx context.Context, resourceID string) error
}

// NotificationChannelRepository manages notification channels.
// ResourceCredentialRepository manages optional auth credentials for protocol-aware resources.
type ResourceCredentialRepository interface {
	Get(ctx context.Context, resourceID string) (*domain.ResourceCredential, error)
	Upsert(ctx context.Context, cred *domain.ResourceCredential) error
	Delete(ctx context.Context, resourceID string) error
	Exists(ctx context.Context, resourceID string) (bool, error)
}

type NotificationChannelRepository interface {
	Create(ctx context.Context, channel *domain.NotificationChannel) error
	FindByID(ctx context.Context, id string) (*domain.NotificationChannel, error)
	List(ctx context.Context, limit, offset int) ([]*domain.NotificationChannel, error)
	Update(ctx context.Context, channel *domain.NotificationChannel) error
	Delete(ctx context.Context, id string) error
	FindByType(ctx context.Context, channelType domain.NotificationChannelType) ([]*domain.NotificationChannel, error)
	FindDefaultChannels(ctx context.Context) ([]*domain.NotificationChannel, error)
	FindByResourceID(ctx context.Context, resourceID string) ([]*domain.NotificationChannel, error)
	FindByComponentID(ctx context.Context, componentID string) ([]*domain.NotificationChannel, error)
}

// MaintenanceRepository manages maintenance windows.
type MaintenanceRepository interface {
	Create(ctx context.Context, m *domain.Maintenance) (*domain.Maintenance, error)
	FindByID(ctx context.Context, id string) (*domain.Maintenance, error)
	List(ctx context.Context, status string, limit, offset int) ([]*domain.Maintenance, error)
	Update(ctx context.Context, m *domain.Maintenance) error
	Delete(ctx context.Context, id string) error
	FindActiveForResource(ctx context.Context, resourceID string, now time.Time) ([]*domain.Maintenance, error)
}

// StatusPageSettingsRepository manages status page configuration.
type StatusPageSettingsRepository interface {
	Get(ctx context.Context) (*domain.StatusPageSettings, error)
	Upsert(ctx context.Context, settings *domain.StatusPageSettings) error
}

// UserRepository manages user accounts and authentication.
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

// APIKeyRepository manages API key persistence and lookup.
type APIKeyRepository interface {
	Create(ctx context.Context, key *domain.APIKey) error
	FindByID(ctx context.Context, id, userID string) (*domain.APIKey, error)
	FindByKeyHash(ctx context.Context, keyHash string) (*domain.APIKey, error)
	ListByUserID(ctx context.Context, userID string) ([]domain.APIKey, error)
	UpdateLastUsed(ctx context.Context, id string, at time.Time, ip string) error
	Revoke(ctx context.Context, id, userID string) error
	CountByUserID(ctx context.Context, userID string) (int64, error)
}

// IncidentDiagnosticsRepository manages detailed diagnostic information for incidents.
type IncidentDiagnosticsRepository interface {
	Create(ctx context.Context, d *domain.IncidentDiagnostics) (*domain.IncidentDiagnostics, error)
	FindByIncidentID(ctx context.Context, incidentID string) (*domain.IncidentDiagnostics, error)
	Update(ctx context.Context, d *domain.IncidentDiagnostics) error
	Delete(ctx context.Context, id string) error
}

// ExpiryNotificationLogRepository manages deduplication records for expiry alerts.
type ExpiryNotificationLogRepository interface {
	CountByKey(ctx context.Context, resourceID, expiryType string, threshold int) (int64, error)
	Create(ctx context.Context, log *domain.ExpiryNotificationLog) error
	DeleteByResourceIDAndType(ctx context.Context, resourceID, expiryType string) error
	DeleteOlderThan(ctx context.Context, cutoff time.Time) error
}
