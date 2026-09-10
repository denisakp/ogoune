package main

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/denisakp/ogoune/pkg/agentwire"
)

// maxProcessNameLen bounds what this parser will accept as a process name. The
// kernel's own limit is 15 characters; the margin is for anything a future format
// might legitimately carry. Beyond it, the line is not a report we understand.
const maxProcessNameLen = 64

// Classification of kernel log lines (spec 090).
//
// The input here is not ours. Every subsystem writes to the kernel log, the
// message formats differ between kernel versions, and a line can be truncated
// mid-write. So this parser has exactly two obligations: never panic, and never
// claim a line is ours when it is not. It is fuzzed for the first and
// table-tested for the second.
//
// A recognised line yields a classified report and the raw text is dropped on the
// spot. There is nowhere in kmsgReport to put it, which is the enforcement rather
// than the discipline.

// kmsgReport is one classified kernel report. Not an event: events are the
// per-interval aggregate of these (FR-028).
type kmsgReport struct {
	Kind    string
	Process string
	PID     int
}

// parseKmsgLine classifies one line of /dev/kmsg. The second return is false for
// everything that is not an out-of-memory kill or a segmentation fault, which is
// the overwhelming majority of what the kernel log carries.
//
// The line format is "priority,sequence,timestamp,flags;message". Everything
// before the first semicolon is metadata this parser does not need; a line
// without one is either truncated or from a source that does not follow the
// convention, and is skipped either way.
func parseKmsgLine(line string) (kmsgReport, bool) {
	msg := line
	if i := strings.IndexByte(line, ';'); i >= 0 {
		msg = line[i+1:]
	} else if strings.ContainsRune(line, ',') {
		// Has metadata separators but no terminator: a truncated line. Skipping
		// beats guessing where the message would have started.
		return kmsgReport{}, false
	}
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return kmsgReport{}, false
	}

	if r, ok := parseOOMKill(msg); ok {
		return r, true
	}
	if r, ok := parseSegfault(msg); ok {
		return r, true
	}
	if r, ok := parseFatalSignal(msg); ok {
		return r, true
	}
	return kmsgReport{}, false
}

// parseOOMKill recognises the kernel's out-of-memory kill report.
//
// Two shapes are in circulation and both are handled, because a monitoring agent
// does not get to pick its kernel version:
//
//	Killed process 4711 (postgres) total-vm:...
//	oom-kill:constraint=...,task=postgres,pid=4711,...
func parseOOMKill(msg string) (kmsgReport, bool) {
	if strings.HasPrefix(msg, "oom-kill:") {
		r := kmsgReport{Kind: agentwire.KindOOMKill}
		for _, part := range strings.Split(msg, ",") {
			k, v, found := strings.Cut(part, "=")
			if !found {
				continue
			}
			switch strings.TrimSpace(k) {
			case "task":
				r.Process = v
			case "pid":
				r.PID, _ = strconv.Atoi(v)
			}
		}
		if !validProcessName(r.Process) {
			return kmsgReport{}, false
		}
		return r, true
	}

	if rest, found := strings.CutPrefix(msg, "Killed process "); found {
		pidStr, after, ok := strings.Cut(rest, " ")
		if !ok {
			return kmsgReport{}, false
		}
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			return kmsgReport{}, false
		}
		name, ok := betweenParens(after)
		if !ok || !validProcessName(name) {
			return kmsgReport{}, false
		}
		return kmsgReport{Kind: agentwire.KindOOMKill, Process: name, PID: pid}, true
	}

	return kmsgReport{}, false
}

// parseSegfault recognises the segmentation fault report:
//
//	myapp[1234]: segfault at 0 ip ... error 4 in libc.so.6
func parseSegfault(msg string) (kmsgReport, bool) {
	if !strings.Contains(msg, "segfault at ") {
		return kmsgReport{}, false
	}
	head, _, ok := strings.Cut(msg, ":")
	if !ok {
		return kmsgReport{}, false
	}
	name, pidStr, ok := cutBracket(strings.TrimSpace(head))
	if !ok {
		return kmsgReport{}, false
	}
	pid, err := strconv.Atoi(pidStr)
	if err != nil || !validProcessName(name) {
		return kmsgReport{}, false
	}
	return kmsgReport{Kind: agentwire.KindSegfault, Process: name, PID: pid}, true
}

