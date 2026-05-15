package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/dto"
	"github.com/denisakp/ogoune/internal/repository"
	"github.com/denisakp/ogoune/pkg/notifier"
)

// NotificationService provides business logic for notification operations.
type NotificationService struct {
	resources repository.ResourceRepository
	channels  repository.NotificationChannelRepository
}

// NewNotificationService creates a new NotificationService.
func NewNotificationService(
	resources repository.ResourceRepository,
	channels repository.NotificationChannelRepository,
) *NotificationService {
	return &NotificationService{
		resources: resources,
		channels:  channels,
	}
}

// CreateNotificationChannel creates a new notification channel
func (s *NotificationService) CreateNotificationChannel(ctx context.Context, payload *dto.CreateNotificationChannelPayload) (*domain.NotificationChannel, error) {
	// Validate channel type
	if !payload.Type.IsValid() {
		return nil, fmt.Errorf("%w: invalid channel type: %s", ErrValidationFailed, payload.Type)
	}

	// Validate config based on type
	if err := s.validateChannelConfig(payload.Type, payload.Config); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrValidationFailed, err)
	}

	// Create domain model
	channel := &domain.NotificationChannel{
		Name:             payload.Name,
		Type:             payload.Type,
		Config:           payload.Config,
		EnabledByDefault: payload.EnabledByDefault,
	}

	// Persist to database
	if err := s.channels.Create(ctx, channel); err != nil {
		return nil, fmt.Errorf("failed to create notification channel: %w", err)
	}

	return channel, nil
}

// GetNotificationChannel retrieves a notification channel by ID
func (s *NotificationService) GetNotificationChannel(ctx context.Context, id string) (*domain.NotificationChannel, error) {
	channel, err := s.channels.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("%w: notification channel not found", ErrResourceNotFound)
		}
		return nil, err
	}
	return channel, nil
}

// ListNotificationChannels retrieves all notification channels with pagination
func (s *NotificationService) ListNotificationChannels(ctx context.Context, limit, offset int) ([]*domain.NotificationChannel, error) {
	return s.channels.List(ctx, limit, offset)
}

// UpdateNotificationChannel updates an existing notification channel
func (s *NotificationService) UpdateNotificationChannel(ctx context.Context, id string, payload *dto.UpdateNotificationChannelPayload) (*domain.NotificationChannel, error) {
	// Fetch existing channel
	channel, err := s.channels.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("%w: notification channel not found", ErrResourceNotFound)
		}
		return nil, err
	}

	// Update fields if provided
	if payload.Name != nil {
		channel.Name = *payload.Name
	}

	if payload.Type != nil {
		if !payload.Type.IsValid() {
			return nil, fmt.Errorf("%w: invalid channel type: %s", ErrValidationFailed, *payload.Type)
		}
		channel.Type = *payload.Type
	}

	if payload.Config != nil {
		// Validate config based on type
		if err := s.validateChannelConfig(channel.Type, payload.Config); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrValidationFailed, err)
		}
		channel.Config = payload.Config
	}

	if payload.EnabledByDefault != nil {
		channel.EnabledByDefault = *payload.EnabledByDefault
	}

	// Persist updates
	if err := s.channels.Update(ctx, channel); err != nil {
		return nil, fmt.Errorf("failed to update notification channel: %w", err)
	}

	return channel, nil
}

// DeleteNotificationChannel deletes a notification channel
func (s *NotificationService) DeleteNotificationChannel(ctx context.Context, id string) error {
	// Verify channel exists
	_, err := s.channels.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("%w: notification channel not found", ErrResourceNotFound)
		}
		return err
	}

	// Delete channel
	if err := s.channels.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete notification channel: %w", err)
	}

	return nil
}

