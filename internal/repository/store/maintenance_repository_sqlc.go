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

type MaintenanceRepositorySQLC struct {
	pgQ      *pgsqlc.Queries
	sqliteQ  *sqlitesqlc.Queries
	pgPool   *pgxpool.Pool
	sqliteDB *sql.DB
}

func NewMaintenanceRepositorySQLC(rt SqlcRuntime) port.MaintenanceRepository {
	r := &MaintenanceRepositorySQLC{}
	if pool := rt.PgxPool(); pool != nil {
		r.pgPool = pool
		r.pgQ = pgsqlc.New(pool)
	} else if db := rt.SQLiteDB(); db != nil {
		r.sqliteDB = db
		r.sqliteQ = sqlitesqlc.New(db)
	}
	return r
}

func (r *MaintenanceRepositorySQLC) unconfigured() error {
	return fmt.Errorf("maintenance_repository_sqlc: unconfigured runtime")
}

// ---------- helpers ----------

func resourceIDsFromMaintenance(m *domain.Maintenance) []string {
	if len(m.Resources) == 0 {
		return nil
	}
	out := make([]string, 0, len(m.Resources))
	for _, res := range m.Resources {
		out = append(out, res.ID)
	}
	return out
}

func pgIntPtrFromInt(p *int) pgtype.Int4 {
	if p == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: int32(*p), Valid: true}
}

func sqliteNullInt64Ptr(p *int) sql.NullInt64 {
	if p == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(*p), Valid: true}
}

// ---------- Public surface ----------

func (r *MaintenanceRepositorySQLC) Create(ctx context.Context, m *domain.Maintenance) (*domain.Maintenance, error) {
	m.EnsureID()
	now := time.Now()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = now
	}
	resourceIDs := resourceIDsFromMaintenance(m)
	switch {
	case r.pgQ != nil:
		return m, pgsqlc.WithTx(ctx, r.pgPool, func(q *pgsqlc.Queries) error {
			if err := q.CreateMaintenance(ctx, pgsqlc.CreateMaintenanceParams{
				ID:             m.ID,
				CreatedAt:      pgtype.Timestamptz{Time: m.CreatedAt, Valid: true},
				UpdatedAt:      pgtype.Timestamptz{Time: m.UpdatedAt, Valid: true},
				Title:          m.Title,
				Description:    pgTextFromPtr(m.Description),
				Strategy:       string(m.Strategy),
				Status:         m.Status,
				StartAt:        pgTimestampFromPtr(m.StartAt),
				EndAt:          pgTimestampFromPtr(m.EndAt),
				CronExpr:       pgTextFromPtr(m.CronExpr),
				WindowMinutes:  pgIntPtrFromInt(m.WindowMinutes),
				Timezone:       pgTextFromPtr(m.Timezone),
				EffectiveFrom:  pgTimestampFromPtr(m.EffectiveFrom),
				EffectiveUntil: pgTimestampFromPtr(m.EffectiveUntil),
				StartedAt:      pgTimestampFromPtr(m.StartedAt),
				EndedAt:        pgTimestampFromPtr(m.EndedAt),
			}); err != nil {
				return err
			}
			for _, rid := range resourceIDs {
				if err := q.LinkMaintenanceResource(ctx, pgsqlc.LinkMaintenanceResourceParams{
					MaintenanceID: m.ID, ResourceID: rid,
				}); err != nil {
					return err
				}
			}
			return nil
		})
	case r.sqliteQ != nil:
		return m, sqlitesqlc.WithTx(ctx, r.sqliteDB, func(q *sqlitesqlc.Queries) error {
			if err := q.CreateMaintenance(ctx, sqlitesqlc.CreateMaintenanceParams{
				ID:             m.ID,
				CreatedAt:      m.CreatedAt,
				UpdatedAt:      m.UpdatedAt,
				Title:          m.Title,
				Description:    nullStringFromPtr(m.Description),
				Strategy:       string(m.Strategy),
				Status:         m.Status,
				StartAt:        nullTimeFromPtr(m.StartAt),
				EndAt:          nullTimeFromPtr(m.EndAt),
				CronExpr:       nullStringFromPtr(m.CronExpr),
				WindowMinutes:  sqliteNullInt64Ptr(m.WindowMinutes),
				Timezone:       nullStringFromPtr(m.Timezone),
				EffectiveFrom:  nullTimeFromPtr(m.EffectiveFrom),
				EffectiveUntil: nullTimeFromPtr(m.EffectiveUntil),
				StartedAt:      nullTimeFromPtr(m.StartedAt),
				EndedAt:        nullTimeFromPtr(m.EndedAt),
			}); err != nil {
				return err
			}
			for _, rid := range resourceIDs {
				if err := q.LinkMaintenanceResource(ctx, sqlitesqlc.LinkMaintenanceResourceParams{
					MaintenanceID: m.ID, ResourceID: rid,
				}); err != nil {
					return err
				}
			}
			return nil
		})
	default:
		return nil, r.unconfigured()
	}
}

