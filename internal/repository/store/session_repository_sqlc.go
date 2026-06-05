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
	"github.com/denisakp/ogoune/internal/repository"
	pgsqlc "github.com/denisakp/ogoune/internal/repository/sqlc/pg"
	sqlitesqlc "github.com/denisakp/ogoune/internal/repository/sqlc/sqlite"
)

type SessionRepositorySQLC struct {
	pgQ     *pgsqlc.Queries
	sqliteQ *sqlitesqlc.Queries
}

func NewSessionRepositorySQLC(rt SqlcRuntime) port.SessionRepository {
	r := &SessionRepositorySQLC{}
	if pool := rt.PgxPool(); pool != nil {
		r.pgQ = pgsqlc.New(pool)
	} else if db := rt.SQLiteDB(); db != nil {
		r.sqliteQ = sqlitesqlc.New(db)
	}
	return r
}

func (r *SessionRepositorySQLC) unconfigured() error {
	return fmt.Errorf("session_repository_sqlc: unconfigured runtime")
}

func (r *SessionRepositorySQLC) Create(ctx context.Context, s *domain.Session) error {
	s.EnsureID()
	now := time.Now()
	if s.CreatedAt.IsZero() {
		s.CreatedAt = now
	}
	if s.LastActiveAt.IsZero() {
		s.LastActiveAt = now
	}
	switch {
	case r.pgQ != nil:
		_, err := r.pgQ.CreateSession(ctx, pgsqlc.CreateSessionParams{
			ID:           s.ID,
			UserID:       s.UserID,
			Browser:      s.Browser,
			Os:           s.OS,
			Ip:           s.IP,
			Location:     pgTextFromPtr(s.Location),
			LastActiveAt: pgtype.Timestamptz{Time: s.LastActiveAt, Valid: true},
			CreatedAt:    pgtype.Timestamptz{Time: s.CreatedAt, Valid: true},
			RevokedAt:    pgTimestampFromPtr(s.RevokedAt),
		})
		if err != nil {
			return fmt.Errorf("sqlc: create session: %w", err)
		}
		return nil
	case r.sqliteQ != nil:
		_, err := r.sqliteQ.CreateSession(ctx, sqlitesqlc.CreateSessionParams{
			ID:           s.ID,
			UserID:       s.UserID,
			Browser:      s.Browser,
			Os:           s.OS,
			Ip:           s.IP,
			Location:     nullStringFromPtr(s.Location),
			LastActiveAt: s.LastActiveAt,
			CreatedAt:    s.CreatedAt,
			RevokedAt:    nullTimeFromPtr(s.RevokedAt),
		})
		if err != nil {
			return fmt.Errorf("sqlc: create session: %w", err)
		}
		return nil
	default:
		return r.unconfigured()
	}
}

func (r *SessionRepositorySQLC) FindByID(ctx context.Context, id string) (*domain.Session, error) {
	switch {
	case r.pgQ != nil:
		row, err := r.pgQ.FindSessionByID(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, repository.ErrNotFound
			}
			return nil, fmt.Errorf("sqlc: find session: %w", err)
		}
		return sessionFromPG(row), nil
	case r.sqliteQ != nil:
		row, err := r.sqliteQ.FindSessionByID(ctx, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, repository.ErrNotFound
			}
			return nil, fmt.Errorf("sqlc: find session: %w", err)
		}
		return sessionFromSQLite(row), nil
	default:
		return nil, r.unconfigured()
	}
}

func (r *SessionRepositorySQLC) ListActiveByUser(ctx context.Context, userID string) ([]*domain.Session, error) {
	switch {
	case r.pgQ != nil:
		rows, err := r.pgQ.ListActiveSessionsByUser(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("sqlc: list sessions: %w", err)
		}
		out := make([]*domain.Session, len(rows))
		for i, row := range rows {
			out[i] = sessionFromPG(row)
		}
		return out, nil
	case r.sqliteQ != nil:
		rows, err := r.sqliteQ.ListActiveSessionsByUser(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("sqlc: list sessions: %w", err)
		}
		out := make([]*domain.Session, len(rows))
		for i, row := range rows {
			out[i] = sessionFromSQLite(row)
		}
		return out, nil
	default:
		return nil, r.unconfigured()
	}
}

