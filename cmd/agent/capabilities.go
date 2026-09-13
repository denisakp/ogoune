package main

import (
	"log/slog"

	"github.com/denisakp/ogoune/pkg/agentwire"
)

// Capability declaration. Before every frame the agent asks three
// questions about the machine it is on -- can the kernel log be read, can the
// cgroup OOM counter be read, can segfaults be captured -- and puts the answers
// on the frame. It describes SOURCES, never events: a quiet machine and a blind
// agent are different answers, and until now they looked the same.
//
// Probing is not opening. The collector's live sources were opened once at
// startup and are never replaced: the kernel-log reader sits at the position it
// reached, the cgroup reader holds its baseline, and swapping either would
// replay old kills as if they had just happened. The probe opens and closes,
// and reports what it saw.
//
// The one thing a probe does change: if a source the collector lacks becomes
// readable -- the container was recreated with kernel-log access, say -- the
// collector ADOPTS a freshly opened one. Without that, the declaration would
// say "available" about a source nobody was reading, which is the false-health
// gap this feature exists to close.

// capabilityProbe answers the three questions for this platform.
type capabilityProbe interface {
	Probe() agentwire.Capabilities
}

// sourceOpener opens a kernel source by name (agentwire.SourceKmsg,
// agentwire.SourceCgroup), or returns nil when it cannot. Injected so the
// adoption rule is testable without a kernel.
type sourceOpener func(name string) kernelSource

// adopt adds a source the collector does not yet have. Returns whether it did.
// Idempotent per name: a second adoption of the same source is a no-op, so a
// probe that keeps saying "available" does not keep opening readers.
func (c *eventCollector) adopt(src kernelSource) bool {
	if c == nil || src == nil {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, s := range c.sources {
		if s.Name() == src.Name() {
			return false
		}
	}
	c.sources = append(c.sources, src)
	// The one operator-visible line this feature adds to the log: capture
	// started mid-run. The operator who just restarted the container with
	// kernel-log access deserves it here as well as on the host page.
	slog.Info("agent: kernel event source adopted", "source", src.Name())
	return true
}

func (c *eventCollector) has(name string) bool {
	if c == nil {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, s := range c.sources {
		if s.Name() == name {
			return true
		}
	}
	return false
}

// probeAndAdopt runs the probe and reconciles the collector with what it said:
// a capability that probes available while the collector has no reader for it
// gets one opened and adopted. A failed probe never removes a live reader -- a
// descriptor already open stays readable, and the declaration still reports
// what the probe saw, which is the honest answer for a NEW agent on this box.
func probeAndAdopt(p capabilityProbe, c *eventCollector, open sourceOpener) agentwire.Capabilities {
	caps := p.Probe()
	if open == nil || c == nil {
		return caps
	}
	for name, cap := range map[string]agentwire.Capability{
		agentwire.SourceKmsg:   caps.Kmsg,
		agentwire.SourceCgroup: caps.CgroupOOM,
	} {
		switch {
		case cap.Available && !c.has(name):
			if src := open(name); src != nil {
				c.adopt(src)
			} else {
				// Probed readable, failed to open a moment later. Rare and
				// transient; the next interval tries again.
				slog.Debug("agent: source probed available but did not open", "source", name)
			}
		case !cap.Available && c.has(name):
			slog.Debug("agent: source probed unavailable while a reader is live", "source", name, "reason", cap.Reason)
		}
	}
	return caps
}

// openKernelSource is the production opener: the same constructors startup
// uses, so an adopted reader starts at the end of the log / at the current
// counter exactly as one opened at boot would.
func openKernelSource(name string) kernelSource {
	switch name {
	case agentwire.SourceKmsg:
		return newKmsgSourceIfAny()
	case agentwire.SourceCgroup:
		return newCgroupOOMSourceIfAny()
	}
	return nil
}
