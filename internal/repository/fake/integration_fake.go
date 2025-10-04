package fake

import (
	"context"
	"sync"

	domain "github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/repository"
)

type IntegrationFake struct {
	mu           sync.RWMutex
	integrations map[string]*domain.Integration
}

// NewIntegrationFake creates a new in-memory fake implementation of IntegrationRepository
func NewIntegrationFake() repository.IntegrationRepository {
	return &IntegrationFake{
		integrations: make(map[string]*domain.Integration),
	}
}

func (f *IntegrationFake) Create(ctx context.Context, i *domain.Integration) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if i.ID == "" {
		return ErrInvalidInput
	}

	if _, exists := f.integrations[i.ID]; exists {
		return ErrDuplicate
	}

	// Create a copy to avoid external mutations
	integration := *i
	f.integrations[i.ID] = &integration
	return nil
}

func (f *IntegrationFake) FindByID(ctx context.Context, id string) (*domain.Integration, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	integration, exists := f.integrations[id]
	if !exists {
		return nil, ErrNotFound
	}

	// Return a copy to avoid external mutations
	result := *integration
	return &result, nil
}

func (f *IntegrationFake) List(ctx context.Context, limit, offset int) ([]*domain.Integration, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var integrations []*domain.Integration
	i := 0
	for _, integration := range f.integrations {
		if i < offset {
			i++
			continue
		}
		if len(integrations) >= limit {
			break
		}
		// Return a copy to avoid external mutations
		result := *integration
		integrations = append(integrations, &result)
		i++
	}

	return integrations, nil
}

func (f *IntegrationFake) Update(ctx context.Context, i *domain.Integration) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.integrations[i.ID]; !exists {
		return ErrNotFound
	}

	// Create a copy to avoid external mutations
	integration := *i
	f.integrations[i.ID] = &integration
	return nil
}

func (f *IntegrationFake) Delete(ctx context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.integrations[id]; !exists {
		return ErrNotFound
	}

	delete(f.integrations, id)
	return nil
}

func (f *IntegrationFake) FindActiveByType(ctx context.Context, t domain.IntegrationType, limit, offset int) ([]*domain.Integration, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var integrations []*domain.Integration
	skipped := 0
	for _, integration := range f.integrations {
		integType := integration.GetType()
		if integType == t && integration.IsActive {
			if skipped < offset {
				skipped++
				continue
			}
			if len(integrations) >= limit {
				break
			}
			// Return a copy to avoid external mutations
			result := *integration
			integrations = append(integrations, &result)
		}
	}

	return integrations, nil
}

// ListActive retrieves all active integrations without pagination.
func (f *IntegrationFake) ListActive(ctx context.Context) ([]*domain.Integration, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var integrations []*domain.Integration
	for _, integration := range f.integrations {
		if integration.IsActive {
			// Return a copy to avoid external mutations
			result := *integration
			integrations = append(integrations, &result)
		}
	}

	return integrations, nil
}
