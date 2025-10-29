package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/denisakp/pulseguard/internal/config"
	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/repository"
	"github.com/denisakp/pulseguard/pkg/notifier"
)

// NotificationService provides business logic for notification operations.
type NotificationService struct {
	resources    repository.ResourceRepository
	integrations repository.IntegrationRepository
}

// NewNotificationService creates a new NotificationService with the given repository dependencies.
func NewNotificationService(
	resources repository.ResourceRepository,
	integrations repository.IntegrationRepository,
) *NotificationService {
	return &NotificationService{
		resources:    resources,
		integrations: integrations,
	}
}

// TestNotification sends a test notification for a specific resource using SMTP.
// It retrieves the resource, fetches active SMTP integrations, and sends a test email.
func (s *NotificationService) TestNotification(ctx context.Context, resourceID string) error {
	// Fetch the resource
	resource, err := s.resources.FindByID(ctx, resourceID)
	if err != nil {
		return fmt.Errorf("failed to find resource: %w", err)
	}

	// load config
	cfg := config.Load()

	// Send test notification to all SMTP integrations
	smtpNotifier := notifier.NewSMTPNotifier()

	// default SMTP integration
	smtpConfig, _ := json.Marshal(map[string]string{
		"type":          "smtp",
		"recipient":     cfg.DefaultRecipientEmail,
		"sender":        "no-reply@pulseguard.io",
		"smtp_host":     cfg.SMTPHost,
		"smtp_port":     cfg.SMTPPort,
		"smtp_user":     cfg.SMTPUser,
		"smtp_password": cfg.SMTPPassword,
	})

	smtpIntegration := &domain.Integration{
		Config:   smtpConfig,
		IsActive: true,
		Name:     "default-smtp-integration",
		Base:     domain.Base{ID: "smtp-default"},
	}

	if err := smtpNotifier.SendTestNotification(ctx, *smtpIntegration, *resource); err != nil {
		return fmt.Errorf("failed to send test notification via integration %s: %w", smtpIntegration, err)
	}

	return nil
}
