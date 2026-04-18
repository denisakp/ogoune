package v1

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"errors"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/dto"
	dtoV1 "github.com/denisakp/ogoune/internal/dto/v1"
	"github.com/denisakp/ogoune/internal/service"
	"github.com/go-chi/chi/v5"
)

// ChannelV1ServiceInterface defines the service methods used by the v1 notification channel handler.
type ChannelV1ServiceInterface interface {
	ListNotificationChannels(ctx context.Context, limit, offset int) ([]*domain.NotificationChannel, error)
	GetNotificationChannel(ctx context.Context, id string) (*domain.NotificationChannel, error)
	CreateNotificationChannel(ctx context.Context, payload *dto.CreateNotificationChannelPayload) (*domain.NotificationChannel, error)
	UpdateNotificationChannel(ctx context.Context, id string, payload *dto.UpdateNotificationChannelPayload) (*domain.NotificationChannel, error)
	DeleteNotificationChannel(ctx context.Context, id string) error
}

// NotificationChannelHandler handles v1 CRUD endpoints for notification channels.
type NotificationChannelHandler struct {
	service ChannelV1ServiceInterface
}

// NewNotificationChannelHandler creates a new NotificationChannelHandler.
func NewNotificationChannelHandler(svc ChannelV1ServiceInterface) *NotificationChannelHandler {
	return &NotificationChannelHandler{service: svc}
}

// sensitiveConfigKeys is a set of config keys that must be stripped from responses.
var sensitiveConfigKeys = map[string]bool{
	"password":    true,
	"auth_token":  true,
	"token":       true,
	"account_sid": true,
	"secret":      true,
}

