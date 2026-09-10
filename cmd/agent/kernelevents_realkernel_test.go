//go:build linux

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/denisakp/ogoune/pkg/agentwire"
)

// Kernel capture against a REAL kernel log.
//
// Everything else in this package tests the classifier against strings and the
// collector against a fake source. That is what let a blocking read on
// /dev/kmsg ship: the agent stopped streaming metrics on every host where the
// kernel log was readable, and no test could see it, because in a container the
// open fails and there is no source to block on.
//
// These tests need what only a real Linux kernel provides. They are skipped
// where it is unavailable — UNLESS OGOUNE_REQUIRE_KERNEL_CAPTURE is set, which
// CI does. A verification that silently skips is not a verification, and a green
// build must not be able to mean "we did not look".

const requireEnv = "OGOUNE_REQUIRE_KERNEL_CAPTURE"

// skipOrFail skips when the kernel log is out of reach, or fails if the caller
// declared it must be reachable.
func skipOrFail(t *testing.T, reason string) {
	t.Helper()
	if os.Getenv(requireEnv) != "" {
		t.Fatalf("%s is set, so this must run, but %s", requireEnv, reason)
	}
	t.Skipf("kernel capture unavailable: %s", reason)
}

// realKmsgSource opens the actual kernel log, or skips.
func realKmsgSource(t *testing.T) *kmsgSource {
	t.Helper()
	src := newKmsgSource()
	if src == nil {
		skipOrFail(t, "/dev/kmsg could not be opened (root, and not a container, is required)")
	}
	// Closed in the background, with a bound. A Drain blocked in a read holds the
	// source's mutex, so a direct Close would wait on it forever and the test
	// would die of the package timeout instead of the assertion that actually
	// failed. Leaking a descriptor in that case is fine -- the process is about
	// to end, and a legible failure is worth more.
	t.Cleanup(func() {
		closed := make(chan struct{})
		go func() { src.Close(); close(closed) }()
		select {
		case <-closed:
		case <-time.After(time.Second):
		}
	})
	return src.(*kmsgSource)
}

// injectKernelLine writes one record to the kernel log.
//
// Only the message body is written: the kernel prepends its own
// "priority,sequence,timestamp,flags;" prefix, which is exactly what the parser
// strips. Writing a full record would leave the fake prefix inside the message.
func injectKernelLine(t *testing.T, msg string) {
	t.Helper()
	f, err := os.OpenFile(kmsgPath, os.O_WRONLY, 0)
	if err != nil {
		skipOrFail(t, fmt.Sprintf("cannot write to %s: %v", kmsgPath, err))
	}
	defer f.Close()
	// The trailing newline is required: a write without one leaves the record
	// pending in the kernel's continuation buffer and it never becomes readable.
	if _, err := f.WriteString(msg + "\n"); err != nil {
		skipOrFail(t, fmt.Sprintf("write to %s failed: %v", kmsgPath, err))
	}
}

// marker keeps assertions specific: the kernel log is shared with everything
// else on the machine, and CI runners are busy.
func marker(t *testing.T) string {
	t.Helper()
	return "ogtest" + strings.ReplaceAll(fmt.Sprintf("%d", time.Now().UnixNano()%1e9), "-", "")
}

// drainWithin fails if Drain has not returned by the deadline.
//
// This is the bug that shipped, in one assertion: the read was blocking, so on a
// quiet machine this call never came back and the agent's collector — which
// calls it on the metrics path — never sent another frame.
func drainWithin(t *testing.T, src *kmsgSource, d time.Duration) []kmsgReport {
	t.Helper()
	done := make(chan []kmsgReport, 1)
	go func() { done <- src.Drain() }()
	select {
	case reports := <-done:
		return reports
	case <-time.After(d):
		t.Fatalf("Drain did not return within %s on a real kernel log; "+
			"a blocking read here stops the agent streaming metrics entirely", d)
		return nil
	}
}

func TestRealKernel_DrainReturnsOnAQuietKernel(t *testing.T) {
	src := realKmsgSource(t)
	// Twice: the first call may consume whatever was pending, the second is
	// certain to find nothing and is the one that used to block forever.
	drainWithin(t, src, 3*time.Second)
	drainWithin(t, src, 3*time.Second)
}

func TestRealKernel_ClassifiesAnInjectedOOMKill(t *testing.T) {
	src := realKmsgSource(t)
	drainWithin(t, src, 3*time.Second) // position past anything already pending

	name := marker(t)
	injectKernelLine(t, fmt.Sprintf("Killed process 4711 (%s) total-vm:83252kB, anon-rss:65536kB", name))

	report := waitForReport(t, src, func(r kmsgReport) bool {
		return r.Kind == agentwire.KindOOMKill && r.Process == name
	})
	if report.PID != 4711 {
		t.Errorf("pid = %d, want 4711", report.PID)
	}
}

