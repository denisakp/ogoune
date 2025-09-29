package repository

import (
	"context"
	"errors"

	domain "github.com/denisakp/pulseguard/internal/domain"
)

// Common repository errors (wrapped at boundaries where appropriate)
var (
	ErrNotFound     = errors.New("repository: not found")
	ErrDuplicate    = errors.New("repository: duplicate")
	ErrInvalidInput = errors.New("repository: invalid input")
)

// MonitorFilter captures future filtering/pagination (minimal for now)
type MonitorFilter struct {
	Limit  int
	Offset int
}

// MonitorStore defines persistence operations for monitors
type MonitorStore interface {
	Create(ctx context.Context, m *domain.Resource) error
	GetByID(ctx context.Context, id string) (*domain.Resource, error)
	List(ctx context.Context, f MonitorFilter) ([]*domain.Resource, error)
	Update(ctx context.Context, m *domain.Resource) error
	Delete(ctx context.Context, id string) error
}

// CheckResult encapsulates result data for a resource check (aliasing existing incident/result model not yet added)
// NOTE: For now we use Incident/IncidentEventStep for events; dedicated CheckResult model may be introduced later.
// Placeholder structure (kept internal to repository layer until solidified) could be added in future feature.

// CheckResultStore defines operations for recording and retrieving recent check state
type CheckResultStore interface {
	// Record persists a new result for a resource
	Record(ctx context.Context, incident *domain.Incident) error
	// LatestByResource returns the most recent N incidents/results for a resource
	LatestByResource(ctx context.Context, resourceID string, limit int) ([]*domain.Incident, error)
	// LatestStatus returns the most recent incident/result (if any)
	LatestStatus(ctx context.Context, resourceID string) (*domain.Incident, error)
}
