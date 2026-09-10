package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/correlation"
	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/narrative"
	"github.com/denisakp/ogoune/internal/repository/fake"
)

// explanationFixture is the host-context fixture plus a correlator, so the two
// enrichments can be exercised together -- and, more to the point, apart.
type explanationFixture struct {
	*hostContextFixture
	events *fake.HostEventFake
}

func newExplanationFixture(t *testing.T, rawWindow time.Duration) *explanationFixture {
	t.Helper()
	base := newHostContextFixture(t, rawWindow)
	events := fake.NewHostEventFake()
	now := base.now
	base.svc = base.svc.WithCorrelator(
		correlation.New(events, base.hosts).WithClock(func() time.Time { return now }),
	)
	return &explanationFixture{hostContextFixture: base, events: events}
}

func (f *explanationFixture) seedEvent(t *testing.T, hostID, id, kind string, at time.Time) {
	t.Helper()
	require.NoError(t, f.events.Create(context.Background(), &domain.HostEvent{
		Base:        domain.Base{ID: id},
		HostID:      hostID,
		OccurredAt:  at,
		Kind:        kind,
		Source:      "kmsg",
		Occurrences: 1,
		Detail:      &domain.HostEventDetail{Process: "postgres", PID: 4711},
	}))
}

func TestGetIncidentByID_Explanation_Populated(t *testing.T) {
	f := newExplanationFixture(t, 7*24*time.Hour)
	started := f.now.Add(-2 * time.Minute)
	inc := f.seedIncident(t, "inc-expl", started, false)
	f.seedEvent(t, "host-inc-expl", "e1", "oom_kill", started.Add(-13*time.Second))

	got, err := f.svc.GetIncidentByID(context.Background(), inc.ID)
	require.NoError(t, err)
	require.NotNil(t, got.Explanation)

	e := got.Explanation
	assert.Equal(t, "host-inc-expl", e.HostID)
	assert.Equal(t, "web-inc-expl", e.HostName)
	assert.Equal(t, "e1", e.Event.ID)
	assert.True(t, e.Precedes)
	assert.Zero(t, e.OtherEvents)
	assert.Equal(t, started, e.IncidentAt)
	require.Len(t, got.HostEvents, 1, "the evidence is served with the claim")

	assert.NotEmpty(t, narrative.Sentence(e))
}

// The case a design nesting the events under host_context would have lost, and
// the normal shape of every incident older than about a week: metrics purge at
// 7d, events survive to 90d (ADR 0011, FR-006, SC-007).
func TestGetIncidentByID_Explanation_SurvivesPurgedMetrics(t *testing.T) {
	f := newExplanationFixture(t, 7*24*time.Hour)
	started := f.now.Add(-30 * 24 * time.Hour)
	inc := f.seedIncident(t, "inc-old", started, false)
	// No host_metrics rows at all: thinning and purge already took them.
	f.seedEvent(t, "host-inc-old", "e-old", "oom_kill", started.Add(-20*time.Second))

	got, err := f.svc.GetIncidentByID(context.Background(), inc.ID)
	require.NoError(t, err)

	assert.Nil(t, got.HostContext, "the metrics are gone")
	require.NotNil(t, got.Explanation, "the kernel event is not, and it is the whole point")
	assert.Equal(t, "e-old", got.Explanation.Event.ID)
}

// The converse: metrics without events must not manufacture a sentence.
func TestGetIncidentByID_Explanation_MetricsWithoutEventsStaySilent(t *testing.T) {
	f := newExplanationFixture(t, 7*24*time.Hour)
	started := f.now.Add(-time.Minute)
	inc := f.seedIncident(t, "inc-metrics-only", started, false)
	f.seedSample(t, "host-inc-metrics-only", started.Add(-10*time.Second), 91.0, 44.0, nil)

	got, err := f.svc.GetIncidentByID(context.Background(), inc.ID)
	require.NoError(t, err)
	require.NotNil(t, got.HostContext)
	assert.Nil(t, got.Explanation, "a busy host is not an explanation")
	assert.Empty(t, got.HostEvents)
}

