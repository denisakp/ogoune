// Package handler — public status endpoints (spec 060).
package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/denisakp/ogoune/internal/dto"
)

// PublicStatusProvider is the minimum surface the handler needs from the
// service layer. Defined inline to keep the handler trivially testable.
type PublicStatusProvider interface {
	GetCurrent(ctx context.Context) (*dto.PublicStatus, error)
}

type PublicStatusHandler struct {
	svc PublicStatusProvider
}

func NewPublicStatusHandler(svc PublicStatusProvider) *PublicStatusHandler {
	return &PublicStatusHandler{svc: svc}
}

// GetCurrent — GET /status.
//
//	@Summary	Public status page snapshot
//	@Tags		public-status
//	@Produce	json
//	@Success	200	{object}	dto.PublicStatus
//	@Failure	500	{object}	dto.ProblemDetails
//	@Router		/status [get]
func (h *PublicStatusHandler) GetCurrent(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.GetCurrent(r.Context())
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

func writeProblem(w http.ResponseWriter, status int, code, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"type":   "about:blank",
		"title":  http.StatusText(status),
		"status": status,
		"code":   code,
		"detail": detail,
	})
}
