package service

import (
	"context"
	"fmt"
	"log"

	"github.com/denisakp/pulseguard/internal/repository"
	"github.com/denisakp/pulseguard/pkg/notifier"
)

// NotificationService provides business logic for notification operations.
type NotificationService struct {
	resources     repository.ResourceRepository
	smtpIsEnabled bool
	smtpRecipient string
	smtpSender    string
	smtpHost      string
	smtpPort      string
	smtpUser      string
	smtpPassword  string
}

// NewNotificationService creates a new NotificationService with SMTP configuration.
func NewNotificationService(
	resources repository.ResourceRepository,
	smtpIsEnabled bool,
	smtpRecipient string,
	smtpSender string,
	smtpHost string,
	smtpPort string,
	smtpUser string,
	smtpPassword string,
) *NotificationService {
	return &NotificationService{
		resources:     resources,
		smtpIsEnabled: smtpIsEnabled,
		smtpRecipient: smtpRecipient,
		smtpSender:    smtpSender,
		smtpHost:      smtpHost,
		smtpPort:      smtpPort,
		smtpUser:      smtpUser,
		smtpPassword:  smtpPassword,
	}
}

// TestNotification sends a test notification for a specific resource using SMTP (if enabled).
// It retrieves the resource and sends a test email notification.
func (s *NotificationService) TestNotification(ctx context.Context, resourceID string) error {
	// Fetch the resource
	resource, err := s.resources.FindByID(ctx, resourceID)
	if err != nil {
		return fmt.Errorf("failed to find resource: %w", err)
	}

	// Check if SMTP is enabled
	if !s.smtpIsEnabled {
		return fmt.Errorf("SMTP notifications are not configured")
	}

	// Create SMTP notifier with configured credentials
	smtpNotifier := notifier.NewSMTPNotifier(
		s.smtpRecipient,
		s.smtpSender,
		s.smtpHost,
		s.smtpPort,
		s.smtpUser,
		s.smtpPassword,
	)

	// Send test notification
	if err := smtpNotifier.SendTestNotification(ctx, *resource); err != nil {
		return fmt.Errorf("failed to send test notification: %w", err)
	}

	log.Printf("Successfully sent test notification for resource %s", resourceID)
	return nil
}
