package main

import (
	"testing"
	"time"

	"github.com/denisakp/ogoune/pkg/agentwire"
)

// fakeSource feeds the collector a scripted batch of reports.
type fakeSource struct {
	name    string
	batches [][]kmsgReport
	closed  bool
}

func (f *fakeSource) Name() string { return f.name }
func (f *fakeSource) Drain() []kmsgReport {
	if len(f.batches) == 0 {
		return nil
	}
	b := f.batches[0]
	f.batches = f.batches[1:]
	return b
}
func (f *fakeSource) Close() { f.closed = true }

func reports(kind string, procs ...string) []kmsgReport {
	out := make([]kmsgReport, 0, len(procs))
	for i, p := range procs {
		out = append(out, kmsgReport{Kind: kind, Process: p, PID: 100 + i})
	}
	return out
}

var now = time.Date(2026, 9, 9, 14, 2, 47, 0, time.UTC)

// T016 -- many reports of one kind in an interval become ONE event with a count.
// Thirty-seven identical rows on a host page is noise; one row saying it happened
// thirty-seven times is a finding.
func TestEventCollector_AggregatesPerKind(t *testing.T) {
	src := &fakeSource{name: "kmsg", batches: [][]kmsgReport{
		reports(agentwire.KindOOMKill, "postgres", "postgres", "postgres", "node"),
	}}
	c := newEventCollector([]kernelSource{src})

	got := c.Collect(now)

	if len(got) != 1 {
		t.Fatalf("events = %d, want 1 aggregated event", len(got))
	}
	e := got[0]
	if e.Kind != agentwire.KindOOMKill {
		t.Errorf("kind = %q", e.Kind)
	}
	if e.Occurrences != 4 {
		t.Errorf("occurrences = %d, want 4", e.Occurrences)
	}
	if e.Process != "postgres" {
		t.Errorf("representative process = %q, want the first seen", e.Process)
	}
	if !e.OccurredAt.Equal(now) {
		t.Errorf("occurred_at = %v", e.OccurredAt)
	}
}

// T016 -- two kinds are two events. Aggregation is per kind, not per interval.
func TestEventCollector_SeparatesKinds(t *testing.T) {
	batch := append(reports(agentwire.KindOOMKill, "postgres"), reports(agentwire.KindSegfault, "myapp")...)
	c := newEventCollector([]kernelSource{&fakeSource{name: "kmsg", batches: [][]kmsgReport{batch}}})

	got := c.Collect(now)

	if len(got) != 2 {
		t.Fatalf("events = %d, want one per kind", len(got))
	}
	kinds := map[string]bool{got[0].Kind: true, got[1].Kind: true}
	if !kinds[agentwire.KindOOMKill] || !kinds[agentwire.KindSegfault] {
		t.Errorf("kinds = %v", kinds)
	}
}

// T016 -- the distinct list holds the distinct names and nothing more. This is
// what tells a storm killing forty copies of one process from one killing forty
// different processes.
func TestEventCollector_DistinctProcesses(t *testing.T) {
	c := newEventCollector([]kernelSource{&fakeSource{name: "kmsg", batches: [][]kmsgReport{reports(agentwire.KindOOMKill, "postgres", "postgres", "node", "postgres", "python3")}}})

	got := c.Collect(now)

	if len(got) != 1 {
		t.Fatalf("events = %d", len(got))
	}
	if got[0].Occurrences != 5 {
		t.Errorf("occurrences = %d, want 5", got[0].Occurrences)
	}
	if len(got[0].DistinctProcesses) != 3 {
		t.Fatalf("distinct = %v, want 3 names", got[0].DistinctProcesses)
	}
	if got[0].DistinctTruncated {
		t.Error("three names is under the bound, nothing was truncated")
	}
}

