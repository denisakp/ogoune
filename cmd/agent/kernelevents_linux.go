package main

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/sys/unix"

	"github.com/denisakp/ogoune/pkg/agentwire"
)

// Linux kernel event sources (spec 090). Plain file reads: no eBPF, no kernel
// module, no capability beyond reading what the operating system already
// publishes.
//
// Neither source can fail in a way the caller sees. A reader that cannot read
// reports nothing, which the collector cannot distinguish from a quiet host — and
// does not need to.

const (
	kmsgPath          = "/dev/kmsg"
	cgroupEventsPath  = "/sys/fs/cgroup/memory.events"
	// One read(2) returns one record, and a buffer smaller than the record loses
	// it. The kernel caps a record well below this.
	kmsgReadBufferMax = 8192
	// kmsgDrainMax bounds how many reports one drain will return. A storm is
	// absorbed by the accumulator's aggregation, but the reader still must not
	// hand it an unbounded slice.
	kmsgDrainMax = 512
)

// newKernelSources opens whatever this host allows. Returns an empty slice rather
// than an error when nothing is readable: that is the common case, not a failure.
func newKernelSources() []kernelSource {
	var out []kernelSource
	if s := newKmsgSource(); s != nil {
		out = append(out, s)
	}
	if s := newCgroupOOMSource(); s != nil {
		out = append(out, s)
	}
	return out
}

// --- /dev/kmsg -------------------------------------------------------------

// kmsgSource reads the kernel log through a RAW non-blocking file descriptor.
//
// Both properties are load-bearing, and getting either wrong stops the agent
// streaming metrics -- which is worse than never capturing an event at all.
//
//   - NON-BLOCKING. A plain open(2) of /dev/kmsg blocks on read until the kernel
//     emits something. Draining it on the metrics path then parks the collector
//     forever on a quiet machine. That is not hypothetical: it is what shipped,
//     and it was invisible in CI and in containers because the open fails there,
//     leaving no source to block on. With O_NONBLOCK an empty log returns EAGAIN.
//   - RAW fd, not an *os.File. os.NewFile registers a pollable descriptor with
//     the Go runtime poller, which turns a would-be EAGAIN back into a blocking
//     wait. Reading through unix.Read keeps the syscall's own semantics.
//
// Reads are also record-oriented rather than line-buffered: each read(2) on
// /dev/kmsg returns exactly one record, and a buffer too small to hold it loses
// that record entirely. bufio would happily split one across reads.
type kmsgSource struct {
	mu  sync.Mutex
	fd  int
	buf []byte
}

// newKmsgSource opens the kernel log positioned at its END.
//
// This is the single most important line in the feature after "must not harm the
// agent". /dev/kmsg is a ring buffer holding everything since boot: opened at the
// start, it would yield a wave of kills that happened days ago, timestamped as if
// they had just occurred. An agent restarted by a package upgrade would then
// manufacture a fresh-looking history of failures that were resolved long ago,
// and anything correlating incidents against them would be confidently wrong.
//
// Events that occur while the agent is down are lost, permanently and on purpose.
// The alternative — persisting a read cursor — trades a fabrication risk for
// state management on the operator's machine, and is still wrong after a reboot
// clears the buffer.
func newKmsgSource() kernelSource {
	fd, err := unix.Open(kmsgPath, unix.O_RDONLY|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)
	if err != nil {
		// The common case in a container, which is the documented default
		// deployment. Debug rather than warn: the collector reports unavailability
		// once, at a level the operator will see, and this is the detail behind it.
		slog.Debug("agent: kernel log unavailable", "path", kmsgPath, "error", err)
		return nil
	}
	if _, err := unix.Seek(fd, 0, io.SeekEnd); err != nil {
		slog.Debug("agent: cannot seek kernel log to end", "error", err)
		_ = unix.Close(fd)
		return nil
	}
	return &kmsgSource{fd: fd, buf: make([]byte, kmsgReadBufferMax)}
}

func (s *kmsgSource) Name() string { return agentwire.SourceKmsg }

func (s *kmsgSource) Drain() []kmsgReport {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.fd < 0 {
		return nil
	}

	var out []kmsgReport
	// Bounded by reads attempted, not by reports produced: a burst of records the
	// classifier ignores must not keep this loop running either.
	for attempts := 0; len(out) < kmsgDrainMax && attempts < kmsgDrainMax; attempts++ {
		n, err := unix.Read(s.fd, s.buf)
		switch {
		case errors.Is(err, unix.EINTR):
			continue
		case errors.Is(err, unix.EPIPE):
			// Records were overwritten while we were away. The kernel has already
			// moved the read position to the oldest surviving record, so the right
			// answer is to keep going, not to give up on the interval.
			continue
		case err != nil:
			// EAGAIN on a quiet machine, which is the normal exit from this loop.
			// Anything else means the log became unreadable; either way the answer
			// is "nothing more this interval" and metrics are unaffected.
			return out
		case n <= 0:
			return out
		}
		if r, ok := parseKmsgLine(strings.TrimRight(string(s.buf[:n]), "\n")); ok {
			out = append(out, r)
		}
	}
	return out
}

func (s *kmsgSource) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.fd >= 0 {
		_ = unix.Close(s.fd)
		s.fd = -1
	}
}

// --- cgroup v2 memory.events ----------------------------------------------

// cgroupOOMSource watches the cgroup v2 out-of-memory counter.
//
// It exists because an agent running inside a container frequently cannot read
// /dev/kmsg — and a container is the deployment the documentation presents first.
// This keeps the most valuable signal working there. It reports counts rather than
// process names, so its events are less detailed than the kernel log's; less
// detail beats no signal.
type cgroupOOMSource struct {
	mu   sync.Mutex
	path string
	last int64
}

// newCgroupOOMSource reads the counter once at startup to establish a baseline.
//
// Without that baseline the first drain would report every kill since boot as if
// it had just happened — the same fabrication the kernel-log reader avoids by
// seeking to the end, arriving by a different route.
func newCgroupOOMSource() kernelSource {
	v, err := readOOMKillCounter(cgroupEventsPath)
	if err != nil {
		slog.Debug("agent: cgroup memory events unavailable", "path", cgroupEventsPath, "error", err)
		return nil
	}
	return &cgroupOOMSource{path: cgroupEventsPath, last: v}
}

func (s *cgroupOOMSource) Name() string { return agentwire.SourceCgroup }

func (s *cgroupOOMSource) Drain() []kmsgReport {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, err := readOOMKillCounter(s.path)
	if err != nil {
		return nil
	}
	delta := v - s.last
	s.last = v
	if delta <= 0 {
		// A counter that went backwards means the cgroup was recreated. Rebaseline
		// silently rather than reporting a negative burst.
		return nil
	}
	if delta > kmsgDrainMax {
		delta = kmsgDrainMax
	}

	// The counter says how many, not which: these reports carry no process name,
	// and the accumulator folds them into a count.
	out := make([]kmsgReport, 0, delta)
	for i := int64(0); i < delta; i++ {
		out = append(out, kmsgReport{Kind: agentwire.KindOOMKill})
	}
	return out
}

func (s *cgroupOOMSource) Close() {}

// readOOMKillCounter parses the oom_kill line of a cgroup v2 memory.events file.
func readOOMKillCounter(path string) (int64, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	for _, line := range strings.Split(string(b), "\n") {
		key, val, found := strings.Cut(strings.TrimSpace(line), " ")
		if !found || key != "oom_kill" {
			continue
		}
		return strconv.ParseInt(val, 10, 64)
	}
	return 0, os.ErrNotExist
}
