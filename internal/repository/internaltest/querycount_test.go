package internaltest

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"

	pgsqlc "github.com/denisakp/ogoune/internal/repository/sqlc/pg"
	sqlitesqlc "github.com/denisakp/ogoune/internal/repository/sqlc/sqlite"
)

// ---- Fake DBTX impls (no real DB) ----

type fakePGDBTX struct {
	execErr  error
	queryErr error
}

func (f *fakePGDBTX) Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, f.execErr
}
func (f *fakePGDBTX) Query(context.Context, string, ...interface{}) (pgx.Rows, error) {
	return nil, f.queryErr
}
func (f *fakePGDBTX) QueryRow(context.Context, string, ...interface{}) pgx.Row {
	return nil
}

type fakeSQLiteDBTX struct {
	execErr  error
	queryErr error
}

func (f *fakeSQLiteDBTX) ExecContext(context.Context, string, ...interface{}) (sql.Result, error) {
	return nil, f.execErr
}
func (f *fakeSQLiteDBTX) PrepareContext(context.Context, string) (*sql.Stmt, error) {
	return nil, nil
}
func (f *fakeSQLiteDBTX) QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error) {
	return nil, f.queryErr
}
func (f *fakeSQLiteDBTX) QueryRowContext(context.Context, string, ...interface{}) *sql.Row {
	return nil
}

// Compile-time interface checks (mirror production verify.go pattern).
var (
	_ pgsqlc.DBTX     = (*PGQueryCounter)(nil)
	_ pgsqlc.DBTX     = (*fakePGDBTX)(nil)
	_ sqlitesqlc.DBTX = (*SQLiteQueryCounter)(nil)
	_ sqlitesqlc.DBTX = (*fakeSQLiteDBTX)(nil)
)

// ---- Tests ----

func TestCounter_IncrementsPerCall(t *testing.T) {
	c := &Counter{}
	pg := &PGQueryCounter{inner: &fakePGDBTX{}, c: c}
	ctx := context.Background()

	_, _ = pg.Exec(ctx, "q1")
	_, _ = pg.Exec(ctx, "q2")
	_, _ = pg.Exec(ctx, "q3")
	_, _ = pg.Query(ctx, "q4")
	_ = pg.QueryRow(ctx, "q5")

	assert.EqualValues(t, 5, c.Snapshot())
}

func TestCounter_ResetZeroesAtomically(t *testing.T) {
	c := &Counter{}
	pg := &PGQueryCounter{inner: &fakePGDBTX{}, c: c}
	ctx := context.Background()

	const goroutines = 10
	const opsPerGoroutine = 100
	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				_, _ = pg.Exec(ctx, "q")
			}
		}()
	}
	wg.Wait()

	assert.EqualValues(t, goroutines*opsPerGoroutine, c.Snapshot())
	c.Reset()
	assert.EqualValues(t, 0, c.Snapshot())
}

func TestPGCounter_ForwardsErrors(t *testing.T) {
	wantExec := errors.New("exec boom")
	wantQuery := errors.New("query boom")
	c := &Counter{}
	pg := &PGQueryCounter{inner: &fakePGDBTX{execErr: wantExec, queryErr: wantQuery}, c: c}
	ctx := context.Background()

	_, err := pg.Exec(ctx, "x")
	assert.ErrorIs(t, err, wantExec)
	_, err = pg.Query(ctx, "x")
	assert.ErrorIs(t, err, wantQuery)
	assert.EqualValues(t, 2, c.Snapshot(), "counter must increment even on error")
}

func TestSQLiteCounter_ForwardsErrors(t *testing.T) {
	wantExec := errors.New("exec boom")
	wantQuery := errors.New("query boom")
	c := &Counter{}
	sq := &SQLiteQueryCounter{inner: &fakeSQLiteDBTX{execErr: wantExec, queryErr: wantQuery}, c: c}
	ctx := context.Background()

	_, err := sq.ExecContext(ctx, "x")
	assert.ErrorIs(t, err, wantExec)
	_, err = sq.QueryContext(ctx, "x")
	assert.ErrorIs(t, err, wantQuery)
	assert.EqualValues(t, 2, c.Snapshot(), "counter must increment even on error")
}

func TestSQLiteCounter_PrepareIncrements(t *testing.T) {
	c := &Counter{}
	sq := &SQLiteQueryCounter{inner: &fakeSQLiteDBTX{}, c: c}
	ctx := context.Background()
	_, _ = sq.PrepareContext(ctx, "x")
	_ = sq.QueryRowContext(ctx, "x")
	assert.EqualValues(t, 2, c.Snapshot())
}
