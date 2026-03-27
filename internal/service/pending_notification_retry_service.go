package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository"
	"github.com/denisakp/ogoune/pkg/notifier"
)

const defaultPendingNotificationRetryLimit = 1000

// PendingNotificationRetrySummary captures one startup retry pass outcome.
type PendingNotificationRetrySummary struct {
	RetriedCount        int
	ExpiredCount        int
	FailedCount         int
	SkippedClaimedCount int
	ScannedCount        int
}

// PendingNotificationRetryService retries pending notification events during startup.
type PendingNotificationRetryService struct {
	notifications repository.NotificationRepository
	incidents     repository.IncidentRepository
	channels      repository.NotificationChannelRepository
	components    repository.ComponentRepository
	claimOwner    string
	staleAfter    time.Duration
	now           func() time.Time
}

// NewPendingNotificationRetryService creates a startup retry service.
func NewPendingNotificationRetryService(
	notifications repository.NotificationRepository,
	incidents repository.IncidentRepository,
	channels repository.NotificationChannelRepository,
	components repository.ComponentRepository,
	claimOwner string,
	staleAfter time.Duration,
) *PendingNotificationRetryService {
	if staleAfter <= 0 {
		staleAfter = 24 * time.Hour
	}
	if claimOwner == "" {
		host, err := os.Hostname()
		if err != nil || host == "" {
			host = "unknown-host"
		}
		claimOwner = fmt.Sprintf("%s-%d", host, os.Getpid())
	}

	return &PendingNotificationRetryService{
		notifications: notifications,
		incidents:     incidents,
		channels:      channels,
		components:    components,
		claimOwner:    claimOwner,
		staleAfter:    staleAfter,
		now:           time.Now,
	}
}

// RetryPendingNotifications performs a bounded single-pass retry run.
func (s *PendingNotificationRetryService) RetryPendingNotifications(ctx context.Context, limit int) (PendingNotificationRetrySummary, error) {
	summary := PendingNotificationRetrySummary{}
	if limit <= 0 {
		limit = defaultPendingNotificationRetryLimit
	}

	pendingEvents, err := s.notifications.FindPending(ctx, limit, 0)
	if err != nil {
		return summary, fmt.Errorf("failed to find pending notification events: %w", err)
	}

	summary.ScannedCount = len(pendingEvents)
	if len(pendingEvents) == 0 {
		log.Println("[STARTUP] No pending notifications found")
		return summary, nil
	}

	log.Printf("[STARTUP] Found %d pending notification(s) to evaluate", len(pendingEvents))
	cutoff := s.now().Add(-s.staleAfter)

	for _, event := range pendingEvents {
		if event == nil {
			continue
		}

		claimed, err := s.notifications.ClaimPending(ctx, event.ID, s.claimOwner, s.now())
		if err != nil {
			log.Printf("[STARTUP] [WARNING] Failed to claim pending notification %s: %v", event.ID, err)
			continue
		}
		if !claimed {
			summary.SkippedClaimedCount++
			continue
		}

		if event.CreatedAt.Before(cutoff) {
			reason := fmt.Sprintf("expired stale pending notification event older than %s", s.staleAfter)
			if err := s.notifications.MarkAsExpired(ctx, event.ID, reason, s.now()); err != nil {
				log.Printf("[STARTUP] [WARNING] Failed to mark stale notification %s expired: %v", event.ID, err)
				continue
			}
			summary.ExpiredCount++
			continue
		}

		if event.Type != domain.NotificationEventTypeDown && event.Type != domain.NotificationEventTypeUp {
			reason := fmt.Sprintf("unsupported pending notification type: %s", event.Type)
			if err := s.notifications.MarkAsFailed(ctx, event.ID, reason, s.now()); err != nil {
				log.Printf("[STARTUP] [WARNING] Failed to mark unsupported notification %s as failed: %v", event.ID, err)
				continue
			}
			summary.FailedCount++
			continue
		}

		incident, err := s.incidents.FindByID(ctx, event.IncidentID)
		if err != nil {
			reason := fmt.Sprintf("incident lookup failed for %s: %v", event.IncidentID, err)
			if markErr := s.notifications.MarkAsFailed(ctx, event.ID, reason, s.now()); markErr != nil {
				log.Printf("[STARTUP] [WARNING] Failed to mark notification %s as failed after incident lookup error: %v", event.ID, markErr)
				continue
			}
			log.Printf("[STARTUP] [WARNING] %s", reason)
			summary.FailedCount++
			continue
		}

		resource := &incident.Resource
		channels := s.resolveNotificationChannels(ctx, resource)
		if len(channels) == 0 {
			reason := fmt.Sprintf("no notification channels available for resource %s", resource.ID)
			if err := s.notifications.MarkAsFailed(ctx, event.ID, reason, s.now()); err != nil {
				log.Printf("[STARTUP] [WARNING] Failed to mark notification %s as failed after missing channels: %v", event.ID, err)
				continue
			}
			log.Printf("[STARTUP] [WARNING] %s", reason)
			summary.FailedCount++
			continue
		}

		if err := s.dispatchNotification(ctx, notifier.NotificationPayload{Incident: incident}, channels); err != nil {
			reason := fmt.Sprintf("retry dispatch failed: %v", err)
			if markErr := s.notifications.MarkAsFailed(ctx, event.ID, reason, s.now()); markErr != nil {
				log.Printf("[STARTUP] [WARNING] Failed to mark notification %s as failed after retry error: %v", event.ID, markErr)
				continue
			}
			log.Printf("[STARTUP] [WARNING] Failed to retry notification %s: %v", event.ID, err)
			summary.FailedCount++
			continue
		}

		if err := s.notifications.MarkAsSent(ctx, event.ID, s.now()); err != nil {
			log.Printf("[STARTUP] [WARNING] Notification %s dispatched but failed to mark as sent: %v", event.ID, err)
			continue
		}

		summary.RetriedCount++
	}

	log.Printf("[STARTUP] Pending notifications summary: retried=%d expired=%d failed=%d skipped_claimed=%d",
		summary.RetriedCount,
		summary.ExpiredCount,
		summary.FailedCount,
		summary.SkippedClaimedCount,
	)

	return summary, nil
}

func (s *PendingNotificationRetryService) dispatchNotification(ctx context.Context, payload notifier.NotificationPayload, channels []*domain.NotificationChannel) error {
	for _, channel := range channels {
		if channel == nil {
			continue
		}

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
			return fmt.Errorf("failed to send notification: %w", err)
		}
	}

	return nil
}

func (s *PendingNotificationRetryService) resolveNotificationChannels(ctx context.Context, r *domain.Resource) []*domain.NotificationChannel {
	var channels []*domain.NotificationChannel

	resourceChannels, err := s.channels.FindByResourceID(ctx, r.ID)
	if err == nil && len(resourceChannels) > 0 {
		return resourceChannels
	}

	if r.ComponentID != nil && s.components != nil {
		component, err := s.components.FindByID(ctx, *r.ComponentID)
		if err == nil && component != nil {
			componentChannels, err := s.channels.FindByComponentID(ctx, component.ID)
			if err == nil && len(componentChannels) > 0 {
				return componentChannels
			}
		}
	}

	defaultChannels, err := s.channels.FindDefaultChannels(ctx)
	if err == nil && len(defaultChannels) > 0 {
		return defaultChannels
	}

	return channels
}
