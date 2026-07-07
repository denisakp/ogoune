package store

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/database"
	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	pgsqlc "github.com/denisakp/ogoune/internal/repository/sqlc/pg"
)

// TestTagsRepository_QuerierInjection_Postgres demonstrates the FR-010
// template: createWithQ accepts a *pgsqlc.Queries bound to a transaction,
// the row participates in the tx, and aborting the tx rolls the insert back.
//
// Requires a Postgres backend (testcontainers or POSTGRES_TEST_DSN).
func TestTagsRepository_QuerierInjection_Postgres(t *testing.T) {
	fx := internaltest.SetupPostgres(t)
	if fx == nil {
		return // Skipped by SetupPostgres when no backend available.
	}
	require.Equal(t, database.DriverPostgres, fx.Runtime.Driver)

	ctx := context.Background()
	repo := NewTagsRepositorySQLC(fx.Runtime).(*TagsRepositorySQLC)
	pool := fx.Runtime.PgxPool()

	id := "01TAGQI00000000000000001A"
	tag := &domain.Tags{
		Base: domain.Base{ID: id, CreatedAt: time.Now()},
		Name: "Querier-Injected",
	}

	abortErr := errors.New("test abort")
	err := pgsqlc.WithTx(ctx, pool, func(q *pgsqlc.Queries) error {
		if err := repo.createWithQ(ctx, q, tag); err != nil {
			return err
		}
		// Confirm the row IS visible inside the tx — same Queries handle.
		row, err := q.FindTagByID(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, "Querier-Injected", row.Name)
		return abortErr // force rollback
	})
	require.ErrorIs(t, err, abortErr)

	// After rollback, the row must be absent when read through the pool.
	_, err = pgsqlc.New(pool).FindTagByID(ctx, id)
	assert.ErrorIs(t, err, pgx.ErrNoRows)
}
