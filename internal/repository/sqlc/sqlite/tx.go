package sqlitesqlc

import (
	"context"
	"database/sql"
	"fmt"
)

// WithTx runs fn inside a *sql.Tx (default isolation).
//
// Behavior mirrors the pg.WithTx variant:
//   - fn returns nil → Commit; the Commit error (if any) is returned.
//   - fn returns err → Rollback; fn's error is returned.
//   - fn panics → Rollback, then re-panic.
//
// fn receives a *Queries bound to the transaction via New(tx).
func WithTx(ctx context.Context, db *sql.DB, fn func(*Queries) error) (err error) {
	if db == nil {
		return fmt.Errorf("sqlitesqlc.WithTx: nil db")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("sqlitesqlc.WithTx: begin: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()
	if err = fn(New(tx)); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("sqlitesqlc.WithTx: commit: %w", err)
	}
	return nil
}
