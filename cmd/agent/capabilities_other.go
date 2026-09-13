//go:build !linux

package main

import "github.com/denisakp/ogoune/pkg/agentwire"

type platformProbe struct{}

func newCapabilityProbe() capabilityProbe { return platformProbe{} }

// Not Linux: nothing to read, nothing to enable. Still declared -- "not
// supported here" is a fact, and silence is what this feature ends.
func (platformProbe) Probe() agentwire.Capabilities {
	off := agentwire.Capability{Available: false, Reason: agentwire.ReasonPlatform}
	return agentwire.Capabilities{Kmsg: off, CgroupOOM: off, Segfault: off}
}

func newKmsgSourceIfAny() kernelSource      { return nil }
func newCgroupOOMSourceIfAny() kernelSource { return nil }
