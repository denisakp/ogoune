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

type IncidentEventStepRepositorySQLC struct {
	pgQ     *pgsqlc.Queries
	sqliteQ *sqlitesqlc.Queries
}

func NewIncidentEventStepRepositorySQLC(rt SqlcRuntime) port.IncidentEventStepRepository {
	r := &IncidentEventStepRepositorySQLC{}
	if pool := rt.PgxPool(); pool != nil {
		r.pgQ = pgsqlc.New(pool)
	} else if db := rt.SQLiteDB(); db != nil {
		r.sqliteQ = sqlitesqlc.New(db)
	}
	return r
}

func (r *IncidentEventStepRepositorySQLC) unconfigured() error {
	return fmt.Errorf("incident_event_step_repository_sqlc: unconfigured runtime")
}

func (r *IncidentEventStepRepositorySQLC) Create(ctx context.Context, s *domain.IncidentEventStep) (*domain.IncidentEventStep, error) {
	s.EnsureID()
	now := time.Now()
	if s.CreatedAt.IsZero() {
		s.CreatedAt = now
	}
	if s.UpdatedAt.IsZero() {
		s.UpdatedAt = now
	}
	switch {
	case r.pgQ != nil:
		err := r.pgQ.CreateIncidentEventStep(ctx, pgsqlc.CreateIncidentEventStepParams{
			ID:         s.ID,
			CreatedAt:  pgtype.Timestamptz{Time: s.CreatedAt, Valid: true},
			UpdatedAt:  pgtype.Timestamptz{Time: s.UpdatedAt, Valid: true},
			IncidentID: s.IncidentID,
			Step:       string(s.Step),
			Message:    pgTextFromPtr(s.Message),
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: create event step: %w", err)
		}
		return s, nil
	case r.sqliteQ != nil:
		err := r.sqliteQ.CreateIncidentEventStep(ctx, sqlitesqlc.CreateIncidentEventStepParams{
			ID:         s.ID,
			CreatedAt:  s.CreatedAt,
			UpdatedAt:  s.UpdatedAt,
			IncidentID: s.IncidentID,
			Step:       string(s.Step),
			Message:    nullStringFromPtr(s.Message),
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: create event step: %w", err)
		}
		return s, nil
	default:
		return nil, r.unconfigured()
	}
}

func (r *IncidentEventStepRepositorySQLC) FindByID(ctx context.Context, id string) (*domain.IncidentEventStep, error) {
	switch {
	case r.pgQ != nil:
		row, err := r.pgQ.FindIncidentEventStepByID(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, repository.ErrNotFound
			}
			return nil, fmt.Errorf("sqlc: find event step: %w", err)
		}
		return eventStepFromPGJoin(row), nil
	case r.sqliteQ != nil:
		row, err := r.sqliteQ.FindIncidentEventStepByID(ctx, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, repository.ErrNotFound
			}
			return nil, fmt.Errorf("sqlc: find event step: %w", err)
		}
		return eventStepFromSQLiteJoin(row), nil
	default:
		return nil, r.unconfigured()
	}
}

func (r *IncidentEventStepRepositorySQLC) FindLastByIncidentAndStep(ctx context.Context, incidentID string, stepType domain.IncidentEventStepType) (*domain.IncidentEventStep, error) {
	switch {
	case r.pgQ != nil:
		row, err := r.pgQ.FindLastIncidentEventStepByIncidentAndStep(ctx, pgsqlc.FindLastIncidentEventStepByIncidentAndStepParams{
			IncidentID: incidentID, Step: string(stepType),
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, repository.ErrNotFound
			}
			return nil, fmt.Errorf("sqlc: find last event step: %w", err)
		}
		return eventStepFromPG(row), nil
	case r.sqliteQ != nil:
		row, err := r.sqliteQ.FindLastIncidentEventStepByIncidentAndStep(ctx, sqlitesqlc.FindLastIncidentEventStepByIncidentAndStepParams{
			IncidentID: incidentID, Step: string(stepType),
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, repository.ErrNotFound
			}
			return nil, fmt.Errorf("sqlc: find last event step: %w", err)
		}
		return eventStepFromSQLite(row), nil
	default:
		return nil, r.unconfigured()
	}
}