// The cgroup form, which is what a container out-of-memory kill actually emits.
func TestRealKernel_ClassifiesTheCgroupOOMForm(t *testing.T) {
	src := realKmsgSource(t)
	drainWithin(t, src, 3*time.Second)

	name := marker(t)
	injectKernelLine(t, fmt.Sprintf(
		"oom-kill:constraint=CONSTRAINT_MEMCG,nodemask=(null),cpuset=child,task=%s,pid=1673237,uid=0", name))

	waitForReport(t, src, func(r kmsgReport) bool {
		return r.Kind == agentwire.KindOOMKill && r.Process == name
	})
}

// Both segfault dialects. arm64 does not emit the x86 line, which is why
// segfault capture was silently dead on every arm64 host.
func TestRealKernel_ClassifiesBothSegfaultForms(t *testing.T) {
	cases := map[string]func(string) string{
		"x86": func(n string) string {
			return fmt.Sprintf("%s[1234]: segfault at 0 ip 00007f errorno 4 in libc.so.6", n)
		},
		"arm64": func(n string) string {
			return fmt.Sprintf("%s: %s: potentially unexpected fatal signal 11.", n, n)
		},
	}
	for label, build := range cases {
		t.Run(label, func(t *testing.T) {
			src := realKmsgSource(t)
			drainWithin(t, src, 3*time.Second)

			name := marker(t)
			injectKernelLine(t, build(name))

			waitForReport(t, src, func(r kmsgReport) bool {
				return r.Kind == agentwire.KindSegfault && r.Process == name
			})
		})
	}
}

// Restarting the agent must not replay the log.
//
// /dev/kmsg holds everything since boot. A reader that starts at the beginning
// would report kills from days ago as if they had just happened, and anything
// correlating incidents against them would be confidently wrong. This is the
// property that makes a package upgrade safe, and only a real kernel log can
// demonstrate it.
func TestRealKernel_NewSourceSeesNothingFromBeforeItOpened(t *testing.T) {
	first := realKmsgSource(t)
	drainWithin(t, first, 3*time.Second)

	name := marker(t)
	injectKernelLine(t, fmt.Sprintf("Killed process 9999 (%s) total-vm:1kB", name))
	waitForReport(t, first, func(r kmsgReport) bool { return r.Process == name })

	// A restart: a brand-new reader, opened after that kill was logged.
	restarted := realKmsgSource(t)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		for _, r := range drainWithin(t, restarted, 3*time.Second) {
			if r.Process == name {
				t.Fatal("a restarted agent replayed a kill from before it started; " +
					"that fabricates history an operator would act on")
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// The collector must survive a real source without costing the metrics frame.
func TestRealKernel_CollectorReturnsPromptly(t *testing.T) {
	sources := newKernelSources()
	if len(sources) == 0 {
		skipOrFail(t, "no kernel event source is available")
	}
	for _, s := range sources {
		defer s.Close()
	}

	c := newEventCollector(sources)
	done := make(chan []agentwire.KernelEvent, 1)
	go func() { done <- c.Collect(time.Now().UTC()) }()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Collect did not return against a real kernel log; " +
			"it runs on the metrics path, so this is the agent going silent")
	}
}

// A genuine out-of-memory kill, produced by the kernel rather than injected.
// Skipped where systemd-run or cgroup memory limits are unavailable; the
// injected tests above cover the parsing either way.
func TestRealKernel_CapturesAGenuineOOMKill(t *testing.T) {
	if _, err := exec.LookPath("systemd-run"); err != nil {
		t.Skip("systemd-run unavailable; injected OOM lines are covered above")
	}
	src := realKmsgSource(t)
	drainWithin(t, src, 3*time.Second)

	script := t.TempDir() + "/eat.sh"
	if err := os.WriteFile(script, []byte("#!/bin/sh\nexec dd if=/dev/zero of=/dev/null bs=1M count=99999999 iflag=fullblock\n"), 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}
	// A hard memory ceiling with no swap: the kernel has to kill it.
	cmd := exec.Command("systemd-run", "--scope", "--quiet",
		"-p", "MemoryMax=32M", "-p", "MemorySwapMax=0",
		"/bin/sh", "-c", "a=$(head -c 64000000 /dev/zero | tr '\\0' 'x'); echo ${#a}")
	_ = cmd.Run() // the kill is the expected outcome, so the exit status says nothing

	waitForReport(t, src, func(r kmsgReport) bool { return r.Kind == agentwire.KindOOMKill })
}

// waitForReport drains until a matching report appears or the deadline passes.
func waitForReport(t *testing.T, src *kmsgSource, match func(kmsgReport) bool) kmsgReport {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		for _, r := range drainWithin(t, src, 3*time.Second) {
			if match(r) {
				return r
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatal("no matching kernel report reached the agent within the deadline")
	return kmsgReport{}
}
