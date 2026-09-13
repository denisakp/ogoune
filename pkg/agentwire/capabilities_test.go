package agentwire

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const baseFrame = `"schema_version":2,"cpu_pct":1,"mem_pct":2,"net_in":3,"net_out":4,"disks":[]`

func TestDecode_CapabilitiesAbsent_IsNil(t *testing.T) {
	f, err := Decode([]byte(`{` + baseFrame + `}`))
	require.NoError(t, err)
	assert.Nil(t, f.Capabilities, "an agent predating the feature sends none; nil means not reported")
}

func TestDecode_CapabilitiesPresent(t *testing.T) {
	f, err := Decode([]byte(`{` + baseFrame + `,"capabilities":{"kmsg":{"available":false,"reason":"unreadable"},"cgroup_oom":{"available":true},"segfault":{"available":false,"reason":"unreadable"}}}`))
	require.NoError(t, err)
	require.NotNil(t, f.Capabilities)
	assert.Equal(t, Capability{Available: false, Reason: ReasonUnreadable}, f.Capabilities.Kmsg)
	assert.Equal(t, Capability{Available: true}, f.Capabilities.CgroupOOM)
	assert.Equal(t, Capability{Available: false, Reason: ReasonUnreadable}, f.Capabilities.Segfault)
}

func TestNormalize(t *testing.T) {
	cases := map[string]struct {
		in      Capabilities
		want    Capabilities
		changed bool
	}{
		"clean declaration untouched": {
			in:   Capabilities{Kmsg: Capability{Available: true}, CgroupOOM: Capability{Available: true}, Segfault: Capability{Available: false, Reason: ReasonSettingOff}},
			want: Capabilities{Kmsg: Capability{Available: true}, CgroupOOM: Capability{Available: true}, Segfault: Capability{Available: false, Reason: ReasonSettingOff}},
		},
		"unknown reason becomes unreadable": {
			in:      Capabilities{Kmsg: Capability{Available: false, Reason: "weird"}, CgroupOOM: Capability{Available: true}, Segfault: Capability{Available: false, Reason: ReasonUnreadable}},
			want:    Capabilities{Kmsg: Capability{Available: false, Reason: ReasonUnreadable}, CgroupOOM: Capability{Available: true}, Segfault: Capability{Available: false, Reason: ReasonUnreadable}},
			changed: true,
		},
		"empty reason on an unavailable capability becomes unreadable": {
			in:      Capabilities{CgroupOOM: Capability{Available: false}, Kmsg: Capability{Available: false, Reason: ReasonUnreadable}, Segfault: Capability{Available: false, Reason: ReasonUnreadable}},
			want:    Capabilities{CgroupOOM: Capability{Available: false, Reason: ReasonUnreadable}, Kmsg: Capability{Available: false, Reason: ReasonUnreadable}, Segfault: Capability{Available: false, Reason: ReasonUnreadable}},
			changed: true,
		},
		"available with a reason drops the reason": {
			in:      Capabilities{Kmsg: Capability{Available: true, Reason: "leftover"}, CgroupOOM: Capability{Available: true}, Segfault: Capability{Available: true}},
			want:    Capabilities{Kmsg: Capability{Available: true}, CgroupOOM: Capability{Available: true}, Segfault: Capability{Available: true}},
			changed: true,
		},
		"segfault cannot be available without the kernel log": {
			in:      Capabilities{Kmsg: Capability{Available: false, Reason: ReasonUnreadable}, CgroupOOM: Capability{Available: true}, Segfault: Capability{Available: true}},
			want:    Capabilities{Kmsg: Capability{Available: false, Reason: ReasonUnreadable}, CgroupOOM: Capability{Available: true}, Segfault: Capability{Available: false, Reason: ReasonUnreadable}},
			changed: true,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := tc.in
			changed := got.Normalize()
			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.changed, changed)
		})
	}
	var nilCaps *Capabilities
	assert.False(t, nilCaps.Normalize(), "nil receiver is a no-op")
}

func TestDecode_NormalisesOnTheWayIn(t *testing.T) {
	f, err := Decode([]byte(`{` + baseFrame + `,"capabilities":{"kmsg":{"available":false,"reason":"weird"},"cgroup_oom":{"available":true,"reason":"x"},"segfault":{"available":true}}}`))
	require.NoError(t, err, "a declaration is never a reason to reject a frame")
	assert.Equal(t, ReasonUnreadable, f.Capabilities.Kmsg.Reason)
	assert.Equal(t, "", f.Capabilities.CgroupOOM.Reason)
	assert.False(t, f.Capabilities.Segfault.Available)
}

// SC-006, the old-backend half: the schema version did not move, so a backend
// that predates the feature decodes a new agent's frame exactly as it decodes
// the same frame with the object stripped. Simulated by stripping the key and
// comparing every other decoded field.
func TestDecode_OldBackendIgnoresTheObject(t *testing.T) {
	withCaps := []byte(`{` + baseFrame + `,"events":[],"capabilities":{"kmsg":{"available":true},"cgroup_oom":{"available":true},"segfault":{"available":true}}}`)
	var raw map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(withCaps, &raw))
	delete(raw, "capabilities")
	stripped, err := json.Marshal(raw)
	require.NoError(t, err)

	a, err := Decode(withCaps)
	require.NoError(t, err)
	b, err := Decode(stripped)
	require.NoError(t, err)

	assert.Equal(t, SchemaVersion, a.SchemaVersion, "no version bump: the old backend's ceiling still admits the frame")
	a.Capabilities = nil
	assert.Equal(t, b, a, "everything but the declaration is identical")
}

func TestEncode_RoundTripsCapabilities(t *testing.T) {
	in := Frame{CPUPct: 1, MemPct: 2, NetIn: 3, NetOut: 4, Capabilities: &Capabilities{
		Kmsg: Capability{Available: true}, CgroupOOM: Capability{Available: true}, Segfault: Capability{Available: false, Reason: ReasonSettingOff},
	}}
	b, err := Encode(in)
	require.NoError(t, err)
	out, err := Decode(b)
	require.NoError(t, err)
	assert.Equal(t, in.Capabilities, out.Capabilities)
	assert.Contains(t, string(b), `"reason":"setting_off"`)
	assert.NotContains(t, string(b), `"reason":""`, "available capabilities carry no reason key")
}
