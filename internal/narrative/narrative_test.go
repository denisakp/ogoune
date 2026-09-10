package narrative

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/domain"
)

var incidentAt = time.Date(2026, 9, 10, 14, 3, 0, 0, time.UTC)

func expl(mut func(*domain.IncidentExplanation)) *domain.IncidentExplanation {
	e := &domain.IncidentExplanation{
		HostID:     "01JHOST",
		HostName:   "web-01",
		IncidentAt: incidentAt,
		Cause:      "HTTP check failed: 502 Bad Gateway",
		Event: domain.HostEvent{
			Base:        domain.Base{ID: "01JEVENT"},
			HostID:      "01JHOST",
			OccurredAt:  incidentAt.Add(-13 * time.Second),
			Kind:        "oom_kill",
			Source:      "kmsg",
			Occurrences: 1,
			Detail:      &domain.HostEventDetail{Process: "postgres", PID: 4711},
		},
		Precedes:   true,
		WindowFrom: incidentAt.Add(-domain.HostContextWindowBefore),
		WindowTo:   incidentAt.Add(domain.HostContextWindowAfter),
	}
	if mut != nil {
		mut(e)
	}
	return e
}

func TestSentence_NominalCase(t *testing.T) {
	got := Sentence(expl(nil))
	assert.Equal(t,
		"HTTP check failed: 502 Bad Gateway at 2026-09-10 14:03:00 UTC. "+
			"The kernel OOM-killed postgres (pid 4711) on web-01 at 2026-09-10 14:02:47 UTC, 13 seconds earlier.",
		got)
}

// Both timestamps are what let an operator overrule the sentence. A wording
// change that drops one turns an interpretation into an assertion (FR-010).
func TestSentence_AlwaysCarriesBothTimes(t *testing.T) {
	cases := map[string]*domain.IncidentExplanation{
		"precedes": expl(nil),
		"follows": expl(func(e *domain.IncidentExplanation) {
			e.Event.OccurredAt = incidentAt.Add(12 * time.Second)
			e.Precedes = false
		}),
		"simultaneous": expl(func(e *domain.IncidentExplanation) {
			e.Event.OccurredAt = incidentAt
		}),
	}
	for name, e := range cases {
		t.Run(name, func(t *testing.T) {
			got := Sentence(e)
			assert.Contains(t, got, e.IncidentAt.UTC().Format(timeLayout))
			assert.Contains(t, got, e.Event.OccurredAt.UTC().Format(timeLayout))
		})
	}
}

// An event inside the window but after the failure must never read as having
// preceded it.
func TestSentence_FollowsNeverReadsAsBefore(t *testing.T) {
	got := Sentence(expl(func(e *domain.IncidentExplanation) {
		e.Event.OccurredAt = incidentAt.Add(90 * time.Second)
		e.Precedes = false
	}))
	assert.Contains(t, got, "1 minute 30 seconds later")
	assert.NotContains(t, got, "earlier")
}

// Two clocks agreeing to the second is not precision worth claiming: the event's
// timestamp comes from the host's kernel, the incident's from this server.
func TestSentence_SimultaneousOmitsTheGap(t *testing.T) {
	got := Sentence(expl(func(e *domain.IncidentExplanation) {
		e.Event.OccurredAt = incidentAt
	}))
	assert.NotContains(t, got, "earlier")
	assert.NotContains(t, got, "later")
	assert.NotContains(t, got, "0 second")
}

// "200 reports" and "200 events" answer different questions. Merging them would
// misdescribe both (R8).
func TestSentence_ReportsAndOtherEventsAreDistinct(t *testing.T) {
	got := Sentence(expl(func(e *domain.IncidentExplanation) {
		e.Event.Occurrences = 200
		e.OtherEvents = 2
	}))
	assert.Contains(t, got, "Reported 200 times.")
	assert.Contains(t, got, "2 other kernel events fell in the same window.")
}

func TestSentence_OtherEventsPlural(t *testing.T) {
	assert.NotContains(t, Sentence(expl(nil)), "other kernel event",
		"no count when there are no others")
	assert.Contains(t, Sentence(expl(func(e *domain.IncidentExplanation) { e.OtherEvents = 1 })),
		"1 other kernel event fell")
	assert.Contains(t, Sentence(expl(func(e *domain.IncidentExplanation) { e.OtherEvents = 5 })),
		"5 other kernel events fell")
}

func TestSentence_SingleOccurrenceSaysNothingAboutCounts(t *testing.T) {
	assert.NotContains(t, Sentence(expl(nil)), "Reported")
}

func TestSentence_TargetPrecisionFollowsTheReport(t *testing.T) {
	cases := []struct {
		name   string
		detail *domain.HostEventDetail
		want   string
	}{
		{"process and pid", &domain.HostEventDetail{Process: "postgres", PID: 4711}, "postgres (pid 4711)"},
		{"process without pid", &domain.HostEventDetail{Process: "postgres"}, "OOM-killed postgres on"},
		{"nothing named", nil, "a process"},
		{"empty detail", &domain.HostEventDetail{}, "a process"},
		{
			"distinct processes",
			&domain.HostEventDetail{DistinctProcesses: []string{"postgres", "node"}},
			"postgres, node",
		},
		{
			"distinct truncated is marked",
			&domain.HostEventDetail{DistinctProcesses: []string{"postgres", "node"}, DistinctTruncated: true},
			"postgres, node and more",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Sentence(expl(func(e *domain.IncidentExplanation) { e.Event.Detail = tc.detail }))
			assert.Contains(t, got, tc.want)
		})
	}
}

