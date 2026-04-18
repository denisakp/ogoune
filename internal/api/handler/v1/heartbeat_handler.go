package v1

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	dtoV1 "github.com/denisakp/ogoune/internal/dto/v1"
	"github.com/denisakp/ogoune/internal/service"
	"github.com/go-chi/chi/v5"
)

// HeartbeatV1ServiceInterface defines the heartbeat service methods used by the v1 heartbeat handler.
type HeartbeatV1ServiceInterface interface {
	GetResourceByHeartbeatSlug(ctx context.Context, slug string) (*domain.Resource, error)
	MarkHeartbeatPing(ctx context.Context, resourceID string, at time.Time) error
	HandleHeartbeatRecovery(ctx context.Context, resource *domain.Resource) error
}

// HeartbeatV1Handler handles v1 heartbeat ping endpoints.
type HeartbeatV1Handler struct {
	service HeartbeatV1ServiceInterface
}

// NewHeartbeatV1Handler creates a new HeartbeatV1Handler.
func NewHeartbeatV1Handler(svc HeartbeatV1ServiceInterface) *HeartbeatV1Handler {
	return &HeartbeatV1Handler{service: svc}
}

// Ping handles POST /api/v1/heartbeat/ping/{slug}
//
// @Summary     Record a heartbeat ping
// @Tags        heartbeat
// @Produce     json
// @Param       slug path string true "Heartbeat slug"
// @Success     200 {object} dtoV1.SingleResponse[dtoV1.HeartbeatPingResponse]
// @Failure     403 {object} dtoV1.ErrorResponse "monitor is paused"
// @Failure     404 {object} dtoV1.ErrorResponse "monitor not found"
// @Failure     429 {object} dtoV1.ErrorResponse "rate limit exceeded"
// @Router      /heartbeat/ping/{slug} [post]
func (h *HeartbeatV1Handler) Ping(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	now := time.Now().UTC()

	resource, err := h.service.GetResourceByHeartbeatSlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, service.ErrResourceNotFound) {
			respondError(w, http.StatusNotFound, "RESOURCE_NOT_FOUND", "monitor not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to process ping")
		return
	}

	if !resource.IsActive {
		respondError(w, http.StatusForbidden, "MONITOR_PAUSED", "monitor is paused")
		return
	}

	if err := h.service.MarkHeartbeatPing(r.Context(), resource.ID, now); err != nil {
		if errors.Is(err, service.ErrResourceNotFound) {
			respondError(w, http.StatusNotFound, "RESOURCE_NOT_FOUND", "monitor not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to process ping")
		return
	}

	// Non-fatal: attempt recovery if monitor was down.
	if resource.Status == domain.StatusDown {
		_ = h.service.HandleHeartbeatRecovery(r.Context(), resource)
	}

	respond(w, http.StatusOK, dtoV1.HeartbeatPingResponse{
		ReceivedAt: now.Format(time.RFC3339),
	})
}
