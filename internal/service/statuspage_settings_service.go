package service

import (
	"context"
	"strings"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
)

// StatusPageSettingsService handles status page settings logic
type StatusPageSettingsService struct {
	repo port.StatusPageSettingsRepository
}

// NewStatusPageSettingsService creates a new service instance
func NewStatusPageSettingsService(repo port.StatusPageSettingsRepository) *StatusPageSettingsService {
	return &StatusPageSettingsService{repo: repo}
}

// GetSettings retrieves the current status page settings
func (s *StatusPageSettingsService) GetSettings(ctx context.Context) (*domain.StatusPageSettings, error) {
	return s.repo.Get(ctx)
}

// UpdateSettings updates the status page settings
func (s *StatusPageSettingsService) UpdateSettings(ctx context.Context, settings *domain.StatusPageSettings) error {
	// Trim whitespace from string fields
	settings.Name = strings.TrimSpace(settings.Name)
	settings.HomepageURL = strings.TrimSpace(settings.HomepageURL)
	settings.CustomDomain = strings.TrimSpace(settings.CustomDomain)
	settings.GoogleAnalyticsID = strings.TrimSpace(settings.GoogleAnalyticsID)

	// Set default name if empty
	if settings.Name == "" {
		settings.Name = "Status Page"
	}

	return s.repo.Upsert(ctx, settings)
}
