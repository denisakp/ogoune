package icmp

import (
	"testing"
	"time"
)

func TestProbeInvalidHost(t *testing.T) {
	result := Probe("not-a-valid-host-xyzzy-12345.invalid", 2*time.Second)
	if result.Reachable {
		t.Error("expected unreachable for non-existent host")
	}
	if result.Error == "" {
		t.Error("expected error message for non-existent host")
	}
}

func TestProbeEmptyHost(t *testing.T) {
	result := Probe("", 2*time.Second)
	if result.Reachable {
		t.Error("expected unreachable for empty host")
	}
	if result.Error == "" {
		t.Error("expected error message for empty host")
	}
}

func TestProbeResultShape(t *testing.T) {
	result := Probe("256.256.256.256", 1*time.Second)
	if result.RTTMs < 0 {
		t.Errorf("expected RTTMs >= 0, got %d", result.RTTMs)
	}
}

func TestProbeTimeoutIsRespected(t *testing.T) {
	start := time.Now()
	timeout := 500 * time.Millisecond
	result := Probe("192.0.2.1", timeout)
	elapsed := time.Since(start)

	if result.Reachable {
		t.Skip("192.0.2.1 unexpectedly reachable")
	}

	maxAllowed := timeout * 5
	if elapsed > maxAllowed {
		t.Errorf("probe took %v but timeout was %v (max allowed %v)", elapsed, timeout, maxAllowed)
	}
}

func TestProbeLocalhostWhenCapabilityAvailable(t *testing.T) {
	cap := Detect()
	if !cap.Available {
		t.Skip("ICMP capability not available on this host")
	}

	result := Probe("127.0.0.1", 2*time.Second)
	if !result.Reachable {
		t.Errorf("expected localhost reachable, got error: %s", result.Error)
	}
	if result.RTTMs < 0 {
		t.Errorf("expected non-negative RTT, got %d", result.RTTMs)
	}
}