// mapChannelResponse maps a domain.NotificationChannel to a v1 ChannelResponse.
// Sensitive config fields are removed before returning.
func mapChannelResponse(ch *domain.NotificationChannel) dtoV1.ChannelResponse {
	var config json.RawMessage
	if len(ch.Config) > 0 {
		var configMap map[string]interface{}
		if err := json.Unmarshal(ch.Config, &configMap); err == nil {
			for key := range sensitiveConfigKeys {
				delete(configMap, key)
			}
			if sanitized, err := json.Marshal(configMap); err == nil {
				config = sanitized
			}
		}
	}
	if config == nil {
		config = json.RawMessage("{}")
	}

	return dtoV1.ChannelResponse{
		ID:        ch.ID,
		Type:      string(ch.Type),
		Config:    config,
		IsDefault: ch.EnabledByDefault,
		IsEnabled: ch.EnabledByDefault,
		CreatedAt: ch.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: ch.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

// List handles GET /api/v1/notification-channels
//
// @Summary     List notification channels
// @Tags        notification-channels
// @Security    BearerAuth
// @Produce     json
// @Param       page     query int false "Page number (default 1)"
// @Param       per_page query int false "Items per page (1-100, default 20)"
// @Success     200 {object} map[string]interface{}
// @Failure     401 {object} dtoV1.ErrorResponse
// @Router      /notification-channels [get]
func (h *NotificationChannelHandler) List(w http.ResponseWriter, r *http.Request) {
	params, errs := parsePagination(r)
	if len(errs) > 0 {
		respondError(w, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "invalid pagination parameters", errs...)
		return
	}

	offset := (params.Page - 1) * params.PerPage
	items, err := h.service.ListNotificationChannels(r.Context(), params.PerPage, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list channels")
		return
	}

	allItems, err := h.service.ListNotificationChannels(r.Context(), 10000, 0)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to count channels")
		return
	}
	total := len(allItems)

	data := make([]dtoV1.ChannelResponse, 0, len(items))
	for _, ch := range items {
		data = append(data, mapChannelResponse(ch))
	}

	respondPaginated(w, data, dtoV1.MetaResponse{
		Page:    params.Page,
		PerPage: params.PerPage,
		Total:   total,
	})
}

// Get handles GET /api/v1/notification-channels/{id}
//
// @Summary     Get a notification channel by ID
// @Tags        notification-channels
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "Channel ID"
// @Success     200 {object} dtoV1.SingleResponse[dtoV1.ChannelResponse]
// @Failure     404 {object} dtoV1.ErrorResponse
// @Router      /notification-channels/{id} [get]
func (h *NotificationChannelHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ch, err := h.service.GetNotificationChannel(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrResourceNotFound) {
			respondError(w, http.StatusNotFound, "RESOURCE_NOT_FOUND", "channel not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get channel")
		return
	}
	if ch == nil {
		respondError(w, http.StatusNotFound, "RESOURCE_NOT_FOUND", "channel not found")
		return
	}
	respond(w, http.StatusOK, mapChannelResponse(ch))
}

// Create handles POST /api/v1/notification-channels
//
// @Summary     Create a notification channel
// @Tags        notification-channels
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       body body dtoV1.CreateChannelRequest true "Channel payload"
// @Success     201 {object} dtoV1.SingleResponse[dtoV1.ChannelResponse]
// @Failure     422 {object} dtoV1.ErrorResponse
// @Failure     403 {object} dtoV1.ErrorResponse
// @Router      /notification-channels [post]
func (h *NotificationChannelHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dtoV1.CreateChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "invalid request body")
		return
	}

	var fieldErrs []dtoV1.FieldError
	if strings.TrimSpace(req.Name) == "" {
		fieldErrs = append(fieldErrs, dtoV1.FieldError{Field: "name", Message: "required"})
	}
	if strings.TrimSpace(req.Type) == "" {
		fieldErrs = append(fieldErrs, dtoV1.FieldError{Field: "type", Message: "required"})
	}
	if len(fieldErrs) > 0 {
		respondError(w, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "validation failed", fieldErrs...)
		return
	}

	channelType := domain.NotificationChannelType(req.Type)
	if !channelType.IsValid() {
		respondError(w, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "invalid channel type",
			dtoV1.FieldError{Field: "type", Message: "must be smtp, slack, or sms"})
		return
	}

	payload := &dto.CreateNotificationChannelPayload{
		Name:   req.Name,
		Type:   channelType,
		Config: req.Config,
	}

	created, err := h.service.CreateNotificationChannel(r.Context(), payload)
	if err != nil {
		if errors.Is(err, service.ErrValidationFailed) {
			respondError(w, http.StatusUnprocessableEntity, "VALIDATION_FAILED", err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create channel")
		return
	}
	respond(w, http.StatusCreated, mapChannelResponse(created))
}

// Update handles PUT /api/v1/notification-channels/{id}
//
// @Summary     Update a notification channel
// @Tags        notification-channels
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id   path string true "Channel ID"
// @Param       body body dtoV1.UpdateChannelRequest true "Update payload"
// @Success     200 {object} dtoV1.SingleResponse[dtoV1.ChannelResponse]
// @Failure     404 {object} dtoV1.ErrorResponse
// @Failure     403 {object} dtoV1.ErrorResponse
// @Router      /notification-channels/{id} [put]
func (h *NotificationChannelHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req dtoV1.UpdateChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "invalid request body")
		return
	}

	payload := &dto.UpdateNotificationChannelPayload{
		Name:   req.Name,
		Config: req.Config,
	}
	if req.Type != nil {
		t := domain.NotificationChannelType(*req.Type)
		payload.Type = &t
	}

	updated, err := h.service.UpdateNotificationChannel(r.Context(), id, payload)
	if err != nil {
		if errors.Is(err, service.ErrResourceNotFound) {
			respondError(w, http.StatusNotFound, "RESOURCE_NOT_FOUND", "channel not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update channel")
		return
	}
	respond(w, http.StatusOK, mapChannelResponse(updated))
}

// Delete handles DELETE /api/v1/notification-channels/{id}
//
// @Summary     Delete a notification channel
// @Tags        notification-channels
// @Security    BearerAuth
// @Param       id path string true "Channel ID"
// @Success     204
// @Failure     404 {object} dtoV1.ErrorResponse
// @Failure     403 {object} dtoV1.ErrorResponse
// @Router      /notification-channels/{id} [delete]
func (h *NotificationChannelHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.service.DeleteNotificationChannel(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrResourceNotFound) {
			respondError(w, http.StatusNotFound, "RESOURCE_NOT_FOUND", "channel not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to delete channel")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
