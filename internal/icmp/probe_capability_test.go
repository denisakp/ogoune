package icmp

import (
	"testing"
	"time"
)

// TestProbeMissingCapabilityReturnsError verifies that when ICMP capability is
// unavailable, Probe() returns a non-panicking, well-formed error result.
func TestProbeMissingCapabilityIsHandled(t *testing.T) {
	// We cannot force capability absence in unit tests, but we can verify that
	// Probe always returns a well-formed result regardless of capability state.
	result := Probe("127.0.0.1", 100*time.Millisecond)
	// It should either succeed (if capability available) or fail gracefully.
	if result.Reachable && result.RTTMs < 0 {
		t.Error("reachable result must have non-negative RTT")
	}
	if !result.Reachable && result.RTTMs != 0 {
		t.Error("unreachable result must have zero RTT")
	}
}

// TestProbeIPv6LocalhostWhenCapabilityAvailable validates IPv6 behavior.
func TestProbeIPv6LocalhostWhenCapabilityAvailable(t *testing.T) {
	cap := Detect()
	if !cap.Available {
		t.Skip("ICMP capability not available on this host — skipping IPv6 test")
	}
	// ::1 is IPv6 loopback. Probe may not have IPv6 socket support and should
	// return a graceful error rather than panic.
	result := Probe("::1", 2*time.Second)
	// Accept either reachable or graceful failure — just no panic.
	if !result.Reachable && result.Error == "" {
		t.Error("unreachable IPv6 probe must include an error message")
	}
}
