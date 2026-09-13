package main

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/pkg/agentwire"
)

type fixedProbe struct{ caps agentwire.Capabilities }

func (p fixedProbe) Probe() agentwire.Capabilities { return p.caps }

func on() agentwire.Capability { return agentwire.Capability{Available: true} }
func off() agentwire.Capability {
	return agentwire.Capability{Available: false, Reason: agentwire.ReasonUnreadable}
}

// countingOpener records how many times each source was opened.
func countingOpener(opens map[string]int) sourceOpener {
	return func(name string) kernelSource {
		opens[name]++
		return &fakeSource{name: name}
	}
}

func TestProbeAndAdopt_AdoptsAMissingSourceOnce(t *testing.T) {
	c := newEventCollector(nil) // the container start: nothing readable
	opens := map[string]int{}
	p := fixedProbe{agentwire.Capabilities{Kmsg: on(), CgroupOOM: off(), Segfault: on()}}

	got := probeAndAdopt(p, c, countingOpener(opens))
	assert.Equal(t, p.caps, got, "the declaration is what the probe said")
	assert.True(t, c.has(agentwire.SourceKmsg), "kmsg became readable: adopted")
	assert.False(t, c.has(agentwire.SourceCgroup))

	// The probe keeps saying available; the collector must not keep opening.
	probeAndAdopt(p, c, countingOpener(opens))
	probeAndAdopt(p, c, countingOpener(opens))
	assert.Equal(t, 1, opens[agentwire.SourceKmsg], "one reader per source per lifetime")
	assert.Equal(t, 0, opens[agentwire.SourceCgroup])
}

func TestProbeAndAdopt_AFailedProbeNeverRemovesALiveSource(t *testing.T) {
	live := &fakeSource{name: agentwire.SourceKmsg}
	c := newEventCollector([]kernelSource{live})
	opens := map[string]int{}
	p := fixedProbe{agentwire.Capabilities{Kmsg: off(), CgroupOOM: off(), Segfault: off()}}

	got := probeAndAdopt(p, c, countingOpener(opens))
	assert.False(t, got.Kmsg.Available, "the declaration reports what the probe saw")
	assert.True(t, c.has(agentwire.SourceKmsg), "an open descriptor stays readable; the reader stays")
	assert.False(t, live.closed)
	assert.Empty(t, opens)
}

func TestProbeAndAdopt_OpenerFailureIsTransient(t *testing.T) {
	c := newEventCollector(nil)
	p := fixedProbe{agentwire.Capabilities{Kmsg: on(), CgroupOOM: on(), Segfault: on()}}
	calls := 0
	flaky := func(name string) kernelSource {
		calls++
		if calls == 1 {
			return nil // probed readable, failed to open a moment later
		}
		return &fakeSource{name: name}
	}
	probeAndAdopt(p, c, flaky)
	probeAndAdopt(p, c, flaky)
	assert.True(t, c.has(agentwire.SourceKmsg))
	assert.True(t, c.has(agentwire.SourceCgroup), "the next interval tries again")
}

func TestProbeAndAdopt_WithoutCollectorOrOpenerStillDeclares(t *testing.T) {
	p := fixedProbe{agentwire.Capabilities{Kmsg: off(), CgroupOOM: off(), Segfault: off()}}
	assert.Equal(t, p.caps, probeAndAdopt(p, nil, nil))
	assert.Equal(t, p.caps, probeAndAdopt(p, newEventCollector(nil), nil))
}

func TestAdopt_IdempotentPerNameAndLogsOnce(t *testing.T) {
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})))
	defer slog.SetDefault(prev)

	c := newEventCollector(nil)
	require.True(t, c.adopt(&fakeSource{name: agentwire.SourceCgroup}))
	require.False(t, c.adopt(&fakeSource{name: agentwire.SourceCgroup}))
	require.False(t, c.adopt(nil))
	assert.Equal(t, 1, bytes.Count(buf.Bytes(), []byte("kernel event source adopted")), "one line per adoption")
}

// FR-004 / FR-005: an agent with nothing readable declares three unavailables
// on every frame, and the startup "capture unavailable" line is emitted once,
// however many probes follow.
func TestCollector_AllUnavailable_DeclaresEveryFrame_LogsUnavailableOnce(t *testing.T) {
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})))
	defer slog.SetDefault(prev)

	events := newEventCollector(nil) // logs "capture unavailable" once, here
	p := fixedProbe{agentwire.Capabilities{Kmsg: off(), CgroupOOM: off(), Segfault: off()}}
	for i := 0; i < 5; i++ {
		caps := probeAndAdopt(p, events, func(string) kernelSource { return nil })
		assert.False(t, caps.Kmsg.Available)
		assert.False(t, caps.CgroupOOM.Available)
		assert.False(t, caps.Segfault.Available)
	}
	assert.Equal(t, 1, bytes.Count(buf.Bytes(), []byte("kernel event capture unavailable")))
	assert.True(t, events.unavailableLogged)
}
