package correlation_test

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

var started = time.Date(2026, 9, 10, 14, 3, 0, 0, time.UTC)

// fixedNow sits after the whole window, so the far-edge clamp is inert unless a
// test deliberately moves it.
func fixedNow() time.Time { return started.Add(time.Hour) }

type harness struct {
	events *fake.HostEventFake
	hosts  *fake.HostFake
	c      *correlation.Correlator
}

func newHarness(t *testing.T, hostID string) *harness {
	t.Helper()
	events := fake.NewHostEventFake()
	hosts := fake.NewHostFake()
	if hostID != "" {
		require.NoError(t, hosts.Create(context.Background(), &domain.Host{
			Base: domain.Base{ID: hostID},
			Name: "web-01",
		}))
		hosts.FindByIDCalls = 0
	}
	return &harness{
		events: events,
		hosts:  hosts,
		c:      correlation.New(events, hosts).WithClock(fixedNow),
	}
}

func (h *harness) add(t *testing.T, hostID, id, kind string, at time.Time) {
	t.Helper()
	require.NoError(t, h.events.Create(context.Background(), &domain.HostEvent{
		Base:        domain.Base{ID: id},
		HostID:      hostID,
		OccurredAt:  at,
		Kind:        kind,
		Source:      "kmsg",
		Occurrences: 1,
	}))
}

func incidentOn(hostID string) *domain.Incident {
	inc := &domain.Incident{
		Base:      domain.Base{ID: "01JINCIDENT"},
		Cause:     "HTTP check failed: 502 Bad Gateway",
		StartedAt: started,
	}
	if hostID != "" {
		id := hostID
		inc.Resource.HostID = &id
	}
	return inc
}

// The zero-cost guarantee. Most monitors have no host, and this feature must be
// free for every operator who runs no agent -- which means returning before any
// query, not after finding nothing.
func TestForIncident_NoHostIssuesNoQuery(t *testing.T) {
	h := newHarness(t, "")

	for _, inc := range []*domain.Incident{incidentOn(""), func() *domain.Incident {
		i := incidentOn("")
		empty := ""
		i.Resource.HostID = &empty
		return i
	}()} {
		got := h.c.ForIncident(context.Background(), inc)
		assert.Nil(t, got.Explanation)
		assert.Empty(t, got.Events)
	}

	assert.Zero(t, h.events.ListInWindowCalls, "no host attached must cost zero event reads")
	assert.Zero(t, h.hosts.FindByIDCalls, "and zero host lookups")
}

func TestForIncident_NilInputsAreSafe(t *testing.T) {
	h := newHarness(t, "h1")
	assert.Nil(t, h.c.ForIncident(context.Background(), nil).Explanation)

	var nilC *correlation.Correlator
	assert.Nil(t, nilC.ForIncident(context.Background(), incidentOn("h1")).Explanation)
}

// Absence is silence: no event means no sentence, not a sentence saying nothing
// happened (FR-008).
func TestForIncident_NoEventsYieldsNothing(t *testing.T) {
	h := newHarness(t, "h1")
	got := h.c.ForIncident(context.Background(), incidentOn("h1"))
	assert.Nil(t, got.Explanation)
	assert.Empty(t, got.Events)
	assert.Zero(t, h.hosts.FindByIDCalls, "no event means the host is never looked up either")
}

// The window is the entire rule (FR-003, SC-005).
func TestForIncident_WindowIsTheWholeRule(t *testing.T) {
	h := newHarness(t, "h1")
	h.add(t, "h1", "e-too-early", "oom_kill", started.Add(-domain.HostContextWindowBefore-time.Second))
	h.add(t, "h1", "e-too-late", "oom_kill", started.Add(domain.HostContextWindowAfter+time.Second))

	got := h.c.ForIncident(context.Background(), incidentOn("h1"))
	assert.Nil(t, got.Explanation, "an event just outside the window is not correlated, however suggestive")
	assert.Empty(t, got.Events)
}

func TestForIncident_WindowEdges(t *testing.T) {
	h := newHarness(t, "h1")
	atStart := started.Add(-domain.HostContextWindowBefore)
	h.add(t, "h1", "e-at-start", "oom_kill", atStart)
	h.add(t, "h1", "e-at-end", "oom_kill", started.Add(domain.HostContextWindowAfter))

	got := h.c.ForIncident(context.Background(), incidentOn("h1"))
	require.NotNil(t, got.Explanation)
	require.Len(t, got.Events, 1, "the start edge is in, the end edge is out")
	assert.Equal(t, "e-at-start", got.Explanation.Event.ID)
	assert.Equal(t, atStart, got.Explanation.WindowFrom)
	assert.Equal(t, started.Add(domain.HostContextWindowAfter), got.Explanation.WindowTo)
}

