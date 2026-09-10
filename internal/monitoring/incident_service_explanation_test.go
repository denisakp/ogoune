package monitoring

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/correlation"
	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/fake"
)

// explanationRig is the standard monitoring fixture plus a webhook channel whose
// bodies are captured, and a correlator over in-memory host repositories.
type explanationRig struct {
	svc      *IncidentService
	events   *fake.HostEventFake
	hosts    *fake.HostFake
	steps    *fake.IncidentEventStepFake
	bodies   *[][]byte
	resource *domain.Resource
	close    func()
}

func newExplanationRig(t *testing.T, hostID string, attach bool) *explanationRig {
	t.Helper()
	ctx := context.Background()

	svc, _, steps, _, channelRepo, asynqClient := setupTestService()

	var bodies [][]byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("reading webhook body: %v", err)
		}
		bodies = append(bodies, b)
		w.WriteHeader(http.StatusOK)
	}))

	channel := &domain.NotificationChannel{
		Base:   domain.Base{ID: "notif-expl"},
		Name:   "expl",
		Type:   domain.NotificationChannelType("webhook"),
		Config: []byte(`{"url":"` + server.URL + `"}`),
	}
	require.NoError(t, channelRepo.Create(ctx, channel))

	events := fake.NewHostEventFake()
	hosts := fake.NewHostFake()
	if hostID != "" {
		require.NoError(t, hosts.Create(ctx, &domain.Host{Base: domain.Base{ID: hostID}, Name: "web-01"}))
		hosts.FindByIDCalls = 0
	}
	svc = svc.WithCorrelator(correlation.New(events, hosts))

	resource := &domain.Resource{
		Base:     domain.Base{ID: "res-expl"},
		Name:     "storefront",
		Target:   "https://shop.example.com/health",
		Type:     domain.ResourceHTTP,
		Status:   domain.StatusDown,
		Interval: 60,
		Timeout:  10,
		IsActive: true,
	}
	if attach {
		id := hostID
		resource.HostID = &id
	}
	channelRepo.AssociateChannelWithResource(resource.ID, channel.ID)

	return &explanationRig{
		svc: svc, events: events, hosts: hosts, steps: steps,
		bodies: &bodies, resource: resource,
		close: func() { server.Close(); asynqClient.Close() },
	}
}

func (r *explanationRig) seedEvent(t *testing.T, hostID, id, kind string, at time.Time) {
	t.Helper()
	require.NoError(t, r.events.Create(context.Background(), &domain.HostEvent{
		Base:        domain.Base{ID: id},
		HostID:      hostID,
		OccurredAt:  at,
		Kind:        kind,
		Source:      "kmsg",
		Occurrences: 1,
		Detail:      &domain.HostEventDetail{Process: "postgres", PID: 4711},
	}))
}

func (r *explanationRig) trigger(t *testing.T) {
	t.Helper()
	require.NoError(t, r.svc.CreateIncident(context.Background(),
		r.resource, domain.CheckResult{Status: "down", ResponseData: "502 Bad Gateway"}))
}

func explanationOf(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	var body map[string]any
	require.NoError(t, json.Unmarshal(raw, &body))
	e, _ := body["explanation"].(map[string]any)
	return e
}

// The alert carries the sentence when the event had already landed -- which the
// confirmation window makes the normal case (SC-002).
func TestCreateIncident_AlertCarriesTheExplanation(t *testing.T) {
	rig := newExplanationRig(t, "host-1", true)
	defer rig.close()
	// The incident's StartedAt is time.Now() inside CreateIncident, so seed
	// relative to now rather than to a fixture instant.
	rig.seedEvent(t, "host-1", "e1", "oom_kill", time.Now().Add(-13*time.Second))

	rig.trigger(t)

	require.Len(t, *rig.bodies, 1)
	e := explanationOf(t, (*rig.bodies)[0])
	require.NotNil(t, e, "the alert must carry the explanation")
	assert.Equal(t, "oom_kill", e["event_kind"])
	assert.Equal(t, "web-01", e["host_name"])
	assert.Contains(t, e["text"], "OOM-killed postgres (pid 4711) on web-01")
}

// T037 -- the zero-cost guarantee on the dispatch path. Most monitors have no
// host, and this feature must not make their alerts cost a query.
func TestCreateIncident_NoHostIssuesNoCorrelationQuery(t *testing.T) {
	rig := newExplanationRig(t, "host-1", false)
	defer rig.close()
	// An event exists for the host, but the monitor is not attached to it.
	rig.seedEvent(t, "host-1", "e1", "oom_kill", time.Now().Add(-time.Second))

	rig.trigger(t)

	require.Len(t, *rig.bodies, 1)
	assert.Nil(t, explanationOf(t, (*rig.bodies)[0]))
	assert.NotContains(t, string((*rig.bodies)[0]), "explanation",
		"absent, not null")
	assert.Zero(t, rig.events.ListInWindowCalls, "no host attached must cost zero event reads")
	assert.Zero(t, rig.hosts.FindByIDCalls, "and zero host lookups")
}