func (r *MaintenanceRepositorySQLC) FindByID(ctx context.Context, id string) (*domain.Maintenance, error) {
	switch {
	case r.pgQ != nil:
		row, err := r.pgQ.FindMaintenanceByID(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, repository.ErrNotFound
			}
			return nil, fmt.Errorf("sqlc: find maintenance: %w", err)
		}
		out := maintenanceFromPG(row)
		// Preload resources (ID-only, mirrors GORM's preload that populates the slice).
		rids, err := r.pgQ.ListMaintenanceResourceIDsByMaintenanceID(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("sqlc: preload maintenance resources: %w", err)
		}
		out.Resources = resourceStubsFromIDs(rids)
		return out, nil
	case r.sqliteQ != nil:
		row, err := r.sqliteQ.FindMaintenanceByID(ctx, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, repository.ErrNotFound
			}
			return nil, fmt.Errorf("sqlc: find maintenance: %w", err)
		}
		out := maintenanceFromSQLite(row)
		rids, err := r.sqliteQ.ListMaintenanceResourceIDsByMaintenanceID(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("sqlc: preload maintenance resources: %w", err)
		}
		out.Resources = resourceStubsFromIDs(rids)
		return out, nil
	default:
		return nil, r.unconfigured()
	}
}

func (r *MaintenanceRepositorySQLC) List(ctx context.Context, status string, limit, offset int) ([]*domain.Maintenance, error) {
	switch {
	case r.pgQ != nil:
		var rows []pgsqlc.Maintenance
		var err error
		if status != "" {
			rows, err = r.pgQ.ListMaintenancesByStatus(ctx, pgsqlc.ListMaintenancesByStatusParams{
				Status: status, Limit: int32(limit), Offset: int32(offset),
			})
		} else {
			rows, err = r.pgQ.ListMaintenancesAll(ctx, pgsqlc.ListMaintenancesAllParams{
				Limit: int32(limit), Offset: int32(offset),
			})
		}
		if err != nil {
			return nil, fmt.Errorf("sqlc: list maintenances: %w", err)
		}
		out := make([]*domain.Maintenance, len(rows))
		for i, row := range rows {
			m := maintenanceFromPG(row)
			rids, err := r.pgQ.ListMaintenanceResourceIDsByMaintenanceID(ctx, m.ID)
			if err != nil {
				return nil, fmt.Errorf("sqlc: preload list resources: %w", err)
			}
			m.Resources = resourceStubsFromIDs(rids)
			out[i] = m
		}
		return out, nil
	case r.sqliteQ != nil:
		var rows []sqlitesqlc.Maintenance
		var err error
		if status != "" {
			rows, err = r.sqliteQ.ListMaintenancesByStatus(ctx, sqlitesqlc.ListMaintenancesByStatusParams{
				Status: status, Limit: int64(limit), Offset: int64(offset),
			})
		} else {
			rows, err = r.sqliteQ.ListMaintenancesAll(ctx, sqlitesqlc.ListMaintenancesAllParams{
				Limit: int64(limit), Offset: int64(offset),
			})
		}
		if err != nil {
			return nil, fmt.Errorf("sqlc: list maintenances: %w", err)
		}
		out := make([]*domain.Maintenance, len(rows))
		for i, row := range rows {
			m := maintenanceFromSQLite(row)
			rids, err := r.sqliteQ.ListMaintenanceResourceIDsByMaintenanceID(ctx, m.ID)
			if err != nil {
				return nil, fmt.Errorf("sqlc: preload list resources: %w", err)
			}
			m.Resources = resourceStubsFromIDs(rids)
			out[i] = m
		}
		return out, nil
	default:
		return nil, r.unconfigured()
	}
}

