//go:build !linux

package main

// Kernel event capture does not apply outside Linux (spec 090). The agent is
// Linux-only by a settled decision, and the sources it reads exist nowhere else.
//
// This file exists so the rest of the agent never branches on platform: it asks
// for sources and gets none, which is the same answer a Linux host gives when it
// cannot read them — the common case there too. The collector reports
// unavailability once and metrics are untouched.
//
// The platform seam is at file level rather than a build tag on the whole agent.
// That was considered and rejected: tagging the agent would drop its tests out of
// a macOS developer's `go test ./...` while CI, which builds on Linux, would still
// run them — the kind of asymmetry that hides a failure until someone else finds
// it.
func newKernelSources() []kernelSource {
	return nil
}
