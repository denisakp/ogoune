package main

import (
	"bytes"
	"os"
	"strings"

	"golang.org/x/sys/unix"

	"github.com/denisakp/ogoune/pkg/agentwire"
)

// Probe paths are variables so the probe can be pointed at a fake filesystem in
// tests; the readers keep their own constants.
var (
	probeKmsgPath           = kmsgPath
	probeCgroupEventsPath   = cgroupEventsPath
	probeExceptionTracePath = "/proc/sys/debug/exception-trace"
)

type linuxProbe struct{}

func newCapabilityProbe() capabilityProbe { return linuxProbe{} }

func (linuxProbe) Probe() agentwire.Capabilities {
	kmsg := probeKmsg()
	cgroup := probeCgroupOOM()
	return agentwire.Capabilities{
		Kmsg:      kmsg,
		CgroupOOM: cgroup,
		Segfault:  probeSegfault(kmsg),
	}
}

// probeKmsg: can the kernel log be opened for reading? Open and close; the live
// reader, if any, is untouched.
func probeKmsg() agentwire.Capability {
	fd, err := unix.Open(probeKmsgPath, unix.O_RDONLY|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)
	if err != nil {
		return agentwire.Capability{Available: false, Reason: agentwire.ReasonUnreadable}
	}
	_ = unix.Close(fd)
	return agentwire.Capability{Available: true}
}

// probeCgroupOOM: is the cgroup v2 memory.events file readable and does it
// carry an oom_kill line? A cgroup v1 host has neither.
func probeCgroupOOM() agentwire.Capability {
	b, err := os.ReadFile(probeCgroupEventsPath)
	if err != nil || !bytes.Contains(b, []byte("oom_kill")) {
		return agentwire.Capability{Available: false, Reason: agentwire.ReasonUnreadable}
	}
	return agentwire.Capability{Available: true}
}

// probeSegfault needs both the kernel log and the kernel asked to report
// userspace faults. Reason precedence matters: when the log is the missing
// piece the reason is the log, not the setting, so the operator is not sent to
// flip a sysctl that would change nothing.
func probeSegfault(kmsg agentwire.Capability) agentwire.Capability {
	if !kmsg.Available {
		return agentwire.Capability{Available: false, Reason: agentwire.ReasonUnreadable}
	}
	b, err := os.ReadFile(probeExceptionTracePath)
	if err != nil {
		return agentwire.Capability{Available: false, Reason: agentwire.ReasonUnreadable}
	}
	if strings.TrimSpace(string(b)) != "1" {
		return agentwire.Capability{Available: false, Reason: agentwire.ReasonSettingOff}
	}
	return agentwire.Capability{Available: true}
}

func newKmsgSourceIfAny() kernelSource      { return newKmsgSource() }
func newCgroupOOMSourceIfAny() kernelSource { return newCgroupOOMSource() }
