package main

import (
	"log/slog"
	"sync"
	"time"

	"github.com/denisakp/ogoune/pkg/agentwire"
)

// Kernel event capture (spec 090).
//
// One rule shapes everything here: capture is best-effort and metrics are not.
// Where the two conflict, metrics win. Every failure path costs the events and
// nothing else — no error reaches the streaming loop, nothing here can terminate
// the agent, and nothing retries in a way that could.
//
// That is not defensive polish. The documented default deployment is a container,
// and a container usually cannot read /dev/kmsg, so "capture is unavailable" is
// the path most installations take on day one.

// maxEventsPerInterval bounds how many aggregated events one interval may
// produce. Two kinds today, so this is slack rather than a real constraint; it
// exists so a future kind cannot turn a log storm into an unbounded frame.
const maxEventsPerInterval = 16

// kernelSource is a reader of kernel reports. Implemented on Linux, inert
// everywhere else — the platform seam lives at file level so the rest of the
// agent never branches on it.
type kernelSource interface {
	// Name identifies the source in stored events, so an operator investigating
	// missing events knows which reader was working.
	Name() string
	// Drain returns the reports seen since the previous call. It never blocks and
	// never returns an error: a source that cannot read reports nothing, which is
	// the same thing as a quiet host as far as the caller is concerned.
	Drain() []kmsgReport
	// Close releases whatever the source holds.
	Close()
}

// eventCollector aggregates kernel reports into per-interval events.
//
// Per kind per interval, with a count and a bounded list of the distinct
// processes affected: a storm killing forty copies of one process and a storm
// killing forty different ones are different problems, and thirty-seven identical
// rows on a host page is noise rather than a finding (FR-028).
type eventCollector struct {
	mu      sync.Mutex
	sources []kernelSource
	dropped int

	// unavailableLogged makes the "capture unavailable" report happen exactly
	// once per process lifetime, whatever the uptime (FR-008).
	unavailableLogged bool
}

func newEventCollector(sources []kernelSource) *eventCollector {
	c := &eventCollector{sources: sources}
	if len(sources) == 0 {
		c.reportUnavailable("no kernel event source is available on this host")
	}
	return c
}

// reportUnavailable logs the unavailability once and remembers it. Called at
// startup and never on a per-interval path.
func (c *eventCollector) reportUnavailable(reason string) {
	if c.unavailableLogged {
		return
	}
	c.unavailableLogged = true
	slog.Info("agent: kernel event capture unavailable, metrics unaffected", "reason", reason)
}

// Collect drains every source and returns the interval's aggregated events.
//
// Returns nil rather than an error under every condition. A caller that gets
// nothing back cannot tell a quiet host from an unreadable one, and does not need
// to: both mean "no events this interval", and neither is a reason to disturb the
// metrics frame this rides on.
func (c *eventCollector) Collect(now time.Time) []agentwire.KernelEvent {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.sources) == 0 {
		return nil
	}

	// Aggregate by kind. Fixed keys, so what is retained cannot grow with what the
	// kernel does — the moment the agent must stay smallest is exactly the moment
	// a host is producing thousands of reports.
	agg := map[string]*eventAccumulator{}
	for _, src := range c.sources {
		for _, r := range src.Drain() {
			a, ok := agg[r.Kind]
			if !ok {
				if len(agg) >= maxEventsPerInterval {
					c.dropped++
					continue
				}
				a = &eventAccumulator{kind: r.Kind, source: src.Name(), seen: map[string]bool{}}
				agg[r.Kind] = a
			}
			a.add(r)
		}
	}

	out := make([]agentwire.KernelEvent, 0, len(agg))
	for _, a := range agg {
		out = append(out, a.event(now))
	}
	return out
}

// Dropped reports how many kinds were refused because the per-interval bound was
// reached, so a burst larger than what is shown is visible rather than silent.
func (c *eventCollector) Dropped() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.dropped
}

// Close releases every source.
func (c *eventCollector) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, s := range c.sources {
		s.Close()
	}
}

// eventAccumulator folds many reports of one kind into a single event.
type eventAccumulator struct {
	kind        string
	source      string
	occurrences int

	// first is the most representative report: the one that opened the interval.
	// Picking the first rather than the last is arbitrary but stable, and stable
	// matters more here than which one it is.
	first kmsgReport

	// seen and distinct track which processes were affected. seen is bounded
	// alongside distinct, so neither grows with the storm.
	seen      map[string]bool
	distinct  []string
	truncated bool
}

func (a *eventAccumulator) add(r kmsgReport) {
	if a.occurrences == 0 {
		a.first = r
	}
	a.occurrences++

	if r.Process == "" {
		return
	}
	if a.seen[r.Process] {
		return
	}
	if len(a.distinct) >= agentwire.MaxDistinctProcesses {
		// Marked, never silently shortened: a partial list read as complete is
		// worse than an obviously partial one (FR-028b).
		a.truncated = true
		return
	}
	a.seen[r.Process] = true
	a.distinct = append(a.distinct, r.Process)
}

func (a *eventAccumulator) event(now time.Time) agentwire.KernelEvent {
	return agentwire.KernelEvent{
		Kind:              a.kind,
		OccurredAt:        now,
		Source:            a.source,
		Occurrences:       a.occurrences,
		Process:           a.first.Process,
		PID:               a.first.PID,
		DistinctProcesses: a.distinct,
		DistinctTruncated: a.truncated,
	}
}
