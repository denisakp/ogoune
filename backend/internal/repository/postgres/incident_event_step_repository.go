package postgres

import (
	"context"
	"errors"
	"fmt"

	domain "github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/repository"
	"gorm.io/gorm"
)

type IncidentEventStepRepositoryImpl struct {
	db *gorm.DB
}

// NewIncidentEventStepRepository creates a new IncidentEventStepRepository using GORM
func NewIncidentEventStepRepository(db *gorm.DB) repository.IncidentEventStepRepository {
	return &IncidentEventStepRepositoryImpl{db: db}
}

// Create persists a new incident event step record to the database.
func (r *IncidentEventStepRepositoryImpl) Create(ctx context.Context, s *domain.IncidentEventStep) (*domain.IncidentEventStep, error) {
	if err := r.db.WithContext(ctx).Create(s).Error; err != nil {
		return nil, fmt.Errorf("failed to create incident event step: %w", err)
	}
	return s, nil
}

// FindByID retrieves an incident event step by its ID.
func (r *IncidentEventStepRepositoryImpl) FindByID(ctx context.Context, id string) (*domain.IncidentEventStep, error) {
	var step domain.IncidentEventStep
	err := r.db.WithContext(ctx).Preload("Incident").First(&step, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find incident event step by ID: %w", err)
	}
	return &step, nil
}

// FindLastByIncidentAndStep retrieves the most recent incident event step for an incident and step type.
func (r *IncidentEventStepRepositoryImpl) FindLastByIncidentAndStep(ctx context.Context, incidentID string, stepType domain.IncidentEventStepType) (*domain.IncidentEventStep, error) {
	var step domain.IncidentEventStep
	err := r.db.WithContext(ctx).
		Where("incident_id = ? AND step = ?", incidentID, stepType).
		Order("created_at DESC").
		Limit(1).
		First(&step).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find latest incident event step: %w", err)
	}
	return &step, nil
}

// List retrieves all incident event steps with pagination, ordered by creation time descending.
func (r *IncidentEventStepRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*domain.IncidentEventStep, error) {
	var steps []*domain.IncidentEventStep
	err := r.db.WithContext(ctx).
		Preload("Incident").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&steps).Error

	if err != nil {
		return nil, fmt.Errorf("failed to list incident event steps: %w", err)
	}
	return steps, nil
}

// Update modifies an existing incident event step record in the database.
func (r *IncidentEventStepRepositoryImpl) Update(ctx context.Context, s *domain.IncidentEventStep) error {
	result := r.db.WithContext(ctx).Save(s)
	if result.Error != nil {
		return fmt.Errorf("failed to update incident event step: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

// Delete removes an incident event step record from the database by its ID.
func (r *IncidentEventStepRepositoryImpl) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&domain.IncidentEventStep{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete incident event step: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}
