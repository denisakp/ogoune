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

// AnnouncementV1ServiceInterface is the slice of *AnnouncementService used by the handler.
type AnnouncementV1ServiceInterface interface {
	ListActive(ctx context.Context) ([]*domain.Announcement, error)
	Create(ctx context.Context, in *domain.Announcement) (*domain.Announcement, error)
	Delete(ctx context.Context, id string) error
}

// AnnouncementHandler exposes /api/v1/announcements (operator banners).
type AnnouncementHandler struct {
	service AnnouncementV1ServiceInterface
}

func NewAnnouncementHandler(svc AnnouncementV1ServiceInterface) *AnnouncementHandler {
	return &AnnouncementHandler{service: svc}
}

// List handles GET /api/v1/announcements.
//
// @Summary  List active announcement banners
// @Tags     announcements
// @Security BearerAuth
// @Produce  json
// @Success  200 {object} map[string]interface{}
// @Failure  401 {object} dtoV1.ErrorResponse
// @Router   /announcements [get]
func (h *AnnouncementHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListActive(r.Context())
	if err != nil {
		respondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list announcements")
		return
	}
	data := make([]dtoV1.AnnouncementResponse, len(items))
	for i, a := range items {
		data[i] = mapAnnouncement(a)
	}
	respond(w, http.StatusOK, data)
}

// Create handles POST /api/v1/announcements.
//
// @Summary  Publish an announcement banner
// @Tags     announcements
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    body body dtoV1.CreateAnnouncementRequest true "Announcement"
// @Success  201 {object} dtoV1.SingleResponse[dtoV1.AnnouncementResponse]
// @Failure  403 {object} dtoV1.ErrorResponse
// @Failure  422 {object} dtoV1.ErrorResponse
// @Router   /announcements [post]
func (h *AnnouncementHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dtoV1.CreateAnnouncementRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	dismissible := true
	if req.Dismissible != nil {
		dismissible = *req.Dismissible
	}
	created, err := h.service.Create(r.Context(), &domain.Announcement{
		Severity:    domain.AnnouncementSeverity(req.Severity),
		Title:       req.Title,
		Description: req.Description,
		Dismissible: dismissible,
	})
	if err != nil {
		if errors.Is(err, service.ErrAnnouncementValidation) {
			respondError(w, r, http.StatusUnprocessableEntity, "VALIDATION_FAILED", err.Error())
			return
		}
		respondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create announcement")
		return
	}
	respond(w, http.StatusCreated, mapAnnouncement(created))
}

// Delete handles DELETE /api/v1/announcements/{id}.
//
// @Summary  Retract an announcement banner
// @Tags     announcements
// @Security BearerAuth
// @Param    id path string true "Announcement ID"
// @Success  204
// @Failure  403 {object} dtoV1.ErrorResponse
// @Failure  404 {object} dtoV1.ErrorResponse
// @Router   /announcements/{id} [delete]
func (h *AnnouncementHandler) Delete(w http.ResponseWriter, r *http.Request) {
	err := h.service.Delete(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		if errors.Is(err, service.ErrAnnouncementNotFound) {
			respondError(w, r, http.StatusNotFound, "ANNOUNCEMENT_NOT_FOUND", "announcement not found")
			return
		}
		respondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to delete announcement")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func mapAnnouncement(a *domain.Announcement) dtoV1.AnnouncementResponse {
	return dtoV1.AnnouncementResponse{
		ID:          a.ID,
		Severity:    string(a.Severity),
		Title:       a.Title,
		Description: a.Description,
		Dismissible: a.Dismissible,
		CreatedAt:   a.CreatedAt.UTC().Format(time.RFC3339),
	}
}
