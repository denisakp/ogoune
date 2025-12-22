package notifier

import (
	"encoding/json"
	"fmt"
)

// SMTPConfig represents the JSON configuration for an SMTP notification channel.
type SMTPConfig struct {
	Recipient string `json:"recipient"`
	Sender    string `json:"sender"`
	Host      string `json:"host"`
	Port      string `json:"port"`
	User      string `json:"user"`
	Password  string `json:"password"`
}

// NewSMTPNotifierFromConfig deserializes a JSON config string and creates an SMTP notifier.
func NewSMTPNotifierFromConfig(configJSON string) (*SMTPNotifier, error) {
	var cfg SMTPConfig
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal SMTP config: %w", err)
	}

	// Validate required fields
	if cfg.Recipient == "" {
		return nil, fmt.Errorf("SMTP config missing recipient")
	}
	if cfg.Sender == "" {
		return nil, fmt.Errorf("SMTP config missing sender")
	}
	if cfg.Host == "" {
		return nil, fmt.Errorf("SMTP config missing host")
	}
	if cfg.Port == "" {
		return nil, fmt.Errorf("SMTP config missing port")
	}
	if cfg.User == "" {
		return nil, fmt.Errorf("SMTP config missing user")
	}
	if cfg.Password == "" {
		return nil, fmt.Errorf("SMTP config missing password")
	}

	return NewSMTPNotifier(cfg.Recipient, cfg.Sender, cfg.Host, cfg.Port, cfg.User, cfg.Password), nil
}

// WebhookConfig represents the JSON configuration for a Webhook notification channel.
type WebhookConfig struct {
	URL    string `json:"url"`
	Secret string `json:"secret,omitempty"`
}

// NewWebhookNotifierFromConfig deserializes a JSON config string and creates a webhook notifier.
func NewWebhookNotifierFromConfig(configJSON string) (*WebHookNotifier, error) {
	var cfg WebhookConfig
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal webhook config: %w", err)
	}

	// Validate required fields
	if cfg.URL == "" {
		return nil, fmt.Errorf("webhook config missing URL")
	}

	var secret *string
	if cfg.Secret != "" {
		secret = &cfg.Secret
	}

	return NewWebHookNotifier(cfg.URL, secret), nil
}
