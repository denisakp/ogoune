package dto

import (
	"time"

	"github.com/denisakp/ogoune/internal/domain"
)

// CreateAPIKeyRequest describes the payload to create a new API key.
type CreateAPIKeyRequest struct {
	Name      string             `json:"name"`
	Scope     domain.APIKeyScope `json:"scope"`
	ExpiresAt *time.Time         `json:"expires_at,omitempty"`
}

// CreateAPIKeyResponse is returned once when a new API key is created.
type CreateAPIKeyResponse struct {
	ID        string             `json:"id"`
	Name      string             `json:"name"`
	Key       string             `json:"key"`
	KeyPrefix string             `json:"key_prefix"`
	Scope     domain.APIKeyScope `json:"scope"`
	ExpiresAt *time.Time         `json:"expires_at,omitempty"`
	CreatedAt time.Time          `json:"created_at"`
}

// APIKeyListItem represents an API key returned by list endpoints.
type APIKeyListItem struct {
	ID         string             `json:"id"`
	Name       string             `json:"name"`
	KeyPrefix  string             `json:"key_prefix"`
	Scope      domain.APIKeyScope `json:"scope"`
	ExpiresAt  *time.Time         `json:"expires_at,omitempty"`
	LastUsedAt *time.Time         `json:"last_used_at,omitempty"`
	LastUsedIP string             `json:"last_used_ip"`
	IsActive   bool               `json:"is_active"`
	CreatedAt  time.Time          `json:"created_at"`
}
