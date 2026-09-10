package main

import (
	"testing"

	"github.com/denisakp/ogoune/pkg/agentwire"
)

// T015 -- classification against lines from real kernels. Two OOM-kill shapes are
// in circulation and a monitoring agent does not get to pick its kernel version,
// so both are covered.
func TestParseKmsgLine_Classifies(t *testing.T) {
	cases := []struct {
		name    string
		line    string
		kind    string
		process string
		pid     int
	}{
		{
			name:    "oom kill, killed-process shape",
			line:    "6,2461,1234567890,-;Killed process 4711 (postgres) total-vm:2097152kB, anon-rss:1048576kB",
			kind:    agentwire.KindOOMKill,
			process: "postgres",
			pid:     4711,
		},
		{
			name:    "oom kill, oom-kill shape",
			line:    "6,2460,1234567889,-;oom-kill:constraint=CONSTRAINT_MEMCG,task=node,pid=1234,uid=0",
			kind:    agentwire.KindOOMKill,
			process: "node",
			pid:     1234,
		},
		{
			name:    "segfault",
			line:    "6,999,1234567891,-;myapp[1234]: segfault at 0 ip 00007f error 4 in libc.so.6",
			kind:    agentwire.KindSegfault,
			process: "myapp",
			pid:     1234,
		},
		{
			name:    "no metadata prefix at all",
			line:    "Killed process 55 (redis-server) total-vm:1kB",
			kind:    agentwire.KindOOMKill,
			process: "redis-server",
			pid:     55,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := parseKmsgLine(c.line)
			if !ok {
				t.Fatalf("line not classified: %q", c.line)
			}
			if got.Kind != c.kind || got.Process != c.process || got.PID != c.pid {
				t.Errorf("got %+v, want kind=%s process=%s pid=%d", got, c.kind, c.process, c.pid)
			}
		})
	}
}

// T015 -- the obligation that matters more than recognising ours: never claim a
// line is ours when it is not. The kernel log carries every subsystem's output.
func TestParseKmsgLine_IgnoresEverythingElse(t *testing.T) {
	notOurs := []string{
		"6,100,1,-;eth0: link up, 1000Mbps full-duplex",
		"4,101,2,-;usb 1-1: new high-speed USB device number 4",
		"6,102,3,-;audit: type=1400 apparmor=\"DENIED\"",
		"6,103,4,-;systemd[1]: Started Session 3 of user root.",
		"3,104,5,-;EXT4-fs (sda1): mounted filesystem with ordered data mode",
		// close to ours without being ours
		"6,105,6,-;process 4711 was not killed",
		"6,106,7,-;segfault handling is disabled",
		"6,107,8,-;Killed process without a name",
		"6,108,9,-;oom-kill:constraint=CONSTRAINT_NONE,pid=1",
		// degenerate
		"",
		"   ",
		";",
		"6,109,10,-;",
		"no semicolon at all but has, commas",
	}

	for _, line := range notOurs {
		if r, ok := parseKmsgLine(line); ok {
			t.Errorf("misclassified %q as %+v", line, r)
		}
	}
}

// A truncated line — the kernel log can be cut mid-write — is skipped rather than
// guessed at.
func TestParseKmsgLine_TruncatedLine(t *testing.T) {
	if _, ok := parseKmsgLine("6,2461,1234567890,-"); ok {
		t.Error("a line cut before its message must not be classified")
	}
	if _, ok := parseKmsgLine("6,2461,1234567890,-;Killed process 47"); ok {
		t.Error("a message cut mid-report must not be classified")
	}
}

// The arm64 fatal-signal form (spec 090 follow-up). Found by running the agent
// on an arm64 host: that kernel never emits the x86 "segfault at ..." line, so
// segfault capture was silently dead on every arm64 machine.
func TestParseKmsgLine_FatalSignalArm64(t *testing.T) {
	cases := []struct {
		name    string
		line    string
		wantOK  bool
		process string
	}{
		{
			name:    "real arm64 line, captured from /dev/kmsg",
			line:    "6,2893,129388087502,-;python3.14: python3: potentially unexpected fatal signal 11.",
			wantOK:  true,
			process: "python3",
		},
		{
			name:    "single name form",
			line:    "6,10,1,-;myapp: potentially unexpected fatal signal 11.",
			wantOK:  true,
			process: "myapp",
		},
		{
			// SIGBUS is real and fatal, but it is not a segmentation fault. Naming
			// it one would be a lie; it needs a kind of its own to be reported.
			name:   "a different fatal signal is not a segfault",
			line:   "6,11,1,-;myapp: myapp: potentially unexpected fatal signal 7.",
			wantOK: false,
		},
		{
			name:   "no signal number",
			line:   "6,12,1,-;myapp: potentially unexpected fatal signal .",
			wantOK: false,
		},
		{
			name:   "no process name",
			line:   "6,13,1,-;potentially unexpected fatal signal 11.",
			wantOK: false,
		},
		{
			name:   "process name that is not one",
			line:   "6,14,1,-;;;: potentially unexpected fatal signal 11.",
			wantOK: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := parseKmsgLine(tc.line)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v (report %+v)", ok, tc.wantOK, got)
			}
			if !tc.wantOK {
				return
			}
			if got.Kind != agentwire.KindSegfault {
				t.Errorf("kind = %q, want %q", got.Kind, agentwire.KindSegfault)
			}
			if got.Process != tc.process {
				t.Errorf("process = %q, want %q", got.Process, tc.process)
			}
			// This format carries no pid, and reporting a zero would read as one.
			if got.PID != 0 {
				t.Errorf("pid = %d, want 0 (the format has none)", got.PID)
			}
		})
	}
}
