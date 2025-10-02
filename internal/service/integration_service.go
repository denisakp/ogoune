package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/repository"
)

// IntegrationService provides business logic for integration management operations.
type IntegrationService struct {
	integrations repository.IntegrationRepository
}

// NewIntegrationService creates a new IntegrationService with the given repository dependency.
func NewIntegrationService(integrations repository.IntegrationRepository) *IntegrationService {
	return &IntegrationService{
		integrations: integrations,
	}
}

// CreateIntegration creates a new integration in the system.
func (s *IntegrationService) CreateIntegration(ctx context.Context, integration *domain.Integration) error {
	if integration == nil {
		return fmt.Errorf("%w: integration cannot be nil", ErrValidationFailed)
	}

	if integration.Name == "" {
		return fmt.Errorf("%w: integration name is required", ErrValidationFailed)
	}

	if integration.Target == "" {
		return fmt.Errorf("%w: integration target is required", ErrValidationFailed)
	}

	if integration.Type == "" {
		return fmt.Errorf("%w: integration type is required", ErrValidationFailed)
	}

	// Validate integration type
	validTypes := map[domain.IntegrationType]bool{
		domain.IntegrationSMTP:       true,
		domain.IntegrationSlack:      true,
		domain.IntegrationGoogleChat: true,
	}

	if !validTypes[integration.Type] {
		return fmt.Errorf("%w: invalid integration type '%s'", ErrValidationFailed, integration.Type)
	}

	if err := s.integrations.Create(ctx, integration); err != nil {
		return fmt.Errorf("failed to create integration: %w", err)
	}

	return nil
}

// ListIntegrations retrieves all integrations with pagination.
func (s *IntegrationService) ListIntegrations(ctx context.Context, limit, offset int) ([]*domain.Integration, error) {
	return s.integrations.List(ctx, limit, offset)
}

// GetIntegrationByID retrieves an integration by its ID.
func (s *IntegrationService) GetIntegrationByID(ctx context.Context, id string) (*domain.Integration, error) {
	integration, err := s.integrations.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("%w: integration not found", ErrResourceNotFound)
		}
		return nil, err
	}
	return integration, nil
}

// UpdateIntegration updates an existing integration.
func (s *IntegrationService) UpdateIntegration(ctx context.Context, id string, name, target string, isActive *bool) (*domain.Integration, error) {
	// Fetch existing integration
	integration, err := s.integrations.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("%w: integration not found", ErrResourceNotFound)
		}
		return nil, err
	}

	// Apply updates
	if name != "" {
		integration.Name = name
	}

	if target != "" {
		integration.Target = target
	}

	if isActive != nil {
		integration.IsActive = *isActive
	}

	if err := s.integrations.Update(ctx, integration); err != nil {
		return nil, fmt.Errorf("failed to update integration: %w", err)
	}

	return integration, nil
}

// ListActiveIntegrations retrieves all active integrations.
func (s *IntegrationService) ListActiveIntegrations(ctx context.Context) ([]*domain.Integration, error) {
	// Query all integrations and filter active ones
	allIntegrations, err := s.integrations.List(ctx, 1000, 0)
	if err != nil {
		return nil, err
	}

	var activeIntegrations []*domain.Integration
	for _, integration := range allIntegrations {
		if integration.IsActive {
			activeIntegrations = append(activeIntegrations, integration)
		}
	}

	return activeIntegrations, nil
}
