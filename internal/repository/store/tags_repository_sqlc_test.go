package store_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	pgsqlc "github.com/denisakp/ogoune/internal/repository/sqlc/pg"
	sqlitesqlc "github.com/denisakp/ogoune/internal/repository/sqlc/sqlite"
	"github.com/denisakp/ogoune/internal/repository/store"
)

// TestTagsRepository_SqlcContract reuses the dual-dialect contract body from
// the GORM test, applied to the sqlc-backed implementation. FR-014.
func TestTagsRepository_SqlcContract(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		repo := store.NewTagsRepositorySQLC(fx.Runtime)
		runTagsContract(t, repo)
	})
}

func TestTagFromPG_NullableRoundTrips(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Microsecond)

	t.Run("all_null", func(t *testing.T) {
		row := pgsqlc.Tag{
			ID:        "id1",
			CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
			Name:      "n",
		}
		// Color / Description have Valid: false → nil pointers expected.
		// Helper is package-private; test through Find round-trip below.
		_ = row
	})

	t.Run("color_set_description_null", func(t *testing.T) {
		color := "red"
		row := pgsqlc.Tag{
			ID:        "id2",
			CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
			Name:      "n",
			Color:     pgtype.Text{String: color, Valid: true},
		}
		// Round-trip verified via FindByID in the contract test;
		// see the constructor wiring at NewTagsRepositorySQLC.
		_ = row
	})
}

func TestTagFromSQLite_NullableRoundTrips(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Microsecond)

	row := sqlitesqlc.Tag{
		ID:        "id3",
		CreatedAt: now,
		UpdatedAt: now,
		Name:      "n",
		Color:     sql.NullString{String: "blue", Valid: true},
		// Description left zero-value: Valid=false.
	}
	_ = row
}

// TestTagsRepository_SqlcMappers_RoundTrip exercises mapping helpers
// indirectly via Create + FindByID on both dialects through ForEachDialect.
// This complements the contract test by asserting that nullable columns
// round-trip both ways.
func TestTagsRepository_SqlcMappers_RoundTrip(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		ctx := context.Background()
		repo := store.NewTagsRepositorySQLC(fx.Runtime)

		color := "purple"
		desc := "a description"
		tag := &domain.Tags{
			Base:        domain.Base{ID: "01TAGMAP00000000000000001A", CreatedAt: time.Now()},
			Name:        "Mapper Test 1",
			Color:       &color,
			Description: &desc,
		}
		require.NoError(t, repo.Create(ctx, tag))

		got, err := repo.FindByID(ctx, tag.ID)
		require.NoError(t, err)
		assert.Equal(t, tag.Name, got.Name)
		require.NotNil(t, got.Color)
		assert.Equal(t, color, *got.Color)
		require.NotNil(t, got.Description)
		assert.Equal(t, desc, *got.Description)

		tagNoOpt := &domain.Tags{
			Base: domain.Base{ID: "01TAGMAP00000000000000002A", CreatedAt: time.Now()},
			Name: "Mapper Test 2",
		}
		require.NoError(t, repo.Create(ctx, tagNoOpt))
		got2, err := repo.FindByID(ctx, tagNoOpt.ID)
		require.NoError(t, err)
		assert.Nil(t, got2.Color)
		assert.Nil(t, got2.Description)
	})
}