// T017 -- the distinct list is bounded, and exceeding the bound is MARKED rather
// than silently shortened. A partial list read as complete is worse than an
// obviously partial one.
func TestEventCollector_DistinctListIsBoundedAndMarked(t *testing.T) {
	procs := make([]string, 0, 40)
	for i := 0; i < 40; i++ {
		procs = append(procs, "proc"+string(rune('a'+i%26))+string(rune('0'+i/26)))
	}
	c := newEventCollector([]kernelSource{&fakeSource{name: "kmsg", batches: [][]kmsgReport{reports(agentwire.KindOOMKill, procs...)}}})

	got := c.Collect(now)

	if len(got) != 1 {
		t.Fatalf("events = %d, want 1: forty distinct processes is still one storm", len(got))
	}
	e := got[0]
	if e.Occurrences != 40 {
		t.Errorf("occurrences = %d, want 40: the count is not bounded, only the list is", e.Occurrences)
	}
	if len(e.DistinctProcesses) != agentwire.MaxDistinctProcesses {
		t.Errorf("distinct = %d, want the bound %d", len(e.DistinctProcesses), agentwire.MaxDistinctProcesses)
	}
	if !e.DistinctTruncated {
		t.Error("more distinct processes were seen than the list holds; that must be marked")
	}
}

// A quiet host produces nothing, and that is not an error.
func TestEventCollector_QuietHost(t *testing.T) {
	c := newEventCollector([]kernelSource{&fakeSource{name: "kmsg"}})
	if got := c.Collect(now); len(got) != 0 {
		t.Errorf("events = %v, want none", got)
	}
}

// T034 / FR-007 -- no source at all is the common case, not an error: the
// documented default deployment is a container, and a container usually cannot
// read the kernel log. It must cost nothing.
func TestEventCollector_NoSourcesIsSilentAndHarmless(t *testing.T) {
	c := newEventCollector(nil)

	for i := 0; i < 100; i++ {
		if got := c.Collect(now); got != nil {
			t.Fatalf("collect returned %v with no sources", got)
		}
	}
	if c.Dropped() != 0 {
		t.Errorf("dropped = %d, want 0: nothing was captured, so nothing was dropped", c.Dropped())
	}
}

// Draining is per interval: a report seen once is not reported again.
func TestEventCollector_DoesNotRepeatAcrossIntervals(t *testing.T) {
	src := &fakeSource{name: "kmsg", batches: [][]kmsgReport{
		reports(agentwire.KindOOMKill, "postgres"),
		nil,
	}}
	c := newEventCollector([]kernelSource{src})

	if got := c.Collect(now); len(got) != 1 {
		t.Fatalf("first interval: events = %d, want 1", len(got))
	}
	if got := c.Collect(now.Add(time.Minute)); len(got) != 0 {
		t.Errorf("second interval: events = %d, want 0 — a report is not re-sent", len(got))
	}
}

func TestEventCollector_CloseReleasesSources(t *testing.T) {
	src := &fakeSource{name: "kmsg"}
	newEventCollector([]kernelSource{src}).Close()
	if !src.closed {
		t.Error("source not closed")
	}
}

// One kill, two kernel lines. Observed on a real host: a cgroup out-of-memory
// kill emits both "oom-kill:...,task=x,pid=N" and "Killed process N (x)", and
// counting both told the operator two processes had died.
func TestEventCollector_OneKillReportedTwiceCountsOnce(t *testing.T) {
	src := &fakeSource{name: "kmsg", batches: [][]kmsgReport{{
		{Kind: agentwire.KindOOMKill, Process: "python3", PID: 1673237},
		{Kind: agentwire.KindOOMKill, Process: "python3", PID: 1673237},
	}}}

	events := newEventCollector([]kernelSource{src}).Collect(time.Now())

	if len(events) != 1 {
		t.Fatalf("events = %d, want 1", len(events))
	}
	if events[0].Occurrences != 1 {
		t.Errorf("occurrences = %d, want 1: two lines about one dead process are one kill", events[0].Occurrences)
	}
	if events[0].PID != 1673237 {
		t.Errorf("pid = %d, want 1673237", events[0].PID)
	}
	if len(events[0].DistinctProcesses) != 1 || events[0].DistinctProcesses[0] != "python3" {
		t.Errorf("distinct = %v, want [python3]", events[0].DistinctProcesses)
	}
}

// Distinct pids are distinct kills, however similar the processes look.
func TestEventCollector_DistinctPIDsCountSeparately(t *testing.T) {
	src := &fakeSource{name: "kmsg", batches: [][]kmsgReport{{
		{Kind: agentwire.KindOOMKill, Process: "worker", PID: 101},
		{Kind: agentwire.KindOOMKill, Process: "worker", PID: 102},
		{Kind: agentwire.KindOOMKill, Process: "worker", PID: 103},
	}}}

	events := newEventCollector([]kernelSource{src}).Collect(time.Now())

	if len(events) != 1 {
		t.Fatalf("events = %d, want 1", len(events))
	}
	if events[0].Occurrences != 3 {
		t.Errorf("occurrences = %d, want 3", events[0].Occurrences)
	}
	if len(events[0].DistinctProcesses) != 1 {
		t.Errorf("distinct = %v, want one entry: three kills of one program, not three programs",
			events[0].DistinctProcesses)
	}
}