func (r *SessionRepositorySQLC) UpdateLastActive(ctx context.Context, id string, at time.Time) error {
	switch {
	case r.pgQ != nil:
		err := r.pgQ.UpdateSessionLastActive(ctx, pgsqlc.UpdateSessionLastActiveParams{
			ID:           id,
			LastActiveAt: pgtype.Timestamptz{Time: at, Valid: true},
		})
		if err != nil {
			return fmt.Errorf("sqlc: update last active: %w", err)
		}
		return nil
	case r.sqliteQ != nil:
		err := r.sqliteQ.UpdateSessionLastActive(ctx, sqlitesqlc.UpdateSessionLastActiveParams{
			ID:           id,
			LastActiveAt: at,
		})
		if err != nil {
			return fmt.Errorf("sqlc: update last active: %w", err)
		}
		return nil
	default:
		return r.unconfigured()
	}
}

func (r *SessionRepositorySQLC) Revoke(ctx context.Context, id string, at time.Time) error {
	switch {
	case r.pgQ != nil:
		n, err := r.pgQ.RevokeSession(ctx, pgsqlc.RevokeSessionParams{
			ID:        id,
			RevokedAt: pgtype.Timestamptz{Time: at, Valid: true},
		})
		if err != nil {
			return fmt.Errorf("sqlc: revoke session: %w", err)
		}
		if n == 0 {
			return repository.ErrNotFound
		}
		return nil
	case r.sqliteQ != nil:
		n, err := r.sqliteQ.RevokeSession(ctx, sqlitesqlc.RevokeSessionParams{
			ID:        id,
			RevokedAt: sql.NullTime{Time: at, Valid: true},
		})
		if err != nil {
			return fmt.Errorf("sqlc: revoke session: %w", err)
		}
		if n == 0 {
			return repository.ErrNotFound
		}
		return nil
	default:
		return r.unconfigured()
	}
}

func (r *SessionRepositorySQLC) RevokeAllExcept(ctx context.Context, userID, currentID string, at time.Time) (int64, error) {
	switch {
	case r.pgQ != nil:
		n, err := r.pgQ.RevokeAllSessionsExcept(ctx, pgsqlc.RevokeAllSessionsExceptParams{
			UserID:    userID,
			ID:        currentID,
			RevokedAt: pgtype.Timestamptz{Time: at, Valid: true},
		})
		if err != nil {
			return 0, fmt.Errorf("sqlc: revoke all except: %w", err)
		}
		return n, nil
	case r.sqliteQ != nil:
		n, err := r.sqliteQ.RevokeAllSessionsExcept(ctx, sqlitesqlc.RevokeAllSessionsExceptParams{
			UserID:    userID,
			ID:        currentID,
			RevokedAt: sql.NullTime{Time: at, Valid: true},
		})
		if err != nil {
			return 0, fmt.Errorf("sqlc: revoke all except: %w", err)
		}
		return n, nil
	default:
		return 0, r.unconfigured()
	}
}

func sessionFromPG(row pgsqlc.Session) *domain.Session {
	out := &domain.Session{
		ID:           row.ID,
		UserID:       row.UserID,
		Browser:      row.Browser,
		OS:           row.Os,
		IP:           row.Ip,
		LastActiveAt: row.LastActiveAt.Time,
		CreatedAt:    row.CreatedAt.Time,
	}
	if row.Location.Valid {
		s := row.Location.String
		out.Location = &s
	}
	if row.RevokedAt.Valid {
		t := row.RevokedAt.Time
		out.RevokedAt = &t
	}
	return out
}

func sessionFromSQLite(row sqlitesqlc.Session) *domain.Session {
	out := &domain.Session{
		ID:           row.ID,
		UserID:       row.UserID,
		Browser:      row.Browser,
		OS:           row.Os,
		IP:           row.Ip,
		LastActiveAt: row.LastActiveAt,
		CreatedAt:    row.CreatedAt,
	}
	if row.Location.Valid {
		s := row.Location.String
		out.Location = &s
	}
	if row.RevokedAt.Valid {
		t := row.RevokedAt.Time
		out.RevokedAt = &t
	}
	return out
}
