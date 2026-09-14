package domain

import "testing"

func ptr(s string) *string { return &s }

// The whole rule in one table: two fields, three states, and the one
// combination that must never fall back.
func TestIncident_ResolveHost(t *testing.T) {
	cases := []struct {
		name       string
		recorded   bool
		hostID     *string
		monitorNow *string // Resource.HostID -- what the monitor points at today
		wantID     string
		wantSource HostLinkSource
	}{
		{
			name: "recorded machine, monitor unchanged",
			recorded: true, hostID: ptr("host-A"), monitorNow: ptr("host-A"),
			wantID: "host-A", wantSource: HostLinkSourceRecorded,
		},
		{
			// The defect this feature removes: the monitor moved, the incident
			// must not follow it.
			name: "recorded machine, monitor since moved",
			recorded: true, hostID: ptr("host-A"), monitorNow: ptr("host-B"),
			wantID: "host-A", wantSource: HostLinkSourceRecorded,
		},
		{
			name: "recorded machine, monitor since detached",
			recorded: true, hostID: ptr("host-A"), monitorNow: nil,
			wantID: "host-A", wantSource: HostLinkSourceRecorded,
		},
		{
			// The case that matters most (spec Q2): we looked, there was none.
			// A machine attached LATER must not be picked up.
			name: "recorded absence, monitor has a machine today",
			recorded: true, hostID: nil, monitorNow: ptr("host-B"),
			wantID: "", wantSource: HostLinkSourceNone,
		},
		{
			name: "recorded absence, monitor still has none",
			recorded: true, hostID: nil, monitorNow: nil,
			wantID: "", wantSource: HostLinkSourceNone,
		},
		{
			name: "predates recording, monitor has a machine",
			recorded: false, hostID: nil, monitorNow: ptr("host-B"),
			wantID: "host-B", wantSource: HostLinkSourceInferred,
		},
		{
			name: "predates recording, monitor has none",
			recorded: false, hostID: nil, monitorNow: nil,
			wantID: "", wantSource: HostLinkSourceNone,
		},
		{
			// An empty string is not a machine.
			name: "recorded empty string is treated as absence",
			recorded: true, hostID: ptr(""), monitorNow: ptr("host-B"),
			wantID: "", wantSource: HostLinkSourceNone,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			inc := &Incident{HostID: tc.hostID, HostLinkRecorded: tc.recorded}
			inc.Resource.HostID = tc.monitorNow

			gotID, gotSource := inc.ResolveHost()
			if gotSource != tc.wantSource {
				t.Errorf("source = %q, want %q", gotSource, tc.wantSource)
			}
			if gotID != tc.wantID {
				t.Errorf("host = %q, want %q", gotID, tc.wantID)
			}
		})
	}
}

func TestIncident_ResolveHost_NilReceiver(t *testing.T) {
	var inc *Incident
	id, src := inc.ResolveHost()
	if id != "" || src != HostLinkSourceNone {
		t.Errorf("nil incident resolved to (%q, %q), want none", id, src)
	}
}
