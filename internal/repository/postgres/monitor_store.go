package postgres

import (
    "context"
    "fmt"

    domain "github.com/denisakp/pulseguard/internal/domain"
    "github.com/denisakp/pulseguard/internal/repository"
    dbpkg "github.com/denisakp/pulseguard/internal/repository/postgres/database"
    "gorm.io/gorm"
)

// MonitorRepository implements repository.MonitorStore backed by Postgres (GORM)
type MonitorRepository struct {
    db *gorm.DB
}

// NewMonitorRepository returns a new monitor repository instance.
func NewMonitorRepository() (*MonitorRepository, error) {
    db, err := dbpkg.Instance()
    if err != nil {
        return nil, err
    }
    return &MonitorRepository{db: db}, nil
}

func (r *MonitorRepository) Create(ctx context.Context, m *domain.Resource) error {
    if m == nil {
        return fmt.Errorf("create monitor: %w", repository.ErrInvalidInput)
    }
    // TODO: implement persistence logic
    return fmt.Errorf("create monitor: not implemented")
}

func (r *MonitorRepository) GetByID(ctx context.Context, id string) (*domain.Resource, error) {
    // TODO: implement fetch logic
    return nil, fmt.Errorf("get monitor: not implemented")
}

func (r *MonitorRepository) List(ctx context.Context, f repository.MonitorFilter) ([]*domain.Resource, error) {
    // TODO: implement list logic
    return nil, fmt.Errorf("list monitors: not implemented")
}

func (r *MonitorRepository) Update(ctx context.Context, m *domain.Resource) error {
    // TODO: implement update logic
    return fmt.Errorf("update monitor: not implemented")
}

func (r *MonitorRepository) Delete(ctx context.Context, id string) error {
    // TODO: implement delete logic
    return fmt.Errorf("delete monitor: not implemented")
}
