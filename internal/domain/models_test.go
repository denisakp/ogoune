package domain

import (
	"testing"
)

func TestIncidentDiagnostics_WithICMP(t *testing.T) {
	tests := []struct {
		name          string
		diag          *IncidentDiagnostics
		icmpAvailable *bool
		icmpReachable *bool
		icmpRttMs     *int
		rootCause     string
		wantNil       bool
		checkFields   func(*testing.T, *IncidentDiagnostics)
	}{
		{
			name:    "nil diagnostics",
			diag:    nil,
			wantNil: true,
		},
		{
			name: "merges all ICMP fields",
			diag: &IncidentDiagnostics{
				IncidentID: "test-incident",
			},
			icmpAvailable: ptrBool(true),
			icmpReachable: ptrBool(true),
			icmpRttMs:     ptrInt(45),
			rootCause:     "service_down",
			checkFields: func(t *testing.T, d *IncidentDiagnostics) {
				if d.ICMPAvailable == nil || !*d.ICMPAvailable {
					t.Error("expected ICMPAvailable=true")
				}
				if d.ICMPReachable == nil || !*d.ICMPReachable {
					t.Error("expected ICMPReachable=true")
				}
				if d.ICMPRttMs == nil || *d.ICMPRttMs != 45 {
					t.Error("expected ICMPRttMs=45")
				}
				if d.RootCauseHint != "service_down" {
					t.Error("expected RootCauseHint=service_down")
				}
			},
		},
		{
			name: "host unreachable case (no RTT)",
			diag: &IncidentDiagnostics{
				IncidentID: "test-incident-2",
			},
			icmpAvailable: ptrBool(true),
			icmpReachable: ptrBool(false),
			icmpRttMs:     nil,
			rootCause:     "host_unreachable",
			checkFields: func(t *testing.T, d *IncidentDiagnostics) {
				if d.ICMPAvailable == nil || !*d.ICMPAvailable {
					t.Error("expected ICMPAvailable=true")
				}
				if d.ICMPReachable == nil || *d.ICMPReachable {
					t.Error("expected ICMPReachable=false")
				}
				if d.ICMPRttMs != nil {
					t.Error("expected ICMPRttMs=nil")
				}
				if d.RootCauseHint != "host_unreachable" {
					t.Error("expected RootCauseHint=host_unreachable")
				}
			},
		},
		{
			name: "capability unavailable case",
			diag: &IncidentDiagnostics{
				IncidentID: "test-incident-3",
			},
			icmpAvailable: ptrBool(false),
			icmpReachable: nil,
			icmpRttMs:     nil,
			rootCause:     "icmp_unavailable",
			checkFields: func(t *testing.T, d *IncidentDiagnostics) {
				if d.ICMPAvailable == nil || *d.ICMPAvailable {
					t.Error("expected ICMPAvailable=false")
				}
				if d.RootCauseHint != "icmp_unavailable" {
					t.Error("expected RootCauseHint=icmp_unavailable")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.diag.WithICMP(tt.icmpAvailable, tt.icmpReachable, tt.icmpRttMs, tt.rootCause)
			if tt.wantNil && result != nil {
				t.Fatal("expected nil result for nil diagnostics")
			}
			if !tt.wantNil && result == nil {
				t.Fatal("expected non-nil result")
			}
			if tt.checkFields != nil {
				tt.checkFields(t, result)
			}
		})
	}
}

// Helper functions for pointer values
func ptrBool(b bool) *bool {
	return &b
}

func ptrInt(i int) *int {
	return &i
}
