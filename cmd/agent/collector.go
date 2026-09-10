package main

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"

	"github.com/denisakp/ogoune/pkg/agentwire"
)

// Collector produces one metrics frame per call. Behind an interface so the
// stream loop can be unit-tested with a fake, keeping gopsutil off the test path.
type Collector interface {
	Collect(ctx context.Context) (agentwire.Frame, error)
}

// gopsutilCollector reads host metrics via gopsutil. A single failing metric is
// logged and skipped (its field stays zero / the mount is omitted) rather than
// failing the whole frame (FR-014).
type gopsutilCollector struct {
	agentVersion string
	// events is optional. Nil means kernel capture is not attached at all, which
	// is exactly how a host that cannot read its kernel log behaves: the frame
	// simply carries no events (spec 090).
	events *eventCollector
	// eventsInFlight guards the deadline below: if a previous collection is
	// somehow still running, skip this interval rather than stacking another
	// goroutine behind it every tick.
	eventsInFlight atomic.Bool
	// eventsStalled reports the stall once, not once per interval.
	eventsStalled atomic.Bool
}

// eventCollectDeadline bounds how long kernel-event capture may take before the
// metrics frame leaves without it.
//
// This exists because the claim "capture never blocks" was once a comment rather
// than a mechanism, and a blocking read on /dev/kmsg parked the collector
// forever -- the agent stopped streaming metrics entirely on any host where the
// kernel log was actually readable. The read is non-blocking now; this is the
// guarantee that a future mistake degrades to "no events this interval" instead
// of to dead monitoring.
const eventCollectDeadline = 2 * time.Second

func newGopsutilCollector(agentVersion string) *gopsutilCollector {
	return &gopsutilCollector{agentVersion: agentVersion}
}

// WithKernelEvents attaches kernel event capture. Separate from the constructor
// so capture stays optional and every existing call site keeps working.
func (c *gopsutilCollector) WithKernelEvents(e *eventCollector) *gopsutilCollector {
	c.events = e
	return c
}

func (c *gopsutilCollector) Collect(ctx context.Context) (agentwire.Frame, error) {
	f := agentwire.Frame{AgentVersion: c.agentVersion}

	if info, err := host.InfoWithContext(ctx); err != nil {
		slog.Debug("collect: host info failed", "error", err)
	} else {
		f.OS = osLabel(info)
	}

	// CPU: percent since the previous call (≈ over the last interval).
	if pcts, err := cpu.PercentWithContext(ctx, 0, false); err != nil || len(pcts) == 0 {
		slog.Debug("collect: cpu failed", "error", err)
	} else {
		f.CPUPct = pcts[0]
	}

	if vm, err := mem.VirtualMemoryWithContext(ctx); err != nil {
		slog.Debug("collect: mem failed", "error", err)
	} else {
		f.MemPct = vm.UsedPercent
	}

	if counters, err := net.IOCountersWithContext(ctx, false); err != nil || len(counters) == 0 {
		slog.Debug("collect: net failed", "error", err)
	} else {
		f.NetIn = int64(counters[0].BytesRecv)
		f.NetOut = int64(counters[0].BytesSent)
	}

	f.Disks = collectDisks(ctx)

	// Kernel events observed during this interval (spec 090). Best-effort by
	// contract, and enforced rather than asserted: capture cannot return an
	// error, and it cannot delay this frame past eventCollectDeadline.
	f.Events = c.collectEvents()

	return f, nil
}

// collectEvents returns this interval's kernel events, or nothing if capture is
// absent, already running, or slow.
//
// A stalled collection leaks its goroutine rather than being cancelled: a read
// already inside a syscall cannot be interrupted from here. That is the right
// trade -- one parked goroutine costs a few kilobytes, while waiting on it costs
// the operator the monitoring they actually rely on.
func (c *gopsutilCollector) collectEvents() []agentwire.KernelEvent {
	if c.events == nil {
		return nil
	}
	if !c.eventsInFlight.CompareAndSwap(false, true) {
		return nil
	}

	done := make(chan []agentwire.KernelEvent, 1)
	go func() {
		defer c.eventsInFlight.Store(false)
		done <- c.events.Collect(time.Now().UTC())
	}()

	select {
	case events := <-done:
		return events
	case <-time.After(eventCollectDeadline):
		if c.eventsStalled.CompareAndSwap(false, true) {
			slog.Warn("agent: kernel event capture is slow; metrics continue without events",
				"deadline", eventCollectDeadline)
		}
		return nil
	}
}

// pseudoFSTypes are non-storage filesystems we never report usage for (kernel
// virtual FS, cgroups, read-only snap images, etc.). Everything else — ext4,
// xfs, btrfs, zfs, vfat, and container/VM filesystems like overlay, virtiofs,
// and 9p — is reported. Using Partitions(all=true) + this denylist (instead of
// Partitions(physical=true)) is what makes disk usage show up inside containers
// and VMs, whose root FS is not a /dev-backed "physical" device.
var pseudoFSTypes = map[string]struct{}{
	"proc": {}, "sysfs": {}, "tmpfs": {}, "devtmpfs": {}, "devpts": {},
	"cgroup": {}, "cgroup2": {}, "mqueue": {}, "debugfs": {}, "tracefs": {},
	"securityfs": {}, "pstore": {}, "bpf": {}, "configfs": {}, "fusectl": {},
	"hugetlbfs": {}, "rpc_pipefs": {}, "binfmt_misc": {}, "autofs": {},
	"nsfs": {}, "ramfs": {}, "squashfs": {}, "efivarfs": {}, "fuse.gvfsd-fuse": {},
	"selinuxfs": {}, "sysctlfs": {}, "cgroupfs": {},
}

