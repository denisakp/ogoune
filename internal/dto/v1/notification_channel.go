package v1

import "encoding/json"

// ChannelResponse is the v1 API representation of a notification channel.
// Sensitive config values (passwords, tokens) MUST be omitted.
// @name ChannelResponse
type ChannelResponse struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Config    json.RawMessage `json:"config"`
	IsDefault bool            `json:"is_default"`
	IsEnabled bool            `json:"is_enabled"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
}

// CreateChannelRequest is the request body for POST /api/v1/notification-channels.
// @name CreateChannelRequest
type CreateChannelRequest struct {
	Name      string          `json:"name"`
	Type      string          `json:"type"`
	Config    json.RawMessage `json:"config"`
	IsDefault bool            `json:"is_default"`
	IsEnabled *bool           `json:"is_enabled,omitempty"` // defaults to true if omitted
}

// UpdateChannelRequest is the request body for PUT /api/v1/notification-channels/:id.
// All fields are optional (PATCH semantics).
// @name UpdateChannelRequest
type UpdateChannelRequest struct {
	Name      *string         `json:"name,omitempty"`
	Type      *string         `json:"type,omitempty"`
	Config    json.RawMessage `json:"config,omitempty"`
	IsDefault *bool           `json:"is_default,omitempty"`
	IsEnabled *bool           `json:"is_enabled,omitempty"`
}
