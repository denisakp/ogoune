package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/denisakp/ogoune/pkg/agentwire"
)

// The probe against a fake filesystem. The kernel-log probe is exercised with a
// regular file standing in for /dev/kmsg: open(2) is what is being asked about,
// and a regular file answers it the same way.
func redirectProbePaths(t *testing.T, kmsg, cgroup, sysctl string) {
	t.Helper()
	pk, pc, ps := probeKmsgPath, probeCgroupEventsPath, probeExceptionTracePath
	probeKmsgPath, probeCgroupEventsPath, probeExceptionTracePath = kmsg, cgroup, sysctl
	t.Cleanup(func() { probeKmsgPath, probeCgroupEventsPath, probeExceptionTracePath = pk, pc, ps })
}

func writeFile(t *testing.T, dir, name, content string, mode os.FileMode) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), mode); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLinuxProbe_Table(t *testing.T) {
	dir := t.TempDir()
	readable := writeFile(t, dir, "kmsg", "", 0o600)
	missing := filepath.Join(dir, "absent")
	cgroupOK := writeFile(t, dir, "memory.events", "low 0\nhigh 0\nmax 0\noom 0\noom_kill 3\n", 0o600)
	cgroupNoLine := writeFile(t, dir, "memory.events.v1", "cache 0\n", 0o600)
	on := writeFile(t, dir, "trace-on", "1\n", 0o600)
	offv := writeFile(t, dir, "trace-off", "0\n", 0o600)

	cases := map[string]struct {
		kmsg, cgroup, sysctl string
		want                 agentwire.Capabilities
	}{
		"native, setting on: everything": {readable, cgroupOK, on, agentwire.Capabilities{
			Kmsg: agentwire.Capability{Available: true}, CgroupOOM: agentwire.Capability{Available: true}, Segfault: agentwire.Capability{Available: true}}},
		"native, setting off: segfault reason is the setting": {readable, cgroupOK, offv, agentwire.Capabilities{
			Kmsg: agentwire.Capability{Available: true}, CgroupOOM: agentwire.Capability{Available: true},
			Segfault: agentwire.Capability{Available: false, Reason: agentwire.ReasonSettingOff}}},
		"container: kmsg missing, setting on -- segfault reason is the LOG, not the setting": {missing, cgroupOK, on, agentwire.Capabilities{
			Kmsg:      agentwire.Capability{Available: false, Reason: agentwire.ReasonUnreadable},
			CgroupOOM: agentwire.Capability{Available: true},
			Segfault:  agentwire.Capability{Available: false, Reason: agentwire.ReasonUnreadable}}},
		"cgroup v1 host: memory.events without oom_kill": {readable, cgroupNoLine, on, agentwire.Capabilities{
			Kmsg: agentwire.Capability{Available: true}, CgroupOOM: agentwire.Capability{Available: false, Reason: agentwire.ReasonUnreadable},
			Segfault: agentwire.Capability{Available: true}}},
		"nothing readable": {missing, missing, missing, agentwire.Capabilities{
			Kmsg:      agentwire.Capability{Available: false, Reason: agentwire.ReasonUnreadable},
			CgroupOOM: agentwire.Capability{Available: false, Reason: agentwire.ReasonUnreadable},
			Segfault:  agentwire.Capability{Available: false, Reason: agentwire.ReasonUnreadable}}},
		"sysctl file unreadable with kmsg readable: unreadable, not setting_off": {readable, cgroupOK, missing, agentwire.Capabilities{
			Kmsg: agentwire.Capability{Available: true}, CgroupOOM: agentwire.Capability{Available: true},
			Segfault: agentwire.Capability{Available: false, Reason: agentwire.ReasonUnreadable}}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			redirectProbePaths(t, tc.kmsg, tc.cgroup, tc.sysctl)
			got := newCapabilityProbe().Probe()
			assert.Equal(t, tc.want, got)
			assert.False(t, (&got).Normalize(), "the agent never sends something the backend has to fix")
		})
	}
}

func TestLinuxProbe_PermissionDenied(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root reads everything")
	}
	dir := t.TempDir()
	denied := writeFile(t, dir, "kmsg", "", 0o000)
	cgroupOK := writeFile(t, dir, "memory.events", "oom_kill 0\n", 0o600)
	on := writeFile(t, dir, "trace-on", "1\n", 0o600)
	redirectProbePaths(t, denied, cgroupOK, on)

	got := newCapabilityProbe().Probe()
	assert.Equal(t, agentwire.Capability{Available: false, Reason: agentwire.ReasonUnreadable}, got.Kmsg)
	assert.Equal(t, agentwire.Capability{Available: false, Reason: agentwire.ReasonUnreadable}, got.Segfault)
}
