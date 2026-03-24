package worker

import (
	"context"
	"log"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/repository"
	"github.com/hibiken/asynq"
)

// TypeExpiryCheck is the Asynq task type for the daily expiry:check job.
const TypeExpiryCheck = "expiry:check"

// enricher abstracts the EnrichmentService so unit tests can provide a fake.
type enricher interface {
	Enrich(ctx context.Context, resource *domain.Resource) (*domain.ResourceMetaData, error)
}

// expiryChecker abstracts ExpiryNotificationService for unit tests.
type expiryChecker interface {
	CheckAndNotify(ctx context.Context, resource *domain.Resource, channels []*domain.NotificationChannel) error
	ResetLogs(ctx context.Context, resourceID string, expiryType string) error
	CleanupOldLogs(ctx context.Context) error
}

// activeResourceLister is the narrow slice of ResourceRepository that the handler needs.
type activeResourceLister interface {
	FindActive(ctx context.Context, limit, offset int) ([]*domain.Resource, error)
}

// ExpiryTaskHandler processes the daily expiry:check task.
// It iterates over every active HTTP resource, enriches its metadata,
// detects renewals, dispatches threshold notifications, and cleans up
// stale log entries. Per-resource errors are logged but do not abort the run.
type ExpiryTaskHandler struct {
	resources activeResourceLister
	channels  repository.NotificationChannelRepository
	enricher  enricher
	expiry    expiryChecker
}

// NewExpiryTaskHandler creates a new ExpiryTaskHandler.
func NewExpiryTaskHandler(
	resources activeResourceLister,
	channels repository.NotificationChannelRepository,
	enricher enricher,
	expiry expiryChecker,
) *ExpiryTaskHandler {
	return &ExpiryTaskHandler{
		resources: resources,
		channels:  channels,
		enricher:  enricher,
		expiry:    expiry,
	}
}

// ProcessTask implements asynq.Handler for the "expiry:check" task type.
// It scans all active HTTP resources, enriches them, checks expiry thresholds,
// and sends notifications. Errors for individual resources are logged and skipped.
func (h *ExpiryTaskHandler) ProcessTask(ctx context.Context, _ *asynq.Task) error {
	// Fetch all active resources. Large limit to process all at once.
	allResources, err := h.resources.FindActive(ctx, 10000, 0)
	if err != nil {
		return err
	}

	for _, resource := range allResources {
		// Only HTTP resources have SSL / domain metadata.
		if resource.Type != domain.ResourceHTTP {
			continue
		}

		if err := h.processResource(ctx, resource); err != nil {
			log.Printf("[ExpiryTaskHandler] resource %s (%s): %v", resource.ID, resource.Name, err)
			// Intentionally continue — one failing resource must not block others.
		}
	}

	// House-keep old log entries regardless of per-resource outcomes.
	if err := h.expiry.CleanupOldLogs(ctx); err != nil {
		log.Printf("[ExpiryTaskHandler] CleanupOldLogs failed: %v", err)
	}

	return nil
}

// processResource enriches a single resource, detects renewals, then checks thresholds.
func (h *ExpiryTaskHandler) processResource(ctx context.Context, resource *domain.Resource) error {
	metadata, err := h.enricher.Enrich(ctx, resource)
	if err != nil {
		return err
	}
	if metadata == nil {
		return nil
	}

	// Renewal detection — if the new date is newer than what was stored, the
	// certificate/domain has been renewed; reset dedup logs so a fresh alert fires.
	if resource.Metadata != nil {
		if h.sslRenewed(resource.Metadata, metadata) {
			if err := h.expiry.ResetLogs(ctx, resource.ID, "ssl"); err != nil {
				log.Printf("[ExpiryTaskHandler] ResetLogs(ssl) for %s: %v", resource.ID, err)
			}
		}
		if h.domainRenewed(resource.Metadata, metadata) {
			if err := h.expiry.ResetLogs(ctx, resource.ID, "domain"); err != nil {
				log.Printf("[ExpiryTaskHandler] ResetLogs(domain) for %s: %v", resource.ID, err)
			}
		}
	}

	// Attach the freshly enriched metadata so CheckAndNotify uses the latest dates.
	resource.Metadata = metadata

	// Resolve notification channels for this resource.
	channels, err := h.channels.FindByResourceID(ctx, resource.ID)
	if err != nil {
		log.Printf("[ExpiryTaskHandler] FindByResourceID for %s: %v", resource.ID, err)
		channels = nil
	}

	return h.expiry.CheckAndNotify(ctx, resource, channels)
}

// sslRenewed returns true when the freshly-fetched SSL date is strictly later than stored.
func (h *ExpiryTaskHandler) sslRenewed(stored, fresh *domain.ResourceMetaData) bool {
	if stored.SSLExpirationDate == nil || fresh.SSLExpirationDate == nil {
		return false
	}
	return fresh.SSLExpirationDate.After(*stored.SSLExpirationDate)
}

// domainRenewed returns true when the freshly-fetched domain date is strictly later than stored.
func (h *ExpiryTaskHandler) domainRenewed(stored, fresh *domain.ResourceMetaData) bool {
	if stored.DomainExpirationDate == nil || fresh.DomainExpirationDate == nil {
		return false
	}
	return fresh.DomainExpirationDate.After(*stored.DomainExpirationDate)
}