// The cgroup counter says how many, never which. Nothing to deduplicate on, so
// every report counts.
func TestEventCollector_ReportsWithoutPIDsAllCount(t *testing.T) {
	src := &fakeSource{name: "cgroup", batches: [][]kmsgReport{{
		{Kind: agentwire.KindOOMKill},
		{Kind: agentwire.KindOOMKill},
		{Kind: agentwire.KindOOMKill},
	}}}

	events := newEventCollector([]kernelSource{src}).Collect(time.Now())

	if len(events) != 1 {
		t.Fatalf("events = %d, want 1", len(events))
	}
	if events[0].Occurrences != 3 {
		t.Errorf("occurrences = %d, want 3", events[0].Occurrences)
	}
}

// A pid-bearing report is more useful than one without, whichever arrived first.
func TestEventCollector_PrefersTheReportThatNamesAPID(t *testing.T) {
	src := &fakeSource{name: "kmsg", batches: [][]kmsgReport{{
		{Kind: agentwire.KindOOMKill, Process: "python3"},
		{Kind: agentwire.KindOOMKill, Process: "python3", PID: 4711},
	}}}

	events := newEventCollector([]kernelSource{src}).Collect(time.Now())

	if len(events) != 1 {
		t.Fatalf("events = %d, want 1", len(events))
	}
	if events[0].PID != 4711 {
		t.Errorf("pid = %d, want 4711", events[0].PID)
	}
}

// The two readers watch the same kernel: adding their reports double-counts.
// Observed on real hardware, where one kill produced two kernel-log lines and a
// cgroup counter increment — reported as three.
func TestEventCollector_TwoReadersOneKillCountsOnce(t *testing.T) {
	kmsg := &fakeSource{name: "kmsg", batches: [][]kmsgReport{{
		{Kind: agentwire.KindOOMKill, Process: "python3", PID: 4711}, // oom-kill:...task=,pid=
		{Kind: agentwire.KindOOMKill, Process: "python3", PID: 4711}, // Killed process N (x)
	}}}
	cgroup := &fakeSource{name: "cgroup", batches: [][]kmsgReport{{
		{Kind: agentwire.KindOOMKill}, // the counter says "one more", never which
	}}}

	events := newEventCollector([]kernelSource{kmsg, cgroup}).Collect(time.Now())

	if len(events) != 1 {
		t.Fatalf("events = %d, want 1", len(events))
	}
	if events[0].Occurrences != 1 {
		t.Errorf("occurrences = %d, want 1: one process died", events[0].Occurrences)
	}
	if events[0].PID != 4711 {
		t.Errorf("pid = %d, want 4711: the report that named one is the useful one", events[0].PID)
	}
}

// When the kernel log lost records to ring-buffer overwrite, the counter's
// higher number is the truthful one.
func TestEventCollector_CounterWinsWhenTheLogMissedKills(t *testing.T) {
	kmsg := &fakeSource{name: "kmsg", batches: [][]kmsgReport{{
		{Kind: agentwire.KindOOMKill, Process: "worker", PID: 1},
		{Kind: agentwire.KindOOMKill, Process: "worker", PID: 2},
	}}}
	cgroup := &fakeSource{name: "cgroup", batches: [][]kmsgReport{{
		{Kind: agentwire.KindOOMKill}, {Kind: agentwire.KindOOMKill},
		{Kind: agentwire.KindOOMKill}, {Kind: agentwire.KindOOMKill},
		{Kind: agentwire.KindOOMKill},
	}}}

	events := newEventCollector([]kernelSource{kmsg, cgroup}).Collect(time.Now())

	if len(events) != 1 {
		t.Fatalf("events = %d, want 1", len(events))
	}
	if events[0].Occurrences != 5 {
		t.Errorf("occurrences = %d, want 5: the counter cannot miss what the log did", events[0].Occurrences)
	}
	if len(events[0].DistinctProcesses) != 1 {
		t.Errorf("distinct = %v, want one entry", events[0].DistinctProcesses)
	}
}
