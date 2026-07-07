package pgsqlc

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WithTx runs fn inside a pgx transaction (default TxOptions; READ COMMITTED).
//
// Behavior:
//   - BeginTx with default options.
//   - fn returns nil → Commit; the Commit error (if any) is returned.
//   - fn returns err → Rollback; fn's error is returned (Rollback errors are
//     intentionally suppressed unless ctx is canceled).
//   - fn panics → Rollback, then re-panic.
//
// fn receives a *Queries bound to the transaction via New(tx).
func WithTx(ctx context.Context, pool *pgxpool.Pool, fn func(*Queries) error) (err error) {
	if pool == nil {
		return fmt.Errorf("pgsqlc.WithTx: nil pool")
	}
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("pgsqlc.WithTx: begin: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()
	if err = fn(New(tx)); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("pgsqlc.WithTx: commit: %w", err)
	}
	return nil
}
