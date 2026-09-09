package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/fake"
)

// hostContextFixture wires an IncidentService with in-memory repositories and a
// frozen clock, so window arithmetic and the resolution boundary are testable
// without a database or wall-clock flake.
type hostContextFixture struct {
	svc       *IncidentService
	incidents *fake.IncidentFake
	hosts     *fake.HostFake
	metrics   *fake.HostMetricFake
	now       time.Time
}

func newHostContextFixture(t *testing.T, rawWindow time.Duration) *hostContextFixture {
	t.Helper()

	incidents := fake.NewIncidentFake()
	hosts := fake.NewHostFake()
	metrics := fake.NewHostMetricFake()
	now := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)

	svc := NewIncidentService(incidents, fake.NewIncidentEventStepFake(), metrics, hosts, rawWindow)
	svc.now = func() time.Time { return now }

	return &hostContextFixture{svc: svc, incidents: incidents, hosts: hosts, metrics: metrics, now: now}
}

// seedIncident creates a host, a monitor attached to it (unless detached is
// true), and an incident that started at startedAt.
func (f *hostContextFixture) seedIncident(t *testing.T, id string, startedAt time.Time, detached bool) *domain.Incident {
	t.Helper()
	ctx := context.Background()

	host := &domain.Host{Base: domain.Base{ID: "host-" + id}, Name: "web-" + id}
	require.NoError(t, f.hosts.Create(ctx, host))

	resource := domain.Resource{Base: domain.Base{ID: "res-" + id}, Name: "site-" + id}
	if !detached {
		hostID := host.ID
		resource.HostID = &hostID
	}

	inc := &domain.Incident{
		Base:       domain.Base{ID: id},
		ResourceID: resource.ID,
		Resource:   resource,
		StartedAt:  startedAt,
	}
	_, err := f.incidents.Create(ctx, inc)
	require.NoError(t, err)
	return inc
}

func (f *hostContextFixture) seedSample(t *testing.T, hostID string, at time.Time, cpu, mem float64, disks []domain.DiskUsage) {
	t.Helper()
	require.NoError(t, f.metrics.Insert(context.Background(), &domain.HostMetricSample{
		HostID:    hostID,
		SampledAt: at,
		CPUPct:    cpu,
		MemPct:    mem,
		Disks:     disks,
	}))
}

// T013 — the populated path, plus the invariant that a non-nil context always
// carries a sample count and a resolution marker (SC-004).
func TestGetIncidentByID_HostContext_Populated(t *testing.T) {
	f := newHostContextFixture(t, 7*24*time.Hour)
	started := f.now.Add(-2 * time.Minute)
	inc := f.seedIncident(t, "inc-populated", started, false)

	// Peaks deliberately on different samples, so a per-field maximum is exercised.
	f.seedSample(t, "host-inc-populated", started.Add(-time.Minute), 93.5, 20.0,
		[]domain.DiskUsage{{Mount: "/", UsedPct: 40.0}})
	f.seedSample(t, "host-inc-populated", started.Add(-30*time.Second), 10.0, 88.25,
		[]domain.DiskUsage{{Mount: "/", UsedPct: 55.0}, {Mount: "/var", UsedPct: 96.5}})

	got, err := f.svc.GetIncidentByID(context.Background(), inc.ID)
	require.NoError(t, err)
	require.NotNil(t, got.HostContext)

	hc := got.HostContext
	assert.Equal(t, "host-inc-populated", hc.HostID)
	assert.Equal(t, "web-inc-populated", hc.HostName)
	assert.InDelta(t, 93.5, hc.PeakCPUPct, 0.001)
	assert.InDelta(t, 88.25, hc.PeakMemPct, 0.001)
	require.NotNil(t, hc.WorstDisk)
	assert.Equal(t, "/var", hc.WorstDisk.Mount, "the worst mount across the whole window, not the first seen")
	assert.InDelta(t, 96.5, hc.WorstDisk.UsedPct, 0.001)

	// SC-004 / SC-005 invariants.
	assert.GreaterOrEqual(t, hc.SampleCount, 1, "a non-nil context always has at least one sample")
	assert.NotEmpty(t, hc.Resolution, "a non-nil context always carries a resolution marker")
}

// T013 — a monitor with no host attached yields no context and no error (FR-011).
func TestGetIncidentByID_HostContext_NoHostAttached(t *testing.T) {
	f := newHostContextFixture(t, 7*24*time.Hour)
	inc := f.seedIncident(t, "inc-nohost", f.now.Add(-2*time.Minute), true)

	got, err := f.svc.GetIncidentByID(context.Background(), inc.ID)
	require.NoError(t, err)
	assert.Nil(t, got.HostContext)
}

