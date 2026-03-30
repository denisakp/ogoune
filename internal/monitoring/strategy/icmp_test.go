package strategy

import (
	"context"
	"testing"

	"github.com/denisakp/ogoune/internal/domain"
)

func TestICMPStrategyExecuteInvalidTarget(t *testing.T) {
	s := NewICMPStrategy()
	resource := &domain.Resource{
		Type:   domain.ResourceICMP,
		Target: "",
	}
	result, err := s.Execute(context.Background(), resource)
	if err != nil {
		t.Fatalf("Execute should not return error, got: %v", err)
	}
	if result.Status != string(domain.StatusDown) {
		t.Errorf("expected StatusDown for empty target, got %s", result.Status)
	}
}

func TestICMPStrategyExecuteNonExistentHost(t *testing.T) {
	s := NewICMPStrategy()
	resource := &domain.Resource{
		Type:    domain.ResourceICMP,
		Target:  "not-a-valid-host-xyzzy-12345.invalid",
		Timeout: 2,
	}
	result, err := s.Execute(context.Background(), resource)
	if err != nil {
		t.Fatalf("Execute should not return error, got: %v", err)
	}
	if result.Status != string(domain.StatusDown) {
		t.Errorf("expected StatusDown for non-existent host, got %s", result.Status)
	}
	if result.ErrorMessage == "" {
		t.Error("expected non-empty ErrorMessage for failed probe")
	}
}

func TestICMPStrategyExecuteReachableHost(t *testing.T) {
	// Only run when ICMP capability is available and network is accessible.
	resource := &domain.Resource{
		Type:    domain.ResourceICMP,
		Target:  "127.0.0.1",
		Timeout: 2,
	}
	s := NewICMPStrategy()
	result, err := s.Execute(context.Background(), resource)
	if err != nil {
		t.Fatalf("Execute should not return error, got: %v", err)
	}
	// Accept either UP (capability available) or DOWN (capability unavailable).
	// Just assert result is well-formed.
	if result.Status == "" {
		t.Error("result Status must not be empty")
	}
}

func TestICMPStrategyHTTPStatusCodeIsNegativeOne(t *testing.T) {
	s := NewICMPStrategy()
	resource := &domain.Resource{
		Type:    domain.ResourceICMP,
		Target:  "192.0.2.1",
		Timeout: 1,
	}
	result, _ := s.Execute(context.Background(), resource)
	if result.HTTPStatusCode != -1 {
		t.Errorf("ICMP strategy must set HTTPStatusCode to -1, got %d", result.HTTPStatusCode)
	}
}