func (r *MaintenanceRepositorySQLC) Update(ctx context.Context, m *domain.Maintenance) error {
	m.UpdatedAt = time.Now()
	targetIDs := resourceIDsFromMaintenance(m)
	switch {
	case r.pgQ != nil:
		return pgsqlc.WithTx(ctx, r.pgPool, func(q *pgsqlc.Queries) error {
			if err := q.UpdateMaintenance(ctx, pgsqlc.UpdateMaintenanceParams{
				ID:             m.ID,
				Title:          m.Title,
				Description:    pgTextFromPtr(m.Description),
				Strategy:       string(m.Strategy),
				Status:         m.Status,
				StartAt:        pgTimestampFromPtr(m.StartAt),
				EndAt:          pgTimestampFromPtr(m.EndAt),
				CronExpr:       pgTextFromPtr(m.CronExpr),
				WindowMinutes:  pgIntPtrFromInt(m.WindowMinutes),
				Timezone:       pgTextFromPtr(m.Timezone),
				EffectiveFrom:  pgTimestampFromPtr(m.EffectiveFrom),
				EffectiveUntil: pgTimestampFromPtr(m.EffectiveUntil),
				StartedAt:      pgTimestampFromPtr(m.StartedAt),
				EndedAt:        pgTimestampFromPtr(m.EndedAt),
				UpdatedAt:      pgtype.Timestamptz{Time: m.UpdatedAt, Valid: true},
			}); err != nil {
				return err
			}
			currentIDs, err := q.ListMaintenanceResourceIDsByMaintenanceID(ctx, m.ID)
			if err != nil {
				return err
			}
			toAdd, toRemove := diffJunctionSets(currentIDs, targetIDs)
			for _, rid := range toRemove {
				if err := q.UnlinkMaintenanceResource(ctx, pgsqlc.UnlinkMaintenanceResourceParams{
					MaintenanceID: m.ID, ResourceID: rid,
				}); err != nil {
					return err
				}
			}
			for _, rid := range toAdd {
				if err := q.LinkMaintenanceResource(ctx, pgsqlc.LinkMaintenanceResourceParams{
					MaintenanceID: m.ID, ResourceID: rid,
				}); err != nil {
					return err
				}
			}
			return nil
		})
	case r.sqliteQ != nil:
		return sqlitesqlc.WithTx(ctx, r.sqliteDB, func(q *sqlitesqlc.Queries) error {
			if err := q.UpdateMaintenance(ctx, sqlitesqlc.UpdateMaintenanceParams{
				ID:             m.ID,
				Title:          m.Title,
				Description:    nullStringFromPtr(m.Description),
				Strategy:       string(m.Strategy),
				Status:         m.Status,
				StartAt:        nullTimeFromPtr(m.StartAt),
				EndAt:          nullTimeFromPtr(m.EndAt),
				CronExpr:       nullStringFromPtr(m.CronExpr),
				WindowMinutes:  sqliteNullInt64Ptr(m.WindowMinutes),
				Timezone:       nullStringFromPtr(m.Timezone),
				EffectiveFrom:  nullTimeFromPtr(m.EffectiveFrom),
				EffectiveUntil: nullTimeFromPtr(m.EffectiveUntil),
				StartedAt:      nullTimeFromPtr(m.StartedAt),
				EndedAt:        nullTimeFromPtr(m.EndedAt),
				UpdatedAt:      m.UpdatedAt,
			}); err != nil {
				return err
			}
			currentIDs, err := q.ListMaintenanceResourceIDsByMaintenanceID(ctx, m.ID)
			if err != nil {
				return err
			}
			toAdd, toRemove := diffJunctionSets(currentIDs, targetIDs)
			for _, rid := range toRemove {
				if err := q.UnlinkMaintenanceResource(ctx, sqlitesqlc.UnlinkMaintenanceResourceParams{
					MaintenanceID: m.ID, ResourceID: rid,
				}); err != nil {
					return err
				}
			}
			for _, rid := range toAdd {
				if err := q.LinkMaintenanceResource(ctx, sqlitesqlc.LinkMaintenanceResourceParams{
					MaintenanceID: m.ID, ResourceID: rid,
				}); err != nil {
					return err
				}
			}
			return nil
		})
	default:
		return r.unconfigured()
	}
}

func (r *MaintenanceRepositorySQLC) Delete(ctx context.Context, id string) error {
	switch {
	case r.pgQ != nil:
		n, err := r.pgQ.DeleteMaintenance(ctx, id)
		if err != nil {
			return fmt.Errorf("sqlc: delete maintenance: %w", err)
		}
		if n == 0 {
			return repository.ErrNotFound
		}
		return nil
	case r.sqliteQ != nil:
		n, err := r.sqliteQ.DeleteMaintenance(ctx, id)
		if err != nil {
			return fmt.Errorf("sqlc: delete maintenance: %w", err)
		}
		if n == 0 {
			return repository.ErrNotFound
		}
		return nil
	default:
		return r.unconfigured()
	}
}

