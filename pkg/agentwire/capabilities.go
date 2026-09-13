package agentwire

// Capabilities is what the agent can observe on the machine it runs on, probed
// before every frame and restated on each one (spec 093). It says whether each
// SOURCE can be read -- never whether an event has been seen through it. A quiet
// machine and a blind agent are different answers, and this is the difference.
//
// Optional on the frame. An agent predating the feature sends none, and the
// backend records "not reported" for that host; a backend predating the feature
// ignores the object, because encoding/json drops keys it does not know.
type Capabilities struct {
	// Kmsg: the kernel log (/dev/kmsg) can be opened for reading. Usually false
	// in a container, which is the recommended install.
	Kmsg Capability `json:"kmsg"`
	// CgroupOOM: the cgroup v2 memory.events counter can be read. Often true even
	// where the kernel log is not, so OOM kills are still detected -- without the
	// name of the killed process.
	CgroupOOM Capability `json:"cgroup_oom"`
	// Segfault: userspace faults can be captured. Needs BOTH the kernel log and
	// the kernel asked to report them (debug.exception-trace=1). Best-effort by
	// product decision: the weakest signal does not dictate install requirements.
	Segfault Capability `json:"segfault"`
}

// Capability is one source's readability plus, when unavailable, a reason drawn
// from a fixed vocabulary. The agent never sends prose; the interface maps the
// reason to a sentence and, where there is one, to the action that enables it.
type Capability struct {
	Available bool   `json:"available"`
	Reason    string `json:"reason,omitempty"`
}

// Reasons a capability is unavailable. Anything else on the wire is normalised
// to ReasonUnreadable: a newer agent with a reason this backend does not know
// still yields a truthful-enough state, never a rejected frame.
const (
	// ReasonUnreadable: the source exists but cannot be opened or read. The
	// container case for the kernel log; also what segfault reports when the
	// log is the missing piece, so the operator is sent to fix the right thing.
	ReasonUnreadable = "unreadable"
	// ReasonSettingOff: segfault only -- the kernel is not asked to report
	// userspace faults. The one reason with a one-line fix.
	ReasonSettingOff = "setting_off"
	// ReasonPlatform: not Linux. Nothing to enable.
	ReasonPlatform = "platform"
)

func knownReason(r string) bool {
	switch r {
	case ReasonUnreadable, ReasonSettingOff, ReasonPlatform:
		return true
	}
	return false
}

// Normalize brings a declaration into the shape the backend reasons about, in
// place. It never rejects: an unknown reason becomes ReasonUnreadable, an
// available capability carries no reason, and a segfault capability cannot be
// available without the kernel log it depends on. Returns whether anything was
// changed, so a caller with a logger can say so.
func (c *Capabilities) Normalize() bool {
	if c == nil {
		return false
	}
	changed := false
	for _, cap := range []*Capability{&c.Kmsg, &c.CgroupOOM, &c.Segfault} {
		switch {
		case cap.Available && cap.Reason != "":
			cap.Reason = ""
			changed = true
		case !cap.Available && !knownReason(cap.Reason):
			cap.Reason = ReasonUnreadable
			changed = true
		}
	}
	if c.Segfault.Available && !c.Kmsg.Available {
		c.Segfault = Capability{Available: false, Reason: ReasonUnreadable}
		changed = true
	}
	return changed
}
