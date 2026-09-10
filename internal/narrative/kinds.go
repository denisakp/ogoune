package narrative

import "github.com/denisakp/ogoune/pkg/agentwire"

// kindPhrase is what the kernel did, as a verb phrase taking a target: "the
// kernel " + phrase. One entry per kind this version knows how to say out loud.
//
// The kind set is open and carries agent-supplied text (pkg/agentwire/event.go).
// Interpolating an unrecognised kind into operator-facing prose would let a
// future -- or malformed -- agent choose the wording of an alert, so selection
// draws only from this map. An unrecognised kind still counts toward "and N
// other events" and still appears in the served event list, so nothing is
// hidden; it is simply never named in a sentence .
//
// Adding a kind means adding a line here. That is the whole extension point.
var kindPhrase = map[string]string{
	agentwire.KindOOMKill:  "OOM-killed",
	agentwire.KindSegfault: "recorded a segmentation fault in",
}

// Phrasable reports whether a sentence can be produced for this kind.
func Phrasable(kind string) bool {
	_, ok := kindPhrase[kind]
	return ok
}
