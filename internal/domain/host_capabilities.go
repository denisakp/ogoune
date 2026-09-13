package domain

// CapabilityReason is why a capability is unavailable, from the agent's fixed
// vocabulary (spec 093). Empty when available.
type CapabilityReason string

const (
	CapabilityReasonUnreadable CapabilityReason = "unreadable"  // the source cannot be opened or read
	CapabilityReasonSettingOff CapabilityReason = "setting_off" // segfault only: the kernel is not asked to report faults
	CapabilityReasonPlatform   CapabilityReason = "platform"    // not Linux
)

// Capability is whether one observation source can be read, and if not, why.
type Capability struct {
	Available bool             `json:"available"`
	Reason    CapabilityReason `json:"reason,omitempty"`
}

// HostCapabilities is what an agent declares it can observe on its machine:
// the kernel log, the cgroup OOM counter, and segfault capture (which needs
// both the log and a kernel setting). It describes sources, never events.
type HostCapabilities struct {
	Kmsg      Capability `json:"kmsg"`
	CgroupOOM Capability `json:"cgroup_oom"`
	Segfault  Capability `json:"segfault"`
}

// OOMDetail is the interface's answer to "are out-of-memory kills detected,
// and how well?" -- derived, not declared. The kernel log names the killed
// process; the cgroup counter only counts.
type OOMDetail string

const (
	OOMDetailWithProcess    OOMDetail = "with_process"
	OOMDetailWithoutProcess OOMDetail = "without_process"
	OOMDetailNone           OOMDetail = "none"
)

// OOMDetail derives how OOM kills are detected from which sources are readable.
func (c HostCapabilities) OOMDetail() OOMDetail {
	switch {
	case c.Kmsg.Available:
		return OOMDetailWithProcess
	case c.CgroupOOM.Available:
		return OOMDetailWithoutProcess
	default:
		return OOMDetailNone
	}
}

// HostCapabilitiesState is the situation a declaration -- or its absence -- was
// read in. The two not-declared states are different facts: an agent that has
// never connected says nothing about the machine; an agent that connects but
// does not declare is simply too old to. Neither may be rendered as
// "unavailable", which would be a false statement about the machine.
type HostCapabilitiesState string

const (
	HostCapabilitiesDeclared    HostCapabilitiesState = "declared"
	HostCapabilitiesNotReported HostCapabilitiesState = "not_reported" // agent connected, no declaration
	HostCapabilitiesNotKnown    HostCapabilitiesState = "not_known"    // never connected; pre-feature incident; lookup failed
	HostCapabilitiesNoMachine   HostCapabilitiesState = "no_machine"   // incident only: the monitor had no host
)

// CapabilitiesState derives the host's current state from two columns that
// move together: a declaration present means declared; none but a last-seen
// time means the agent connects without declaring; neither means it never
// connected. Nil-safe.
func (h *Host) CapabilitiesState() HostCapabilitiesState {
	switch {
	case h == nil:
		return HostCapabilitiesNotKnown
	case h.Capabilities != nil:
		return HostCapabilitiesDeclared
	case h.LastSeenAt != nil:
		return HostCapabilitiesNotReported
	default:
		return HostCapabilitiesNotKnown
	}
}

// FrozenCapabilitiesState is the state copied onto the incident when it opened.
// A nil stored state is a row from before the feature, and reads as not known:
// nothing was recorded, and the host's current declaration is not a substitute
// for what it was then. Nil-safe.
func (i *Incident) FrozenCapabilitiesState() HostCapabilitiesState {
	if i == nil || i.HostCapabilitiesState == nil {
		return HostCapabilitiesNotKnown
	}
	return *i.HostCapabilitiesState
}
