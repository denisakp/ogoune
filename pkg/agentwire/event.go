package agentwire

import "time"

// Kernel event kinds (spec 090). An open set: a backend must store a kind it does
// not recognise rather than drop it, so an older backend paired with a newer agent
// loses nothing.
const (
	KindOOMKill  = "oom_kill"
	KindSegfault = "segfault"
)

// Event sources. Recorded because the same kill can be seen by more than one, and
// because an operator investigating missing events needs to know which reader was
// working.
const (
	SourceKmsg   = "kmsg"
	SourceCgroup = "cgroup"
)

// MaxDistinctProcesses bounds the distinct-process list carried on an event. The
// list exists to tell a storm killing forty copies of one process from a storm
// killing forty different ones -- materially different problems. The bound exists
// so the list cannot grow with the storm.
const MaxDistinctProcesses = 8

// KernelEvent is what the kernel reported about processes on a host during one
// collection interval. One per kind per interval, carrying how many reports it
// aggregates: an OOM storm is one event with Occurrences=200, not two hundred
// events (spec 090, FR-028).
//
// There is deliberately no field for the raw kernel line. /dev/kmsg carries every
// subsystem's output, in formats this project does not control and that change
// between kernel versions; storing it would mean keeping content nobody has
// examined. Classification extracts what is actionable and discards the rest
// (FR-027a). Having nowhere to put a raw line is the enforcement.
type KernelEvent struct {
	Kind        string    `json:"kind"`
	OccurredAt  time.Time `json:"occurred_at"`
	Source      string    `json:"source"`
	Occurrences int       `json:"occurrences"`

	// Process and PID describe the most representative report in the interval.
	Process string `json:"process,omitempty"`
	PID     int    `json:"pid,omitempty"`
	// Cgroup is the cgroup path or container identifier, when the kernel named one.
	Cgroup string `json:"cgroup,omitempty"`

	// DistinctProcesses is bounded by MaxDistinctProcesses. DistinctTruncated says
	// more were seen than the list holds -- marked rather than silently shortened,
	// so a partial list is never read as complete (FR-028b).
	DistinctProcesses []string `json:"distinct_processes,omitempty"`
	DistinctTruncated bool     `json:"distinct_truncated,omitempty"`
}
