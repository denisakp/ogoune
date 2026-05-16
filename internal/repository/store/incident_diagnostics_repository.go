package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository"
	"gorm.io/gorm"
)

// IncidentDiagnosticsRepositoryImpl provides GORM-based implementation of IncidentDiagnosticsRepository
type IncidentDiagnosticsRepositoryImpl struct {
	db *gorm.DB
}

// NewIncidentDiagnosticsRepository creates a new IncidentDiagnosticsRepository using GORM
func NewIncidentDiagnosticsRepository(db *gorm.DB) port.IncidentDiagnosticsRepository {
	return &IncidentDiagnosticsRepositoryImpl{db: db}
}

// Create persists a new incident diagnostics record to the database.
// Returns the created diagnostics with its generated ID, or an error if creation fails.
func (r *IncidentDiagnosticsRepositoryImpl) Create(ctx context.Context, d *domain.IncidentDiagnostics) (*domain.IncidentDiagnostics, error) {
	if d == nil {
		return nil, fmt.Errorf("incident diagnostics cannot be nil")
	}

	if err := r.db.WithContext(ctx).Create(d).Error; err != nil {
		return nil, fmt.Errorf("failed to create incident diagnostics: %w", err)
	}
	return d, nil
}

// FindByIncidentID retrieves diagnostics by incident ID.
// Returns ErrNotFound if no diagnostics exist for the incident.
func (r *IncidentDiagnosticsRepositoryImpl) FindByIncidentID(ctx context.Context, incidentID string) (*domain.IncidentDiagnostics, error) {
	var diagnostics domain.IncidentDiagnostics
	err := r.db.WithContext(ctx).
		Where("incident_id = ?", incidentID).
		First(&diagnostics).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find diagnostics by incident ID: %w", err)
	}
	return &diagnostics, nil
}

// Update modifies an existing incident diagnostics record in the database.
// Returns ErrNotFound if the diagnostics record doesn't exist.
func (r *IncidentDiagnosticsRepositoryImpl) Update(ctx context.Context, d *domain.IncidentDiagnostics) error {
	if d == nil {
		return fmt.Errorf("incident diagnostics cannot be nil")
	}

	result := r.db.WithContext(ctx).Save(d)
	if result.Error != nil {
		return fmt.Errorf("failed to update incident diagnostics: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

// Delete removes an incident diagnostics record from the database by its ID.
// Returns ErrNotFound if the diagnostics record doesn't exist.
func (r *IncidentDiagnosticsRepositoryImpl) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&domain.IncidentDiagnostics{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete incident diagnostics: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}
