package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/repository"
	"gorm.io/gorm"
)

// IncidentRepositoryImpl provides GORM-based implementation of IncidentRepository
type IncidentRepositoryImpl struct {
	db *gorm.DB
}

// NewIncidentRepository creates a new IncidentRepository using GORM
func NewIncidentRepository(db *gorm.DB) repository.IncidentRepository {
	return &IncidentRepositoryImpl{db: db}
}

func (r *IncidentRepositoryImpl) Create(ctx context.Context, incident *domain.Incident) error {
	if err := r.db.WithContext(ctx).Create(incident).Error; err != nil {
		return fmt.Errorf("failed to create incident: %w", err)
	}
	return nil
}

func (r *IncidentRepositoryImpl) FindByID(ctx context.Context, id string) (*domain.Incident, error) {
	var incident domain.Incident
	err := r.db.WithContext(ctx).Preload("Resource").First(&incident, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find incident by ID: %w", err)
	}
	return &incident, nil
}

func (r *IncidentRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*domain.Incident, error) {
	var incidents []*domain.Incident
	err := r.db.WithContext(ctx).
		Preload("Resource").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&incidents).Error

	if err != nil {
		return nil, fmt.Errorf("failed to list incidents: %w", err)
	}
	return incidents, nil
}

func (r *IncidentRepositoryImpl) Update(ctx context.Context, incident *domain.Incident) error {
	result := r.db.WithContext(ctx).Save(incident)
	if result.Error != nil {
		return fmt.Errorf("failed to update incident: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *IncidentRepositoryImpl) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&domain.Incident{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete incident: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *IncidentRepositoryImpl) FindUnresolved(ctx context.Context, limit, offset int) ([]*domain.Incident, error) {
	var incidents []*domain.Incident
	err := r.db.WithContext(ctx).
		Preload("Resource").
		Where("is_resolved = ?", false).
		Order("started_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&incidents).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find unresolved incidents: %w", err)
	}
	return incidents, nil
}

func (r *IncidentRepositoryImpl) FindByResource(ctx context.Context, resourceID string, limit, offset int) ([]*domain.Incident, error) {
	var incidents []*domain.Incident
	err := r.db.WithContext(ctx).
		Preload("Resource").
		Where("resource_id = ?", resourceID).
		Order("started_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&incidents).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find incidents by resource: %w", err)
	}
	return incidents, nil
}
