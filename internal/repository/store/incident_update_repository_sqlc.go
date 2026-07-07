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

type IncidentUpdateRepositorySQLC struct {
	pgQ     *pgsqlc.Queries
	sqliteQ *sqlitesqlc.Queries
}

func NewIncidentUpdateRepositorySQLC(rt SqlcRuntime) port.IncidentUpdateRepository {
	r := &IncidentUpdateRepositorySQLC{}
	if pool := rt.PgxPool(); pool != nil {
		r.pgQ = pgsqlc.New(pool)
	} else if db := rt.SQLiteDB(); db != nil {
		r.sqliteQ = sqlitesqlc.New(db)
	}
	return r
}

func (r *IncidentUpdateRepositorySQLC) unconfigured() error {
	return fmt.Errorf("incident_updates_sqlc: unconfigured runtime")
}

func (r *IncidentUpdateRepositorySQLC) Create(ctx context.Context, u *domain.IncidentUpdate) (*domain.IncidentUpdate, error) {
	now := time.Now()
	u.EnsureID()
	if u.PostedAt.IsZero() {
		u.PostedAt = now
	}
	if u.CreatedAt.IsZero() {
		u.CreatedAt = now
	}
	u.UpdatedAt = now
	switch {
	case r.pgQ != nil:
		if err := r.pgQ.CreateIncidentUpdate(ctx, pgsqlc.CreateIncidentUpdateParams{
			ID:         u.ID,
			IncidentID: u.IncidentID,
			Status:     string(u.Status),
			Message:    u.Message,
			PostedBy:   u.PostedBy,
			PostedAt:   pgtype.Timestamptz{Time: u.PostedAt, Valid: true},
			CreatedAt:  pgtype.Timestamptz{Time: u.CreatedAt, Valid: true},
			UpdatedAt:  pgtype.Timestamptz{Time: u.UpdatedAt, Valid: true},
		}); err != nil {
			return nil, fmt.Errorf("sqlc: create incident_update: %w", err)
		}
		return u, nil
	case r.sqliteQ != nil:
		if err := r.sqliteQ.CreateIncidentUpdate(ctx, sqlitesqlc.CreateIncidentUpdateParams{
			ID:         u.ID,
			IncidentID: u.IncidentID,
			Status:     string(u.Status),
			Message:    u.Message,
			PostedBy:   u.PostedBy,
			PostedAt:   u.PostedAt,
			CreatedAt:  u.CreatedAt,
			UpdatedAt:  u.UpdatedAt,
		}); err != nil {
			return nil, fmt.Errorf("sqlc: create incident_update: %w", err)
		}
		return u, nil
	default:
		return nil, r.unconfigured()
	}
}

func (r *IncidentUpdateRepositorySQLC) FindByID(ctx context.Context, id string) (*domain.IncidentUpdate, error) {
	switch {
	case r.pgQ != nil:
		row, err := r.pgQ.GetIncidentUpdate(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, repository.ErrNotFound
			}
			return nil, fmt.Errorf("sqlc: get incident_update: %w", err)
		}
		return incidentUpdateFromPG(row), nil
	case r.sqliteQ != nil:
		row, err := r.sqliteQ.GetIncidentUpdate(ctx, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, repository.ErrNotFound
			}
			return nil, fmt.Errorf("sqlc: get incident_update: %w", err)
		}
		return incidentUpdateFromSQLite(row), nil
	default:
		return nil, r.unconfigured()
	}
}

func (r *IncidentUpdateRepositorySQLC) ListByIncident(ctx context.Context, incidentID string) ([]*domain.IncidentUpdate, error) {
	switch {
	case r.pgQ != nil:
		rows, err := r.pgQ.ListIncidentUpdates(ctx, incidentID)
		if err != nil {
			return nil, fmt.Errorf("sqlc: list incident_updates: %w", err)
		}
		out := make([]*domain.IncidentUpdate, 0, len(rows))
		for _, row := range rows {
			out = append(out, incidentUpdateFromPG(row))
		}
		return out, nil
	case r.sqliteQ != nil:
		rows, err := r.sqliteQ.ListIncidentUpdates(ctx, incidentID)
		if err != nil {
			return nil, fmt.Errorf("sqlc: list incident_updates: %w", err)
		}
		out := make([]*domain.IncidentUpdate, 0, len(rows))
		for _, row := range rows {
			out = append(out, incidentUpdateFromSQLite(row))
		}
		return out, nil
	default:
		return nil, r.unconfigured()
	}
}

func (r *IncidentUpdateRepositorySQLC) Update(ctx context.Context, u *domain.IncidentUpdate) error {
	u.UpdatedAt = time.Now()
	switch {
	case r.pgQ != nil:
		return r.pgQ.UpdateIncidentUpdate(ctx, pgsqlc.UpdateIncidentUpdateParams{
			ID:        u.ID,
			Status:    string(u.Status),
			Message:   u.Message,
			PostedAt:  pgtype.Timestamptz{Time: u.PostedAt, Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: u.UpdatedAt, Valid: true},
		})
	case r.sqliteQ != nil:
		return r.sqliteQ.UpdateIncidentUpdate(ctx, sqlitesqlc.UpdateIncidentUpdateParams{
			ID:        u.ID,
			Status:    string(u.Status),
			Message:   u.Message,
			PostedAt:  u.PostedAt,
			UpdatedAt: u.UpdatedAt,
		})
	default:
		return r.unconfigured()
	}
}

func (r *IncidentUpdateRepositorySQLC) Delete(ctx context.Context, id string) error {
	switch {
	case r.pgQ != nil:
		return r.pgQ.DeleteIncidentUpdate(ctx, id)
	case r.sqliteQ != nil:
		return r.sqliteQ.DeleteIncidentUpdate(ctx, id)
	default:
		return r.unconfigured()
	}
}

func incidentUpdateFromPG(row pgsqlc.IncidentUpdate) *domain.IncidentUpdate {
	return &domain.IncidentUpdate{
		Base: domain.Base{
			ID:        row.ID,
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
		},
		IncidentID: row.IncidentID,
		Status:     domain.IncidentUpdateStatus(row.Status),
		Message:    row.Message,
		PostedBy:   row.PostedBy,
		PostedAt:   row.PostedAt.Time,
	}
}

func incidentUpdateFromSQLite(row sqlitesqlc.IncidentUpdate) *domain.IncidentUpdate {
	return &domain.IncidentUpdate{
		Base: domain.Base{
			ID:        row.ID,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		IncidentID: row.IncidentID,
		Status:     domain.IncidentUpdateStatus(row.Status),
		Message:    row.Message,
		PostedBy:   row.PostedBy,
		PostedAt:   row.PostedAt,
	}
}