// TestNotificationChannel sends a test notification using the specified channel
func (s *NotificationService) TestNotificationChannel(ctx context.Context, id string) error {
	// Fetch channel
	channel, err := s.channels.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("%w: notification channel not found", ErrResourceNotFound)
		}
		return err
	}

	// For MVP, only support SMTP
	if channel.Type != domain.NotificationChannelTypeSMTP {
		return fmt.Errorf("only SMTP channels are supported in this version")
	}

	// Parse SMTP config
	var smtpConfig dto.SMTPConfig
	if err := json.Unmarshal(channel.Config, &smtpConfig); err != nil {
		return fmt.Errorf("invalid SMTP configuration: %w", err)
	}

	// Create test resource for notification
	testResource := domain.Resource{
		Base: domain.Base{
			ID: "test-resource-id",
		},
		Name:   "Test Resource",
		Type:   domain.ResourceHTTP,
		Target: "https://example.com",
		Status: domain.StatusUp,
	}

	// Create SMTP notifier with channel config
	smtpNotifier := notifier.NewSMTPNotifier(
		smtpConfig.Recipients[0], // Use first recipient for test
		smtpConfig.Sender,
		smtpConfig.Host,
		fmt.Sprintf("%d", smtpConfig.Port),
		smtpConfig.Username,
		smtpConfig.Password,
	)

	// Send test notification
	if err := smtpNotifier.SendTestNotification(ctx, testResource); err != nil {
		return fmt.Errorf("failed to send test notification: %w", err)
	}

	return nil
}

// validateChannelConfig validates the configuration JSON for a given channel type
func (s *NotificationService) validateChannelConfig(channelType domain.NotificationChannelType, configJSON json.RawMessage) error {
	switch channelType {
	case domain.NotificationChannelTypeSMTP:
		var config dto.SMTPConfig
		if err := json.Unmarshal(configJSON, &config); err != nil {
			return fmt.Errorf("invalid SMTP config format: %w", err)
		}
		// Validate required SMTP fields
		if config.Host == "" {
			return errors.New("SMTP host is required")
		}
		if config.Port == 0 {
			return errors.New("SMTP port is required")
		}
		if config.Sender == "" {
			return errors.New("SMTP sender is required")
		}
		if len(config.Recipients) == 0 {
			return errors.New("at least one recipient is required")
		}
		return nil

	case domain.NotificationChannelTypeSlack:
		var config dto.SlackConfig
		if err := json.Unmarshal(configJSON, &config); err != nil {
			return fmt.Errorf("invalid Slack config format: %w", err)
		}
		if config.WebhookURL == "" {
			return errors.New("slack webhook URL is required")
		}
		return nil

	case domain.NotificationChannelTypeSMS:
		var config dto.SMSConfig
		if err := json.Unmarshal(configJSON, &config); err != nil {
			return fmt.Errorf("invalid SMS config format: %w", err)
		}
		if config.Provider == "" {
			return errors.New("SMS provider is required")
		}
		if config.FromNumber == "" {
			return errors.New("SMS from number is required")
		}
		if len(config.ToNumbers) == 0 {
			return errors.New("at least one SMS recipient is required")
		}
		return nil

	default:
		return fmt.Errorf("unsupported channel type: %s", channelType)
	}
}

// ValidateAndTestChannelConfig validates and tests channel configuration without requiring it to be saved.
// This is useful for testing config before creating a channel.
func (s *NotificationService) ValidateAndTestChannelConfig(ctx context.Context, channelType domain.NotificationChannelType, configJSON json.RawMessage) error {
	// Validate config first
	if err := s.validateChannelConfig(channelType, configJSON); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// For MVP, only test SMTP
	if channelType != domain.NotificationChannelTypeSMTP {
		// For non-SMTP types, just validate without sending
		return nil
	}

	// Parse SMTP config
	var smtpConfig dto.SMTPConfig
	if err := json.Unmarshal(configJSON, &smtpConfig); err != nil {
		return fmt.Errorf("invalid SMTP configuration: %w", err)
	}

	// Create test resource for notification
	testResource := domain.Resource{
		Base: domain.Base{
			ID: "test-resource-id",
		},
		Name:   "Test Resource",
		Type:   domain.ResourceHTTP,
		Target: "https://example.com",
		Status: domain.StatusUp,
	}

	// Create SMTP notifier with channel config
	smtpNotifier := notifier.NewSMTPNotifier(
		smtpConfig.Recipients[0], // Use first recipient for test
		smtpConfig.Sender,
		smtpConfig.Host,
		fmt.Sprintf("%d", smtpConfig.Port),
		smtpConfig.Username,
		smtpConfig.Password,
	)

	// Send test notification
	if err := smtpNotifier.SendTestNotification(ctx, testResource); err != nil {
		return fmt.Errorf("failed to send test notification: %w", err)
	}

	return nil
}