func (r *MaintenanceRepositorySQLC) FindActiveForResource(ctx context.Context, resourceID string, now time.Time) ([]*domain.Maintenance, error) {
	switch {
	case r.pgQ != nil:
		rows, err := r.pgQ.FindActiveMaintenancesForResource(ctx, pgsqlc.FindActiveMaintenancesForResourceParams{
			ResourceID: resourceID,
			StartAt:    pgtype.Timestamptz{Time: now, Valid: true},
			EndAt:      pgtype.Timestamptz{Time: now, Valid: true},
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: find active maintenances: %w", err)
		}
		out := make([]*domain.Maintenance, len(rows))
		for i, row := range rows {
			m := maintenanceFromPG(row)
			rids, err := r.pgQ.ListMaintenanceResourceIDsByMaintenanceID(ctx, m.ID)
			if err != nil {
				return nil, fmt.Errorf("sqlc: preload active resources: %w", err)
			}
			m.Resources = resourceStubsFromIDs(rids)
			out[i] = m
		}
		return out, nil
	case r.sqliteQ != nil:
		rows, err := r.sqliteQ.FindActiveMaintenancesForResource(ctx, sqlitesqlc.FindActiveMaintenancesForResourceParams{
			ResourceID: resourceID,
			StartAt:    sql.NullTime{Time: now, Valid: true},
			EndAt:      sql.NullTime{Time: now, Valid: true},
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: find active maintenances: %w", err)
		}
		out := make([]*domain.Maintenance, len(rows))
		for i, row := range rows {
			m := maintenanceFromSQLite(row)
			rids, err := r.sqliteQ.ListMaintenanceResourceIDsByMaintenanceID(ctx, m.ID)
			if err != nil {
				return nil, fmt.Errorf("sqlc: preload active resources: %w", err)
			}
			m.Resources = resourceStubsFromIDs(rids)
			out[i] = m
		}
		return out, nil
	default:
		return nil, r.unconfigured()
	}
}

// ---------- mapping helpers ----------

func resourceStubsFromIDs(ids []string) []*domain.Resource {
	if len(ids) == 0 {
		return nil
	}
	out := make([]*domain.Resource, len(ids))
	for i, id := range ids {
		out[i] = &domain.Resource{Base: domain.Base{ID: id}}
	}
	return out
}

func maintenanceFromPG(row pgsqlc.Maintenance) *domain.Maintenance {
	out := &domain.Maintenance{
		Base: domain.Base{
			ID:        row.ID,
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
		},
		Title:    row.Title,
		Strategy: domain.MaintenanceStrategy(row.Strategy),
		Status:   row.Status,
	}
	if row.Description.Valid {
		s := row.Description.String
		out.Description = &s
	}
	if row.StartAt.Valid {
		t := row.StartAt.Time
		out.StartAt = &t
	}
	if row.EndAt.Valid {
		t := row.EndAt.Time
		out.EndAt = &t
	}
	if row.CronExpr.Valid {
		s := row.CronExpr.String
		out.CronExpr = &s
	}
	if row.WindowMinutes.Valid {
		v := int(row.WindowMinutes.Int32)
		out.WindowMinutes = &v
	}
	if row.Timezone.Valid {
		s := row.Timezone.String
		out.Timezone = &s
	}
	if row.EffectiveFrom.Valid {
		t := row.EffectiveFrom.Time
		out.EffectiveFrom = &t
	}
	if row.EffectiveUntil.Valid {
		t := row.EffectiveUntil.Time
		out.EffectiveUntil = &t
	}
	if row.StartedAt.Valid {
		t := row.StartedAt.Time
		out.StartedAt = &t
	}
	if row.EndedAt.Valid {
		t := row.EndedAt.Time
		out.EndedAt = &t
	}
	return out
}

func maintenanceFromSQLite(row sqlitesqlc.Maintenance) *domain.Maintenance {
	out := &domain.Maintenance{
		Base: domain.Base{
			ID:        row.ID,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		Title:    row.Title,
		Strategy: domain.MaintenanceStrategy(row.Strategy),
		Status:   row.Status,
	}
	if row.Description.Valid {
		s := row.Description.String
		out.Description = &s
	}
	if row.StartAt.Valid {
		t := row.StartAt.Time
		out.StartAt = &t
	}
	if row.EndAt.Valid {
		t := row.EndAt.Time
		out.EndAt = &t
	}
	if row.CronExpr.Valid {
		s := row.CronExpr.String
		out.CronExpr = &s
	}
	if row.WindowMinutes.Valid {
		v := int(row.WindowMinutes.Int64)
		out.WindowMinutes = &v
	}
	if row.Timezone.Valid {
		s := row.Timezone.String
		out.Timezone = &s
	}
	if row.EffectiveFrom.Valid {
		t := row.EffectiveFrom.Time
		out.EffectiveFrom = &t
	}
	if row.EffectiveUntil.Valid {
		t := row.EffectiveUntil.Time
		out.EffectiveUntil = &t
	}
	if row.StartedAt.Valid {
		t := row.StartedAt.Time
		out.StartedAt = &t
	}
	if row.EndedAt.Valid {
		t := row.EndedAt.Time
		out.EndedAt = &t
	}
	return out
}