// The documented rule: the latest event at or before the failure.
func TestForIncident_NamesTheLastEventBeforeTheFailure(t *testing.T) {
	h := newHarness(t, "h1")
	h.add(t, "h1", "e-old", "oom_kill", started.Add(-4*time.Minute))
	h.add(t, "h1", "e-closest", "segfault", started.Add(-10*time.Second))
	h.add(t, "h1", "e-after", "oom_kill", started.Add(20*time.Second))

	got := h.c.ForIncident(context.Background(), incidentOn("h1"))
	require.NotNil(t, got.Explanation)
	assert.Equal(t, "e-closest", got.Explanation.Event.ID)
	assert.True(t, got.Explanation.Precedes)
	assert.Equal(t, 2, got.Explanation.OtherEvents, "named one, counted the rest")
	assert.Len(t, got.Events, 3, "and served all of them")
}

// An event exactly at the incident start counts as preceding it: the window's
// bound is "at or before".
func TestForIncident_SimultaneousCountsAsPreceding(t *testing.T) {
	h := newHarness(t, "h1")
	h.add(t, "h1", "e-same", "oom_kill", started)

	got := h.c.ForIncident(context.Background(), incidentOn("h1"))
	require.NotNil(t, got.Explanation)
	assert.True(t, got.Explanation.Precedes)
}

func TestForIncident_FallsBackToTheEarliestAfter(t *testing.T) {
	h := newHarness(t, "h1")
	h.add(t, "h1", "e-later", "oom_kill", started.Add(50*time.Second))
	h.add(t, "h1", "e-sooner", "oom_kill", started.Add(20*time.Second))

	got := h.c.ForIncident(context.Background(), incidentOn("h1"))
	require.NotNil(t, got.Explanation)
	assert.Equal(t, "e-sooner", got.Explanation.Event.ID)
	assert.False(t, got.Explanation.Precedes, "it followed the failure and must not read otherwise")
}

// Determinism has to hold at the tie too, or "generated on read" means an
// incident that changes on refresh (FR-018a, SC-010).
func TestForIncident_TiesBreakOnTheSmallerID(t *testing.T) {
	at := started.Add(-30 * time.Second)
	for _, order := range [][]string{{"e-bbb", "e-aaa"}, {"e-aaa", "e-bbb"}} {
		h := newHarness(t, "h1")
		for _, id := range order {
			h.add(t, "h1", id, "oom_kill", at)
		}
		got := h.c.ForIncident(context.Background(), incidentOn("h1"))
		require.NotNil(t, got.Explanation)
		assert.Equal(t, "e-aaa", got.Explanation.Event.ID,
			"insertion order must not decide which event a sentence names")
	}
}

// The same tie-break has to hold on the "earliest after" branch. Wanting the
// earliest means preferring the earlier timestamp but still the smaller id, and
// one comparator answering both questions gets exactly this case wrong.
func TestForIncident_TiesBreakOnTheSmallerIDAfterToo(t *testing.T) {
	at := started.Add(30 * time.Second)
	for _, order := range [][]string{{"e-bbb", "e-aaa"}, {"e-aaa", "e-bbb"}} {
		h := newHarness(t, "h1")
		for _, id := range order {
			h.add(t, "h1", id, "oom_kill", at)
		}
		got := h.c.ForIncident(context.Background(), incidentOn("h1"))
		require.NotNil(t, got.Explanation)
		assert.Equal(t, "e-aaa", got.Explanation.Event.ID)
		assert.False(t, got.Explanation.Precedes)
	}
}

// Unphrasable kinds are shown and counted, never named (FR-007, R3).
func TestForIncident_UnphrasableKindsAreCountedNotNamed(t *testing.T) {
	h := newHarness(t, "h1")
	h.add(t, "h1", "e-unknown-1", "something_new", started.Add(-20*time.Second))
	h.add(t, "h1", "e-unknown-2", "another_thing", started.Add(-10*time.Second))
	h.add(t, "h1", "e-known", "oom_kill", started.Add(-40*time.Second))

	got := h.c.ForIncident(context.Background(), incidentOn("h1"))
	require.NotNil(t, got.Explanation)
	assert.Equal(t, "e-known", got.Explanation.Event.ID,
		"a nearer event of an unphrasable kind must not win the naming")
	assert.Equal(t, 2, got.Explanation.OtherEvents, "unphrasable kinds still count")
	assert.Len(t, got.Events, 3, "and are still served, so nothing is hidden")
}

