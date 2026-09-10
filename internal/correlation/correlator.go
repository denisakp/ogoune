package correlation

import (
	"context"
	"log/slog"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/narrative"
	"github.com/denisakp/ogoune/internal/port"
)

// maxWindowEvents bounds how many events one window will read.
//
// A storm-prone host can hold many rows in a six-minute window. The sentence
// names one and counts the rest, and the served list is meant to be read by a
// human, so neither grows with the storm. The bound is what stops the query
// from growing with it either.
const maxWindowEvents = 50

// Result is everything one correlation lookup produced: the events that fell in
// the incident's window, and the explanation naming one of them.
//
// Both come from a single read. The incident detail path needs the events too --
// a sentence whose evidence is not served alongside it cannot be checked -- and
// issuing a second query for them would be paying twice for one answer.
type Result struct {
	// Events are the window's events, newest first. Includes kinds this version
	// cannot phrase: they are shown and counted, just never named.
	Events []*domain.HostEvent
	// Explanation is nil whenever there is nothing to say.
	Explanation *domain.IncidentExplanation
}

// Correlator joins an incident to what the kernel reported on its host.
type Correlator struct {
	events port.HostEventRepository
	hosts  port.HostRepository
	now    func() time.Time
}

func New(events port.HostEventRepository, hosts port.HostRepository) *Correlator {
	return &Correlator{events: events, hosts: hosts, now: time.Now}
}

// WithClock replaces the clock. Only the window's far edge is clamped against
// it, so this exists for tests rather than for configuration.
func (c *Correlator) WithClock(fn func() time.Time) *Correlator {
	if fn != nil {
		c.now = fn
	}
	return c
}

// ForIncident returns the window's events and, when one of them can be named, an
// explanation.
//
// Best-effort by contract: every failure path returns a zero Result rather than
// an error. The caller is rendering an incident or dispatching an alert, and
// neither must ever be lost to its own enrichment.
func (c *Correlator) ForIncident(ctx context.Context, incident *domain.Incident) Result {
	if c == nil || c.events == nil || c.hosts == nil || incident == nil {
		return Result{}
	}
	// The normal state of most monitors, and the reason this feature costs
	// nothing to operators who run no agent: return before issuing any query,
	// not after finding nothing.
	if incident.Resource.HostID == nil || *incident.Resource.HostID == "" {
		return Result{}
	}
	hostID := *incident.Resource.HostID

	from := incident.StartedAt.Add(-domain.HostContextWindowBefore)
	to := incident.StartedAt.Add(domain.HostContextWindowAfter)
	// The window may still be elapsing when an operator opens a fresh incident,
	// or when the alert is dispatched. Read what exists rather than wait for it.
	if now := c.now(); to.After(now) {
		to = now
	}
	if !to.After(from) {
		return Result{}
	}

	events, err := c.events.ListInWindow(ctx, hostID, from, to, maxWindowEvents)
	if err != nil {
		slog.Debug("correlation: window lookup failed",
			"incident_id", incident.ID, "host_id", hostID, "error", err)
		return Result{}
	}
	if len(events) == 0 {
		// Absence is silence: no event means no sentence, not a sentence saying
		// nothing happened (FR-008).
		return Result{}
	}

	named := selectEvent(events, incident.StartedAt)
	if named == nil {
		// Events exist but none of their kinds can be phrased. They are still
		// served and still visible; there is simply no sentence (FR-007).
		return Result{Events: events}
	}

	host, err := c.hosts.FindByID(ctx, hostID)
	if err != nil || host == nil {
		slog.Debug("correlation: host lookup failed",
			"incident_id", incident.ID, "host_id", hostID, "error", err)
		return Result{Events: events}
	}

	return Result{
		Events: events,
		Explanation: &domain.IncidentExplanation{
			HostID:      hostID,
			HostName:    host.Name,
			IncidentAt:  incident.StartedAt,
			Cause:       incident.Cause,
			Event:       *named,
			Precedes:    !named.OccurredAt.After(incident.StartedAt),
			OtherEvents: len(events) - 1,
			WindowFrom:  from,
			WindowTo:    to,
		},
	}
}

// selectEvent picks the event a sentence will name: the LATEST one at or before
// the incident started; if none precedes it, the EARLIEST one after.
//
// "The last thing the kernel reported before the check failed" is explainable to
// an operator in one clause, which is the property that matters -- a rule whose
// output cannot be explained is a model, and this feature is forbidden one
// (FR-003). Nearest-in-absolute-time was rejected for naming a segfault one
// second after the failure over an out-of-memory kill three seconds before it,
// which reads backwards.
//
// Total and deterministic: identical timestamps break on the smaller id, so the
// same stored facts always yield the same sentence (FR-018a, SC-010). Only kinds
// this version can phrase are eligible; the rest are counted, never named.
func selectEvent(events []*domain.HostEvent, startedAt time.Time) *domain.HostEvent {
	var before, after *domain.HostEvent
	for _, e := range events {
		if e == nil || !narrative.Phrasable(e.Kind) {
			continue
		}
		if e.OccurredAt.After(startedAt) {
			if after == nil || beats(e, after, false) {
				after = e
			}
			continue
		}
		if before == nil || beats(e, before, true) {
			before = e
		}
	}
	if before != nil {
		return before
	}
	return after
}

// beats reports whether cand should replace cur, given whether the caller wants
// the latest of a set or the earliest.
//
// The tie-break goes to the smaller id in BOTH directions, which is why it
// cannot be folded into a single "is earlier than" comparison: wanting the
// latest means preferring the later timestamp but the smaller id, and one
// comparator answering both questions gets the tie backwards. Ids are ULIDs, so
// the tie-break is stable rather than meaningful -- stable is all determinism
// needs (FR-018a).
func beats(cand, cur *domain.HostEvent, wantLatest bool) bool {
	if cand.OccurredAt.Equal(cur.OccurredAt) {
		return cand.ID < cur.ID
	}
	if wantLatest {
		return cand.OccurredAt.After(cur.OccurredAt)
	}
	return cand.OccurredAt.Before(cur.OccurredAt)
}
