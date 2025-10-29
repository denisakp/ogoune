package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/denisakp/pulseguard/internal/api/response"
	"github.com/denisakp/pulseguard/internal/dto"
)

// NotificationServiceInterface defines the methods required by NotificationHandler.
type NotificationServiceInterface interface {
	TestNotification(ctx context.Context, resourceID string) error
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