// T013 — a host with no samples in the window is absent, not zero-filled (FR-010).
func TestGetIncidentByID_HostContext_NoSamplesIsAbsentNotZeroed(t *testing.T) {
	f := newHostContextFixture(t, 7*24*time.Hour)
	started := f.now.Add(-2 * time.Minute)
	inc := f.seedIncident(t, "inc-nosamples", started, false)

	// A sample well outside the window must not resurrect the context.
	f.seedSample(t, "host-inc-nosamples", started.Add(-2*time.Hour), 99.0, 99.0, nil)

	got, err := f.svc.GetIncidentByID(context.Background(), inc.ID)
	require.NoError(t, err)
	assert.Nil(t, got.HostContext, "absence is expressed by absence, never by a zeroed object")
}

// T013 — a host that reported no mounts still gets CPU and memory (spec 089 edge case).
func TestGetIncidentByID_HostContext_NoDisksKeepsTheRest(t *testing.T) {
	f := newHostContextFixture(t, 7*24*time.Hour)
	started := f.now.Add(-2 * time.Minute)
	inc := f.seedIncident(t, "inc-nodisks", started, false)
	f.seedSample(t, "host-inc-nodisks", started, 71.0, 62.0, nil)

	got, err := f.svc.GetIncidentByID(context.Background(), inc.ID)
	require.NoError(t, err)
	require.NotNil(t, got.HostContext)
	assert.Nil(t, got.HostContext.WorstDisk)
	assert.InDelta(t, 71.0, got.HostContext.PeakCPUPct, 0.001, "one missing signal must not suppress the others")
}

// T014 — a repository failure degrades to an absent context, never an error
// reaching the caller (FR-012).
func TestGetIncidentByID_HostContext_RepositoryErrorDegrades(t *testing.T) {
	f := newHostContextFixture(t, 7*24*time.Hour)
	started := f.now.Add(-2 * time.Minute)
	inc := f.seedIncident(t, "inc-repoerr", started, false)
	f.seedSample(t, "host-inc-repoerr", started, 50.0, 50.0, nil)
	f.metrics.AggregateWindowErr = errors.New("connection reset")

	got, err := f.svc.GetIncidentByID(context.Background(), inc.ID)
	require.NoError(t, err, "the enrichment must never fail the request that carries it")
	require.NotNil(t, got)
	assert.Nil(t, got.HostContext)
}

// T015 — a stored sample whose disk document will not decode must not cost the
// operator the peaks that decoded fine. FR-012 and SC-002 both name this case.
func TestGetIncidentByID_HostContext_MalformedDiskDocumentDegrades(t *testing.T) {
	f := newHostContextFixture(t, 7*24*time.Hour)
	started := f.now.Add(-2 * time.Minute)
	inc := f.seedIncident(t, "inc-malformed", started, false)

	// The fake stores decoded values, so an undecodable document is modelled as a
	// sample carrying no usable mounts alongside one that does.
	f.seedSample(t, "host-inc-malformed", started.Add(-time.Minute), 64.0, 33.0, nil)
	f.seedSample(t, "host-inc-malformed", started, 22.0, 41.0,
		[]domain.DiskUsage{{Mount: "/", UsedPct: 77.0}})

	got, err := f.svc.GetIncidentByID(context.Background(), inc.ID)
	require.NoError(t, err)
	require.NotNil(t, got.HostContext)
	assert.InDelta(t, 64.0, got.HostContext.PeakCPUPct, 0.001, "an unusable disk document must not lose the CPU peak")
	require.NotNil(t, got.HostContext.WorstDisk)
	assert.Equal(t, "/", got.HostContext.WorstDisk.Mount)
}

// T016 — the correlation window is a code constant. Its bounds must not shift
// with any configuration value (FR-002).
func TestGetIncidentByID_HostContext_WindowIsACodeConstant(t *testing.T) {
	started := time.Date(2026, 3, 10, 11, 58, 0, 0, time.UTC)

	// Two services with wildly different retention configuration must compute
	// exactly the same window bounds.
	for _, rawWindow := range []time.Duration{time.Minute, 30 * 24 * time.Hour} {
		f := newHostContextFixture(t, rawWindow)
		inc := f.seedIncident(t, "inc-const", started, false)
		f.seedSample(t, "host-inc-const", started, 12.0, 13.0, nil)

		got, err := f.svc.GetIncidentByID(context.Background(), inc.ID)
		require.NoError(t, err)
		require.NotNil(t, got.HostContext)

		assert.Equal(t, started.Add(-hostContextWindowBefore), got.HostContext.WindowFrom.UTC(),
			"window start is startedAt - 5m regardless of configuration")
		assert.Equal(t, started.Add(hostContextWindowAfter), got.HostContext.WindowTo.UTC(),
			"window end is startedAt + 1m regardless of configuration")
	}

	assert.Equal(t, 5*time.Minute, hostContextWindowBefore)
	assert.Equal(t, time.Minute, hostContextWindowAfter)
}

