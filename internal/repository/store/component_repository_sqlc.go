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

// ComponentRepositorySQLC implements port.ComponentRepository via sqlc.
//
// Note: Wave-2 scope intentionally does NOT eagerly hydrate the Resources
// field; the service layer (ComponentService.toComponentResponse) already
// falls back to a separate FindByComponentID(resourceRepo) lookup when
// Resources is nil. Wave 3 will revisit when resource_repository migrates.
type ComponentRepositorySQLC struct {
	pgQ     *pgsqlc.Queries
	sqliteQ *sqlitesqlc.Queries
}

func NewComponentRepositorySQLC(rt SqlcRuntime) port.ComponentRepository {
	r := &ComponentRepositorySQLC{}
	if pool := rt.PgxPool(); pool != nil {
		r.pgQ = pgsqlc.New(pool)
	} else if db := rt.SQLiteDB(); db != nil {
		r.sqliteQ = sqlitesqlc.New(db)
	}
	return r
}

func (r *ComponentRepositorySQLC) unconfigured() error {
	return fmt.Errorf("component_repository_sqlc: unconfigured runtime")
}

func (r *ComponentRepositorySQLC) Create(ctx context.Context, c *domain.Component) (*domain.Component, error) {
	c.EnsureID()
	now := time.Now()
	if c.CreatedAt.IsZero() {
		c.CreatedAt = now
	}
	if c.UpdatedAt.IsZero() {
		c.UpdatedAt = now
	}
	if c.LastNotificationStatus == "" {
		c.LastNotificationStatus = domain.ComponentStatus("up")
	}
	if c.GroupingWindowSeconds == 0 {
		c.GroupingWindowSeconds = 30
	}
	switch {
	case r.pgQ != nil:
		err := r.pgQ.CreateComponent(ctx, pgsqlc.CreateComponentParams{
			ID:                     c.ID,
			CreatedAt:              pgtype.Timestamptz{Time: c.CreatedAt, Valid: true},
			UpdatedAt:              pgtype.Timestamptz{Time: c.UpdatedAt, Valid: true},
			Name:                   c.Name,
			Description:            pgTextFromPtr(c.Description),
			LastNotificationStatus: string(c.LastNotificationStatus),
			GroupingWindowSeconds:  int32(c.GroupingWindowSeconds),
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: create component: %w", err)
		}
		return c, nil
	case r.sqliteQ != nil:
		err := r.sqliteQ.CreateComponent(ctx, sqlitesqlc.CreateComponentParams{
			ID:                     c.ID,
			CreatedAt:              c.CreatedAt,
			UpdatedAt:              c.UpdatedAt,
			Name:                   c.Name,
			Description:            nullStringFromPtr(c.Description),
			LastNotificationStatus: string(c.LastNotificationStatus),
			GroupingWindowSeconds:  int64(c.GroupingWindowSeconds),
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: create component: %w", err)
		}
		return c, nil
	default:
		return nil, r.unconfigured()
	}
}

func (r *ComponentRepositorySQLC) FindByID(ctx context.Context, id string) (*domain.Component, error) {
	switch {
	case r.pgQ != nil:
		row, err := r.pgQ.FindComponentByID(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, repository.ErrNotFound
			}
			return nil, fmt.Errorf("sqlc: find component: %w", err)
		}
		return componentFromPG(row), nil
	case r.sqliteQ != nil:
		row, err := r.sqliteQ.FindComponentByID(ctx, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, repository.ErrNotFound
			}
			return nil, fmt.Errorf("sqlc: find component: %w", err)
		}
		return componentFromSQLite(row), nil
	default:
		return nil, r.unconfigured()
	}
}

