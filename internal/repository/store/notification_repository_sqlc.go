package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository"
	pgsqlc "github.com/denisakp/ogoune/internal/repository/sqlc/pg"
	sqlitesqlc "github.com/denisakp/ogoune/internal/repository/sqlc/sqlite"
)

type NotificationRepositorySQLC struct {
	pgQ      *pgsqlc.Queries
	sqliteQ  *sqlitesqlc.Queries
	pgPool   *pgxpool.Pool
	sqliteDB *sql.DB
}

func NewNotificationRepositorySQLC(rt SqlcRuntime) port.NotificationRepository {
	r := &NotificationRepositorySQLC{}
	if pool := rt.PgxPool(); pool != nil {
		r.pgPool = pool
		r.pgQ = pgsqlc.New(pool)
	} else if db := rt.SQLiteDB(); db != nil {
		r.sqliteDB = db
		r.sqliteQ = sqlitesqlc.New(db)
	}
	return r
}

func (r *NotificationRepositorySQLC) unconfigured() error {
	return fmt.Errorf("notification_repository_sqlc: unconfigured runtime")
}

func (r *NotificationRepositorySQLC) Create(ctx context.Context, n *domain.NotificationEvent) error {
	n.EnsureID()
	now := time.Now()
	if n.CreatedAt.IsZero() {
		n.CreatedAt = now
	}
	if n.UpdatedAt.IsZero() {
		n.UpdatedAt = now
	}
	if n.Status == "" {
		n.Status = domain.NotificationEventStatusPending
	}
	switch {
	case r.pgQ != nil:
		return r.pgQ.CreateNotificationEvent(ctx, pgsqlc.CreateNotificationEventParams{
			ID:          n.ID,
			CreatedAt:   pgtype.Timestamptz{Time: n.CreatedAt, Valid: true},
			UpdatedAt:   pgtype.Timestamptz{Time: n.UpdatedAt, Valid: true},
			IncidentID:  n.IncidentID,
			Type:        string(n.Type),
			Status:      string(n.Status),
			ClaimOwner:  pgTextFromPtr(n.ClaimOwner),
			ClaimedAt:   pgTimestampFromPtr(n.ClaimedAt),
			ProcessedAt: pgTimestampFromPtr(n.ProcessedAt),
			LastError:   n.LastError,
		})
	case r.sqliteQ != nil:
		return r.sqliteQ.CreateNotificationEvent(ctx, sqlitesqlc.CreateNotificationEventParams{
			ID:          n.ID,
			CreatedAt:   n.CreatedAt,
			UpdatedAt:   n.UpdatedAt,
			IncidentID:  n.IncidentID,
			Type:        string(n.Type),
			Status:      string(n.Status),
			ClaimOwner:  nullStringFromPtr(n.ClaimOwner),
			ClaimedAt:   nullTimeFromPtr(n.ClaimedAt),
			ProcessedAt: nullTimeFromPtr(n.ProcessedAt),
			LastError:   n.LastError,
		})
	default:
		return r.unconfigured()
	}
}

func (r *NotificationRepositorySQLC) FindByID(ctx context.Context, id string) (*domain.NotificationEvent, error) {
	switch {
	case r.pgQ != nil:
		row, err := r.pgQ.FindNotificationEventByID(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, repository.ErrNotFound
			}
			return nil, fmt.Errorf("sqlc: find notification: %w", err)
		}
		return notificationFromPG(row), nil
	case r.sqliteQ != nil:
		row, err := r.sqliteQ.FindNotificationEventByID(ctx, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, repository.ErrNotFound
			}
			return nil, fmt.Errorf("sqlc: find notification: %w", err)
		}
		return notificationFromSQLite(row), nil
	default:
		return nil, r.unconfigured()
	}
}