// T029 -- the resolution boundary. A window entirely inside the configured
// full-resolution retention window is "full"; entirely outside, or straddling,
// is "reduced" (FR-007, FR-007b).
func TestGetIncidentByID_HostContext_ResolutionBoundary(t *testing.T) {
	rawWindow := 48 * time.Hour

	cases := []struct {
		name    string
		age     time.Duration // how long before "now" the incident started
		want    domain.HostContextResolution
		comment string
	}{
		{"well inside the window", 2 * time.Hour, domain.HostContextFull, "recent incident, samples untouched"},
		{"just inside the window", 47 * time.Hour, domain.HostContextFull, "still within retention's raw span"},
		{"straddling the threshold", 48 * time.Hour, domain.HostContextReduced, "errs toward reduced when it straddles"},
		{"well outside the window", 96 * time.Hour, domain.HostContextReduced, "thinned to one sample per minute"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newHostContextFixture(t, rawWindow)
			started := f.now.Add(-tc.age)
			inc := f.seedIncident(t, "inc-res-"+tc.name, started, false)
			f.seedSample(t, "host-inc-res-"+tc.name, started, 30.0, 40.0, nil)

			got, err := f.svc.GetIncidentByID(context.Background(), inc.ID)
			require.NoError(t, err)
			require.NotNil(t, got.HostContext)
			assert.Equal(t, tc.want, got.HostContext.Resolution, tc.comment)
		})
	}
}

// T030 -- resolution must never be inferred from the samples. A host reporting
// once a minute natively, inside the threshold, is at full resolution: its data
// is intact, it is simply a slow reporter. This is the regression FR-007a exists
// to prevent.
func TestGetIncidentByID_HostContext_ResolutionIgnoresSampleDensity(t *testing.T) {
	f := newHostContextFixture(t, 48*time.Hour)
	started := f.now.Add(-2 * time.Hour)
	inc := f.seedIncident(t, "inc-sparse", started, false)

	// Six samples across the six-minute window: exactly what decimation would
	// leave behind, but here it is the agent's native interval.
	for i := -5; i <= 0; i++ {
		f.seedSample(t, "host-inc-sparse", started.Add(time.Duration(i)*time.Minute), 25.0, 35.0, nil)
	}

	got, err := f.svc.GetIncidentByID(context.Background(), inc.ID)
	require.NoError(t, err)
	require.NotNil(t, got.HostContext)
	assert.Equal(t, domain.HostContextFull, got.HostContext.Resolution,
		"a sparse-but-recent window is full resolution: the agent's interval is configurable")
	assert.Equal(t, 6, got.HostContext.SampleCount)
}

// T033 -- past the retention horizon the samples are gone, so the context is
// absent rather than reduced. No extra code path: the aggregate simply finds
// nothing, and absence is expressed by absence.
func TestGetIncidentByID_HostContext_OutOfRetentionIsAbsent(t *testing.T) {
	f := newHostContextFixture(t, 7*24*time.Hour)
	started := f.now.Add(-90 * 24 * time.Hour)
	inc := f.seedIncident(t, "inc-purged", started, false)
	// Nothing seeded: retention purged this host's samples long ago.

	got, err := f.svc.GetIncidentByID(context.Background(), inc.ID)
	require.NoError(t, err)
	assert.Nil(t, got.HostContext,
		"an out-of-retention incident degrades exactly like a monitor with no host")
}

// T035 -- a window that has only partly elapsed aggregates over what exists, and
// reports the real bounds rather than the nominal ones (FR-004).
func TestGetIncidentByID_HostContext_PartiallyElapsedWindow(t *testing.T) {
	f := newHostContextFixture(t, 7*24*time.Hour)
	started := f.now.Add(-30 * time.Second) // +1m has not happened yet
	inc := f.seedIncident(t, "inc-fresh", started, false)
	f.seedSample(t, "host-inc-fresh", started.Add(-10*time.Second), 44.0, 55.0, nil)

	got, err := f.svc.GetIncidentByID(context.Background(), inc.ID)
	require.NoError(t, err)
	require.NotNil(t, got.HostContext)

	assert.Equal(t, f.now, got.HostContext.WindowTo.UTC(),
		"the window end is clamped to now while it is still elapsing")
	assert.True(t, got.HostContext.WindowTo.Before(started.Add(hostContextWindowAfter)),
		"and is genuinely earlier than the nominal end")
	assert.Equal(t, 1, got.HostContext.SampleCount)
}

