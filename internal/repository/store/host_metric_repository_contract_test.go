package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
)

func TestHostMetricRepository_Contract(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		repo := store.NewHostMetricRepositorySQLC(fx.Runtime)
		runHostMetricContract(t, repo)
	})
}

func newSample(hostID string, at time.Time) *domain.HostMetricSample {
	return &domain.HostMetricSample{
		HostID:    hostID,
		SampledAt: at,
		CPUPct:    10.0,
		MemPct:    20.0,
		NetIn:     100,
		NetOut:    200,
		Disks:     []domain.DiskUsage{{Mount: "/", UsedPct: 42.0}},
	}
}

func runHostMetricContract(t *testing.T, repo port.HostMetricsRepository) {
	t.Helper()
	ctx := context.Background()

	t.Run("Insert_ListInRange_Chronological", func(t *testing.T) {
		hostID := "host-range"
		now := time.Now().UTC().Truncate(time.Second)
		t1 := now.Add(-30 * time.Minute)
		t2 := now.Add(-20 * time.Minute)
		t3 := now.Add(-10 * time.Minute)
		outBefore := now.Add(-90 * time.Minute)
		outAfter := now.Add(10 * time.Minute)

		// Insert deliberately out of chronological order.
		for _, at := range []time.Time{t2, outBefore, t3, t1, outAfter} {
			require.NoError(t, repo.Insert(ctx, newSample(hostID, at)))
		}

		got, err := repo.ListInRange(ctx, hostID, now.Add(-40*time.Minute), now)
		require.NoError(t, err)
		require.Len(t, got, 3, "window excludes out-of-range samples")

		// Chronological ASC.
		assert.WithinDuration(t, t1, got[0].SampledAt.UTC(), time.Second)
		assert.WithinDuration(t, t2, got[1].SampledAt.UTC(), time.Second)
		assert.WithinDuration(t, t3, got[2].SampledAt.UTC(), time.Second)

		// Payload roundtrip on first row.
		assert.InDelta(t, 10.0, got[0].CPUPct, 0.001)
		assert.InDelta(t, 20.0, got[0].MemPct, 0.001)
		assert.EqualValues(t, 100, got[0].NetIn)
		assert.EqualValues(t, 200, got[0].NetOut)
		require.Len(t, got[0].Disks, 1)
		assert.Equal(t, "/", got[0].Disks[0].Mount)
		assert.InDelta(t, 42.0, got[0].Disks[0].UsedPct, 0.001)
	})

	t.Run("DeleteOlderThan_PurgesStrictlyBeforeCutoff", func(t *testing.T) {
		hostID := "host-purge"
		now := time.Now().UTC().Truncate(time.Second)
		old1 := now.Add(-3 * time.Hour)
		old2 := now.Add(-2 * time.Hour)
		recent := now.Add(-10 * time.Minute)
		require.NoError(t, repo.Insert(ctx, newSample(hostID, old1)))
		require.NoError(t, repo.Insert(ctx, newSample(hostID, old2)))
		require.NoError(t, repo.Insert(ctx, newSample(hostID, recent)))

		cutoff := now.Add(-1 * time.Hour)
		n, err := repo.DeleteOlderThan(ctx, cutoff)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, n, int64(2), "at least the two old samples for this host purged")

		remaining, err := repo.ListInRange(ctx, hostID, now.Add(-4*time.Hour), now)
		require.NoError(t, err)
		require.Len(t, remaining, 1, "only the sample at/after cutoff survives")
		assert.WithinDuration(t, recent, remaining[0].SampledAt.UTC(), time.Second)
	})

	t.Run("DeleteByHost", func(t *testing.T) {
		hostID := "host-delete-metrics"
		now := time.Now().UTC().Truncate(time.Second)
		require.NoError(t, repo.Insert(ctx, newSample(hostID, now.Add(-5*time.Minute))))
		require.NoError(t, repo.Insert(ctx, newSample(hostID, now.Add(-1*time.Minute))))

		require.NoError(t, repo.DeleteByHost(ctx, hostID))

		remaining, err := repo.ListInRange(ctx, hostID, now.Add(-1*time.Hour), now.Add(time.Hour))
		require.NoError(t, err)
		assert.Empty(t, remaining)
	})

	t.Run("Decimate_ReducesDensityBeforeCutoff", func(t *testing.T) {
		hostID := "host-decimate"
		// Three samples inside the SAME minute, all older than the cutoff.
		bucket := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)
		require.NoError(t, repo.Insert(ctx, newSample(hostID, bucket.Add(10*time.Second))))
		require.NoError(t, repo.Insert(ctx, newSample(hostID, bucket.Add(20*time.Second))))
		require.NoError(t, repo.Insert(ctx, newSample(hostID, bucket.Add(30*time.Second))))
		// A recent sample AFTER the cutoff — must be left untouched.
		recent := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
		require.NoError(t, repo.Insert(ctx, newSample(hostID, recent)))

		cutoff := time.Date(2026, 1, 1, 11, 0, 0, 0, time.UTC)
		_, err := repo.Decimate(ctx, cutoff)
		require.NoError(t, err)

		got, err := repo.ListInRange(ctx, hostID,
			time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2026, 1, 1, 23, 59, 59, 0, time.UTC))
		require.NoError(t, err)
		require.Len(t, got, 2, "one survivor per minute before cutoff + the post-cutoff sample")

		// Exactly one survivor from the pre-cutoff minute bucket (which one is
		// nondeterministic — MIN(id) over ULIDs minted in the same millisecond).
		survivor := got[0].SampledAt.UTC()
		assert.False(t, survivor.Before(bucket), "survivor within the 10:30 bucket")
		assert.True(t, survivor.Before(bucket.Add(time.Minute)), "survivor within the 10:30 bucket")
		// The post-cutoff sample is untouched.
		assert.WithinDuration(t, recent, got[1].SampledAt.UTC(), time.Second)
	})

	t.Run("AggregateWindow_PeaksAreTrueMaxima", func(t *testing.T) {
		hostID := "host-agg-peaks"
		base := time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)

		// Three samples in the window, with the peaks on different rows so a
		// per-column MAX is genuinely exercised, not a single "worst row".
		hot := newSample(hostID, base.Add(10*time.Second))
		hot.CPUPct, hot.MemPct = 91.5, 30.0
		hot.Disks = []domain.DiskUsage{{Mount: "/", UsedPct: 50.0}}
		require.NoError(t, repo.Insert(ctx, hot))

		full := newSample(hostID, base.Add(20*time.Second))
		full.CPUPct, full.MemPct = 12.0, 98.25
		full.Disks = []domain.DiskUsage{{Mount: "/", UsedPct: 60.0}, {Mount: "/var", UsedPct: 97.5}}
		require.NoError(t, repo.Insert(ctx, full))

		calm := newSample(hostID, base.Add(30*time.Second))
		calm.CPUPct, calm.MemPct = 5.0, 6.0
		calm.Disks = nil // a sample that reported no mounts must not contribute a document
		require.NoError(t, repo.Insert(ctx, calm))

		agg, err := repo.AggregateWindow(ctx, hostID, base, base.Add(time.Minute))
		require.NoError(t, err)
		require.NotNil(t, agg)
		assert.Equal(t, 3, agg.SampleCount)
		assert.InDelta(t, 91.5, agg.PeakCPUPct, 0.001, "peak CPU is a true maximum, not an average or the first value")
		assert.InDelta(t, 98.25, agg.PeakMemPct, 0.001, "peak memory comes from a different row than peak CPU")
		require.Len(t, agg.Disks, 2, "only samples that reported mounts contribute disk documents")
	})

	t.Run("AggregateWindow_ExcludesSamplesOutsideTheWindow", func(t *testing.T) {
		hostID := "host-agg-bounds"
		base := time.Date(2026, 3, 2, 10, 0, 0, 0, time.UTC)

		inside := newSample(hostID, base.Add(30*time.Second))
		inside.CPUPct = 40.0
		require.NoError(t, repo.Insert(ctx, inside))

		// Just outside on both sides, and much hotter — if the range predicate
		// leaks, the peak becomes 99 and this fails loudly.
		before := newSample(hostID, base.Add(-time.Second))
		before.CPUPct = 99.0
		require.NoError(t, repo.Insert(ctx, before))
		after := newSample(hostID, base.Add(2*time.Minute))
		after.CPUPct = 99.0
		require.NoError(t, repo.Insert(ctx, after))

		agg, err := repo.AggregateWindow(ctx, hostID, base, base.Add(time.Minute))
		require.NoError(t, err)
		require.NotNil(t, agg)
		assert.Equal(t, 1, agg.SampleCount)
		assert.InDelta(t, 40.0, agg.PeakCPUPct, 0.001)
	})

	t.Run("AggregateWindow_EmptyWindowIsAbsentNotZeroed", func(t *testing.T) {
		hostID := "host-agg-empty"
		base := time.Date(2026, 3, 3, 10, 0, 0, 0, time.UTC)
		require.NoError(t, repo.Insert(ctx, newSample(hostID, base.Add(-time.Hour))))

		agg, err := repo.AggregateWindow(ctx, hostID, base, base.Add(time.Minute))
		require.NoError(t, err)
		assert.Nil(t, agg, "absence is expressed by absence, never by a zero-filled aggregate")
	})

	t.Run("AggregateWindow_OtherHostsAreNotAggregated", func(t *testing.T) {
		base := time.Date(2026, 3, 4, 10, 0, 0, 0, time.UTC)
		mine := newSample("host-agg-mine", base.Add(10*time.Second))
		mine.CPUPct = 20.0
		require.NoError(t, repo.Insert(ctx, mine))
		theirs := newSample("host-agg-theirs", base.Add(10*time.Second))
		theirs.CPUPct = 99.0
		require.NoError(t, repo.Insert(ctx, theirs))

		agg, err := repo.AggregateWindow(ctx, "host-agg-mine", base, base.Add(time.Minute))
		require.NoError(t, err)
		require.NotNil(t, agg)
		assert.Equal(t, 1, agg.SampleCount)
		assert.InDelta(t, 20.0, agg.PeakCPUPct, 0.001)
	})

	// FR-021b: what the aggregate brings back must be bounded by the window, not
	// by how much history the host has. A host with hundreds of samples outside
	// the window must cost exactly the same as one with none.
	t.Run("AggregateWindow_TransferIsBoundedByTheWindowNotByHistory", func(t *testing.T) {
		hostID := "host-agg-bounded"
		base := time.Date(2026, 3, 5, 12, 0, 0, 0, time.UTC)

		// 300 samples of history spread over the ten hours before the window.
		for i := 0; i < 300; i++ {
			old := newSample(hostID, base.Add(-time.Duration(i+1)*2*time.Minute))
			old.CPUPct = 88.0
			require.NoError(t, repo.Insert(ctx, old))
		}
		// Six samples inside the window, each reporting mounts.
		for i := 0; i < 6; i++ {
			in := newSample(hostID, base.Add(time.Duration(i*10)*time.Second))
			in.CPUPct = float64(10 + i)
			require.NoError(t, repo.Insert(ctx, in))
		}

		agg, err := repo.AggregateWindow(ctx, hostID, base, base.Add(time.Minute))
		require.NoError(t, err)
		require.NotNil(t, agg)
		assert.Equal(t, 6, agg.SampleCount, "sample count reflects the window, not the history")
		assert.InDelta(t, 15.0, agg.PeakCPUPct, 0.001, "history outside the window never reaches the peak")
		assert.LessOrEqual(t, len(agg.Disks), 6,
			"disk documents transferred are bounded by the window: 300 samples of history must not be read back")
	})
}
