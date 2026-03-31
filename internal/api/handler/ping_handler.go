package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/service"
	"github.com/go-chi/chi/v5"
)

// HeartbeatPingServiceInterface defines ping-specific resource operations.
type HeartbeatPingServiceInterface interface {
	GetResourceByHeartbeatSlug(ctx context.Context, slug string) (*domain.Resource, error)
	MarkHeartbeatPing(ctx context.Context, resourceID string, at time.Time) error
	HandleHeartbeatRecovery(ctx context.Context, resource *domain.Resource) error
}

// PingHandler handles public heartbeat ping endpoints.
type PingHandler struct {
	resourceService HeartbeatPingServiceInterface
	rateLimiter     *PingRateLimiter
	now             func() time.Time
}

func NewPingHandler(resourceService HeartbeatPingServiceInterface) *PingHandler {
	return &PingHandler{
		resourceService: resourceService,
		rateLimiter:     NewPingRateLimiter(100, time.Minute),
		now:             time.Now,
	}
}

func (h *PingHandler) Ping(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		respondError(w, http.StatusUnprocessableEntity, "invalid slug format")
		return
	}
	if err := domain.ValidateHeartbeatSlug(slug); err != nil {
		respondError(w, http.StatusUnprocessableEntity, "invalid slug format")
		return
	}

	now := h.now().UTC()
	if !h.rateLimiter.Allow(slug, now) {
		respondError(w, http.StatusTooManyRequests, "rate limit exceeded")
		return
	}

	resource, err := h.resourceService.GetResourceByHeartbeatSlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, service.ErrResourceNotFound) {
			respondError(w, http.StatusNotFound, "monitor not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to process ping")
		return
	}
	if !resource.IsActive {
		respondError(w, http.StatusForbidden, "monitor is paused")
		return
	}

	if err := h.resourceService.MarkHeartbeatPing(r.Context(), resource.ID, now); err != nil {
		if errors.Is(err, service.ErrResourceNotFound) {
			respondError(w, http.StatusNotFound, "monitor not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to process ping")
		return
	}

	// If the monitor was down, resolve any active incident (recovery path).
	if resource.Status == domain.StatusDown {
		if err := h.resourceService.HandleHeartbeatRecovery(r.Context(), resource); err != nil {
			// Non-fatal: log and continue; ping was recorded successfully.
			_ = err
		}
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"monitor": resource.Name,
	})
}
