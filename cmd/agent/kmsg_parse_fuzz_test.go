package main

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// FuzzParseKmsgLine (spec 090, SC-008).
//
// A table test proves the formats we thought of. This proves the one we did not.
// The input is /dev/kmsg: every subsystem writes to it, formats change between
// kernel versions, lines can be cut mid-write, and nothing guarantees valid UTF-8.
// This parser runs on machines the operator cares about, so its floor is not
// "parses correctly" but "cannot take the agent down".
//
//	go test -fuzz=FuzzParseKmsgLine -fuzztime=60s ./cmd/agent/
func FuzzParseKmsgLine(f *testing.F) {
	seeds := []string{
		// real shapes
		"6,2461,1234567890,-;Killed process 4711 (postgres) total-vm:2097152kB",
		"6,2460,1234567889,-;oom-kill:constraint=CONSTRAINT_MEMCG,task=node,pid=1234,uid=0",
		"6,999,1234567891,-;myapp[1234]: segfault at 0 ip 00007f error 4 in libc.so.6",
		// other subsystems
		"6,100,1,-;eth0: link up, 1000Mbps full-duplex",
		"4,101,2,-;usb 1-1: new high-speed USB device number 4",
		// degenerate and truncated
		"",
		";",
		",",
		"6,2461,1234567890,-",
		"6,2461,1234567890,-;Killed process 47",
		"Killed process (",
		"Killed process 1 ()",
		"[]: segfault at ",
		"oom-kill:task=",
		// hostile
		strings.Repeat("A", 4096),
		"6,1,1,-;Killed process 99999999999999999999999 (x)",
		"\x00\x01\x02;segfault at ",
		"\xff\xfe;Killed process 1 (\xff)",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, line string) {
		// The contract under fuzz is narrow and absolute: return, never panic.
		r, ok := parseKmsgLine(line)
		if !ok {
			return
		}

		// When it does claim a line, the claim must be structurally sound —
		// otherwise a malformed line becomes a malformed event downstream.
		if r.Kind == "" {
			t.Fatalf("classified with no kind: %q -> %+v", line, r)
		}
		if r.Process == "" {
			t.Fatalf("classified with no process: %q -> %+v", line, r)
		}
		if !utf8.ValidString(r.Process) {
			// Invalid UTF-8 would break JSON encoding on the wire, turning a
			// kernel log oddity into a dropped frame.
			t.Fatalf("process name is not valid UTF-8: %q -> %q", line, r.Process)
		}
		if strings.ContainsAny(r.Process, ";\n") {
			t.Fatalf("process name carries raw line structure: %q -> %q", line, r.Process)
		}
	})
}
