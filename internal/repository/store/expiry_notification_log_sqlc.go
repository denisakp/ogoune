package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	pgsqlc "github.com/denisakp/ogoune/internal/repository/sqlc/pg"
	sqlitesqlc "github.com/denisakp/ogoune/internal/repository/sqlc/sqlite"
)

type ExpiryNotificationLogRepositorySQLC struct {
	pgQ     *pgsqlc.Queries
	sqliteQ *sqlitesqlc.Queries
}

func NewExpiryNotificationLogRepositorySQLC(rt SqlcRuntime) port.ExpiryNotificationLogRepository {
	r := &ExpiryNotificationLogRepositorySQLC{}
	if pool := rt.PgxPool(); pool != nil {
		r.pgQ = pgsqlc.New(pool)
	} else if db := rt.SQLiteDB(); db != nil {
		r.sqliteQ = sqlitesqlc.New(db)
	}
	return r
}

func (r *ExpiryNotificationLogRepositorySQLC) unconfigured() error {
	return fmt.Errorf("expiry_notification_log_sqlc: unconfigured runtime")
}

func (r *ExpiryNotificationLogRepositorySQLC) CountByKey(ctx context.Context, resourceID, expiryType string, threshold int) (int64, error) {
	switch {
	case r.pgQ != nil:
		return r.pgQ.CountExpiryNotificationLogsByKey(ctx, pgsqlc.CountExpiryNotificationLogsByKeyParams{
			ResourceID: resourceID,
			ExpiryType: expiryType,
			Threshold:  int32(threshold),
		})
	case r.sqliteQ != nil:
		return r.sqliteQ.CountExpiryNotificationLogsByKey(ctx, sqlitesqlc.CountExpiryNotificationLogsByKeyParams{
			ResourceID: resourceID,
			ExpiryType: expiryType,
			Threshold:  int64(threshold),
		})
	default:
		return 0, r.unconfigured()
	}
}

func (r *ExpiryNotificationLogRepositorySQLC) Create(ctx context.Context, log *domain.ExpiryNotificationLog) error {
	log.EnsureID()
	now := time.Now()
	if log.CreatedAt.IsZero() {
		log.CreatedAt = now
	}
	if log.UpdatedAt.IsZero() {
		log.UpdatedAt = now
	}
	if log.SentAt.IsZero() {
		log.SentAt = now
	}
	switch {
	case r.pgQ != nil:
		return r.pgQ.CreateExpiryNotificationLog(ctx, pgsqlc.CreateExpiryNotificationLogParams{
			ID:         log.ID,
			ResourceID: log.ResourceID,
			ExpiryType: log.ExpiryType,
			Threshold:  int32(log.Threshold),
			SentAt:     pgtype.Timestamptz{Time: log.SentAt, Valid: true},
			CreatedAt:  pgtype.Timestamptz{Time: log.CreatedAt, Valid: true},
			UpdatedAt:  pgtype.Timestamptz{Time: log.UpdatedAt, Valid: true},
		})
	case r.sqliteQ != nil:
		return r.sqliteQ.CreateExpiryNotificationLog(ctx, sqlitesqlc.CreateExpiryNotificationLogParams{
			ID:         log.ID,
			ResourceID: log.ResourceID,
			ExpiryType: log.ExpiryType,
			Threshold:  int64(log.Threshold),
			SentAt:     log.SentAt,
			CreatedAt:  log.CreatedAt,
			UpdatedAt:  log.UpdatedAt,
		})
	default:
		return r.unconfigured()
	}
}

func (r *ExpiryNotificationLogRepositorySQLC) DeleteByResourceIDAndType(ctx context.Context, resourceID, expiryType string) error {
	switch {
	case r.pgQ != nil:
		return r.pgQ.DeleteExpiryNotificationLogsByResourceIDAndType(ctx, pgsqlc.DeleteExpiryNotificationLogsByResourceIDAndTypeParams{
			ResourceID: resourceID,
			ExpiryType: expiryType,
		})
	case r.sqliteQ != nil:
		return r.sqliteQ.DeleteExpiryNotificationLogsByResourceIDAndType(ctx, sqlitesqlc.DeleteExpiryNotificationLogsByResourceIDAndTypeParams{
			ResourceID: resourceID,
			ExpiryType: expiryType,
		})
	default:
		return r.unconfigured()
	}
}

func (r *ExpiryNotificationLogRepositorySQLC) DeleteOlderThan(ctx context.Context, cutoff time.Time) error {
	switch {
	case r.pgQ != nil:
		return r.pgQ.DeleteExpiryNotificationLogsOlderThan(ctx, pgtype.Timestamptz{Time: cutoff, Valid: true})
	case r.sqliteQ != nil:
		return r.sqliteQ.DeleteExpiryNotificationLogsOlderThan(ctx, cutoff)
	default:
		return r.unconfigured()
	}
}
