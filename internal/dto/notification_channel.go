package dto

import (
	"encoding/json"

	"github.com/denisakp/ogoune/internal/domain"
)

// SMTPConfig represents SMTP-specific configuration
type SMTPConfig struct {
	Host       string   `json:"host"`
	Port       int      `json:"port"`
	Username   string   `json:"username"`
	Password   string   `json:"password"`
	Sender     string   `json:"sender"`
	Recipients []string `json:"recipients"`
	CC         []string `json:"cc,omitempty"`
	BCC        []string `json:"bcc,omitempty"`
	Subject    string   `json:"subject,omitempty"`
}

// SlackConfig represents Slack-specific configuration
type SlackConfig struct {
	WebhookURL string `json:"webhook_url"`
	Channel    string `json:"channel,omitempty"`
	Username   string `json:"username,omitempty"`
}

// SMSConfig represents SMS-specific configuration
type SMSConfig struct {
	Provider   string   `json:"provider"`
	AccountSID string   `json:"account_sid,omitempty"`
	AuthToken  string   `json:"auth_token,omitempty"`
	FromNumber string   `json:"from_number"`
	ToNumbers  []string `json:"to_numbers"`
}

// CreateNotificationChannelPayload contains fields for creating a notification channel
type CreateNotificationChannelPayload struct {
	Name             string                         `json:"name" binding:"required"`
	Type             domain.NotificationChannelType `json:"type" binding:"required"`
	Config           json.RawMessage                `json:"config" binding:"required"`
	EnabledByDefault bool                           `json:"enabled_by_default"`
}

// UpdateNotificationChannelPayload contains fields for updating a notification channel
type UpdateNotificationChannelPayload struct {
	Name             *string                         `json:"name,omitempty"`
	Type             *domain.NotificationChannelType `json:"type,omitempty"`
	Config           json.RawMessage                 `json:"config,omitempty"`
	EnabledByDefault *bool                           `json:"enabled_by_default,omitempty"`
}

// TestNotificationChannelConfigPayload contains only the fields needed to test a channel config
type TestNotificationChannelConfigPayload struct {
	Type   domain.NotificationChannelType `json:"type" binding:"required"`
	Config json.RawMessage                `json:"config" binding:"required"`
}

// NotificationChannelResponse represents the notification channel response with parsed config
type NotificationChannelResponse struct {
	ID               string                         `json:"id"`
	Name             string                         `json:"name"`
	Type             domain.NotificationChannelType `json:"type"`
	Config           map[string]interface{}         `json:"config"`
	EnabledByDefault bool                           `json:"enabled_by_default"`
	CreatedAt        string                         `json:"created_at"`
	UpdatedAt        string                         `json:"updated_at"`
}

// ToNotificationChannelResponse converts a domain NotificationChannel to DTO response
func ToNotificationChannelResponse(channel *domain.NotificationChannel) (*NotificationChannelResponse, error) {
	var configMap map[string]interface{}
	if err := json.Unmarshal(channel.Config, &configMap); err != nil {
		return nil, err
	}

	return &NotificationChannelResponse{
		ID:               channel.ID,
		Name:             channel.Name,
		Type:             channel.Type,
		Config:           configMap,
		EnabledByDefault: channel.EnabledByDefault,
		CreatedAt:        channel.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        channel.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