// T036 -- the same incident evaluated later covers the full window and a larger
// sample count. Nothing is cached between reads.
func TestGetIncidentByID_HostContext_ReopenPicksUpNewSamples(t *testing.T) {
	f := newHostContextFixture(t, 7*24*time.Hour)
	started := f.now.Add(-30 * time.Second)
	inc := f.seedIncident(t, "inc-reopen", started, false)
	f.seedSample(t, "host-inc-reopen", started.Add(-10*time.Second), 44.0, 55.0, nil)

	first, err := f.svc.GetIncidentByID(context.Background(), inc.ID)
	require.NoError(t, err)
	require.NotNil(t, first.HostContext)
	require.Equal(t, 1, first.HostContext.SampleCount)

	// Time moves past the window's nominal end, and a sample lands inside it.
	later := f.now.Add(2 * time.Minute)
	f.svc.now = func() time.Time { return later }
	f.seedSample(t, "host-inc-reopen", started.Add(30*time.Second), 81.0, 60.0, nil)

	second, err := f.svc.GetIncidentByID(context.Background(), inc.ID)
	require.NoError(t, err)
	require.NotNil(t, second.HostContext)
	assert.Equal(t, 2, second.HostContext.SampleCount, "no cache between reads")
	assert.InDelta(t, 81.0, second.HostContext.PeakCPUPct, 0.001)
	assert.Equal(t, started.Add(hostContextWindowAfter).UTC(), second.HostContext.WindowTo.UTC(),
		"the full window is now in the past, so the nominal end applies")
}

// countingHostContextMetrics records absence reasons for assertion.
type countingHostContextMetrics struct{ reasons map[string]int }

func newCountingMetrics() *countingHostContextMetrics {
	return &countingHostContextMetrics{reasons: map[string]int{}}
}

func (c *countingHostContextMetrics) RecordHostContextAbsent(reason string) {
	c.reasons[reason]++
}

// T042 -- every silent absence is counted, and each reason is distinguishable.
// no_host_attached is deliberately not counted: it is the normal state of most
// monitors, not a signal (spec 089, FR-013a).
func TestGetIncidentByID_HostContext_AbsenceIsCountedByReason(t *testing.T) {
	t.Run("no_samples", func(t *testing.T) {
		f := newHostContextFixture(t, 7*24*time.Hour)
		obs := newCountingMetrics()
		f.svc.WithHostContextMetrics(obs)
		inc := f.seedIncident(t, "inc-count-nosamples", f.now.Add(-2*time.Minute), false)

		_, err := f.svc.GetIncidentByID(context.Background(), inc.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, obs.reasons["no_samples"])
		assert.Zero(t, obs.reasons["out_of_retention"])
		assert.Zero(t, obs.reasons["lookup_error"])
	})

	t.Run("out_of_retention", func(t *testing.T) {
		f := newHostContextFixture(t, 48*time.Hour)
		obs := newCountingMetrics()
		f.svc.WithHostContextMetrics(obs)
		inc := f.seedIncident(t, "inc-count-old", f.now.Add(-90*24*time.Hour), false)

		_, err := f.svc.GetIncidentByID(context.Background(), inc.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, obs.reasons["out_of_retention"])
		assert.Zero(t, obs.reasons["no_samples"])
	})

	t.Run("lookup_error", func(t *testing.T) {
		f := newHostContextFixture(t, 7*24*time.Hour)
		obs := newCountingMetrics()
		f.svc.WithHostContextMetrics(obs)
		inc := f.seedIncident(t, "inc-count-err", f.now.Add(-2*time.Minute), false)
		f.metrics.AggregateWindowErr = errors.New("connection reset")

		_, err := f.svc.GetIncidentByID(context.Background(), inc.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, obs.reasons["lookup_error"])
	})

	t.Run("no_host_attached_is_not_counted", func(t *testing.T) {
		f := newHostContextFixture(t, 7*24*time.Hour)
		obs := newCountingMetrics()
		f.svc.WithHostContextMetrics(obs)
		inc := f.seedIncident(t, "inc-count-nohost", f.now.Add(-2*time.Minute), true)

		_, err := f.svc.GetIncidentByID(context.Background(), inc.ID)
		require.NoError(t, err)
		assert.Empty(t, obs.reasons, "the normal state of most monitors is not a signal")
	})
}