// fatalSignalMarker is the arm64 kernel's way of reporting the same thing
// parseSegfault handles on x86.
const fatalSignalMarker = "potentially unexpected fatal signal "

// sigsegv is the only fatal signal this parser turns into an event. Others --
// SIGBUS, SIGILL, SIGABRT -- are real and would need kinds of their own; naming
// them "segfault" would be a lie, and inventing kinds here is not this parser's
// call.
const sigsegv = 11

// parseFatalSignal recognises the arm64 form of a fatal-signal report:
//
//	python3.14: python3: potentially unexpected fatal signal 11.
//
// This shape exists because arm64 does not emit the x86 "segfault at ..." line.
// The gap was found by running the agent on an arm64 machine, not by reading
// kernel sources -- x86 was all the original parser had ever seen.
//
// Two names appear because the kernel prints the thread's comm and the process
// name; the one adjacent to the marker is the one to report. There is no pid in
// this format, so the report carries none rather than an invented zero.
//
// One thing an operator has to know, and it belongs in the docs rather than
// here: on most distributions this line is not emitted at all unless
// debug.exception-trace is 1. A quiet kernel is indistinguishable from a healthy
// one, so capture silently sees nothing.
func parseFatalSignal(msg string) (kmsgReport, bool) {
	idx := strings.Index(msg, fatalSignalMarker)
	if idx < 0 {
		return kmsgReport{}, false
	}

	rest := msg[idx+len(fatalSignalMarker):]
	numEnd := 0
	for numEnd < len(rest) && rest[numEnd] >= '0' && rest[numEnd] <= '9' {
		numEnd++
	}
	if numEnd == 0 {
		return kmsgReport{}, false
	}
	sig, err := strconv.Atoi(rest[:numEnd])
	if err != nil || sig != sigsegv {
		return kmsgReport{}, false
	}

	// The name is the last colon-separated field before the marker.
	head := strings.TrimSpace(msg[:idx])
	head = strings.TrimSuffix(head, ":")
	if i := strings.LastIndexByte(head, ':'); i >= 0 {
		head = head[i+1:]
	}
	name := strings.TrimSpace(head)
	if !validProcessName(name) {
		return kmsgReport{}, false
	}
	return kmsgReport{Kind: agentwire.KindSegfault, Process: name}, true
}

// betweenParens returns the text inside the first "(...)" pair.
func betweenParens(s string) (string, bool) {
	open := strings.IndexByte(s, '(')
	if open < 0 {
		return "", false
	}
	closeIdx := strings.IndexByte(s[open+1:], ')')
	if closeIdx < 0 {
		return "", false
	}
	name := s[open+1 : open+1+closeIdx]
	if name == "" {
		return "", false
	}
	return name, true
}

// cutBracket splits "name[1234]" into its two parts.
func cutBracket(s string) (string, string, bool) {
	open := strings.IndexByte(s, '[')
	if open <= 0 || !strings.HasSuffix(s, "]") {
		return "", "", false
	}
	return s[:open], s[open+1 : len(s)-1], true
}

// validProcessName rejects anything this parser should not present as a process.
//
// Found by the fuzz campaign on its own seed corpus: /dev/kmsg carries arbitrary
// bytes, and nothing guarantees a name is valid UTF-8. Go's JSON encoder does not
// fail on invalid UTF-8 — it substitutes the replacement character — so an
// unchecked name would travel the wire and reach the operator's screen as
// mojibake, with no error anywhere to explain it. Refusing the report is better
// than displaying garbage attributed to a real kill.
//
// Control characters are refused for the same reason, and length because a
// process name is short: the kernel's own limit is 15 characters, so anything
// long is a line this parser has misread rather than a name.
func validProcessName(name string) bool {
	if name == "" || len(name) > maxProcessNameLen {
		return false
	}
	if !utf8.ValidString(name) {
		return false
	}
	for _, r := range name {
		if unicode.IsControl(r) {
			return false
		}
		// A name carrying the kernel log's own separators means this parser
		// mis-sliced the line and is about to present structure as a process.
		// Found by fuzzing: ";;[0]:segfault at 0" yielded a process named ";".
		if r == ';' {
			return false
		}
	}
	return true
}
