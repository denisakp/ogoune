package dto

import "github.com/denisakp/pulseguard/internal/domain"

type CreateIntegrationPayload struct {
	Name       string                 `json:"name" binding:"required"`
	Config     map[string]interface{} `json:"config" binding:"required"`
	EventTypes []domain.EventType     `json:"event_types" binding:"required"`
	IsActive   bool                   `json:"is_active" binding:"required"`
}

type UpdateIntegrationPayload struct {
	Name       string                 `json:"name" binding:"required"`
	Config     map[string]interface{} `json:"config" binding:"required"`
	EventTypes []domain.EventType     `json:"event_types" binding:"required"`
	IsActive   bool                   `json:"is_active" binding:"required"`
}
