package fake

import (
	"context"
	"sync"

	"github.com/denisakp/pulseguard/internal/domain"
)

// StatusPageSettingsFake is an in-memory implementation for testing
type StatusPageSettingsFake struct {
	mu       sync.RWMutex
	settings *domain.StatusPageSettings
}

// NewStatusPageSettingsFake creates a new fake repository with default settings
func NewStatusPageSettingsFake() *StatusPageSettingsFake {
	return &StatusPageSettingsFake{
		settings: &domain.StatusPageSettings{
			Name:                 "Status Page",
			EnableDetailsPage:    true,
			ShowUptimePercentage: true,
			HidePausedMonitors:   true,
			ShowIncidentHistory:  true,
		},
	}
}

// Get returns the current settings
func (f *StatusPageSettingsFake) Get(ctx context.Context) (*domain.StatusPageSettings, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// Return a copy
	result := *f.settings
	return &result, nil
}

// Upsert creates or updates settings
func (f *StatusPageSettingsFake) Upsert(ctx context.Context, settings *domain.StatusPageSettings) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Keep the ID if it exists
	if f.settings != nil && f.settings.ID != "" {
		settings.ID = f.settings.ID
		settings.CreatedAt = f.settings.CreatedAt
	}

	// Store a copy
	f.settings = &domain.StatusPageSettings{}
	*f.settings = *settings

	return nil
}
