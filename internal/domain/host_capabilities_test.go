package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHostCapabilities_OOMDetail(t *testing.T) {
	on, off := Capability{Available: true}, Capability{Available: false, Reason: CapabilityReasonUnreadable}
	cases := map[string]struct {
		kmsg, cgroup Capability
		want         OOMDetail
	}{
		"kmsg and cgroup: with process name": {on, on, OOMDetailWithProcess},
		"kmsg only: with process name":       {on, off, OOMDetailWithProcess},
		"cgroup only: without process name":  {off, on, OOMDetailWithoutProcess},
		"neither: not detected":              {off, off, OOMDetailNone},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := HostCapabilities{Kmsg: tc.kmsg, CgroupOOM: tc.cgroup}
			assert.Equal(t, tc.want, c.OOMDetail())
		})
	}
}

func TestHost_CapabilitiesState(t *testing.T) {
	now := time.Now()
	cases := map[string]struct {
		host *Host
		want HostCapabilitiesState
	}{
		"nil host":        {nil, HostCapabilitiesNotKnown},
		"never connected": {&Host{}, HostCapabilitiesNotKnown},
		"connected, no declaration (older agent)": {&Host{LastSeenAt: &now}, HostCapabilitiesNotReported},
		"declared":                             {&Host{LastSeenAt: &now, Capabilities: &HostCapabilities{}}, HostCapabilitiesDeclared},
		"declared wins even without last seen": {&Host{Capabilities: &HostCapabilities{}}, HostCapabilitiesDeclared},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.host.CapabilitiesState())
		})
	}
}

func TestIncident_FrozenCapabilitiesState(t *testing.T) {
	var nilInc *Incident
	assert.Equal(t, HostCapabilitiesNotKnown, nilInc.FrozenCapabilitiesState())
	assert.Equal(t, HostCapabilitiesNotKnown, (&Incident{}).FrozenCapabilitiesState(), "pre-feature row")

	for _, st := range []HostCapabilitiesState{HostCapabilitiesDeclared, HostCapabilitiesNotReported, HostCapabilitiesNotKnown, HostCapabilitiesNoMachine} {
		s := st
		assert.Equal(t, st, (&Incident{HostCapabilitiesState: &s}).FrozenCapabilitiesState())
	}
}