func (r *ComponentRepositorySQLC) List(ctx context.Context, limit, offset int) ([]*domain.Component, error) {
	switch {
	case r.pgQ != nil:
		rows, err := r.pgQ.ListComponents(ctx, pgsqlc.ListComponentsParams{
			Limit: int32(limit), Offset: int32(offset),
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: list components: %w", err)
		}
		out := make([]*domain.Component, len(rows))
		for i, row := range rows {
			out[i] = componentFromPG(row)
		}
		return out, nil
	case r.sqliteQ != nil:
		rows, err := r.sqliteQ.ListComponents(ctx, sqlitesqlc.ListComponentsParams{
			Limit: int64(limit), Offset: int64(offset),
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: list components: %w", err)
		}
		out := make([]*domain.Component, len(rows))
		for i, row := range rows {
			out[i] = componentFromSQLite(row)
		}
		return out, nil
	default:
		return nil, r.unconfigured()
	}
}

func (r *ComponentRepositorySQLC) Update(ctx context.Context, c *domain.Component) error {
	c.UpdatedAt = time.Now()
	switch {
	case r.pgQ != nil:
		return r.pgQ.UpdateComponent(ctx, pgsqlc.UpdateComponentParams{
			ID:          c.ID,
			Name:        c.Name,
			Description: pgTextFromPtr(c.Description),
			UpdatedAt:   pgtype.Timestamptz{Time: c.UpdatedAt, Valid: true},
		})
	case r.sqliteQ != nil:
		return r.sqliteQ.UpdateComponent(ctx, sqlitesqlc.UpdateComponentParams{
			ID:          c.ID,
			Name:        c.Name,
			Description: nullStringFromPtr(c.Description),
			UpdatedAt:   c.UpdatedAt,
		})
	default:
		return r.unconfigured()
	}
}

func (r *ComponentRepositorySQLC) Delete(ctx context.Context, id string) error {
	switch {
	case r.pgQ != nil:
		n, err := r.pgQ.DeleteComponent(ctx, id)
		if err != nil {
			return fmt.Errorf("sqlc: delete component: %w", err)
		}
		if n == 0 {
			return repository.ErrNotFound
		}
		return nil
	case r.sqliteQ != nil:
		n, err := r.sqliteQ.DeleteComponent(ctx, id)
		if err != nil {
			return fmt.Errorf("sqlc: delete component: %w", err)
		}
		if n == 0 {
			return repository.ErrNotFound
		}
		return nil
	default:
		return r.unconfigured()
	}
}

func (r *ComponentRepositorySQLC) UpdateLastNotificationStatus(ctx context.Context, id string, status domain.ComponentStatus) error {
	switch {
	case r.pgQ != nil:
		n, err := r.pgQ.UpdateComponentLastNotificationStatus(ctx, pgsqlc.UpdateComponentLastNotificationStatusParams{
			ID: id, LastNotificationStatus: string(status),
		})
		if err != nil {
			return fmt.Errorf("sqlc: update component status: %w", err)
		}
		if n == 0 {
			return repository.ErrNotFound
		}
		return nil
	case r.sqliteQ != nil:
		n, err := r.sqliteQ.UpdateComponentLastNotificationStatus(ctx, sqlitesqlc.UpdateComponentLastNotificationStatusParams{
			ID: id, LastNotificationStatus: string(status),
		})
		if err != nil {
			return fmt.Errorf("sqlc: update component status: %w", err)
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

func componentFromPG(row pgsqlc.Component) *domain.Component {
	out := &domain.Component{
		Base: domain.Base{
			ID:        row.ID,
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
		},
		Name:                   row.Name,
		LastNotificationStatus: domain.ComponentStatus(row.LastNotificationStatus),
		GroupingWindowSeconds:  int(row.GroupingWindowSeconds),
	}
	if row.Description.Valid {
		s := row.Description.String
		out.Description = &s
	}
	return out
}

func componentFromSQLite(row sqlitesqlc.Component) *domain.Component {
	out := &domain.Component{
		Base: domain.Base{
			ID:        row.ID,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		Name:                   row.Name,
		LastNotificationStatus: domain.ComponentStatus(row.LastNotificationStatus),
		GroupingWindowSeconds:  int(row.GroupingWindowSeconds),
	}
	if row.Description.Valid {
		s := row.Description.String
		out.Description = &s
	}
	return out
}
