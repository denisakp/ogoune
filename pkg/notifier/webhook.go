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

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/narrative"
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
func (n *WebHookNotifier) Send(ctx context.Context, payload NotificationPayload) error {
	if n.url == "" {
		return fmt.Errorf("webhook url is empty")
	}

	var body map[string]any

	switch {
	case payload.Flapping != nil:
		flapping := payload.Flapping
		event := "flapping"
		body = map[string]any{
			"event":            event,
			"resource_id":      flapping.Resource.ID,
			"resource_name":    flapping.Resource.Name,
			"target":           flapping.Resource.Target,
			"transition_count": flapping.TransitionCount,
			"window_seconds":   flapping.WindowSeconds,
			"triggered_at":     flapping.TriggeredAt.Format(time.RFC3339),
		}
		if flapping.Stabilized {
			body["event"] = "flapping_stabilized"
			body["final_status"] = flapping.FinalStatus
		}
	case payload.Reminder != nil:
		reminder := payload.Reminder
		body = map[string]any{
			"event":               "reminder",
			"resource_id":         reminder.Resource.ID,
			"resource_name":       reminder.Resource.Name,
			"target":              reminder.Resource.Target,
			"incident_id":         reminder.Incident.ID,
			"incident_started_at": reminder.Incident.StartedAt.Format(time.RFC3339),
			"elapsed_minutes":     reminder.ElapsedMinutes,
			"triggered_at":        reminder.TriggeredAt.Format(time.RFC3339),
		}
	case payload.Component != nil:
		component := payload.Component
		impacted := make([]map[string]string, 0, len(component.Impacted))
		for _, r := range component.Impacted {
			impacted = append(impacted, map[string]string{
				"id":     r.ID,
				"name":   r.Name,
				"status": string(r.Status),
			})
		}

		body = map[string]any{
			"type":      "component",
			"component": component.Component.Name,
			"status":    component.Status,
			"impacted":  impacted,
		}
	case payload.Expiry != nil:
		expiry := payload.Expiry
		body = map[string]any{
			"type":           "expiry",
			"event_type":     "expiry",
			"resource_id":    expiry.Resource.ID,
			"resource_name":  expiry.Resource.Name,
			"resource_url":   expiry.Resource.Target,
			"expiry_type":    expiry.ExpiryType,
			"days_remaining": expiry.DaysRemaining,
			"expires_at":     expiry.ExpiresAt.Format(time.RFC3339),
			"issuer":         expiry.Issuer,
			"threshold":      expiry.Threshold,
			"triggered_at":   expiry.TriggeredAt.Format(time.RFC3339),
		}
	case payload.Incident != nil:
		incident := payload.Incident
		status := "DOWN"
		if incident.ResolvedAt != nil {
			status = "UP"
		}
		body = map[string]any{
			"type":    "incident",
			"status":  status,
			"message": incident.Cause,
		}
		// One key, added only when there is something to say. Absent rather than
		// null when there is not: a key appearing with a null value is a body
		// change, and downstream parsers notice (spec 091, FR-012).
		if e := explanationBody(payload.Explanation); e != nil {
			body["explanation"] = e
		}
		if incident.Resource.Type == domain.ResourceKeyword {
			if incident.Resource.Keyword != nil {
				body["keyword"] = *incident.Resource.Keyword
			}
			if incident.Resource.KeywordMode != nil {
				body["keyword_mode"] = *incident.Resource.KeywordMode
			}
			body["failure_cause"] = incident.Cause
		}
	case payload.Operator != nil:
		op := payload.Operator
		body = map[string]any{
			"type":  "operator",
			"title": op.Title,
			"body":  op.Body,
			"items": op.Items,
		}
	default:
		return fmt.Errorf("notification payload missing incident, component, expiry, flapping, reminder, or operator")
	}

	payloadBytes, err := json.Marshal(body)
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

		req.Header.Set("X-Ogoune-Signature", signature)
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

// explanationBody shapes the causal narrative for a webhook consumer, carrying
// both structured fields and the rendered sentence (spec 091).
//
// Both, because the two kinds of consumer want different things: one forwards to
// a chat channel and needs prose, another feeds a system and needs fields.
// Making either parse the other's format would be a worse contract than sending
// twelve extra bytes.
//
// Returns nil when there is nothing to say, or when the event's kind is one this
// version cannot phrase -- an object whose text came out empty would be worse
// than no object, because a consumer templating on it renders a gap.
func explanationBody(e *domain.IncidentExplanation) map[string]any {
	if e == nil {
		return nil
	}
	text := narrative.Sentence(e)
	if text == "" {
		return nil
	}

	out := map[string]any{
		"text":              text,
		"host_id":           e.HostID,
		"host_name":         e.HostName,
		"event_kind":        e.Event.Kind,
		"event_at":          e.Event.OccurredAt.UTC().Format(time.RFC3339),
		"event_occurrences": e.Event.Occurrences,
		// precedes compares two timestamps. It is not a causal claim, and there
		// is deliberately no field here that would be one.
		"precedes":     e.Precedes,
		"other_events": e.OtherEvents,
	}
	// The kernel does not always name a process, and a key present with an empty
	// value would read as "named nothing" rather than "did not say".
	if e.Event.Detail != nil {
		if p := e.Event.Detail.Process; p != "" {
			out["event_process"] = p
		}
		if pid := e.Event.Detail.PID; pid > 0 {
			out["event_pid"] = pid
		}
	}
	return out
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

		req.Header.Set("X-Ogoune-Signature", signature)
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
