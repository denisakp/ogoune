package notifier

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
)

// WebHookNotifier is a notifier that sends notifications via webhook.
type WebHookNotifier struct {
	client *http.Client
	url    string
	secret *string
}

// NewWebHookNotifier creates a new WebHookNotifier with the provided URL and optional secret.
func NewWebHookNotifier(url string, secret *string) *WebHookNotifier {
	return &WebHookNotifier{
		url:    url,
		secret: secret,
		client: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

// Send sends a notification via webhook.
func (n *WebHookNotifier) Send(ctx context.Context, incident domain.Incident) error {
	if n.url == "" {
		return fmt.Errorf("webhook url is empty")
	}

	status := "DOWN"
	if incident.ResolvedAt != nil {
		status = "UP"
	}

	payloadBytes, err := json.Marshal(map[string]interface{}{
		"status":  status,
		"message": incident.Cause,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// create http request
	req, err := http.NewRequestWithContext(ctx, "POST", n.url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// set the header signature if exists.
	if n.secret != nil && *n.secret != "" {
		// generate HMac signature for the payload
		mac := hmac.New(sha256.New, []byte(*n.secret))
		mac.Write(payloadBytes)
		signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

		req.Header.Set("X-PulseGuard-Signature", signature)
	}

	response, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook notification: %w", err)
	}
	defer response.Body.Close()

	// check the response status
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("webhook returned non-success status: %d", response.StatusCode)
	}

	return nil
}

func (n *WebHookNotifier) SendTestNotification(ctx context.Context) error {
	if n.url == "" {
		return fmt.Errorf("webhook url is empty")
	}

	payloadBytes, err := json.Marshal(map[string]interface{}{
		"status":  "TEST",
		"message": "Test notification",
	})
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// create http request
	req, err := http.NewRequestWithContext(ctx, "POST", n.url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// set the header signature if exists.
	if n.secret != nil && *n.secret != "" {
		// generate HMac signature for the payload
		mac := hmac.New(sha256.New, []byte(*n.secret))
		mac.Write(payloadBytes)
		signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

		req.Header.Set("X-PulseGuard-Signature", signature)
	}

	response, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook notification: %w", err)
	}
	defer response.Body.Close()

	// check the response status
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("webhook returned non-success status: %d", response.StatusCode)
	}

	return nil
}
