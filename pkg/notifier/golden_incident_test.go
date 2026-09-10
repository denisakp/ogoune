package notifier

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/domain"
)

// Golden fixtures for the down-alert payloads (spec 091, FR-012 / SC-003).
//
// These were captured BEFORE the causal narrative existed, and they are the only
// admissible evidence that adding it left the silent case untouched. A test
// asserting that some Explanation field is nil proves nothing about what an
// operator receives; these bytes do.
//
// To re-capture deliberately: UPDATE_GOLDEN=1 go test ./pkg/notifier/ -run Golden
// Never re-capture to make a failing test pass -- a diff here means the payload
// changed for every operator with no correlated event, which is the one thing
// this feature promised not to do.

// goldenIncident is fixed in every respect that reaches a rendered payload, so
// the bytes depend on the code and never on the clock.
func goldenIncident() *domain.Incident {
	started := time.Date(2026, 9, 10, 14, 3, 0, 0, time.UTC)
	return &domain.Incident{
		Base:       domain.Base{ID: "01JGOLDENINCIDENT00000000"},
		ResourceID: "01JGOLDENRESOURCE00000000",
		Resource: domain.Resource{
			Base:   domain.Base{ID: "01JGOLDENRESOURCE00000000"},
			Name:   "web-01 storefront",
			Target: "https://shop.example.com/health",
			Type:   domain.ResourceHTTP,
		},
		Cause:     "HTTP check failed: 502 Bad Gateway",
		StartedAt: started,
	}
}

// goldenExplanation is the correlated counterpart of goldenIncident.
func goldenExplanation() *domain.IncidentExplanation {
	inc := goldenIncident()
	event := domain.HostEvent{
		Base:        domain.Base{ID: "01JGOLDENEVENT0000000000"},
		HostID:      "01JGOLDENHOST00000000000",
		OccurredAt:  inc.StartedAt.Add(-13 * time.Second),
		Kind:        "oom_kill",
		Source:      "kmsg",
		Occurrences: 1,
		Detail:      &domain.HostEventDetail{Process: "postgres", PID: 4711},
	}
	return &domain.IncidentExplanation{
		HostID:      event.HostID,
		HostName:    "web-01",
		IncidentAt:  inc.StartedAt,
		Cause:       inc.Cause,
		Event:       event,
		Precedes:    true,
		OtherEvents: 2,
		WindowFrom:  inc.StartedAt.Add(-domain.HostContextWindowBefore),
		WindowTo:    inc.StartedAt.Add(domain.HostContextWindowAfter),
	}
}

func assertGolden(t *testing.T, name string, got []byte) {
	t.Helper()
	path := filepath.Join("testdata", name)

	if os.Getenv("UPDATE_GOLDEN") == "1" {
		require.NoError(t, os.WriteFile(path, got, 0o644))
		t.Logf("golden updated: %s", path)
		return
	}

	want, err := os.ReadFile(path)
	require.NoErrorf(t, err, "missing golden %s -- capture it with UPDATE_GOLDEN=1 before changing the renderer", path)
	require.Equalf(t, string(want), string(got),
		"%s changed. For an incident with no correlated event the payload must be byte-identical to the pre-feature one (FR-012, SC-003)", name)
}

func TestGoldenDownEmail_NoExplanation(t *testing.T) {
	n := &SMTPNotifier{}
	subject, body := n.incidentEmailContent(*goldenIncident(), nil)

	assertGolden(t, "down_email_subject.golden", []byte(subject))
	assertGolden(t, "down_email_body.golden.html", []byte(body))
}

// captureWebhook posts a payload at a local server and returns the raw body.
func captureWebhook(t *testing.T, payload NotificationPayload) []byte {
	t.Helper()
	var captured []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("reading webhook body: %v", err)
		}
		captured = b
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := NewWebHookNotifier(srv.URL, nil)
	require.NoError(t, n.Send(context.Background(), payload))
	return captured
}

func TestGoldenDownWebhook_NoExplanation(t *testing.T) {
	assertGolden(t, "down_webhook.golden.json",
		captureWebhook(t, NotificationPayload{Incident: goldenIncident()}))
}

// T034 -- absent, not null. A key that appears with a null value is a body
// change, and a consumer's parser notices even when a human would not.
func TestGoldenDownWebhook_ExplanationKeyIsAbsentNotNull(t *testing.T) {
	raw := captureWebhook(t, NotificationPayload{Incident: goldenIncident()})

	var body map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(raw, &body))
	_, present := body["explanation"]
	assert.False(t, present, "the key must not exist at all when there is nothing to say")
	assert.NotContains(t, string(raw), "explanation")
}

// T036 -- an explanation that arrives late changes the incident page, never the
// notifications. This is the notifier half: a payload built without one renders
// without one, whatever was stored afterwards.
func TestGoldenDownEmail_ExplanationKeyIsAbsentFromSubject(t *testing.T) {
	n := &SMTPNotifier{}
	withoutSubject, _ := n.incidentEmailContent(*goldenIncident(), nil)
	withSubject, _ := n.incidentEmailContent(*goldenIncident(), goldenExplanation())
	assert.Equal(t, withoutSubject, withSubject,
		"an operator triages a mailbox by subject line; correlation must not rewrite it")
}

