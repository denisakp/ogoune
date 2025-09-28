package postgres

import (
    "context"
    "fmt"

    domain "github.com/denisakp/pulseguard/internal/domain"
    "github.com/denisakp/pulseguard/internal/repository"
    dbpkg "github.com/denisakp/pulseguard/internal/repository/postgres/database"
    "gorm.io/gorm"
)

// CheckResultRepository implements repository.CheckResultStore backed by Postgres
type CheckResultRepository struct {
    db *gorm.DB
}

// NewCheckResultRepository constructs the repository.
func NewCheckResultRepository() (*CheckResultRepository, error) {
    db, err := dbpkg.Instance()
    if err != nil {
        return nil, err
    }
    return &CheckResultRepository{db: db}, nil
}

func (r *CheckResultRepository) Record(ctx context.Context, incident *domain.Incident) error {
    if incident == nil {
        return fmt.Errorf("record result: %w", repository.ErrInvalidInput)
    }
    // TODO: implement insert logic
    return fmt.Errorf("record result: not implemented")
}

func (r *CheckResultRepository) LatestByResource(ctx context.Context, resourceID string, limit int) ([]*domain.Incident, error) {
    // TODO: implement select logic
    return nil, fmt.Errorf("latest by resource: not implemented")
}

func (r *CheckResultRepository) LatestStatus(ctx context.Context, resourceID string) (*domain.Incident, error) {
    // TODO: implement latest status logic
    return nil, fmt.Errorf("latest status: not implemented")
}
