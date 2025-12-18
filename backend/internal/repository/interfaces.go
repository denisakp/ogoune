package repository

import (
	"context"
	"errors"

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
}
