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

// ReportSettingsRepositorySQLC is the sqlc-backed single-row report configuration store.
type ReportSettingsRepositorySQLC struct {
	pgQ     *pgsqlc.Queries
	sqliteQ *sqlitesqlc.Queries
}

func NewReportSettingsRepositorySQLC(rt SqlcRuntime) port.ReportSettingsRepository {
	r := &ReportSettingsRepositorySQLC{}
	if pool := rt.PgxPool(); pool != nil {
		r.pgQ = pgsqlc.New(pool)
	} else if db := rt.SQLiteDB(); db != nil {
		r.sqliteQ = sqlitesqlc.New(db)
	}
	return r
}

func (r *ReportSettingsRepositorySQLC) unconfigured() error {
	return fmt.Errorf("report_settings_repository_sqlc: unconfigured runtime")
}

func isNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) || err.Error() == "sql: no rows in result set"
}

func (r *ReportSettingsRepositorySQLC) Get(ctx context.Context) (*domain.ReportSettings, error) {
	switch {
	case r.pgQ != nil:
		row, err := r.pgQ.GetReportSettings(ctx)
		if err != nil {
			if isNoRows(err) {
				return nil, repository.ErrNotFound
			}
			return nil, fmt.Errorf("sqlc: get report settings: %w", err)
		}
		return &domain.ReportSettings{
			Base:           domain.Base{ID: row.ID, CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time},
			Enabled:        row.Enabled,
			RecipientEmail: row.RecipientEmail,
			Schedule:       row.Schedule,
			Scope:          row.Scope,
			LastSentAt:     pgtzPtr(row.LastSentAt),
		}, nil
	case r.sqliteQ != nil:
		row, err := r.sqliteQ.GetReportSettings(ctx)
		if err != nil {
			if isNoRows(err) {
				return nil, repository.ErrNotFound
			}
			return nil, fmt.Errorf("sqlc: get report settings: %w", err)
		}
		return &domain.ReportSettings{
			Base:           domain.Base{ID: row.ID, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt},
			Enabled:        row.Enabled != 0,
			RecipientEmail: row.RecipientEmail,
			Schedule:       row.Schedule,
			Scope:          row.Scope,
			LastSentAt:     nullTimePtr(row.LastSentAt),
		}, nil
	default:
		return nil, r.unconfigured()
	}
}

func (r *ReportSettingsRepositorySQLC) Upsert(ctx context.Context, s *domain.ReportSettings) (*domain.ReportSettings, error) {
	s.ID = domain.ReportSettingsSingletonID
	now := time.Now()
	if s.CreatedAt.IsZero() {
		s.CreatedAt = now
	}
	s.UpdatedAt = now
	switch {
	case r.pgQ != nil:
		_, err := r.pgQ.UpsertReportSettings(ctx, pgsqlc.UpsertReportSettingsParams{
			ID:             s.ID,
			Enabled:        s.Enabled,
			RecipientEmail: s.RecipientEmail,
			Schedule:       s.Schedule,
			Scope:          s.Scope,
			LastSentAt:     ptrPgtz(s.LastSentAt),
			CreatedAt:      pgtype.Timestamptz{Time: s.CreatedAt, Valid: true},
			UpdatedAt:      pgtype.Timestamptz{Time: s.UpdatedAt, Valid: true},
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: upsert report settings: %w", err)
		}
	case r.sqliteQ != nil:
		_, err := r.sqliteQ.UpsertReportSettings(ctx, sqlitesqlc.UpsertReportSettingsParams{
			ID:             s.ID,
			Enabled:        boolToInt64(s.Enabled),
			RecipientEmail: s.RecipientEmail,
			Schedule:       s.Schedule,
			Scope:          s.Scope,
			LastSentAt:     ptrNullTime(s.LastSentAt),
			CreatedAt:      s.CreatedAt,
			UpdatedAt:      s.UpdatedAt,
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: upsert report settings: %w", err)
		}
	default:
		return nil, r.unconfigured()
	}
	return r.Get(ctx)
}

// ── shared conversion helpers (report_* dialect mapping) ──

func pgtzPtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	v := t.Time
	return &v
}

func ptrPgtz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func nullTimePtr(t sql.NullTime) *time.Time {
	if !t.Valid {
		return nil
	}
	v := t.Time
	return &v
}

func ptrNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: *t, Valid: true}
}
