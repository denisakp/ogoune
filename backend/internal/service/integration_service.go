package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/dto"
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
func (s *IntegrationService) CreateIntegration(ctx context.Context, payload *dto.CreateIntegrationPayload) (*domain.Integration, error) {
	configBytes, err := json.Marshal(payload.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	eventTypesBytes, err := json.Marshal(payload.EventTypes)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event types: %w", err)
	}

	integration := &domain.Integration{
		Name:       payload.Name,
		Config:     configBytes,
		IsActive:   payload.IsActive,
		EventTypes: eventTypesBytes,
	}

	if err := s.integrations.Create(ctx, integration); err != nil {
		return nil, fmt.Errorf("failed to create integration: %w", err)
	}

	return integration, nil
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
func (s *IntegrationService) UpdateIntegration(ctx context.Context, id string, payload *dto.UpdateIntegrationPayload) (*domain.Integration, error) {
	// Fetch existing integration
	integration, err := s.integrations.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("%w: integration not found", ErrResourceNotFound)
		}
		return nil, err
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
