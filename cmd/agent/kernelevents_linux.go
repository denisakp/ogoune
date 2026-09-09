package main

import (
	"bufio"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"

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

type kmsgSource struct {
	mu sync.Mutex
	f  *os.File
	r  *bufio.Reader
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
	f, err := os.Open(kmsgPath)
	if err != nil {
		// The common case in a container, which is the documented default
		// deployment. Debug rather than warn: the collector reports unavailability
		// once, at a level the operator will see, and this is the detail behind it.
		slog.Debug("agent: kernel log unavailable", "path", kmsgPath, "error", err)
		return nil
	}
	if _, err := f.Seek(0, io.SeekEnd); err != nil {
		slog.Debug("agent: cannot seek kernel log to end", "error", err)
		_ = f.Close()
		return nil
	}
	return &kmsgSource{f: f, r: bufio.NewReaderSize(f, kmsgReadBufferMax)}
}

func (s *kmsgSource) Name() string { return agentwire.SourceKmsg }

func (s *kmsgSource) Drain() []kmsgReport {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.f == nil {
		return nil
	}

	var out []kmsgReport
	for len(out) < kmsgDrainMax {
		// /dev/kmsg is opened non-blocking by the kernel for readers that seek, so
		// a read with nothing pending returns an error rather than waiting. Either
		// way the answer is "nothing more this interval".
		line, err := s.r.ReadString('\n')
		if err != nil {
			break
		}
		if r, ok := parseKmsgLine(strings.TrimRight(line, "\n")); ok {
			out = append(out, r)
		}
	}
	return out
}

func (s *kmsgSource) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.f != nil {
		_ = s.f.Close()
		s.f = nil
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