func TestForIncident_OnlyUnphrasableKindsYieldsEventsButNoSentence(t *testing.T) {
	h := newHarness(t, "h1")
	h.add(t, "h1", "e-unknown", "something_new", started.Add(-20*time.Second))

	got := h.c.ForIncident(context.Background(), incidentOn("h1"))
	assert.Nil(t, got.Explanation, "no sentence rather than a malformed one")
	assert.Len(t, got.Events, 1, "but the event is still visible")
	assert.Zero(t, h.hosts.FindByIDCalls, "and no host lookup is spent on a sentence that will not exist")
}

// Enrichment must never cost the thing it enriches.
func TestForIncident_LookupFailuresDegradeQuietly(t *testing.T) {
	t.Run("event read fails", func(t *testing.T) {
		h := newHarness(t, "h1")
		h.add(t, "h1", "e1", "oom_kill", started.Add(-time.Second))
		h.events.ListInWindowErr = errors.New("boom")

		got := h.c.ForIncident(context.Background(), incidentOn("h1"))
		assert.Nil(t, got.Explanation)
		assert.Empty(t, got.Events)
	})

	t.Run("host lookup fails", func(t *testing.T) {
		h := newHarness(t, "")
		h.add(t, "h-missing", "e1", "oom_kill", started.Add(-time.Second))

		got := h.c.ForIncident(context.Background(), incidentOn("h-missing"))
		assert.Nil(t, got.Explanation, "a sentence naming no host would be unactionable")
		assert.Len(t, got.Events, 1, "the events are still served")
	})
}

// Events of another host never enter this incident's window.
func TestForIncident_ScopedToTheMonitorsHost(t *testing.T) {
	h := newHarness(t, "h1")
	h.add(t, "h2", "e-theirs", "oom_kill", started.Add(-time.Second))

	got := h.c.ForIncident(context.Background(), incidentOn("h1"))
	assert.Nil(t, got.Explanation)
	assert.Empty(t, got.Events)
}

// A fresh incident's window is still elapsing. Read what exists rather than wait
// for it, and report the clamped edge rather than the nominal one.
func TestForIncident_ClampsTheWindowToNow(t *testing.T) {
	now := started.Add(15 * time.Second)
	h := newHarness(t, "h1")
	h.c = correlation.New(h.events, h.hosts).WithClock(func() time.Time { return now })
	h.add(t, "h1", "e-in", "oom_kill", started.Add(-5*time.Second))
	h.add(t, "h1", "e-future", "oom_kill", started.Add(30*time.Second))

	got := h.c.ForIncident(context.Background(), incidentOn("h1"))
	require.NotNil(t, got.Explanation)
	assert.Equal(t, now, got.Explanation.WindowTo, "the reported edge is the one actually read")
	assert.Len(t, got.Events, 1)
}

// An incident whose window has not begun yet cannot correlate, and must not
// spend a query discovering that.
func TestForIncident_WindowEntirelyInTheFutureIsInert(t *testing.T) {
	h := newHarness(t, "h1")
	h.c = correlation.New(h.events, h.hosts).WithClock(func() time.Time {
		return started.Add(-domain.HostContextWindowBefore - time.Hour)
	})

	got := h.c.ForIncident(context.Background(), incidentOn("h1"))
	assert.Nil(t, got.Explanation)
	assert.Zero(t, h.events.ListInWindowCalls)
}

// The whole result is a projection of stored facts, so repeating the lookup must
// repeat the answer.
func TestForIncident_IsRepeatable(t *testing.T) {
	h := newHarness(t, "h1")
	h.add(t, "h1", "e1", "oom_kill", started.Add(-30*time.Second))
	h.add(t, "h1", "e2", "segfault", started.Add(-20*time.Second))

	first := h.c.ForIncident(context.Background(), incidentOn("h1"))
	require.NotNil(t, first.Explanation)
	firstText := narrative.Sentence(first.Explanation)

	for i := 0; i < 10; i++ {
		again := h.c.ForIncident(context.Background(), incidentOn("h1"))
		require.NotNil(t, again.Explanation)
		assert.Equal(t, first.Explanation.Event.ID, again.Explanation.Event.ID)
		assert.Equal(t, firstText, narrative.Sentence(again.Explanation))
	}
}
