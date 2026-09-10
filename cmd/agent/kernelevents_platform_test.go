package main

import (
	"testing"
	"time"
)

// T037 — the platform seam, tested in a file that compiles EVERYWHERE.
//
// This is the one guard against a coverage trap rather than an addition of
// caution. A test for the inert path placed in kernelevents_other.go would only
// compile off Linux, and CI builds on Linux — it would sit in the repository
// looking like coverage while never running. So the assertions here are the ones
// that hold on any platform, and they run on all of them.
//
// What they establish: whatever this host allows, asking for sources is safe, the
// answer is well-formed, and a collector built from it is harmless.
func TestKernelSources_SeamIsSafeOnAnyPlatform(t *testing.T) {
	// On Linux this may open /dev/kmsg or find it unreadable; on anything else it
	// returns nothing. Neither may panic, and neither may report an error — a host
	// that cannot capture is the common case, not a failure.
	sources := newKernelSources()

	for i, s := range sources {
		if s == nil {
			t.Fatalf("source %d is nil: the slice must hold usable sources or be empty", i)
		}
		if s.Name() == "" {
			t.Errorf("source %d has no name; an operator investigating missing events needs to know which reader was working", i)
		}
		// Draining an idle source must be safe and must not block the interval.
		done := make(chan struct{})
		go func() { _ = s.Drain(); close(done) }()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatalf("source %q blocked on Drain; capture must never hold up the collection loop", s.Name())
		}
	}

	c := newEventCollector(sources)
	defer c.Close()

	// Whatever this platform returned, collecting is harmless and repeatable.
	for i := 0; i < 10; i++ {
		_ = c.Collect(time.Now())
	}
}

// T037 — the inert answer and the unreadable answer are the same answer.
//
// A macOS host has no sources because the interfaces do not exist; a container on
// Linux usually has none because it cannot read them. The collector must not be
// able to tell, and must behave identically either way — which is what makes the
// non-Linux path safe without a test that can only run there.
func TestKernelSources_NoSourcesBehavesIdenticallyWhateverTheReason(t *testing.T) {
	c := newEventCollector(nil)
	defer c.Close()

	for i := 0; i < 50; i++ {
		if got := c.Collect(time.Now()); got != nil {
			t.Fatalf("collect returned %v with no sources", got)
		}
	}
	if c.Dropped() != 0 {
		t.Errorf("dropped = %d; nothing was captured, so nothing was dropped", c.Dropped())
	}
}
