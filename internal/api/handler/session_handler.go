package handler

import (
	"errors"
	"net/http"

	"github.com/denisakp/ogoune/internal/api/response"
	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/service"
	"github.com/go-chi/chi/v5"
)

// SessionHandler exposes session management for the signed-in user.
// Spec 059 FR-008/009/009a · contracts/sessions-api.md.
type SessionHandler struct {
	sessionService *service.SessionService
}

func NewSessionHandler(sessionService *service.SessionService) *SessionHandler {
	return &SessionHandler{sessionService: sessionService}
}

type sessionDTO struct {
	ID           string  `json:"id"`
	Browser      string  `json:"browser"`
	OS           string  `json:"os"`
	IP           string  `json:"ip"`
	Location     *string `json:"location"`
	LastActiveAt string  `json:"last_active_at"`
	IsCurrent    bool    `json:"is_current"`
	RevokedAt    *string `json:"revoked_at"`
}

func toSessionDTO(s *domain.Session, currentID string) sessionDTO {
	return sessionDTO{
		ID:           s.ID,
		Browser:      s.Browser,
		OS:           s.OS,
		IP:           s.IP,
		Location:     s.Location,
		LastActiveAt: s.LastActiveAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		IsCurrent:    s.ID == currentID,
		RevokedAt:    nil,
	}
}

func (h *SessionHandler) currentSessionID(r *http.Request) string {
	if v, ok := r.Context().Value("session_id").(string); ok {
		return v
	}
	return ""
}

// List handles GET /me/sessions.
func (h *SessionHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value("user_id").(string)
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	currentID := h.currentSessionID(r)

	rows, _, err := h.sessionService.List(r.Context(), userID, currentID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to list sessions")
		return
	}
	out := make([]sessionDTO, 0, len(rows))
	for _, s := range rows {
		out = append(out, toSessionDTO(s, currentID))
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"data": out})
}

// Revoke handles DELETE /me/sessions/{id}.
func (h *SessionHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value("user_id").(string)
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "session id required")
		return
	}
	currentID := h.currentSessionID(r)
	if err := h.sessionService.Revoke(r.Context(), userID, id, currentID); err != nil {
		switch {
		case errors.Is(err, service.ErrCannotRevokeCurrent):
			response.Error(w, http.StatusUnprocessableEntity, "cannot revoke the current session via this endpoint")
		case errors.Is(err, service.ErrSessionNotFound):
			response.Error(w, http.StatusNotFound, "session not found")
		default:
			response.Error(w, http.StatusInternalServerError, "failed to revoke session")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RevokeOthers handles DELETE /me/sessions/others.
func (h *SessionHandler) RevokeOthers(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value("user_id").(string)
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	currentID := h.currentSessionID(r)
	if _, err := h.sessionService.RevokeAllOthers(r.Context(), userID, currentID); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to revoke sessions")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
