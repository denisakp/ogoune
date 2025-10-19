package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
)

// GoogleChatNotifier sends notifications to Google Chat via webhook.
type GoogleChatNotifier struct {
	client *http.Client
}

// GoogleChatConfig holds the configuration for Google Chat notifications.
type GoogleChatConfig struct {
	WebhookURL string `json:"webhook_url"`
}

// NewGoogleChatNotifier creates a new Google Chat notifier instance.
func NewGoogleChatNotifier() Notifier {
	return &GoogleChatNotifier{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Send sends a notification to Google Chat using the configured webhook URL.
func (n *GoogleChatNotifier) Send(ctx context.Context, integration domain.Integration, incident domain.Incident) error {
	// Parse the integration config to get the webhook URL
	var config GoogleChatConfig
	if err := json.Unmarshal(integration.Config, &config); err != nil {
		return fmt.Errorf("failed to parse google chat config: %w", err)
	}

	if config.WebhookURL == "" {
		return fmt.Errorf("webhook_url is required in google chat integration config")
	}

	// Determine the status message based on the incident resolution status
	status := "DOWN"
	emoji := "🔴"
	if incident.ResolvedAt != nil {
		status = "UP"
		emoji = "🟢"
	}

	// Build the Google Chat message payload
	message := map[string]interface{}{
		"text": fmt.Sprintf(
			"%s *%s* | Resource: %s\n"+
				"Status: *%s*\n"+
				"Target: %s\n"+
				"Reason: %s\n"+
				"Cause: %s\n"+
				"Incident ID: %s",
			emoji,
			status,
			incident.Resource.Name,
			status,
			incident.Resource.Target,
			incident.Reason,
			incident.Cause,
			incident.ID,
		),
	}

	// Marshal the message to JSON
	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal google chat message: %w", err)
	}

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", config.WebhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create google chat request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send google chat notification: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("google chat webhook returned non-success status: %d", resp.StatusCode)
	}

	return nil
}
