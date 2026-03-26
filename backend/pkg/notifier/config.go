package notifier

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type stringOrNumber string

func (s *stringOrNumber) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*s = ""
		return nil
	}

	var asString string
	if err := json.Unmarshal(data, &asString); err == nil {
		*s = stringOrNumber(asString)
		return nil
	}

	var asInt int
	if err := json.Unmarshal(data, &asInt); err == nil {
		*s = stringOrNumber(strconv.Itoa(asInt))
		return nil
	}

	var asFloat float64
	if err := json.Unmarshal(data, &asFloat); err == nil {
		*s = stringOrNumber(strconv.FormatInt(int64(asFloat), 10))
		return nil
	}

	return fmt.Errorf("expected string or number")
}

// SMTPConfig represents the JSON configuration for an SMTP notification channel.
type SMTPConfig struct {
	Recipient  string         `json:"recipient"`
	Recipients []string       `json:"recipients,omitempty"`
	Sender     string         `json:"sender"`
	Host       string         `json:"host"`
	Port       stringOrNumber `json:"port"`
	User       string         `json:"user"`
	Username   string         `json:"username"`
	Password   string         `json:"password"`
}

// NewSMTPNotifierFromConfig deserializes a JSON config string and creates an SMTP notifier.
func NewSMTPNotifierFromConfig(configJSON string) (*SMTPNotifier, error) {
	var cfg SMTPConfig
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal SMTP config: %w", err)
	}

	recipient := cfg.Recipient
	if recipient == "" && len(cfg.Recipients) > 0 {
		recipient = cfg.Recipients[0]
	}

	user := cfg.User
	if user == "" {
		user = cfg.Username
	}

	port := string(cfg.Port)

	// Validate required fields
	if recipient == "" {
		return nil, fmt.Errorf("SMTP config missing recipient")
	}
	if cfg.Sender == "" {
		return nil, fmt.Errorf("SMTP config missing sender")
	}
	if cfg.Host == "" {
		return nil, fmt.Errorf("SMTP config missing host")
	}
	if port == "" {
		return nil, fmt.Errorf("SMTP config missing port")
	}
	if user == "" {
		return nil, fmt.Errorf("SMTP config missing user")
	}
	if cfg.Password == "" {
		return nil, fmt.Errorf("SMTP config missing password")
	}

	return NewSMTPNotifier(recipient, cfg.Sender, cfg.Host, port, user, cfg.Password), nil
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