func TestSentence_SegfaultReadsAsItself(t *testing.T) {
	got := Sentence(expl(func(e *domain.IncidentExplanation) { e.Event.Kind = "segfault" }))
	assert.Contains(t, got, "recorded a segmentation fault in postgres (pid 4711)")
}

// Silence, never a half-built sentence. An unrecognised kind carries
// agent-supplied text and must not choose the wording of an alert (FR-007).
func TestSentence_UnphrasableKindSaysNothing(t *testing.T) {
	assert.Empty(t, Sentence(expl(func(e *domain.IncidentExplanation) { e.Event.Kind = "something_new" })))
	assert.Empty(t, Sentence(expl(func(e *domain.IncidentExplanation) { e.Event.Kind = "" })))
	assert.Empty(t, Sentence(nil))
}

func TestSentence_FallsBackWhenNamesAreMissing(t *testing.T) {
	got := Sentence(expl(func(e *domain.IncidentExplanation) {
		e.HostName = "  "
		e.Cause = ""
	}))
	assert.Contains(t, got, "This check failed at")
	assert.Contains(t, got, "on 01JHOST at", "an operator can act on an id; they cannot act on nothing")
}

// Generated on read is only safe if it is deterministic: otherwise an incident
// would read differently on each refresh (FR-018a, SC-010).
func TestSentence_IsDeterministic(t *testing.T) {
	e := expl(func(e *domain.IncidentExplanation) {
		e.Event.Occurrences = 37
		e.OtherEvents = 3
	})
	first := Sentence(e)
	for i := 0; i < 20; i++ {
		require.Equal(t, first, Sentence(e), "the same facts must render the same bytes")
	}
	assert.NotEmpty(t, first)
}

// The sentence must never assert more than "these two things happened close
// together" in machine-readable form. It says "the kernel did X"; it does not
// say the check failed because of it.
func TestSentence_ClaimsNoMechanism(t *testing.T) {
	got := strings.ToLower(Sentence(expl(nil)))
	for _, forbidden := range []string{"because", "caused by", "root cause", "due to"} {
		assert.NotContainsf(t, got, forbidden,
			"the wording may read as an explanation, but must not assert a mechanism it cannot know: %q", forbidden)
	}
}

func TestHumanGap(t *testing.T) {
	cases := map[time.Duration]string{
		time.Second:                    "1 second",
		13 * time.Second:               "13 seconds",
		time.Minute:                    "1 minute",
		2 * time.Minute:                "2 minutes",
		90 * time.Second:               "1 minute 30 seconds",
		2*time.Minute + 1*time.Second:  "2 minutes 1 second",
		5*time.Minute + 59*time.Second: "5 minutes 59 seconds",
	}
	for d, want := range cases {
		assert.Equalf(t, want, humanGap(d), "gap %s", d)
	}
}

func TestPhrasable(t *testing.T) {
	assert.True(t, Phrasable("oom_kill"))
	assert.True(t, Phrasable("segfault"))
	assert.False(t, Phrasable("something_new"), "the kind set is open; the phrase set is not")
}

// The two timestamps exist to be compared with each other, so they must be in
// the same zone. An incident's start carries the server's local zone while a
// kernel event's is stored in UTC; formatting each as it arrives produced
// "11:19:04 GMT" beside "11:17:35 UTC" in one sentence on a real host.
func TestSentence_BothTimesAreRenderedInUTC(t *testing.T) {
	paris := time.FixedZone("CEST", 2*60*60)
	got := Sentence(expl(func(e *domain.IncidentExplanation) {
		e.IncidentAt = incidentAt.In(paris)
		e.Event.OccurredAt = incidentAt.Add(-13 * time.Second)
	}))

	assert.Contains(t, got, "2026-09-10 14:03:00 UTC")
	assert.Contains(t, got, "2026-09-10 14:02:47 UTC")
	assert.NotContains(t, got, "CEST",
		"a sentence mixing zones makes its own arithmetic unverifiable")
	assert.Contains(t, got, "13 seconds earlier",
		"and the stated gap must follow from the printed times")
}

// The gap must be derivable from the two printed timestamps, because those are
// what an operator checks it against.
func TestSentence_GapMatchesThePrintedTimes(t *testing.T) {
	// 89 seconds apart to the second, with sub-second parts that would round the
	// naive difference down to 88.
	incident := time.Date(2026, 9, 10, 11, 19, 4, 100_000_000, time.UTC)
	event := time.Date(2026, 9, 10, 11, 17, 35, 700_000_000, time.UTC)

	got := Sentence(expl(func(e *domain.IncidentExplanation) {
		e.IncidentAt = incident
		e.Event.OccurredAt = event
	}))

	assert.Contains(t, got, "2026-09-10 11:19:04 UTC")
	assert.Contains(t, got, "2026-09-10 11:17:35 UTC")
	assert.Contains(t, got, "1 minute 29 seconds earlier",
		"11:19:04 minus 11:17:35 is 89 seconds, and that is what the reader will compute")
}