// --- T035: what the renderers say when there IS something to say -------------

func TestDownEmail_CarriesTheExplanation(t *testing.T) {
	n := &SMTPNotifier{}
	_, body := n.incidentEmailContent(*goldenIncident(), goldenExplanation())

	assert.Contains(t, body, "What happened around it")
	assert.Contains(t, body, "OOM-killed postgres (pid 4711) on web-01")
	assert.Contains(t, body, "2026-09-10 14:03:00 UTC", "the failure time is rendered")
	assert.Contains(t, body, "2026-09-10 14:02:47 UTC", "and so is the kernel's")
	assert.Contains(t, body, "2 other kernel events")
	assert.Contains(t, body, "not a proven cause",
		"the email must state what the correlation is, and is not")

	// The pre-existing content is still all there: this adds, it does not replace.
	assert.Contains(t, body, "Resource Down Alert")
	assert.Contains(t, body, goldenIncident().Cause)
}

func TestDownWebhook_CarriesTheExplanation(t *testing.T) {
	raw := captureWebhook(t, NotificationPayload{
		Incident:    goldenIncident(),
		Explanation: goldenExplanation(),
	})

	var body struct {
		Type        string `json:"type"`
		Status      string `json:"status"`
		Message     string `json:"message"`
		Explanation struct {
			Text             string `json:"text"`
			HostID           string `json:"host_id"`
			HostName         string `json:"host_name"`
			EventKind        string `json:"event_kind"`
			EventAt          string `json:"event_at"`
			EventProcess     string `json:"event_process"`
			EventPID         int    `json:"event_pid"`
			EventOccurrences int    `json:"event_occurrences"`
			Precedes         bool   `json:"precedes"`
			OtherEvents      int    `json:"other_events"`
		} `json:"explanation"`
	}
	require.NoError(t, json.Unmarshal(raw, &body))

	// The pre-existing keys keep their meaning.
	assert.Equal(t, "incident", body.Type)
	assert.Equal(t, "DOWN", body.Status)
	assert.Equal(t, goldenIncident().Cause, body.Message)

	e := body.Explanation
	assert.Equal(t, "web-01", e.HostName)
	assert.Equal(t, "oom_kill", e.EventKind)
	assert.Equal(t, "2026-09-10T14:02:47Z", e.EventAt)
	assert.Equal(t, "postgres", e.EventProcess)
	assert.Equal(t, 4711, e.EventPID)
	assert.Equal(t, 1, e.EventOccurrences)
	assert.True(t, e.Precedes)
	assert.Equal(t, 2, e.OtherEvents)
	assert.Contains(t, e.Text, "OOM-killed postgres (pid 4711) on web-01")
	// T040 -- both timestamps reach every surface, this one included. They are
	// what let a consumer, or the human reading a forwarded message, overrule
	// the sentence.
	assert.Contains(t, e.Text, "2026-09-10 14:03:00 UTC")
	assert.Contains(t, e.Text, "2026-09-10 14:02:47 UTC")
}

// The two renderers must agree about which event was named: an operator reading
// the email and a system reading the webhook are looking at one incident.
func TestRenderers_AgreeOnTheNamedEvent(t *testing.T) {
	expl := goldenExplanation()

	n := &SMTPNotifier{}
	_, email := n.incidentEmailContent(*goldenIncident(), expl)
	raw := captureWebhook(t, NotificationPayload{Incident: goldenIncident(), Explanation: expl})

	var body struct {
		Explanation struct {
			Text string `json:"text"`
		} `json:"explanation"`
	}
	require.NoError(t, json.Unmarshal(raw, &body))
	assert.Contains(t, email, body.Explanation.Text,
		"the webhook's sentence is the email's sentence")
}

// A kind this version cannot phrase must not produce an object with a hole in
// it: no sentence means no explanation at all (FR-007, SC-009).
func TestRenderers_UnphrasableKindProducesNothing(t *testing.T) {
	expl := goldenExplanation()
	expl.Event.Kind = "something_new"

	n := &SMTPNotifier{}
	_, silent := n.incidentEmailContent(*goldenIncident(), nil)
	_, unphrasable := n.incidentEmailContent(*goldenIncident(), expl)
	assert.Equal(t, silent, unphrasable, "the email is exactly the silent one")

	raw := captureWebhook(t, NotificationPayload{Incident: goldenIncident(), Explanation: expl})
	assert.NotContains(t, string(raw), "explanation")
}

// SC-004, on the notification side: the wording may interpret, the payload may
// not assert.
func TestRenderers_ClaimNoMechanism(t *testing.T) {
	expl := goldenExplanation()
	raw := captureWebhook(t, NotificationPayload{Incident: goldenIncident(), Explanation: expl})

	var keyed map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(raw, &keyed))
	nested, ok := keyed["explanation"]
	require.True(t, ok)
	var inner map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(nested, &inner))

	for _, forbidden := range []string{
		"cause_id", "caused_by", "score", "confidence", "probability", "severity", "certainty",
	} {
		_, present := inner[forbidden]
		assert.Falsef(t, present, "no field may assert a mechanism: %q", forbidden)
	}
}
