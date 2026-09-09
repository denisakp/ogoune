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

func TestResourceHealthRepository_Contract(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		resources := store.NewResourceRepositorySQLC(fx.Runtime)
		repo := store.NewResourceHealthRepositorySQLC(fx.Runtime)
		runResourceHealthContract(t, resources, repo)
	})
}

func i64(v int64) *int64     { return &v }
func f64(v float64) *float64 { return &v }

// seedMonitor creates the monitor a health record must reference.
func seedMonitor(t *testing.T, resources port.ResourceRepository, id string) {
	t.Helper()
	_, err := resources.Create(context.Background(), &domain.Resource{
		Base:     domain.Base{ID: id},
		Name:     id,
		Type:     domain.ResourceProtocol,
		Target:   "postgres://db.example.com:5432",
		IsActive: true,
		Interval: 60,
		Timeout:  10,
	})
	require.NoError(t, err)
}

func runResourceHealthContract(t *testing.T, resources port.ResourceRepository, repo port.ResourceHealthRepository) {
	t.Helper()
	ctx := context.Background()

	t.Run("AbsentReadReturnsNothing", func(t *testing.T) {
		got, err := repo.FindByResourceID(ctx, "rh-never-written")
		require.NoError(t, err, "an absent record is not an error")
		assert.Nil(t, got, "absence is expressed by absence, never a zero-filled struct")
	})

	t.Run("RoundTripsEveryField", func(t *testing.T) {
		id := "rh-roundtrip"
		seedMonitor(t, resources, id)
		at := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)

		require.NoError(t, repo.Upsert(ctx, &domain.ResourceHealth{
			ResourceID:            id,
			CollectedAt:           at,
			ConnectionsActive:     i64(195),
			ConnectionsMax:        i64(200),
			LongestQuerySeconds:   f64(412.5),
			ReplicationLagSeconds: f64(0.75),
			PrivilegeLimited:      false,
			UnsupportedVersion:    false,
		}))

		got, err := repo.FindByResourceID(ctx, id)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.WithinDuration(t, at, got.CollectedAt.UTC(), time.Second)
		require.NotNil(t, got.ConnectionsActive)
		assert.Equal(t, int64(195), *got.ConnectionsActive)
		require.NotNil(t, got.ConnectionsMax)
		assert.Equal(t, int64(200), *got.ConnectionsMax)
		require.NotNil(t, got.LongestQuerySeconds)
		assert.InDelta(t, 412.5, *got.LongestQuerySeconds, 0.001)
		require.NotNil(t, got.ReplicationLagSeconds)
		assert.InDelta(t, 0.75, *got.ReplicationLagSeconds, 0.001)
		assert.False(t, got.PrivilegeLimited)
		assert.False(t, got.UnsupportedVersion)
	})

	// The distinction the whole feature rests on: a field the credential could not
	// read must come back nil, not 0. A database sitting at zero connections and a
	// database whose figures could not be read must never look alike.
	t.Run("NullableFieldsRoundTripAsNullNotZero", func(t *testing.T) {
		id := "rh-nulls"
		seedMonitor(t, resources, id)

		require.NoError(t, repo.Upsert(ctx, &domain.ResourceHealth{
			ResourceID:        id,
			CollectedAt:       time.Now().UTC(),
			ConnectionsActive: i64(12),
			ConnectionsMax:    i64(100),
			// the two privileged fields were withheld
			PrivilegeLimited: true,
		}))

		got, err := repo.FindByResourceID(ctx, id)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Nil(t, got.LongestQuerySeconds, "a withheld field is null, never 0")
		assert.Nil(t, got.ReplicationLagSeconds, "no replication is null, never 0")
		assert.True(t, got.PrivilegeLimited, "the boolean survives so the interface can explain why")
	})

	// SC-010: storage stays constant per monitor no matter how long it runs.
	t.Run("UpsertReplacesRatherThanAccumulates", func(t *testing.T) {
		id := "rh-upsert"
		seedMonitor(t, resources, id)

		for i := 1; i <= 25; i++ {
			require.NoError(t, repo.Upsert(ctx, &domain.ResourceHealth{
				ResourceID:        id,
				CollectedAt:       time.Now().UTC(),
				ConnectionsActive: i64(int64(i)),
				ConnectionsMax:    i64(200),
			}))
		}

		got, err := repo.FindByResourceID(ctx, id)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.NotNil(t, got.ConnectionsActive)
		assert.Equal(t, int64(25), *got.ConnectionsActive,
			"the latest write wins; 25 checks leave one row, not 25")
	})

	t.Run("DeleteClearsTheRecord", func(t *testing.T) {
		id := "rh-delete"
		seedMonitor(t, resources, id)
		require.NoError(t, repo.Upsert(ctx, &domain.ResourceHealth{
			ResourceID:        id,
			CollectedAt:       time.Now().UTC(),
			ConnectionsActive: i64(7),
		}))

		require.NoError(t, repo.DeleteByResourceID(ctx, id))

		got, err := repo.FindByResourceID(ctx, id)
		require.NoError(t, err)
		assert.Nil(t, got, "a check that collects nothing must leave no stale figures on display")
	})

	t.Run("DeleteIsIdempotent", func(t *testing.T) {
		require.NoError(t, repo.DeleteByResourceID(ctx, "rh-not-there"),
			"deleting an absent record is not an error: it runs after every empty check")
	})

	t.Run("RecordsAreScopedToTheirMonitor", func(t *testing.T) {
		seedMonitor(t, resources, "rh-mine")
		seedMonitor(t, resources, "rh-theirs")
		require.NoError(t, repo.Upsert(ctx, &domain.ResourceHealth{
			ResourceID: "rh-mine", CollectedAt: time.Now().UTC(), ConnectionsActive: i64(5),
		}))
		require.NoError(t, repo.Upsert(ctx, &domain.ResourceHealth{
			ResourceID: "rh-theirs", CollectedAt: time.Now().UTC(), ConnectionsActive: i64(99),
		}))

		got, err := repo.FindByResourceID(ctx, "rh-mine")
		require.NoError(t, err)
		require.NotNil(t, got)
		require.NotNil(t, got.ConnectionsActive)
		assert.Equal(t, int64(5), *got.ConnectionsActive)
	})
}
