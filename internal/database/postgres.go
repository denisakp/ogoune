package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

// openPostgres opens a single physical pgxpool, then wraps it with
// stdlib.OpenDBFromPool to obtain a *sql.DB used by the migrator and the
// startup schema validator. The pgxpool is the production handle for sqlc
// queries (via the generated pgsqlc.New(pool) constructor in repos).
func openPostgres(ctx context.Context, cfg resolvedConfig) (*pgxpool.Pool, *sql.DB, error) {
	slog.Info("opening database connection", "driver", string(cfg.Driver))

	poolCfg, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, nil, fmt.Errorf("db init: failed to parse postgres dsn: %w", err)
	}
	poolCfg.MaxConns = 25
	poolCfg.MaxConnLifetime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, nil, fmt.Errorf("db init: failed to open postgres pool: %w", err)
	}

	sqlDB := stdlib.OpenDBFromPool(pool)

	return pool, sqlDB, nil
}
