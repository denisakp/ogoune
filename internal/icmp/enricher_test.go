package icmp

import (
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
)

func TestExtractHost(t *testing.T) {
	tests := []struct {
		name    string
		rt      domain.ResourceType
		target  string
		host    string
		wantErr bool
	}{
		{name: "http target", rt: domain.ResourceHTTP, target: "https://example.com/path", host: "example.com"},
		{name: "tcp target", rt: domain.ResourceTCP, target: "example.com:443", host: "example.com"},
		{name: "dns target", rt: domain.ResourceDNS, target: "example.com", host: "example.com"},
		{name: "icmp target", rt: domain.ResourceICMP, target: "192.0.2.10", host: "192.0.2.10"},
		{name: "bad http", rt: domain.ResourceHTTP, target: "not-a-url", wantErr: true},
		{name: "bad tcp", rt: domain.ResourceTCP, target: "example.com", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, err := ExtractHost(tt.rt, tt.target)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got none and host=%q", host)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if host != tt.host {
				t.Fatalf("expected host=%q got %q", tt.host, host)
			}
		})
	}
}

func TestEnrichHintMatrix_NonICMP(t *testing.T) {
	originalDetect := detectFunc
	originalProbe := probeFunc
	defer func() {
		detectFunc = originalDetect
		probeFunc = originalProbe
	}()

	t.Run("capability unavailable", func(t *testing.T) {
		detectFunc = func() CapabilityResult { return CapabilityResult{Available: false, Reason: "no raw socket"} }
		probeFunc = func(string, time.Duration) ProbeResult {
			t.Fatal("probe must not be called when capability is unavailable")
			return ProbeResult{}
		}

		result := Enrich(domain.ResourceHTTP, "https://example.com", domain.CheckResult{Status: string(domain.StatusDown)})
		if result.ICMPAvailable == nil || *result.ICMPAvailable {
			t.Fatalf("expected icmp_available=false")
		}
		if result.RootCauseHint != RootCauseICMPUnavailable {
			t.Fatalf("expected hint %q got %q", RootCauseICMPUnavailable, result.RootCauseHint)
		}
	})

	t.Run("host unreachable", func(t *testing.T) {
		detectFunc = func() CapabilityResult { return CapabilityResult{Available: true} }
		probeFunc = func(_ string, timeout time.Duration) ProbeResult {
			if timeout != EnrichmentTimeout {
				t.Fatalf("expected fixed timeout %v, got %v", EnrichmentTimeout, timeout)
			}
			return ProbeResult{Reachable: false, Error: "timeout"}
		}

		result := Enrich(domain.ResourceHTTP, "https://example.com", domain.CheckResult{Status: string(domain.StatusDown)})
		if result.ICMPAvailable == nil || !*result.ICMPAvailable {
			t.Fatalf("expected icmp_available=true")
		}
		if result.ICMPReachable == nil || *result.ICMPReachable {
			t.Fatalf("expected icmp_reachable=false")
		}
		if result.RootCauseHint != RootCauseHostUnreachable {
			t.Fatalf("expected hint %q got %q", RootCauseHostUnreachable, result.RootCauseHint)
		}
	})

	t.Run("service down", func(t *testing.T) {
		detectFunc = func() CapabilityResult { return CapabilityResult{Available: true} }
		probeFunc = func(_ string, timeout time.Duration) ProbeResult {
			if timeout != EnrichmentTimeout {
				t.Fatalf("expected fixed timeout %v, got %v", EnrichmentTimeout, timeout)
			}
			return ProbeResult{Reachable: true, RTTMs: 12}
		}

		result := Enrich(domain.ResourceTCP, "example.com:443", domain.CheckResult{Status: string(domain.StatusDown)})
		if result.ICMPReachable == nil || !*result.ICMPReachable {
			t.Fatalf("expected icmp_reachable=true")
		}
		if result.ICMPRTTMs == nil || *result.ICMPRTTMs != 12 {
			t.Fatalf("expected icmp_rtt_ms=12 got %v", result.ICMPRTTMs)
		}
		if result.RootCauseHint != RootCauseServiceDown {
			t.Fatalf("expected hint %q got %q", RootCauseServiceDown, result.RootCauseHint)
		}
	})
}

func TestEnrichICMP_NoSecondProbe(t *testing.T) {
	originalDetect := detectFunc
	originalProbe := probeFunc
	defer func() {
		detectFunc = originalDetect
		probeFunc = originalProbe
	}()

	probeCalled := false
	detectFunc = func() CapabilityResult {
		return CapabilityResult{Available: true}
	}
	probeFunc = func(_ string, _ time.Duration) ProbeResult {
		probeCalled = true
		return ProbeResult{Reachable: true}
	}

	result := Enrich(
		domain.ResourceICMP,
		"example.com",
		domain.CheckResult{
			Status:       string(domain.StatusDown),
			ErrorMessage: "request timeout",
			ResponseData: "request timeout",
		},
	)

	if probeCalled {
		t.Fatal("expected no second probe for ICMP monitor failures")
	}
	if result.RootCauseHint != RootCauseHostUnreachable {
		t.Fatalf("expected host_unreachable hint, got %q", result.RootCauseHint)
	}
}