func (r *IncidentEventStepRepositorySQLC) List(ctx context.Context, limit, offset int) ([]*domain.IncidentEventStep, error) {
	switch {
	case r.pgQ != nil:
		rows, err := r.pgQ.ListIncidentEventSteps(ctx, pgsqlc.ListIncidentEventStepsParams{
			Limit: int32(limit), Offset: int32(offset),
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: list event steps: %w", err)
		}
		out := make([]*domain.IncidentEventStep, len(rows))
		for i, row := range rows {
			out[i] = eventStepFromPG(row)
		}
		return out, nil
	case r.sqliteQ != nil:
		rows, err := r.sqliteQ.ListIncidentEventSteps(ctx, sqlitesqlc.ListIncidentEventStepsParams{
			Limit: int64(limit), Offset: int64(offset),
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: list event steps: %w", err)
		}
		out := make([]*domain.IncidentEventStep, len(rows))
		for i, row := range rows {
			out[i] = eventStepFromSQLite(row)
		}
		return out, nil
	default:
		return nil, r.unconfigured()
	}
}

func (r *IncidentEventStepRepositorySQLC) Update(ctx context.Context, s *domain.IncidentEventStep) error {
	s.UpdatedAt = time.Now()
	switch {
	case r.pgQ != nil:
		n, err := r.pgQ.UpdateIncidentEventStep(ctx, pgsqlc.UpdateIncidentEventStepParams{
			ID:         s.ID,
			IncidentID: s.IncidentID,
			Step:       string(s.Step),
			Message:    pgTextFromPtr(s.Message),
			UpdatedAt:  pgtype.Timestamptz{Time: s.UpdatedAt, Valid: true},
		})
		if err != nil {
			return fmt.Errorf("sqlc: update event step: %w", err)
		}
		if n == 0 {
			return repository.ErrNotFound
		}
		return nil
	case r.sqliteQ != nil:
		n, err := r.sqliteQ.UpdateIncidentEventStep(ctx, sqlitesqlc.UpdateIncidentEventStepParams{
			ID:         s.ID,
			IncidentID: s.IncidentID,
			Step:       string(s.Step),
			Message:    nullStringFromPtr(s.Message),
			UpdatedAt:  s.UpdatedAt,
		})
		if err != nil {
			return fmt.Errorf("sqlc: update event step: %w", err)
		}
		if n == 0 {
			return repository.ErrNotFound
		}
		return nil
	default:
		return r.unconfigured()
	}
}

func (r *IncidentEventStepRepositorySQLC) Delete(ctx context.Context, id string) error {
	switch {
	case r.pgQ != nil:
		n, err := r.pgQ.DeleteIncidentEventStep(ctx, id)
		if err != nil {
			return fmt.Errorf("sqlc: delete event step: %w", err)
		}
		if n == 0 {
			return repository.ErrNotFound
		}
		return nil
	case r.sqliteQ != nil:
		n, err := r.sqliteQ.DeleteIncidentEventStep(ctx, id)
		if err != nil {
			return fmt.Errorf("sqlc: delete event step: %w", err)
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

func eventStepFromPG(row pgsqlc.IncidentEventStep) *domain.IncidentEventStep {
	out := &domain.IncidentEventStep{
		Base: domain.Base{
			ID:        row.ID,
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
		},
		IncidentID: row.IncidentID,
		Step:       domain.IncidentEventStepType(row.Step),
	}
	if row.Message.Valid {
		s := row.Message.String
		out.Message = &s
	}
	return out
}

func eventStepFromSQLite(row sqlitesqlc.IncidentEventStep) *domain.IncidentEventStep {
	out := &domain.IncidentEventStep{
		Base: domain.Base{
			ID:        row.ID,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		IncidentID: row.IncidentID,
		Step:       domain.IncidentEventStepType(row.Step),
	}
	if row.Message.Valid {
		s := row.Message.String
		out.Message = &s
	}
	return out
}

// eventStepFromPGJoin splits the JOIN row into a step + embedded incident.
func eventStepFromPGJoin(row pgsqlc.FindIncidentEventStepByIDRow) *domain.IncidentEventStep {
	out := &domain.IncidentEventStep{
		Base: domain.Base{
			ID:        row.ID,
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
		},
		IncidentID: row.IncidentID,
		Step:       domain.IncidentEventStepType(row.Step),
		Incident: domain.Incident{
			Base: domain.Base{
				ID:        row.IncidentID,
				CreatedAt: row.ICreatedAt.Time,
				UpdatedAt: row.IUpdatedAt.Time,
			},
			ResourceID: row.ResourceID,
			Cause:      row.Cause,
			StartedAt:  row.StartedAt.Time,
			Details:    row.Details,
		},
	}
	if row.Message.Valid {
		s := row.Message.String
		out.Message = &s
	}
	if row.ResolvedAt.Valid {
		t := row.ResolvedAt.Time
		out.Incident.ResolvedAt = &t
	}
	return out
}

func eventStepFromSQLiteJoin(row sqlitesqlc.FindIncidentEventStepByIDRow) *domain.IncidentEventStep {
	out := &domain.IncidentEventStep{
		Base: domain.Base{
			ID:        row.ID,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		IncidentID: row.IncidentID,
		Step:       domain.IncidentEventStepType(row.Step),
		Incident: domain.Incident{
			Base: domain.Base{
				ID:        row.IncidentID,
				CreatedAt: row.ICreatedAt,
				UpdatedAt: row.IUpdatedAt,
			},
			ResourceID: row.ResourceID,
			Cause:      row.Cause,
			StartedAt:  row.StartedAt,
			Details:    row.Details,
		},
	}
	if row.Message.Valid {
		s := row.Message.String
		out.Message = &s
	}
	if row.ResolvedAt.Valid {
		t := row.ResolvedAt.Time
		out.Incident.ResolvedAt = &t
	}
	return out
}
