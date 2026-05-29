package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// openPostgres opens a single physical pgxpool, wraps it as *sql.DB via
// stdlib.OpenDBFromPool, and hands that *sql.DB to GORM so the GORM handle
// and the returned *pgxpool.Pool share one underlying pool.
func openPostgres(ctx context.Context, cfg resolvedConfig) (*gorm.DB, *pgxpool.Pool, error) {
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

	database, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		Logger: newGormLogger(cfg.GormLogLevel),
	})
	if err != nil {
		pool.Close()
		return nil, nil, fmt.Errorf("db init: failed to connect to postgres: %w", err)
	}

	return database, pool, nil
}