// Resolved once per incident, not once per channel: the answer does not depend
// on the channel, and paying per channel would multiply the cost by a number
// the operator controls.
func TestCreateIncident_CorrelatesOncePerIncidentNotPerChannel(t *testing.T) {
	ctx := context.Background()
	svc, _, _, _, channelRepo, asynqClient := setupTestService()
	defer asynqClient.Close()

	var hits int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	resource := &domain.Resource{
		Base: domain.Base{ID: "res-multichannel"}, Name: "multi",
		Target: "https://example.com", Type: domain.ResourceHTTP,
		Status: domain.StatusDown, Interval: 60, Timeout: 10, IsActive: true,
	}
	hostID := "host-multi"
	resource.HostID = &hostID

	for _, id := range []string{"ch-1", "ch-2", "ch-3"} {
		ch := &domain.NotificationChannel{
			Base:   domain.Base{ID: id},
			Name:   id,
			Type:   domain.NotificationChannelType("webhook"),
			Config: []byte(`{"url":"` + server.URL + `"}`),
		}
		require.NoError(t, channelRepo.Create(ctx, ch))
		channelRepo.AssociateChannelWithResource(resource.ID, ch.ID)
	}

	events := fake.NewHostEventFake()
	hosts := fake.NewHostFake()
	require.NoError(t, hosts.Create(ctx, &domain.Host{Base: domain.Base{ID: hostID}, Name: "web-01"}))
	hosts.FindByIDCalls = 0
	require.NoError(t, events.Create(ctx, &domain.HostEvent{
		Base: domain.Base{ID: "e1"}, HostID: hostID,
		OccurredAt: time.Now().Add(-5 * time.Second), Kind: "oom_kill",
		Source: "kmsg", Occurrences: 1,
	}))

	svc = svc.WithCorrelator(correlation.New(events, hosts))
	require.NoError(t, svc.CreateIncident(ctx, resource,
		domain.CheckResult{Status: "down", ResponseData: "timeout"}))

	assert.Equal(t, 3, hits, "one notification per channel")
	assert.Equal(t, 1, events.ListInWindowCalls, "but one correlation for the incident")
	assert.Equal(t, 1, hosts.FindByIDCalls)
}

// T036 -- an event that lands after the alert changes the incident page and
// nothing else. No sequel notification, no extra event step (FR-011b, SC-012).
func TestCreateIncident_LateEventProducesNoSecondNotification(t *testing.T) {
	rig := newExplanationRig(t, "host-1", true)
	defer rig.close()

	rig.trigger(t)
	require.Len(t, *rig.bodies, 1)
	assert.Nil(t, explanationOf(t, (*rig.bodies)[0]), "nothing had landed yet")

	stepsBefore := rig.steps.FindByIncidentID(lastIncidentID(t, rig))

	// The kill actually preceded the failure; it just reached the backend after
	// the alert went out. Nothing in the system may react to that.
	rig.seedEvent(t, "host-1", "e-late", "oom_kill", time.Now().Add(-20*time.Second))

	assert.Len(t, *rig.bodies, 1, "no second notification of any kind")

	stepsAfter := rig.steps.FindByIncidentID(lastIncidentID(t, rig))
	assert.Len(t, stepsAfter, len(stepsBefore), "and no extra event step")
}

func lastIncidentID(t *testing.T, rig *explanationRig) string {
	t.Helper()
	steps, err := rig.steps.List(context.Background(), 100, 0)
	require.NoError(t, err)
	require.NotEmpty(t, steps)
	return steps[0].IncidentID
}

// T029a -- correlation must be invisible to the incident lifecycle (FR-015).
// Same monitor, same failure, with and without a correlator attached: the event
// step sequence has to match exactly.
func TestCreateIncident_LifecycleIsUnchangedByCorrelation(t *testing.T) {
	stepsFor := func(t *testing.T, withCorrelator bool) []string {
		t.Helper()
		ctx := context.Background()
		svc, incidents, steps, _, channelRepo, asynqClient := setupTestService()
		defer asynqClient.Close()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		ch := &domain.NotificationChannel{
			Base: domain.Base{ID: "ch-lifecycle"}, Name: "lifecycle",
			Type:   domain.NotificationChannelType("webhook"),
			Config: []byte(`{"url":"` + server.URL + `"}`),
		}
		require.NoError(t, channelRepo.Create(ctx, ch))

		hostID := "host-lifecycle"
		resource := &domain.Resource{
			Base: domain.Base{ID: "res-lifecycle"}, Name: "lifecycle",
			Target: "https://example.com", Type: domain.ResourceHTTP,
			Status: domain.StatusDown, Interval: 60, Timeout: 10, IsActive: true,
			HostID: &hostID,
		}
		channelRepo.AssociateChannelWithResource(resource.ID, ch.ID)

		if withCorrelator {
			events := fake.NewHostEventFake()
			hosts := fake.NewHostFake()
			require.NoError(t, hosts.Create(ctx, &domain.Host{Base: domain.Base{ID: hostID}, Name: "web-01"}))
			require.NoError(t, events.Create(ctx, &domain.HostEvent{
				Base: domain.Base{ID: "e1"}, HostID: hostID,
				OccurredAt: time.Now().Add(-8 * time.Second), Kind: "oom_kill",
				Source: "kmsg", Occurrences: 1,
			}))
			svc = svc.WithCorrelator(correlation.New(events, hosts))
		}

		require.NoError(t, svc.CreateIncident(ctx, resource,
			domain.CheckResult{Status: "down", ResponseData: "timeout"}))

		found, err := incidents.FindByResource(ctx, resource.ID, 10, 0)
		require.NoError(t, err)
		require.Len(t, found, 1, "one incident, correlator or not")

		require.NoError(t, svc.ResolveIncident(ctx, resource,
			domain.CheckResult{Status: "up", ResponseData: "200 OK"}))

		all := steps.FindByIncidentID(found[0].ID)
		out := make([]string, 0, len(all))
		for _, s := range all {
			out = append(out, string(s.Step))
		}
		return out
	}

	assert.Equal(t, stepsFor(t, false), stepsFor(t, true),
		"correlation may change what an alert says; it may not change when incidents happen")
}