func TestGetIncidentByID_Explanation_AbsentPaths(t *testing.T) {
	started := time.Date(2026, 3, 10, 11, 58, 0, 0, time.UTC)

	t.Run("no host attached", func(t *testing.T) {
		f := newExplanationFixture(t, 7*24*time.Hour)
		inc := f.seedIncident(t, "inc-detached", started, true)
		// An event exists for the host, but the monitor is not attached to it.
		f.seedEvent(t, "host-inc-detached", "e1", "oom_kill", started.Add(-time.Second))

		got, err := f.svc.GetIncidentByID(context.Background(), inc.ID)
		require.NoError(t, err)
		assert.Nil(t, got.Explanation)
		assert.Empty(t, got.HostEvents)
		assert.Zero(t, f.events.ListInWindowCalls, "and it cost no query")
	})

	t.Run("no events in the window", func(t *testing.T) {
		f := newExplanationFixture(t, 7*24*time.Hour)
		inc := f.seedIncident(t, "inc-noevents", started, false)
		f.seedEvent(t, "host-inc-noevents", "e-far", "oom_kill",
			started.Add(-domain.HostContextWindowBefore-time.Minute))

		got, err := f.svc.GetIncidentByID(context.Background(), inc.ID)
		require.NoError(t, err)
		assert.Nil(t, got.Explanation)
		assert.Empty(t, got.HostEvents)
	})

	t.Run("only kinds this version cannot phrase", func(t *testing.T) {
		f := newExplanationFixture(t, 7*24*time.Hour)
		inc := f.seedIncident(t, "inc-unknown", started, false)
		f.seedEvent(t, "host-inc-unknown", "e-unknown", "something_new", started.Add(-time.Second))

		got, err := f.svc.GetIncidentByID(context.Background(), inc.ID)
		require.NoError(t, err)
		assert.Nil(t, got.Explanation, "no sentence")
		require.Len(t, got.HostEvents, 1, "but the event is still visible")
	})

	// Enrichment must never cost the incident.
	t.Run("event lookup fails", func(t *testing.T) {
		f := newExplanationFixture(t, 7*24*time.Hour)
		inc := f.seedIncident(t, "inc-boom", started, false)
		f.seedEvent(t, "host-inc-boom", "e1", "oom_kill", started.Add(-time.Second))
		f.events.ListInWindowErr = errors.New("database on fire")

		got, err := f.svc.GetIncidentByID(context.Background(), inc.ID)
		require.NoError(t, err, "the incident is still served")
		require.NotNil(t, got)
		assert.Equal(t, inc.ID, got.ID)
		assert.Nil(t, got.Explanation)
	})

	// A service with no correlator behaves exactly as it did before spec 091.
	t.Run("no correlator attached", func(t *testing.T) {
		f := newHostContextFixture(t, 7*24*time.Hour)
		inc := f.seedIncident(t, "inc-nocorrelator", started, false)

		got, err := f.svc.GetIncidentByID(context.Background(), inc.ID)
		require.NoError(t, err)
		assert.Nil(t, got.Explanation)
		assert.Empty(t, got.HostEvents)
	})
}

// Named one, counted the rest -- never a list (FR-005, SC-006).
func TestGetIncidentByID_Explanation_NamesOneCountsTheRest(t *testing.T) {
	f := newExplanationFixture(t, 7*24*time.Hour)
	started := f.now.Add(-time.Minute)
	inc := f.seedIncident(t, "inc-storm", started, false)
	f.seedEvent(t, "host-inc-storm", "e-far", "oom_kill", started.Add(-4*time.Minute))
	f.seedEvent(t, "host-inc-storm", "e-near", "segfault", started.Add(-5*time.Second))
	f.seedEvent(t, "host-inc-storm", "e-unknown", "something_new", started.Add(-2*time.Second))

	got, err := f.svc.GetIncidentByID(context.Background(), inc.ID)
	require.NoError(t, err)
	require.NotNil(t, got.Explanation)
	assert.Equal(t, "e-near", got.Explanation.Event.ID)
	assert.Equal(t, 2, got.Explanation.OtherEvents, "unphrasable kinds still count")
	assert.Len(t, got.HostEvents, 3, "and are still served")
}

// Generated on read, so reading twice must read the same (SC-010).
func TestGetIncidentByID_Explanation_IsStableAcrossReads(t *testing.T) {
	f := newExplanationFixture(t, 7*24*time.Hour)
	started := f.now.Add(-time.Minute)
	inc := f.seedIncident(t, "inc-stable", started, false)
	f.seedEvent(t, "host-inc-stable", "e-a", "oom_kill", started.Add(-10*time.Second))
	f.seedEvent(t, "host-inc-stable", "e-b", "segfault", started.Add(-10*time.Second))

	first, err := f.svc.GetIncidentByID(context.Background(), inc.ID)
	require.NoError(t, err)
	require.NotNil(t, first.Explanation)
	want := narrative.Sentence(first.Explanation)

	for i := 0; i < 5; i++ {
		again, err := f.svc.GetIncidentByID(context.Background(), inc.ID)
		require.NoError(t, err)
		require.NotNil(t, again.Explanation)
		assert.Equal(t, want, narrative.Sentence(again.Explanation))
	}
}
