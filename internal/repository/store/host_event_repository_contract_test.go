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

func TestHostEventRepository_Contract(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		hosts := store.NewHostRepositorySQLC(fx.Runtime)
		repo := store.NewHostEventRepositorySQLC(fx.Runtime)
		runHostEventContract(t, hosts, repo)
	})
}

func seedHost(t *testing.T, hosts port.HostRepository, id string) {
	t.Helper()
	require.NoError(t, hosts.Create(context.Background(), &domain.Host{
		Base: domain.Base{ID: id},
		Name: id,
	}))
}

func evt(hostID, kind string, at time.Time) *domain.HostEvent {
	return &domain.HostEvent{
		HostID:      hostID,
		OccurredAt:  at,
		Kind:        kind,
		Source:      "kmsg",
		Occurrences: 1,
	}
}

func runHostEventContract(t *testing.T, hosts port.HostRepository, repo port.HostEventRepository) {
	t.Helper()
	ctx := context.Background()
	base := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)

	t.Run("RoundTripsClassifiedDetail", func(t *testing.T) {
		id := "he-detail"
		seedHost(t, hosts, id)

		e := evt(id, "oom_kill", base)
		e.Occurrences = 37
		e.Detail = &domain.HostEventDetail{
			Process:           "postgres",
			PID:               4711,
			Cgroup:            "/system.slice/docker-a1b2.scope",
			DistinctProcesses: []string{"postgres", "node"},
			DistinctTruncated: true,
		}
		require.NoError(t, repo.Create(ctx, e))
		assert.NotEmpty(t, e.ID, "the wrapper assigns the id")

		got, err := repo.ListByHost(ctx, id, 10)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, 37, got[0].Occurrences)
		require.NotNil(t, got[0].Detail)
		assert.Equal(t, "postgres", got[0].Detail.Process)
		assert.Equal(t, 4711, got[0].Detail.PID)
		assert.Equal(t, []string{"postgres", "node"}, got[0].Detail.DistinctProcesses)
		assert.True(t, got[0].Detail.DistinctTruncated,
			"the truncation marker survives, so a partial list is never read as complete")
	})

	// A kernel report that named nothing and one whose detail was lost must not
	// look alike: absent detail round-trips as nil, never as an empty object.
	t.Run("AbsentDetailRoundTripsAsNil", func(t *testing.T) {
		id := "he-nodetail"
		seedHost(t, hosts, id)
		require.NoError(t, repo.Create(ctx, evt(id, "segfault", base)))

		got, err := repo.ListByHost(ctx, id, 10)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Nil(t, got[0].Detail)
	})

	t.Run("NewestFirst", func(t *testing.T) {
		id := "he-order"
		seedHost(t, hosts, id)
		for i := 0; i < 5; i++ {
			require.NoError(t, repo.Create(ctx, evt(id, "oom_kill", base.Add(time.Duration(i)*time.Minute))))
		}

		got, err := repo.ListByHost(ctx, id, 10)
		require.NoError(t, err)
		require.Len(t, got, 5)
		for i := 1; i < len(got); i++ {
			assert.Falsef(t, got[i].OccurredAt.After(got[i-1].OccurredAt),
				"position %d is newer than the one before it", i)
		}
		assert.True(t, got[0].OccurredAt.Equal(base.Add(4*time.Minute)),
			"an operator opening a host page wants what just happened")
	})

	t.Run("LimitBoundsTheRead", func(t *testing.T) {
		id := "he-limit"
		seedHost(t, hosts, id)
		for i := 0; i < 20; i++ {
			require.NoError(t, repo.Create(ctx, evt(id, "oom_kill", base.Add(time.Duration(i)*time.Minute))))
		}

		got, err := repo.ListByHost(ctx, id, 5)
		require.NoError(t, err)
		assert.Len(t, got, 5)
	})

	t.Run("ScopedToTheirHost", func(t *testing.T) {
		seedHost(t, hosts, "he-mine")
		seedHost(t, hosts, "he-theirs")
		require.NoError(t, repo.Create(ctx, evt("he-mine", "oom_kill", base)))
		require.NoError(t, repo.Create(ctx, evt("he-theirs", "oom_kill", base)))

		got, err := repo.ListByHost(ctx, "he-mine", 10)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "he-mine", got[0].HostID)
	})

	// Retention: deletion only, never decimation. A kernel event is rare and
	// discrete, and thinning would purge exactly what this feature exists to keep.
	t.Run("RetentionDeletesByCutoff", func(t *testing.T) {
		id := "he-retention"
		seedHost(t, hosts, id)
		require.NoError(t, repo.Create(ctx, evt(id, "oom_kill", base.Add(-100*24*time.Hour))))
		require.NoError(t, repo.Create(ctx, evt(id, "oom_kill", base)))

		_, err := repo.DeleteOlderThan(ctx, base.Add(-90*24*time.Hour))
		require.NoError(t, err)

		got, err := repo.ListByHost(ctx, id, 10)
		require.NoError(t, err)
		require.Len(t, got, 1, "only the old one goes")
		assert.True(t, got[0].OccurredAt.Equal(base))
	})

	t.Run("EventsGoWithTheirHost", func(t *testing.T) {
		id := "he-cascade"
		seedHost(t, hosts, id)
		require.NoError(t, repo.Create(ctx, evt(id, "oom_kill", base)))

		require.NoError(t, hosts.Delete(ctx, id))

		got, err := repo.ListByHost(ctx, id, 10)
		require.NoError(t, err)
		assert.Empty(t, got, "deleting a host takes its events with it")
	})

	// The correlation window (spec 091). Half-open [from, to), so an event landing
	// exactly on the far edge belongs to one window and not two -- asserted in
	// both dialects because the two compare timestamps by different mechanisms.
	t.Run("WindowIsHalfOpen", func(t *testing.T) {
		id := "he-window"
		seedHost(t, hosts, id)
		from := base
		to := base.Add(6 * time.Minute)

		require.NoError(t, repo.Create(ctx, evt(id, "oom_kill", from.Add(-time.Second))))
		atFrom := from
		require.NoError(t, repo.Create(ctx, evt(id, "oom_kill", atFrom)))
		inside := from.Add(3 * time.Minute)
		require.NoError(t, repo.Create(ctx, evt(id, "segfault", inside)))
		require.NoError(t, repo.Create(ctx, evt(id, "oom_kill", to)))
		require.NoError(t, repo.Create(ctx, evt(id, "oom_kill", to.Add(time.Hour))))

		got, err := repo.ListInWindow(ctx, id, from, to, 10)
		require.NoError(t, err)
		require.Len(t, got, 2, "the start edge is included, the end edge is not")
		assert.True(t, got[0].OccurredAt.Equal(inside), "newest first")
		assert.True(t, got[1].OccurredAt.Equal(atFrom),
			"an event exactly at the window start belongs to the window")
	})

	t.Run("WindowIsScopedAndBounded", func(t *testing.T) {
		seedHost(t, hosts, "hw-mine")
		seedHost(t, hosts, "hw-theirs")
		from := base
		to := base.Add(10 * time.Minute)
		for i := 0; i < 8; i++ {
			require.NoError(t, repo.Create(ctx, evt("hw-mine", "oom_kill", from.Add(time.Duration(i)*time.Minute))))
		}
		require.NoError(t, repo.Create(ctx, evt("hw-theirs", "oom_kill", from.Add(time.Minute))))

		got, err := repo.ListInWindow(ctx, "hw-mine", from, to, 3)
		require.NoError(t, err)
		require.Len(t, got, 3, "limit bounds the read: a storm must not grow the result")
		for _, e := range got {
			assert.Equal(t, "hw-mine", e.HostID, "another host's events never enter the window")
		}
	})

	// An empty or inverted window has no events by definition. The wrapper answers
	// without asking the database, and must answer the same way either dialect.
	t.Run("EmptyWindowYieldsNothing", func(t *testing.T) {
		id := "he-emptywindow"
		seedHost(t, hosts, id)
		require.NoError(t, repo.Create(ctx, evt(id, "oom_kill", base)))

		got, err := repo.ListInWindow(ctx, id, base, base, 10)
		require.NoError(t, err)
		assert.Empty(t, got, "a zero-width window holds nothing")

		got, err = repo.ListInWindow(ctx, id, base.Add(time.Minute), base, 10)
		require.NoError(t, err)
		assert.Empty(t, got, "an inverted window holds nothing")
	})

	t.Run("UnknownKindIsStoredNotDropped", func(t *testing.T) {
		id := "he-unknown"
		seedHost(t, hosts, id)
		require.NoError(t, repo.Create(ctx, evt(id, "something_new", base)))

		got, err := repo.ListByHost(ctx, id, 10)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "something_new", got[0].Kind,
			"the kind set is open: an older backend must not lose what a newer agent sends")
	})
}
