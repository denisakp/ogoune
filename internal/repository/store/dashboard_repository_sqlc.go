package store

import (
	"context"
	"encoding/json"
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

type DashboardRepositorySQLC struct {
	pgQ     *pgsqlc.Queries
	sqliteQ *sqlitesqlc.Queries
}

func NewDashboardRepositorySQLC(rt SqlcRuntime) port.DashboardRepository {
	r := &DashboardRepositorySQLC{}
	if pool := rt.PgxPool(); pool != nil {
		r.pgQ = pgsqlc.New(pool)
	} else if db := rt.SQLiteDB(); db != nil {
		r.sqliteQ = sqlitesqlc.New(db)
	}
	return r
}

func (r *DashboardRepositorySQLC) unconfigured() error {
	return fmt.Errorf("dashboard_repository_sqlc: unconfigured runtime")
}

// marshalConfig serializes scope + widgets to JSON bytes for the JSON columns.
func marshalConfig(d *domain.Dashboard) (scope, widgets []byte, err error) {
	if scope, err = json.Marshal(d.Scope); err != nil {
		return nil, nil, fmt.Errorf("marshal scope: %w", err)
	}
	// Ensure a non-null JSON array even when empty.
	w := d.Widgets
	if w == nil {
		w = []domain.WidgetInstance{}
	}
	if widgets, err = json.Marshal(w); err != nil {
		return nil, nil, fmt.Errorf("marshal widgets: %w", err)
	}
	return scope, widgets, nil
}

func (r *DashboardRepositorySQLC) Create(ctx context.Context, d *domain.Dashboard) (*domain.Dashboard, error) {
	d.EnsureID()
	now := time.Now()
	if d.CreatedAt.IsZero() {
		d.CreatedAt = now
	}
	if d.UpdatedAt.IsZero() {
		d.UpdatedAt = now
	}
	scope, widgets, err := marshalConfig(d)
	if err != nil {
		return nil, err
	}
	switch {
	case r.pgQ != nil:
		_, err := r.pgQ.CreateDashboard(ctx, pgsqlc.CreateDashboardParams{
			ID:               d.ID,
			OwnerID:          d.OwnerID,
			Name:             d.Name,
			Scope:            scope,
			Widgets:          widgets,
			DefaultTimeRange: d.DefaultTimeRange,
			RefreshInterval:  d.RefreshInterval,
			Visibility:       d.Visibility,
			CreatedAt:        pgtype.Timestamptz{Time: d.CreatedAt, Valid: true},
			UpdatedAt:        pgtype.Timestamptz{Time: d.UpdatedAt, Valid: true},
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: create dashboard: %w", err)
		}
	case r.sqliteQ != nil:
		_, err := r.sqliteQ.CreateDashboard(ctx, sqlitesqlc.CreateDashboardParams{
			ID:               d.ID,
			OwnerID:          d.OwnerID,
			Name:             d.Name,
			Scope:            string(scope),
			Widgets:          string(widgets),
			DefaultTimeRange: d.DefaultTimeRange,
			RefreshInterval:  d.RefreshInterval,
			Visibility:       d.Visibility,
			CreatedAt:        d.CreatedAt,
			UpdatedAt:        d.UpdatedAt,
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: create dashboard: %w", err)
		}
	default:
		return nil, r.unconfigured()
	}
	// Re-fetch to populate owner_name from the users JOIN.
	return r.FindByID(ctx, d.ID)
}

func (r *DashboardRepositorySQLC) FindByID(ctx context.Context, id string) (*domain.Dashboard, error) {
	switch {
	case r.pgQ != nil:
		row, err := r.pgQ.FindDashboardByID(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, repository.ErrNotFound
			}
			return nil, fmt.Errorf("sqlc: find dashboard: %w", err)
		}
		return dashboardFromPGRow(row)
	case r.sqliteQ != nil:
		row, err := r.sqliteQ.FindDashboardByID(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) || err.Error() == "sql: no rows in result set" {
				return nil, repository.ErrNotFound
			}
			return nil, fmt.Errorf("sqlc: find dashboard: %w", err)
		}
		return dashboardFromSQLiteRow(row)
	default:
		return nil, r.unconfigured()
	}
}

func (r *DashboardRepositorySQLC) List(ctx context.Context, limit, offset int) ([]*domain.Dashboard, error) {
	switch {
	case r.pgQ != nil:
		rows, err := r.pgQ.ListDashboards(ctx, pgsqlc.ListDashboardsParams{Lim: int32(limit), Off: int32(offset)})
		if err != nil {
			return nil, fmt.Errorf("sqlc: list dashboards: %w", err)
		}
		out := make([]*domain.Dashboard, 0, len(rows))
		for _, row := range rows {
			d, err := dashboardFromPGRow(pgsqlc.FindDashboardByIDRow(row))
			if err != nil {
				return nil, err
			}
			out = append(out, d)
		}
		return out, nil
	case r.sqliteQ != nil:
		rows, err := r.sqliteQ.ListDashboards(ctx, sqlitesqlc.ListDashboardsParams{Lim: int64(limit), Off: int64(offset)})
		if err != nil {
			return nil, fmt.Errorf("sqlc: list dashboards: %w", err)
		}
		out := make([]*domain.Dashboard, 0, len(rows))
		for _, row := range rows {
			d, err := dashboardFromSQLiteRow(sqlitesqlc.FindDashboardByIDRow(row))
			if err != nil {
				return nil, err
			}
			out = append(out, d)
		}
		return out, nil
	default:
		return nil, r.unconfigured()
	}
}

func (r *DashboardRepositorySQLC) Update(ctx context.Context, d *domain.Dashboard) error {
	d.UpdatedAt = time.Now()
	scope, widgets, err := marshalConfig(d)
	if err != nil {
		return err
	}
	switch {
	case r.pgQ != nil:
		n, err := r.pgQ.UpdateDashboard(ctx, pgsqlc.UpdateDashboardParams{
			Name:             d.Name,
			Scope:            scope,
			Widgets:          widgets,
			DefaultTimeRange: d.DefaultTimeRange,
			RefreshInterval:  d.RefreshInterval,
			Visibility:       d.Visibility,
			UpdatedAt:        pgtype.Timestamptz{Time: d.UpdatedAt, Valid: true},
			ID:               d.ID,
		})
		if err != nil {
			return fmt.Errorf("sqlc: update dashboard: %w", err)
		}
		if n == 0 {
			return repository.ErrNotFound
		}
		return nil
	case r.sqliteQ != nil:
		n, err := r.sqliteQ.UpdateDashboard(ctx, sqlitesqlc.UpdateDashboardParams{
			Name:             d.Name,
			Scope:            string(scope),
			Widgets:          string(widgets),
			DefaultTimeRange: d.DefaultTimeRange,
			RefreshInterval:  d.RefreshInterval,
			Visibility:       d.Visibility,
			UpdatedAt:        d.UpdatedAt,
			ID:               d.ID,
		})
		if err != nil {
			return fmt.Errorf("sqlc: update dashboard: %w", err)
		}
		if n == 0 {
			return repository.ErrNotFound
		}
		return nil
	default:
		return r.unconfigured()
	}
}

func (r *DashboardRepositorySQLC) UpdateWidgets(ctx context.Context, id string, widgets []domain.WidgetInstance, at time.Time) error {
	if widgets == nil {
		widgets = []domain.WidgetInstance{}
	}
	b, err := json.Marshal(widgets)
	if err != nil {
		return fmt.Errorf("marshal widgets: %w", err)
	}
	switch {
	case r.pgQ != nil:
		n, err := r.pgQ.UpdateDashboardWidgets(ctx, pgsqlc.UpdateDashboardWidgetsParams{
			Widgets:   b,
			UpdatedAt: pgtype.Timestamptz{Time: at, Valid: true},
			ID:        id,
		})
		if err != nil {
			return fmt.Errorf("sqlc: update widgets: %w", err)
		}
		if n == 0 {
			return repository.ErrNotFound
		}
		return nil
	case r.sqliteQ != nil:
		n, err := r.sqliteQ.UpdateDashboardWidgets(ctx, sqlitesqlc.UpdateDashboardWidgetsParams{
			Widgets:   string(b),
			UpdatedAt: at,
			ID:        id,
		})
		if err != nil {
			return fmt.Errorf("sqlc: update widgets: %w", err)
		}
		if n == 0 {
			return repository.ErrNotFound
		}
		return nil
	default:
		return r.unconfigured()
	}
}

func (r *DashboardRepositorySQLC) Delete(ctx context.Context, id string) error {
	switch {
	case r.pgQ != nil:
		n, err := r.pgQ.DeleteDashboard(ctx, id)
		if err != nil {
			return fmt.Errorf("sqlc: delete dashboard: %w", err)
		}
		if n == 0 {
			return repository.ErrNotFound
		}
		return nil
	case r.sqliteQ != nil:
		n, err := r.sqliteQ.DeleteDashboard(ctx, id)
		if err != nil {
			return fmt.Errorf("sqlc: delete dashboard: %w", err)
		}
		if n == 0 {
			return repository.ErrNotFound
		}
		return nil
	default:
		return r.unconfigured()
	}
}

func unmarshalConfig(scope, widgets []byte, d *domain.Dashboard) error {
	if len(scope) > 0 {
		if err := json.Unmarshal(scope, &d.Scope); err != nil {
			return fmt.Errorf("unmarshal scope: %w", err)
		}
	}
	d.Widgets = []domain.WidgetInstance{}
	if len(widgets) > 0 {
		if err := json.Unmarshal(widgets, &d.Widgets); err != nil {
			return fmt.Errorf("unmarshal widgets: %w", err)
		}
	}
	return nil
}

func dashboardFromPGRow(row pgsqlc.FindDashboardByIDRow) (*domain.Dashboard, error) {
	d := &domain.Dashboard{
		Base:             domain.Base{ID: row.ID, CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time},
		OwnerID:          row.OwnerID,
		OwnerName:        row.OwnerName,
		Name:             row.DashboardName,
		DefaultTimeRange: row.DefaultTimeRange,
		RefreshInterval:  row.RefreshInterval,
		Visibility:       row.Visibility,
	}
	if err := unmarshalConfig(row.Scope, row.Widgets, d); err != nil {
		return nil, err
	}
	return d, nil
}

func dashboardFromSQLiteRow(row sqlitesqlc.FindDashboardByIDRow) (*domain.Dashboard, error) {
	d := &domain.Dashboard{
		Base:             domain.Base{ID: row.ID, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt},
		OwnerID:          row.OwnerID,
		OwnerName:        row.OwnerName,
		Name:             row.DashboardName,
		DefaultTimeRange: row.DefaultTimeRange,
		RefreshInterval:  row.RefreshInterval,
		Visibility:       row.Visibility,
	}
	if err := unmarshalConfig([]byte(row.Scope), []byte(row.Widgets), d); err != nil {
		return nil, err
	}
	return d, nil
}
