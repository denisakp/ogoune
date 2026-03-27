package store

import (
	"context"
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/repository"
	"gorm.io/gorm"
)

// MaintenanceRepository provides Postgres implementation for MaintenanceRepository.
type MaintenanceRepository struct {
	db *gorm.DB
}

func NewMaintenanceRepository(db *gorm.DB) *MaintenanceRepository {
	return &MaintenanceRepository{db: db}
}

func (r *MaintenanceRepository) Create(ctx context.Context, m *domain.Maintenance) (*domain.Maintenance, error) {
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return nil, err
	}
	return m, nil
}

func (r *MaintenanceRepository) FindByID(ctx context.Context, id string) (*domain.Maintenance, error) {
	var m domain.Maintenance
	if err := r.db.WithContext(ctx).Preload("Resources").First(&m, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (r *MaintenanceRepository) List(ctx context.Context, status string, limit, offset int) ([]*domain.Maintenance, error) {
	var list []*domain.Maintenance
	query := r.db.WithContext(ctx).Model(&domain.Maintenance{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Preload("Resources").Limit(limit).Offset(offset).Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *MaintenanceRepository) Update(ctx context.Context, m *domain.Maintenance) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *MaintenanceRepository) Delete(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Delete(&domain.Maintenance{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

// FindActiveForResource returns maintenances active for a resource at time 'now'.
func (r *MaintenanceRepository) FindActiveForResource(ctx context.Context, resourceID string, now time.Time) ([]*domain.Maintenance, error) {
	var list []*domain.Maintenance
	q := r.db.WithContext(ctx).
		Model(&domain.Maintenance{}).
		Joins("JOIN maintenance_resources mr ON mr.maintenance_id = maintenances.id").
		Where("mr.resource_id = ?", resourceID).
		Where("status = ?", "active").
		Preload("Resources")
	if err := q.Find(&list).Error; err != nil {
		return nil, err
	}
	// Additionally include scheduled one-time windows that are currently between StartAt/EndAt
	var scheduled []*domain.Maintenance
	q2 := r.db.WithContext(ctx).
		Model(&domain.Maintenance{}).
		Joins("JOIN maintenance_resources mr ON mr.maintenance_id = maintenances.id").
		Where("mr.resource_id = ?", resourceID).
		Where("strategy = ?", domain.OneTime).
		Where("status = ?", "scheduled").
		Where("start_at <= ? AND end_at >= ?", now, now)
	if err := q2.Find(&scheduled).Error; err == nil {
		list = append(list, scheduled...)
	}
	return list, nil
}