// collectDisks returns per-mount usage for real (non-pseudo) filesystems,
// skipping kernel/virtual mounts and anything that errors or reports zero total.
// maxReportedFilesystems bounds what one frame carries after deduplication.
//
// A host with more than this many distinct filesystems exists, but a page
// listing them is not read by anyone. The cap keeps the frame and the stored
// sample bounded; the entries kept are the fullest, because those are the ones
// somebody is going to be paged about.
const maxReportedFilesystems = 32

// collectDisks reports one entry per FILESYSTEM, not per mount point.
//
// A mount table is not a list of disks. A btrfs root with subvolumes, a ZFS
// pool, a Docker or Kubernetes node with overlay layers, any bind mount -- each
// produces many mount points backed by one filesystem, all reporting the same
// capacity. A development VM here had 486 mounts over 18 devices: 434 of them on
// a single /dev/vdb1, every one answering "188G, 145G used, 77%".
//
// Reporting them all cost three things: a host page nobody can read, a stored
// sample per interval carrying hundreds of identical rows, and one statfs syscall
// per mount every collection -- 442 of them, ten seconds apart, to learn the same
// three numbers.
//
// Grouping happens BEFORE the usage lookup, which is what removes the syscalls
// rather than merely the duplicate rows.
func collectDisks(ctx context.Context) []agentwire.DiskUsage {
	parts, err := disk.PartitionsWithContext(ctx, true)
	if err != nil {
		slog.Debug("collect: disk partitions failed", "error", err)
		return nil
	}

	out := make([]agentwire.DiskUsage, 0, 8)
	for _, mount := range representativeMounts(parts) {
		u, err := disk.UsageWithContext(ctx, mount)
		if err != nil {
			slog.Debug("collect: disk usage failed", "mount", mount, "error", err)
			continue
		}
		if u.Total == 0 {
			continue // virtual / empty mount
		}
		out = append(out, agentwire.DiskUsage{Mount: mount, UsedPct: u.UsedPercent})
	}

	return capAndSort(out)
}

// representativeMounts picks one mount per device: the shallowest path, because
// "/" is more use to an operator than "/opt/vendor/data/subvol", with ties
// broken lexicographically so the choice never depends on mount order.
//
// Grouping by device NAME rather than by filesystem identity collapses several
// overlay mounts that share the name "overlay" into one. That is the right
// answer rather than a compromise: each of them reports the backing
// filesystem's capacity, so they are the same numbers under different paths.
func representativeMounts(parts []disk.PartitionStat) []string {
	best := make(map[string]string, len(parts))
	for _, p := range parts {
		if _, pseudo := pseudoFSTypes[p.Fstype]; pseudo {
			continue
		}
		if isSystemMount(p.Mountpoint) || p.Device == "" {
			continue
		}
		if cur, ok := best[p.Device]; !ok || shorterMount(p.Mountpoint, cur) {
			best[p.Device] = p.Mountpoint
		}
	}

	out := make([]string, 0, len(best))
	for _, m := range best {
		out = append(out, m)
	}
	// Map iteration is random; the caller's output must not be.
	sort.Strings(out)
	return out
}

// capAndSort bounds the list and orders it for display.
func capAndSort(in []agentwire.DiskUsage) []agentwire.DiskUsage {
	out := in
	if len(out) > maxReportedFilesystems {
		// Fullest first, so a cap keeps what somebody will be paged about, and
		// said out loud rather than silently: a truncated list read as complete
		// is worse than an obviously partial one.
		sort.Slice(out, func(i, j int) bool { return out[i].UsedPct > out[j].UsedPct })
		slog.Warn("agent: reporting only the fullest filesystems",
			"found", len(out), "reported", maxReportedFilesystems)
		out = out[:maxReportedFilesystems]
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Mount < out[j].Mount })
	return out
}

// shorterMount reports whether a is the better representative of a filesystem:
// the shallower path, and on a tie the lexicographically smaller one so the
// choice never depends on mount order.
func shorterMount(a, b string) bool {
	if len(a) != len(b) {
		return len(a) < len(b)
	}
	return a < b
}

// isSystemMount reports whether a mountpoint is under a kernel/pseudo tree we
// never report, even if its fstype slips past the denylist.
func isSystemMount(mp string) bool {
	for _, prefix := range []string{"/proc", "/sys", "/dev", "/run"} {
		if mp == prefix || strings.HasPrefix(mp, prefix+"/") {
			return true
		}
	}
	return false
}

func osLabel(info *host.InfoStat) string {
	if info.Platform == "" {
		return info.OS
	}
	if info.PlatformVersion == "" {
		return info.Platform
	}
	return fmt.Sprintf("%s %s", info.Platform, info.PlatformVersion)
}
