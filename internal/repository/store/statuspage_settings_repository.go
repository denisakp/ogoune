package store

import (
	"context"

	"github.com/denisakp/ogoune/internal/domain"
	"gorm.io/gorm"
)

// StatusPageSettingsRepository implements the repository interface for StatusPageSettings
type StatusPageSettingsRepository struct {
	db *gorm.DB
}

// NewStatusPageSettingsRepository creates a new StatusPageSettingsRepository
func NewStatusPageSettingsRepository(db *gorm.DB) *StatusPageSettingsRepository {
	return &StatusPageSettingsRepository{db: db}
}

// Get retrieves the status page settings (there's always only one row)
func (r *StatusPageSettingsRepository) Get(ctx context.Context) (*domain.StatusPageSettings, error) {
	var settings domain.StatusPageSettings
	err := r.db.WithContext(ctx).First(&settings).Error
	if err == gorm.ErrRecordNotFound {
		// Return default settings if none exist
		return &domain.StatusPageSettings{
			Name:                 "Status Page",
			EnableDetailsPage:    true,
			ShowUptimePercentage: true,
			HidePausedMonitors:   true,
			ShowIncidentHistory:  true,
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

// Upsert creates or updates the status page settings
func (r *StatusPageSettingsRepository) Upsert(ctx context.Context, settings *domain.StatusPageSettings) error {
	// Check if settings exist
	var existing domain.StatusPageSettings
	err := r.db.WithContext(ctx).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// Create new settings
		return r.db.WithContext(ctx).Create(settings).Error
	}

	if err != nil {
		return err
	}

	// Update existing settings (keep the same ID)
	settings.ID = existing.ID
	settings.CreatedAt = existing.CreatedAt
	return r.db.WithContext(ctx).Save(settings).Error
}
