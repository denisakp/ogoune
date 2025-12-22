package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/denisakp/pulseguard/internal/api/response"
	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/dto"
	"github.com/go-chi/chi/v5"
)

// NotificationServiceInterface defines the methods required by NotificationHandler.
type NotificationServiceInterface interface {
	TestNotification(ctx context.Context, resourceID string) error
	CreateNotificationChannel(ctx context.Context, payload *dto.CreateNotificationChannelPayload) (*domain.NotificationChannel, error)
	GetNotificationChannel(ctx context.Context, id string) (*domain.NotificationChannel, error)
	ListNotificationChannels(ctx context.Context, limit, offset int) ([]*domain.NotificationChannel, error)
	UpdateNotificationChannel(ctx context.Context, id string, payload *dto.UpdateNotificationChannelPayload) (*domain.NotificationChannel, error)
	DeleteNotificationChannel(ctx context.Context, id string) error
	TestNotificationChannel(ctx context.Context, id string) error
	ValidateAndTestChannelConfig(ctx context.Context, channelType domain.NotificationChannelType, config json.RawMessage) error
}

// NotificationHandler handles HTTP requests for notification operations.
type NotificationHandler struct {
	notificationService NotificationServiceInterface
}

// NewNotificationHandler creates a new NotificationHandler with injected dependencies.
func NewNotificationHandler(notificationService NotificationServiceInterface) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// TestNotification handles POST /notifications/test - sends a test notification for a resource.
func (h *NotificationHandler) TestNotification(w http.ResponseWriter, r *http.Request) {
	var payload dto.TestNotificationPayload

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	if payload.ResourceID == "" {
		response.Error(w, http.StatusBadRequest, "ResourceID is required")
		return
	}

	resourceID := payload.ResourceID

	if err := h.notificationService.TestNotification(r.Context(), resourceID); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to send test notification: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Test notification sent successfully",
	})
}

// CreateNotificationChannel handles POST /notification-channels - creates a new notification channel.
func (h *NotificationHandler) CreateNotificationChannel(w http.ResponseWriter, r *http.Request) {
	var payload dto.CreateNotificationChannelPayload

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	channel, err := h.notificationService.CreateNotificationChannel(r.Context(), &payload)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to create notification channel: "+err.Error())
		return
	}

	responseDTO, err := dto.ToNotificationChannelResponse(channel)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to convert response: "+err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, responseDTO)
}

// GetNotificationChannel handles GET /notification-channels/{id} - retrieves a notification channel by ID.
func (h *NotificationHandler) GetNotificationChannel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Channel ID is required")
		return
	}

	channel, err := h.notificationService.GetNotificationChannel(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "Notification channel not found: "+err.Error())
		return
	}

	responseDTO, err := dto.ToNotificationChannelResponse(channel)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to convert response: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, responseDTO)
}

// ListNotificationChannels handles GET /notification-channels - retrieves all notification channels.
func (h *NotificationHandler) ListNotificationChannels(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	limit := 50
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	channels, err := h.notificationService.ListNotificationChannels(r.Context(), limit, offset)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to retrieve notification channels: "+err.Error())
		return
	}

	// Convert to response DTOs
	responseList := make([]*dto.NotificationChannelResponse, 0, len(channels))
	for _, channel := range channels {
		responseDTO, err := dto.ToNotificationChannelResponse(channel)
		if err != nil {
			response.Error(w, http.StatusInternalServerError, "Failed to convert response: "+err.Error())
			return
		}
		responseList = append(responseList, responseDTO)
	}

	response.JSON(w, http.StatusOK, responseList)
}

// UpdateNotificationChannel handles PATCH /notification-channels/{id} - updates a notification channel.
func (h *NotificationHandler) UpdateNotificationChannel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Channel ID is required")
		return
	}

	var payload dto.UpdateNotificationChannelPayload

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	channel, err := h.notificationService.UpdateNotificationChannel(r.Context(), id, &payload)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to update notification channel: "+err.Error())
		return
	}

	responseDTO, err := dto.ToNotificationChannelResponse(channel)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to convert response: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, responseDTO)
}

// DeleteNotificationChannel handles DELETE /notification-channels/{id} - deletes a notification channel.
func (h *NotificationHandler) DeleteNotificationChannel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Channel ID is required")
		return
	}

	if err := h.notificationService.DeleteNotificationChannel(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to delete notification channel: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Notification channel deleted successfully",
	})
}

// TestNotificationChannelConfig handles POST /notification-channels/{id}/test - sends a test notification.
func (h *NotificationHandler) TestNotificationChannelConfig(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Channel ID is required")
		return
	}

	if err := h.notificationService.TestNotificationChannel(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to send test notification: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Test notification sent successfully",
	})
}

// ValidateAndTestChannelConfig handles POST /notification-channels/test-config - validates and tests channel config without saving.
func (h *NotificationHandler) ValidateAndTestChannelConfig(w http.ResponseWriter, r *http.Request) {
	var payload dto.TestNotificationChannelConfigPayload

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	if err := h.notificationService.ValidateAndTestChannelConfig(r.Context(), payload.Type, payload.Config); err != nil {
		response.Error(w, http.StatusBadRequest, "Configuration test failed: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Configuration validated and tested successfully",
	})
}
