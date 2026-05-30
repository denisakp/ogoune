package pgsqlc_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/repository/internaltest"
	pgsqlc "github.com/denisakp/ogoune/internal/repository/sqlc/pg"
)

func pgFixture(t *testing.T) *internaltest.DialectFixture {
	t.Helper()
	fx := internaltest.SetupPostgres(t)
	require.NotNil(t, fx)
	return fx
}

func ts(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func TestWithTx_CommitOnSuccess(t *testing.T) {
	fx := pgFixture(t)
	pool := fx.Runtime.PgxPool()
	ctx := context.Background()
	now := time.Now()

	err := pgsqlc.WithTx(ctx, pool, func(q *pgsqlc.Queries) error {
		_, err := q.CreateTag(ctx, pgsqlc.CreateTagParams{
			ID:        "01TXCOMMIT0000000000000001",
			CreatedAt: ts(now),
			UpdatedAt: ts(now),
			Name:      "commit-on-success",
		})
		return err
	})
	require.NoError(t, err)

	row, err := pgsqlc.New(pool).FindTagByID(ctx, "01TXCOMMIT0000000000000001")
	require.NoError(t, err)
	assert.Equal(t, "commit-on-success", row.Name)
}

func TestWithTx_RollbackOnError(t *testing.T) {
	fx := pgFixture(t)
	pool := fx.Runtime.PgxPool()
	ctx := context.Background()
	now := time.Now()

	sentinel := errors.New("boom")
	err := pgsqlc.WithTx(ctx, pool, func(q *pgsqlc.Queries) error {
		if _, err := q.CreateTag(ctx, pgsqlc.CreateTagParams{
			ID:        "01TXROLL0000000000000001",
			CreatedAt: ts(now),
			UpdatedAt: ts(now),
			Name:      "to-be-rolled-back",
		}); err != nil {
			return err
		}
		return sentinel
	})
	require.ErrorIs(t, err, sentinel)

	_, err = pgsqlc.New(pool).FindTagByID(ctx, "01TXROLL0000000000000001")
	assert.ErrorIs(t, err, pgx.ErrNoRows)
}

func TestWithTx_RollbackOnPanic(t *testing.T) {
	fx := pgFixture(t)
	pool := fx.Runtime.PgxPool()
	ctx := context.Background()
	now := time.Now()

	defer func() {
		r := recover()
		require.NotNil(t, r, "expected panic to propagate")
		_, err := pgsqlc.New(pool).FindTagByID(ctx, "01TXPANIC000000000000001")
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	}()

	_ = pgsqlc.WithTx(ctx, pool, func(q *pgsqlc.Queries) error {
		_, err := q.CreateTag(ctx, pgsqlc.CreateTagParams{
			ID:        "01TXPANIC000000000000001",
			CreatedAt: ts(now),
			UpdatedAt: ts(now),
			Name:      "to-panic",
		})
		require.NoError(t, err)
		panic("boom")
	})
}

func TestWithTx_ConcurrentCallers(t *testing.T) {
	fx := pgFixture(t)
	pool := fx.Runtime.PgxPool()
	ctx := context.Background()
	now := time.Now()

	const n = 16
	var wg sync.WaitGroup
	wg.Add(n)
	errs := make(chan error, n)
	for i := 0; i < n; i++ {
		i := i
		go func() {
			defer wg.Done()
			err := pgsqlc.WithTx(ctx, pool, func(q *pgsqlc.Queries) error {
				_, err := q.CreateTag(ctx, pgsqlc.CreateTagParams{
					ID:        fmt.Sprintf("01TXCONC%018d", i),
					CreatedAt: ts(now),
					UpdatedAt: ts(now),
					Name:      fmt.Sprintf("concurrent-%d", i),
				})
				return err
			})
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		require.NoError(t, e)
	}

	for i := 0; i < n; i++ {
		_, err := pgsqlc.New(pool).FindTagByID(ctx, fmt.Sprintf("01TXCONC%018d", i))
		require.NoError(t, err)
	}
}

func TestWithTx_NilPool(t *testing.T) {
	err := pgsqlc.WithTx(context.Background(), nil, func(q *pgsqlc.Queries) error { return nil })
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil pool")
}
