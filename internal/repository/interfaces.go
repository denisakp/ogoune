package repository

import (
	"context"

	"github.com/denisakp/pulseguard/internal/domain"
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
	FindByName(ctx context.Context, name string) (*domain.Tags, error)
	List(ctx context.Context, limit, offset int) ([]*domain.Tags, error)
	Update(ctx context.Context, t *domain.Tags) error
	Delete(ctx context.Context, id string) error
}

// ResourceRepository manages monitored resources
type ResourceRepository interface {
	Create(ctx context.Context, r *domain.Resource) error
	FindByID(ctx context.Context, id string) (*domain.Resource, error)
	List(ctx context.Context, limit, offset int) ([]*domain.Resource, error)
	Update(ctx context.Context, r *domain.Resource) error
	Delete(ctx context.Context, id string) error // soft delete (active=false)
	FindActive(ctx context.Context, limit, offset int) ([]*domain.Resource, error)
	FindByTag(ctx context.Context, tagName string, limit, offset int) ([]*domain.Resource, error)
}

// IncidentRepository manages incidents (unresolved vs resolved)
type IncidentRepository interface {
	Create(ctx context.Context, i *domain.Incident) error
	FindByID(ctx context.Context, id string) (*domain.Incident, error)
	List(ctx context.Context, limit, offset int) ([]*domain.Incident, error)
	Update(ctx context.Context, i *domain.Incident) error
	Delete(ctx context.Context, id string) error
	FindUnresolved(ctx context.Context, limit, offset int) ([]*domain.Incident, error)
	FindByResource(ctx context.Context, resourceID string, limit, offset int) ([]*domain.Incident, error)
}

// IncidentEventStepRepository manages lifecycle steps
type IncidentEventStepRepository interface {
	Create(ctx context.Context, s *domain.IncidentEventStep) error
	FindByID(ctx context.Context, id string) (*domain.IncidentEventStep, error)
	List(ctx context.Context, limit, offset int) ([]*domain.IncidentEventStep, error)
	Update(ctx context.Context, s *domain.IncidentEventStep) error
	Delete(ctx context.Context, id string) error
}

// IntegrationRepository for outbound integrations
type IntegrationRepository interface {
	Create(ctx context.Context, m *domain.Integration) error
	FindByID(ctx context.Context, id string) (*domain.Integration, error)
	List(ctx context.Context, limit, offset int) ([]*domain.Integration, error)
	Update(ctx context.Context, m *domain.Integration) error
	Delete(ctx context.Context, id string) error
	FindActiveByType(ctx context.Context, t domain.IntegrationType, limit, offset int) ([]*domain.Integration, error)
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
