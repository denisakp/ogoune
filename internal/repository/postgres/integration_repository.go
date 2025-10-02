package postgres

import (
	"context"
	"errors"
	"fmt"

	domain "github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/repository"
	"gorm.io/gorm"
)

type IntegrationRepositoryImpl struct {
	db *gorm.DB
}

// NewIntegrationRepository creates a new IntegrationRepository using GORM
func NewIntegrationRepository(db *gorm.DB) repository.IntegrationRepository {
	return &IntegrationRepositoryImpl{db: db}
}

// Create persists a new integration record to the database.
func (r *IntegrationRepositoryImpl) Create(ctx context.Context, i *domain.Integration) error {
	if err := r.db.WithContext(ctx).Create(i).Error; err != nil {
		return fmt.Errorf("failed to create integration: %w", err)
	}
	return nil
}

// FindByID retrieves an integration by its ID.
func (r *IntegrationRepositoryImpl) FindByID(ctx context.Context, id string) (*domain.Integration, error) {
	var integration domain.Integration
	err := r.db.WithContext(ctx).First(&integration, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find integration by ID: %w", err)
	}
	return &integration, nil
}

// List retrieves all integrations with pagination, ordered by creation time descending.
func (r *IntegrationRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*domain.Integration, error) {
	var integrations []*domain.Integration
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&integrations).Error

	if err != nil {
		return nil, fmt.Errorf("failed to list integrations: %w", err)
	}
	return integrations, nil
}

// Update modifies an existing integration record in the database.
func (r *IntegrationRepositoryImpl) Update(ctx context.Context, i *domain.Integration) error {
	result := r.db.WithContext(ctx).Save(i)
	if result.Error != nil {
		return fmt.Errorf("failed to update integration: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

// Delete removes an integration by its ID.
func (r *IntegrationRepositoryImpl) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&domain.Integration{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete integration: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

// FindActiveByType retrieves all active integrations of a specific type with pagination.
func (r *IntegrationRepositoryImpl) FindActiveByType(ctx context.Context, t domain.IntegrationType, limit, offset int) ([]*domain.Integration, error) {
	var integrations []*domain.Integration
	err := r.db.WithContext(ctx).
		Where("type = ? AND is_active = ?", t, true).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&integrations).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find active integrations by type: %w", err)
	}
	return integrations, nil
}
