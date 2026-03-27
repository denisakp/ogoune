package store

import (
	"context"
	"errors"
	"fmt"
	"time"

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

// Create persists a new incident record to the database.
func (r *IncidentRepositoryImpl) Create(ctx context.Context, incident *domain.Incident) (*domain.Incident, error) {
	if err := r.db.WithContext(ctx).Create(incident).Error; err != nil {
		return nil, fmt.Errorf("failed to create incident: %w", err)
	}
	return incident, nil
}

// FindByID retrieves an incident by its ID.
func (r *IncidentRepositoryImpl) FindByID(ctx context.Context, id string) (*domain.Incident, error) {
	var incident domain.Incident
	err := r.db.WithContext(ctx).
		Preload("Resource").
		Preload("IncidentDiagnostics").
		First(&incident, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find incident by ID: %w", err)
	}
	return &incident, nil
}

// List retrieves all incidents with pagination, ordered by creation time descending.
func (r *IncidentRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*domain.Incident, error) {
	var incidents []*domain.Incident
	err := r.db.WithContext(ctx).
		Preload("Resource").
		Preload("IncidentDiagnostics").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&incidents).Error

	if err != nil {
		return nil, fmt.Errorf("failed to list incidents: %w", err)
	}
	return incidents, nil
}

// Update modifies an existing incident record in the database.
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

// Delete removes an incident record from the database by its ID.
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

// FindUnresolved retrieves all unresolved incidents with pagination, ordered by start time descending.
func (r *IncidentRepositoryImpl) FindUnresolved(ctx context.Context, limit, offset int) ([]*domain.Incident, error) {
	var incidents []*domain.Incident
	err := r.db.WithContext(ctx).
		Preload("Resource").
		Preload("IncidentDiagnostics").
		Where("resolved_at", nil).
		Order("started_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&incidents).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find unresolved incidents: %w", err)
	}
	return incidents, nil
}

// FindByResource retrieves all incidents for a specific resource with pagination, ordered by start time descending.
func (r *IncidentRepositoryImpl) FindByResource(ctx context.Context, resourceID string, limit, offset int) ([]*domain.Incident, error) {
	var incidents []*domain.Incident
	err := r.db.WithContext(ctx).
		Preload("Resource").
		Preload("IncidentDiagnostics").
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

// FindActiveByResourceID returns the most recent unresolved incident for a resource.
// Returns repository.ErrNotFound when no active incident exists.
func (r *IncidentRepositoryImpl) FindActiveByResourceID(ctx context.Context, resourceID string) (*domain.Incident, error) {
	var incident domain.Incident
	err := r.db.WithContext(ctx).
		Where("resource_id = ? AND resolved_at IS NULL", resourceID).
		Order("started_at DESC").
		Limit(1).
		First(&incident).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find active incident for resource %s: %w", resourceID, err)
	}
	return &incident, nil
}

// GetIncidentStats retrieves incident statistics for a given time range.
// Returns: total incidents count, affected monitors count, error
func (r *IncidentRepositoryImpl) GetIncidentStats(ctx context.Context, hours int) (int, int, error) {
	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	var totalIncidents int64
	err := r.db.WithContext(ctx).
		Model(&domain.Incident{}).
		Where("started_at >= ?", since).
		Count(&totalIncidents).Error
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get incident stats: %w", err)
	}

	var affectedMonitors int64
	err = r.db.WithContext(ctx).
		Model(&domain.Incident{}).
		Where("started_at >= ?", since).
		Distinct("resource_id").
		Count(&affectedMonitors).Error
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get incident stats: %w", err)
	}

	return int(totalIncidents), int(affectedMonitors), nil
}
