package service

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/pkg/notifier"
)

// ExpiryNotificationService drives per-resource expiry alerting with deduplication.
// A daily expiry:check task calls CheckAndNotify for every active HTTP resource.
type ExpiryNotificationService struct {
	logs             port.ExpiryNotificationLogRepository
	channels         port.NotificationChannelRepository
	globalThresholds []int
}

// NewExpiryNotificationService creates a new ExpiryNotificationService.
func NewExpiryNotificationService(
	logs port.ExpiryNotificationLogRepository,
	channels port.NotificationChannelRepository,
	globalThresholds []int,
) *ExpiryNotificationService {
	return &ExpiryNotificationService{
		logs:             logs,
		channels:         channels,
		globalThresholds: globalThresholds,
	}
}

// ParseGlobalThresholds converts a comma-separated string (e.g. "30,14,7,1") to []int.
// Values that are non-positive or exceed 365 are silently ignored.
// If the string is empty or produces no valid values, the hardcoded defaults are returned.
func ParseGlobalThresholds(s string) []int {
	if s == "" {
		return domain.DefaultExpiryThresholds()
	}
	parts := strings.Split(s, ",")
	result := make([]int, 0, len(parts))
	for _, p := range parts {
		v, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil || v <= 0 || v > 365 {
			continue
		}
		result = append(result, v)
	}
	if len(result) == 0 {
		return domain.DefaultExpiryThresholds()
	}
	return result
}

// CheckAndNotify evaluates SSL and domain expiry for a single resource and dispatches
// any threshold alerts that have not yet been sent. It is idempotent — repeated calls
// for the same resource/threshold combination are no-ops (deduplication via log table).
func (s *ExpiryNotificationService) CheckAndNotify(ctx context.Context, resource *domain.Resource, resourceChannels []*domain.NotificationChannel) error {
	thresholds := resource.ExpiryThresholds(s.globalThresholds)

	// Check SSL timeline
	if resource.Metadata != nil && resource.Metadata.SSLExpirationDate != nil {
		if err := s.checkExpiry(ctx, resource, "ssl", *resource.Metadata.SSLExpirationDate, resource.Metadata.SSLIssuer, thresholds, resourceChannels); err != nil {
			slog.Error("SSL expiry check failed", "resource_id", resource.ID, "error", err)
		}
	}

	// Check domain timeline (silently skip if date is unavailable — NFR-003)
	if resource.Metadata != nil && resource.Metadata.DomainExpirationDate != nil {
		if err := s.checkExpiry(ctx, resource, "domain", *resource.Metadata.DomainExpirationDate, resource.Metadata.DomainRegistrar, thresholds, resourceChannels); err != nil {
			slog.Error("domain expiry check failed", "resource_id", resource.ID, "error", err)
		}
	}

	return nil
}

// checkExpiry performs the threshold evaluation for a single expiry timeline (ssl or domain).
// It finds the most urgent threshold that has not yet fired and dispatches the notification.
func (s *ExpiryNotificationService) checkExpiry(
	ctx context.Context,
	resource *domain.Resource,
	expiryType string,
	expiresAt time.Time,
	issuer string,
	thresholds []int,
	resourceChannels []*domain.NotificationChannel,
) error {
	daysRemaining := int(time.Until(expiresAt).Hours() / 24)

	// Find the most urgent (smallest) threshold that has been crossed but not yet notified.
	var targetThreshold int
	found := false
	for _, t := range thresholds {
		if daysRemaining <= t {
			count, err := s.logs.CountByKey(ctx, resource.ID, expiryType, t)
			if err != nil {
				return fmt.Errorf("failed to count log entries for threshold %d: %w", t, err)
			}
			if count == 0 {
				// Pick the smallest un-fired threshold that we've crossed
				if !found || t < targetThreshold {
					targetThreshold = t
					found = true
				}
			}
		}
	}

	if !found {
		return nil // All applicable thresholds already fired — nothing to do
	}

	data := &notifier.ExpiryNotification{
		Resource:      *resource,
		ExpiryType:    expiryType,
		DaysRemaining: daysRemaining,
		ExpiresAt:     expiresAt,
		Issuer:        issuer,
		Threshold:     targetThreshold,
		TriggeredAt:   time.Now(),
	}

	return s.sendExpiryNotification(ctx, resource, resourceChannels, data, targetThreshold)
}

// sendExpiryNotification dispatches to all configured channels.
// It writes the dedup log entry ONLY on full success (NFR-006).
func (s *ExpiryNotificationService) sendExpiryNotification(
	ctx context.Context,
	resource *domain.Resource,
	resourceChannels []*domain.NotificationChannel,
	data *notifier.ExpiryNotification,
	threshold int,
) error {
	payload := notifier.NotificationPayload{Expiry: data}

	for _, ch := range resourceChannels {
		if err := s.dispatchNotification(ctx, payload, ch); err != nil {
			slog.Error("failed to dispatch expiry notification", "channel_id", ch.ID, "channel_type", ch.Type, "error", err)
			return err // at-least-once: do not write log on failure
		}
	}

	// All dispatches succeeded — record the dedup entry
	return s.recordNotified(ctx, resource.ID, data.ExpiryType, threshold)
}

// recordNotified persists an ExpiryNotificationLog entry to prevent re-dispatch.
func (s *ExpiryNotificationService) recordNotified(ctx context.Context, resourceID, expiryType string, threshold int) error {
	entry := &domain.ExpiryNotificationLog{
		ResourceID: resourceID,
		ExpiryType: expiryType,
		Threshold:  threshold,
		SentAt:     time.Now(),
	}
	return s.logs.Create(ctx, entry)
}

// ResetLogs clears all dedup log entries for a resource+expiryType pair.
// This is called when a renewal is detected (new expiry date is later than stored date).
func (s *ExpiryNotificationService) ResetLogs(ctx context.Context, resourceID, expiryType string) error {
	return s.logs.DeleteByResourceIDAndType(ctx, resourceID, expiryType)
}

// CleanupOldLogs removes log entries older than one year.
// Called at the end of each daily expiry:check run (NFR-005).
func (s *ExpiryNotificationService) CleanupOldLogs(ctx context.Context) error {
	cutoff := time.Now().Add(-365 * 24 * time.Hour)
	if err := s.logs.DeleteOlderThan(ctx, cutoff); err != nil {
		slog.Error("failed to cleanup old expiry notification logs", "error", err)
		return err
	}
	return nil
}

// dispatchNotification mirrors the pattern from incident_service.go.
func (s *ExpiryNotificationService) dispatchNotification(ctx context.Context, payload notifier.NotificationPayload, channel *domain.NotificationChannel) error {
	var n notifier.Notifier
	var err error

	switch channel.Type {
	case "smtp":
		n, err = notifier.NewSMTPNotifierFromConfig(string(channel.Config))
		if err != nil {
			return fmt.Errorf("failed to build SMTP notifier: %w", err)
		}
	case "webhook":
		n, err = notifier.NewWebhookNotifierFromConfig(string(channel.Config))
		if err != nil {
			return fmt.Errorf("failed to build webhook notifier: %w", err)
		}
	default:
		return fmt.Errorf("unknown notification channel type: %s", channel.Type)
	}

	if err := n.Send(ctx, payload); err != nil {
		return fmt.Errorf("failed to send expiry notification: %w", err)
	}

	slog.Info("sent expiry alert", "channel_type", channel.Type, "channel_id", channel.ID)
	return nil
}
