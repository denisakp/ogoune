package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	pgsqlc "github.com/denisakp/ogoune/internal/repository/sqlc/pg"
	sqlitesqlc "github.com/denisakp/ogoune/internal/repository/sqlc/sqlite"
)

// ResourceHealthRepositorySQLC stores the latest database health per monitor
// (spec 088). At most one row per monitor, replaced on every check, so storage is
// constant no matter how long a monitor runs and no retention job is needed.
type ResourceHealthRepositorySQLC struct {
	pgQ     *pgsqlc.Queries
	sqliteQ *sqlitesqlc.Queries
}

func NewResourceHealthRepositorySQLC(rt SqlcRuntime) port.ResourceHealthRepository {
	r := &ResourceHealthRepositorySQLC{}
	if pool := rt.PgxPool(); pool != nil {
		r.pgQ = pgsqlc.New(pool)
	} else if db := rt.SQLiteDB(); db != nil {
		r.sqliteQ = sqlitesqlc.New(db)
	}
	return r
}

func (r *ResourceHealthRepositorySQLC) unconfigured() error {
	return fmt.Errorf("resource_health_repository_sqlc: unconfigured runtime")
}

// Upsert replaces the monitor's record.
func (r *ResourceHealthRepositorySQLC) Upsert(ctx context.Context, h *domain.ResourceHealth) error {
	if h == nil || h.ResourceID == "" {
		return fmt.Errorf("resource_health: missing resource id")
	}
	if h.CollectedAt.IsZero() {
		h.CollectedAt = time.Now().UTC()
	}

	switch {
	case r.pgQ != nil:
		return r.pgQ.UpsertResourceHealth(ctx, pgsqlc.UpsertResourceHealthParams{
			ResourceID:            h.ResourceID,
			CollectedAt:           pgtype.Timestamptz{Time: h.CollectedAt, Valid: true},
			ConnectionsActive:     pgInt8FromPtr(h.ConnectionsActive),
			ConnectionsMax:        pgInt8FromPtr(h.ConnectionsMax),
			LongestQuerySeconds:   pgFloat8FromPtr(h.LongestQuerySeconds),
			ReplicationLagSeconds: pgFloat8FromPtr(h.ReplicationLagSeconds),
			PrivilegeLimited:      h.PrivilegeLimited,
			UnsupportedVersion:    h.UnsupportedVersion,
		})
	case r.sqliteQ != nil:
		return r.sqliteQ.UpsertResourceHealth(ctx, sqlitesqlc.UpsertResourceHealthParams{
			ResourceID:            h.ResourceID,
			CollectedAt:           h.CollectedAt,
			ConnectionsActive:     nullInt64FromPtr(h.ConnectionsActive),
			ConnectionsMax:        nullInt64FromPtr(h.ConnectionsMax),
			LongestQuerySeconds:   nullFloatFromPtr(h.LongestQuerySeconds),
			ReplicationLagSeconds: nullFloatFromPtr(h.ReplicationLagSeconds),
			PrivilegeLimited:      boolToInt64(h.PrivilegeLimited),
			UnsupportedVersion:    boolToInt64(h.UnsupportedVersion),
		})
	default:
		return r.unconfigured()
	}
}

// FindByResourceID returns nil, nil when the monitor has no record. Absence is
// expressed by absence: a zero-filled struct would be indistinguishable from a
// database sitting at zero connections.
func (r *ResourceHealthRepositorySQLC) FindByResourceID(ctx context.Context, resourceID string) (*domain.ResourceHealth, error) {
	switch {
	case r.pgQ != nil:
		row, err := r.pgQ.FindResourceHealth(ctx, resourceID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, nil
			}
			return nil, fmt.Errorf("sqlc: find resource health: %w", err)
		}
		return &domain.ResourceHealth{
			ResourceID:            row.ResourceID,
			CollectedAt:           row.CollectedAt.Time,
			ConnectionsActive:     ptrInt64FromPGInt8(row.ConnectionsActive),
			ConnectionsMax:        ptrInt64FromPGInt8(row.ConnectionsMax),
			LongestQuerySeconds:   ptrFloatFromPGFloat8(row.LongestQuerySeconds),
			ReplicationLagSeconds: ptrFloatFromPGFloat8(row.ReplicationLagSeconds),
			PrivilegeLimited:      row.PrivilegeLimited,
			UnsupportedVersion:    row.UnsupportedVersion,
		}, nil
	case r.sqliteQ != nil:
		row, err := r.sqliteQ.FindResourceHealth(ctx, resourceID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, nil
			}
			return nil, fmt.Errorf("sqlc: find resource health: %w", err)
		}
		return &domain.ResourceHealth{
			ResourceID:            row.ResourceID,
			CollectedAt:           row.CollectedAt,
			ConnectionsActive:     ptrInt64FromNullInt64(row.ConnectionsActive),
			ConnectionsMax:        ptrInt64FromNullInt64(row.ConnectionsMax),
			LongestQuerySeconds:   ptrFloatFromNullFloat(row.LongestQuerySeconds),
			ReplicationLagSeconds: ptrFloatFromNullFloat(row.ReplicationLagSeconds),
			PrivilegeLimited:      row.PrivilegeLimited != 0,
			UnsupportedVersion:    row.UnsupportedVersion != 0,
		}, nil
	default:
		return nil, r.unconfigured()
	}
}

// DeleteByResourceID clears the record. Called when a check collected nothing, so
// the interface shows no figures rather than yesterday's.
func (r *ResourceHealthRepositorySQLC) DeleteByResourceID(ctx context.Context, resourceID string) error {
	switch {
	case r.pgQ != nil:
		return r.pgQ.DeleteResourceHealth(ctx, resourceID)
	case r.sqliteQ != nil:
		return r.sqliteQ.DeleteResourceHealth(ctx, resourceID)
	default:
		return r.unconfigured()
	}
}
