package dto

type TestNotificationPayload struct {
	ResourceID string `json:"resource_id" binding:"required"`
}