func (r *NotificationRepositorySQLC) List(ctx context.Context, limit, offset int) ([]*domain.NotificationEvent, error) {
	switch {
	case r.pgQ != nil:
		rows, err := r.pgQ.ListNotificationEvents(ctx, pgsqlc.ListNotificationEventsParams{
			Limit: int32(limit), Offset: int32(offset),
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: list notifications: %w", err)
		}
		out := make([]*domain.NotificationEvent, len(rows))
		for i, row := range rows {
			out[i] = notificationFromPG(row)
		}
		return out, nil
	case r.sqliteQ != nil:
		rows, err := r.sqliteQ.ListNotificationEvents(ctx, sqlitesqlc.ListNotificationEventsParams{
			Limit: int64(limit), Offset: int64(offset),
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: list notifications: %w", err)
		}
		out := make([]*domain.NotificationEvent, len(rows))
		for i, row := range rows {
			out[i] = notificationFromSQLite(row)
		}
		return out, nil
	default:
		return nil, r.unconfigured()
	}
}

func (r *NotificationRepositorySQLC) Update(ctx context.Context, n *domain.NotificationEvent) error {
	n.UpdatedAt = time.Now()
	switch {
	case r.pgQ != nil:
		_, err := r.pgQ.UpdateNotificationEvent(ctx, pgsqlc.UpdateNotificationEventParams{
			ID:          n.ID,
			IncidentID:  n.IncidentID,
			Type:        string(n.Type),
			Status:      string(n.Status),
			ClaimOwner:  pgTextFromPtr(n.ClaimOwner),
			ClaimedAt:   pgTimestampFromPtr(n.ClaimedAt),
			ProcessedAt: pgTimestampFromPtr(n.ProcessedAt),
			LastError:   n.LastError,
			UpdatedAt:   pgtype.Timestamptz{Time: n.UpdatedAt, Valid: true},
		})
		return err
	case r.sqliteQ != nil:
		_, err := r.sqliteQ.UpdateNotificationEvent(ctx, sqlitesqlc.UpdateNotificationEventParams{
			ID:          n.ID,
			IncidentID:  n.IncidentID,
			Type:        string(n.Type),
			Status:      string(n.Status),
			ClaimOwner:  nullStringFromPtr(n.ClaimOwner),
			ClaimedAt:   nullTimeFromPtr(n.ClaimedAt),
			ProcessedAt: nullTimeFromPtr(n.ProcessedAt),
			LastError:   n.LastError,
			UpdatedAt:   n.UpdatedAt,
		})
		return err
	default:
		return r.unconfigured()
	}
}

func (r *NotificationRepositorySQLC) Delete(ctx context.Context, id string) error {
	if id == "" {
		return repository.ErrInvalidInput
	}
	switch {
	case r.pgQ != nil:
		return r.pgQ.DeleteNotificationEvent(ctx, id)
	case r.sqliteQ != nil:
		return r.sqliteQ.DeleteNotificationEvent(ctx, id)
	default:
		return r.unconfigured()
	}
}

func (r *NotificationRepositorySQLC) FindPending(ctx context.Context, limit, offset int) ([]*domain.NotificationEvent, error) {
	switch {
	case r.pgQ != nil:
		rows, err := r.pgQ.FindPendingNotificationEvents(ctx, pgsqlc.FindPendingNotificationEventsParams{
			Limit: int32(limit), Offset: int32(offset),
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: find pending: %w", err)
		}
		out := make([]*domain.NotificationEvent, len(rows))
		for i, row := range rows {
			out[i] = notificationFromPG(row)
		}
		return out, nil
	case r.sqliteQ != nil:
		rows, err := r.sqliteQ.FindPendingNotificationEvents(ctx, sqlitesqlc.FindPendingNotificationEventsParams{
			Limit: int64(limit), Offset: int64(offset),
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: find pending: %w", err)
		}
		out := make([]*domain.NotificationEvent, len(rows))
		for i, row := range rows {
			out[i] = notificationFromSQLite(row)
		}
		return out, nil
	default:
		return nil, r.unconfigured()
	}
}

// ClaimPending implements dialect-divergent atomic claim per spec FR-011/FR-012.
//   - PG:    SELECT … FOR UPDATE SKIP LOCKED + UPDATE inside pg.WithTx.
//   - SQLite: single conditional UPDATE with WHERE-guard (single-writer lock serializes).
func (r *NotificationRepositorySQLC) ClaimPending(ctx context.Context, id, claimOwner string, claimedAt time.Time) (bool, error) {
	if id == "" || claimOwner == "" {
		return false, repository.ErrInvalidInput
	}
	switch {
	case r.pgPool != nil:
		var claimed bool
		err := pgsqlc.WithTx(ctx, r.pgPool, func(q *pgsqlc.Queries) error {
			_, err := q.ClaimNotificationEventForUpdate(ctx, id)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return nil // already claimed or not pending; not an error
				}
				return err
			}
			if err := q.UpdateNotificationEventClaim(ctx, pgsqlc.UpdateNotificationEventClaimParams{
				ID:         id,
				ClaimOwner: pgtype.Text{String: claimOwner, Valid: true},
				ClaimedAt:  pgtype.Timestamptz{Time: claimedAt, Valid: true},
			}); err != nil {
				return err
			}
			claimed = true
			return nil
		})
		if err != nil {
			return false, fmt.Errorf("sqlc: claim pending notification: %w", err)
		}
		return claimed, nil
	case r.sqliteQ != nil:
		n, err := r.sqliteQ.ClaimNotificationEvent(ctx, sqlitesqlc.ClaimNotificationEventParams{
			ID:         id,
			ClaimOwner: sql.NullString{String: claimOwner, Valid: true},
			ClaimedAt:  sql.NullTime{Time: claimedAt, Valid: true},
		})
		if err != nil {
			return false, fmt.Errorf("sqlc: claim pending notification: %w", err)
		}
		return n == 1, nil
	default:
		return false, r.unconfigured()
	}
}

func (r *NotificationRepositorySQLC) MarkAsSent(ctx context.Context, id string, processedAt time.Time) error {
	return r.markTerminal(ctx, id, domain.NotificationEventStatusSent, "", processedAt)
}

func (r *NotificationRepositorySQLC) MarkAsFailed(ctx context.Context, id, lastError string, processedAt time.Time) error {
	return r.markTerminal(ctx, id, domain.NotificationEventStatusFailed, lastError, processedAt)
}

func (r *NotificationRepositorySQLC) MarkAsExpired(ctx context.Context, id, lastError string, processedAt time.Time) error {
	return r.markTerminal(ctx, id, domain.NotificationEventStatusExpired, lastError, processedAt)
}

func (r *NotificationRepositorySQLC) markTerminal(ctx context.Context, id string, status domain.NotificationEventStatusType, lastError string, processedAt time.Time) error {
	if id == "" {
		return repository.ErrInvalidInput
	}
	switch {
	case r.pgQ != nil:
		n, err := r.pgQ.MarkNotificationEventTerminal(ctx, pgsqlc.MarkNotificationEventTerminalParams{
			ID:          id,
			Status:      string(status),
			ProcessedAt: pgtype.Timestamptz{Time: processedAt, Valid: true},
			LastError:   lastError,
		})
		if err != nil {
			return fmt.Errorf("sqlc: mark terminal: %w", err)
		}
		if n == 0 {
			return repository.ErrNotFound
		}
		return nil
	case r.sqliteQ != nil:
		n, err := r.sqliteQ.MarkNotificationEventTerminal(ctx, sqlitesqlc.MarkNotificationEventTerminalParams{
			ID:          id,
			Status:      string(status),
			ProcessedAt: sql.NullTime{Time: processedAt, Valid: true},
			LastError:   lastError,
		})
		if err != nil {
			return fmt.Errorf("sqlc: mark terminal: %w", err)
		}
		if n == 0 {
			return repository.ErrNotFound
		}
		return nil
	default:
		return r.unconfigured()
	}
}

// ---------- mapping helpers ----------

func notificationFromPG(row pgsqlc.NotificationEvent) *domain.NotificationEvent {
	out := &domain.NotificationEvent{
		Base: domain.Base{
			ID:        row.ID,
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
		},
		IncidentID: row.IncidentID,
		Type:       domain.NotificationEventType(row.Type),
		Status:     domain.NotificationEventStatusType(row.Status),
		LastError:  row.LastError,
	}
	if row.ClaimOwner.Valid {
		s := row.ClaimOwner.String
		out.ClaimOwner = &s
	}
	if row.ClaimedAt.Valid {
		t := row.ClaimedAt.Time
		out.ClaimedAt = &t
	}
	if row.ProcessedAt.Valid {
		t := row.ProcessedAt.Time
		out.ProcessedAt = &t
	}
	return out
}

func notificationFromSQLite(row sqlitesqlc.NotificationEvent) *domain.NotificationEvent {
	out := &domain.NotificationEvent{
		Base: domain.Base{
			ID:        row.ID,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		IncidentID: row.IncidentID,
		Type:       domain.NotificationEventType(row.Type),
		Status:     domain.NotificationEventStatusType(row.Status),
		LastError:  row.LastError,
	}
	if row.ClaimOwner.Valid {
		s := row.ClaimOwner.String
		out.ClaimOwner = &s
	}
	if row.ClaimedAt.Valid {
		t := row.ClaimedAt.Time
		out.ClaimedAt = &t
	}
	if row.ProcessedAt.Valid {
		t := row.ProcessedAt.Time
		out.ProcessedAt = &t
	}
	return out
}
