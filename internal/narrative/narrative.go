package narrative

import (
	"fmt"
	"strings"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
)

// timeLayout is verbose on purpose. An operator reading an alert at 3am should
// never have to work out which day, or which zone, a timestamp belongs to --
// and the two timestamps in a sentence are what let them overrule it.
const timeLayout = "2006-01-02 15:04:05 MST"

// Sentence renders the explanation as one sentence of plain text.
//
// Pure: no context, no repository, no clock. The same input always produces the
// same bytes, which is what makes "generated on read" safe -- without it, an
// incident could read differently on each refresh (FR-018a, SC-010).
//
// Returns "" when the explanation is nil or names a kind this version cannot
// phrase. An empty string means "say nothing", never a half-built sentence with
// a gap in it: absence is silence (FR-007, FR-008).
func Sentence(e *domain.IncidentExplanation) string {
	if e == nil {
		return ""
	}
	phrase, ok := kindPhrase[e.Event.Kind]
	if !ok {
		return ""
	}

	var b strings.Builder

	// What the check observed, and when. The cause is the operator's own words
	// for the failure; it leads because it is what they were paged about.
	cause := strings.TrimSpace(e.Cause)
	if cause == "" {
		cause = "This check failed"
	}
	fmt.Fprintf(&b, "%s at %s. ", cause, e.IncidentAt.Format(timeLayout))

	// What the kernel reported, and when.
	fmt.Fprintf(&b, "The kernel %s %s on %s at %s",
		phrase, target(e.Event.Detail), hostLabel(e), e.Event.OccurredAt.Format(timeLayout))

	// How far apart, and in which direction. Never "before" when it means
	// "after" (FR-010).
	if gap := relative(e); gap != "" {
		fmt.Fprintf(&b, ", %s", gap)
	}
	b.WriteString(".")

	// How many kernel reports the named event aggregates. A separate question
	// from how many events fell in the window, and merging the two would
	// misdescribe both (R8).
	if e.Event.Occurrences > 1 {
		fmt.Fprintf(&b, " Reported %d times.", e.Event.Occurrences)
	}

	// How many OTHER events were in the window. Named one, counted the rest --
	// never a list (FR-005).
	switch {
	case e.OtherEvents == 1:
		b.WriteString(" 1 other kernel event fell in the same window.")
	case e.OtherEvents > 1:
		fmt.Fprintf(&b, " %d other kernel events fell in the same window.", e.OtherEvents)
	}

	return b.String()
}

// hostLabel prefers the host's name and falls back to its id. A sentence naming
// neither would be unactionable, and an operator can act on an id.
func hostLabel(e *domain.IncidentExplanation) string {
	if name := strings.TrimSpace(e.HostName); name != "" {
		return name
	}
	return e.HostID
}

// target names what the kernel acted on, as precisely as the report allowed.
//
// The kernel does not always say. "a process" is the honest answer when it did
// not, and is preferable to a sentence implying knowledge the event does not
// carry -- the cgroup counter, for instance, reports how many, never which.
func target(d *domain.HostEventDetail) string {
	if d == nil {
		return "a process"
	}
	if p := strings.TrimSpace(d.Process); p != "" {
		if d.PID > 0 {
			return fmt.Sprintf("%s (pid %d)", p, d.PID)
		}
		return p
	}
	// No representative process, but a list of distinct ones: a storm killing
	// forty copies of one process and one killing forty different processes are
	// materially different problems, so the list is worth saying.
	if n := len(d.DistinctProcesses); n > 0 {
		joined := strings.Join(d.DistinctProcesses, ", ")
		if d.DistinctTruncated {
			return joined + " and more"
		}
		if n == 1 {
			return joined
		}
		return joined
	}
	return "a process"
}

// relative says how far the event sits from the failure, and on which side.
//
// Returns "" for a simultaneous pair rather than "0 seconds before", which would
// read as precision the two clocks do not have -- the event's timestamp comes
// from the host's kernel and the incident's from this server.
func relative(e *domain.IncidentExplanation) string {
	d := e.IncidentAt.Sub(e.Event.OccurredAt)
	if d < 0 {
		d = -d
	}
	d = d.Round(time.Second)
	if d == 0 {
		return ""
	}
	side := "later"
	if e.Precedes {
		side = "earlier"
	}
	return fmt.Sprintf("%s %s", humanGap(d), side)
}

// humanGap spells out a duration inside the correlation window, which is at
// most a few minutes wide.
func humanGap(d time.Duration) string {
	secs := int(d.Seconds())
	if secs < 60 {
		return plural(secs, "second")
	}
	mins := secs / 60
	rem := secs % 60
	if rem == 0 {
		return plural(mins, "minute")
	}
	return fmt.Sprintf("%s %s", plural(mins, "minute"), plural(rem, "second"))
}

// plural keeps "1 second" from reading as "1 seconds". A sentence an operator
// is meant to trust cannot be sloppy about the thing it is counting.
func plural(n int, unit string) string {
	if n == 1 {
		return fmt.Sprintf("1 %s", unit)
	}
	return fmt.Sprintf("%d %ss", n, unit)
}
