package sqlitesqlc_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/repository/internaltest"
	sqlitesqlc "github.com/denisakp/ogoune/internal/repository/sqlc/sqlite"
)

func sqliteHandle(t *testing.T) *sql.DB {
	t.Helper()
	fx := internaltest.SetupSQLite(t)
	sqlDB := fx.Runtime.SQLiteDB()
	require.NotNil(t, sqlDB)
	return sqlDB
}

func TestWithTx_CommitOnSuccess(t *testing.T) {
	db := sqliteHandle(t)
	ctx := context.Background()
	now := time.Now()

	err := sqlitesqlc.WithTx(ctx, db, func(q *sqlitesqlc.Queries) error {
		_, err := q.CreateTag(ctx, sqlitesqlc.CreateTagParams{
			ID:        "01TXCOMMIT0000000000000001",
			CreatedAt: now,
			UpdatedAt: now,
			Name:      "commit-on-success",
		})
		return err
	})
	require.NoError(t, err)

	row, err := sqlitesqlc.New(db).FindTagByID(ctx, "01TXCOMMIT0000000000000001")
	require.NoError(t, err)
	assert.Equal(t, "commit-on-success", row.Name)
}

func TestWithTx_RollbackOnError(t *testing.T) {
	db := sqliteHandle(t)
	ctx := context.Background()
	now := time.Now()

	sentinel := errors.New("boom")
	err := sqlitesqlc.WithTx(ctx, db, func(q *sqlitesqlc.Queries) error {
		if _, err := q.CreateTag(ctx, sqlitesqlc.CreateTagParams{
			ID:        "01TXROLL0000000000000001",
			CreatedAt: now,
			UpdatedAt: now,
			Name:      "to-be-rolled-back",
		}); err != nil {
			return err
		}
		return sentinel
	})
	require.ErrorIs(t, err, sentinel)

	_, err = sqlitesqlc.New(db).FindTagByID(ctx, "01TXROLL0000000000000001")
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestWithTx_RollbackOnPanic(t *testing.T) {
	db := sqliteHandle(t)
	ctx := context.Background()
	now := time.Now()

	defer func() {
		r := recover()
		require.NotNil(t, r, "expected panic to propagate")
		_, err := sqlitesqlc.New(db).FindTagByID(ctx, "01TXPANIC000000000000001")
		assert.ErrorIs(t, err, sql.ErrNoRows)
	}()

	_ = sqlitesqlc.WithTx(ctx, db, func(q *sqlitesqlc.Queries) error {
		_, err := q.CreateTag(ctx, sqlitesqlc.CreateTagParams{
			ID:        "01TXPANIC000000000000001",
			CreatedAt: now,
			UpdatedAt: now,
			Name:      "to-panic",
		})
		require.NoError(t, err)
		panic("boom")
	})
}

func TestWithTx_ConcurrentCallers(t *testing.T) {
	db := sqliteHandle(t)
	ctx := context.Background()
	now := time.Now()

	const n = 8
	var wg sync.WaitGroup
	wg.Add(n)
	errs := make(chan error, n)
	for i := 0; i < n; i++ {
		i := i
		go func() {
			defer wg.Done()
			err := sqlitesqlc.WithTx(ctx, db, func(q *sqlitesqlc.Queries) error {
				_, err := q.CreateTag(ctx, sqlitesqlc.CreateTagParams{
					ID:        fmt.Sprintf("01TXCONC%018d", i),
					CreatedAt: now,
					UpdatedAt: now,
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
		_, err := sqlitesqlc.New(db).FindTagByID(ctx, fmt.Sprintf("01TXCONC%018d", i))
		require.NoError(t, err)
	}
}

func TestWithTx_NilDB(t *testing.T) {
	err := sqlitesqlc.WithTx(context.Background(), nil, func(q *sqlitesqlc.Queries) error { return nil })
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil db")
}
