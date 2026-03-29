package icmp

import (
"testing"
)

func TestDetectReturnsResult(t *testing.T) {
	result := Detect()
	// Detect must always return a result struct — it never panics.
	// Available can be true or false depending on the host's capabilities,
// so we only verify the contract: that the call completes and that
// Reason is set when unavailable.
if !result.Available && result.Reason == "" {
t.Error("expected Reason to be non-empty when capability is unavailable")
}
}

func TestDetectResultIsConsistent(t *testing.T) {
r1 := Detect()
r2 := Detect()
if r1.Available != r2.Available {
t.Errorf("Detect() returned inconsistent results: %v vs %v", r1.Available, r2.Available)
}
}
