package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository"
	pgsqlc "github.com/denisakp/ogoune/internal/repository/sqlc/pg"
	sqlitesqlc "github.com/denisakp/ogoune/internal/repository/sqlc/sqlite"
)

// ReportHistoryRepositorySQLC is the sqlc-backed generated-report store.
type ReportHistoryRepositorySQLC struct {
	pgQ     *pgsqlc.Queries
	sqliteQ *sqlitesqlc.Queries
}

func NewReportHistoryRepositorySQLC(rt SqlcRuntime) port.ReportHistoryRepository {
	r := &ReportHistoryRepositorySQLC{}
	if pool := rt.PgxPool(); pool != nil {
		r.pgQ = pgsqlc.New(pool)
	} else if db := rt.SQLiteDB(); db != nil {
		r.sqliteQ = sqlitesqlc.New(db)
	}
	return r
}

func (r *ReportHistoryRepositorySQLC) unconfigured() error {
	return fmt.Errorf("report_history_repository_sqlc: unconfigured runtime")
}

func marshalBreakdown(b []domain.ReportBreakdownLine) ([]byte, error) {
	if b == nil {
		b = []domain.ReportBreakdownLine{}
	}
	out, err := json.Marshal(b)
	if err != nil {
		return nil, fmt.Errorf("marshal breakdown: %w", err)
	}
	return out, nil
}

func unmarshalBreakdown(raw []byte) ([]domain.ReportBreakdownLine, error) {
	out := []domain.ReportBreakdownLine{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &out); err != nil {
			return nil, fmt.Errorf("unmarshal breakdown: %w", err)
		}
	}
	return out, nil
}

func (r *ReportHistoryRepositorySQLC) Create(ctx context.Context, h *domain.ReportHistory) (*domain.ReportHistory, error) {
	h.EnsureID()
	now := time.Now()
	if h.CreatedAt.IsZero() {
		h.CreatedAt = now
	}
	if h.SentAt.IsZero() {
		h.SentAt = now
	}
	bd, err := marshalBreakdown(h.Breakdown)
	if err != nil {
		return nil, err
	}
	switch {
	case r.pgQ != nil:
		row, err := r.pgQ.CreateReportHistory(ctx, pgsqlc.CreateReportHistoryParams{
			ID:                h.ID,
			Period:            h.Period,
			SentAt:            pgtype.Timestamptz{Time: h.SentAt, Valid: true},
			Status:            string(h.Status),
			UptimePct:         h.UptimePct,
			IncidentCount:     int32(h.IncidentCount),
			DowntimeSeconds:   h.DowntimeSeconds,
			RecipientEmail:    h.RecipientEmail,
			ResourceBreakdown: bd,
			CreatedAt:         pgtype.Timestamptz{Time: h.CreatedAt, Valid: true},
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: create report history: %w", err)
		}
		return reportHistoryFromPG(row)
	case r.sqliteQ != nil:
		row, err := r.sqliteQ.CreateReportHistory(ctx, sqlitesqlc.CreateReportHistoryParams{
			ID:                h.ID,
			Period:            h.Period,
			SentAt:            h.SentAt,
			Status:            string(h.Status),
			UptimePct:         h.UptimePct,
			IncidentCount:     int64(h.IncidentCount),
			DowntimeSeconds:   h.DowntimeSeconds,
			RecipientEmail:    h.RecipientEmail,
			ResourceBreakdown: string(bd),
			CreatedAt:         h.CreatedAt,
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: create report history: %w", err)
		}
		return reportHistoryFromSQLite(row)
	default:
		return nil, r.unconfigured()
	}
}

func (r *ReportHistoryRepositorySQLC) ListRecent(ctx context.Context, limit int) ([]*domain.ReportHistory, error) {
	switch {
	case r.pgQ != nil:
		rows, err := r.pgQ.ListRecentReportHistory(ctx, int32(limit))
		if err != nil {
			return nil, fmt.Errorf("sqlc: list report history: %w", err)
		}
		out := make([]*domain.ReportHistory, 0, len(rows))
		for _, row := range rows {
			h, err := reportHistoryFromPG(row)
			if err != nil {
				return nil, err
			}
			out = append(out, h)
		}
		return out, nil
	case r.sqliteQ != nil:
		rows, err := r.sqliteQ.ListRecentReportHistory(ctx, int64(limit))
		if err != nil {
			return nil, fmt.Errorf("sqlc: list report history: %w", err)
		}
		out := make([]*domain.ReportHistory, 0, len(rows))
		for _, row := range rows {
			h, err := reportHistoryFromSQLite(row)
			if err != nil {
				return nil, err
			}
			out = append(out, h)
		}
		return out, nil
	default:
		return nil, r.unconfigured()
	}
}

func (r *ReportHistoryRepositorySQLC) FindByPeriod(ctx context.Context, period string) (*domain.ReportHistory, error) {
	switch {
	case r.pgQ != nil:
		row, err := r.pgQ.FindReportHistoryByPeriod(ctx, period)
		if err != nil {
			if isNoRows(err) {
				return nil, repository.ErrNotFound
			}
			return nil, fmt.Errorf("sqlc: find report by period: %w", err)
		}
		return reportHistoryFromPG(row)
	case r.sqliteQ != nil:
		row, err := r.sqliteQ.FindReportHistoryByPeriod(ctx, period)
		if err != nil {
			if isNoRows(err) {
				return nil, repository.ErrNotFound
			}
			return nil, fmt.Errorf("sqlc: find report by period: %w", err)
		}
		return reportHistoryFromSQLite(row)
	default:
		return nil, r.unconfigured()
	}
}

func reportHistoryFromPG(row pgsqlc.ReportHistory) (*domain.ReportHistory, error) {
	bd, err := unmarshalBreakdown(row.ResourceBreakdown)
	if err != nil {
		return nil, err
	}
	return &domain.ReportHistory{
		Base:            domain.Base{ID: row.ID, CreatedAt: row.CreatedAt.Time},
		Period:          row.Period,
		SentAt:          row.SentAt.Time,
		Status:          domain.ReportStatus(row.Status),
		UptimePct:       row.UptimePct,
		IncidentCount:   int(row.IncidentCount),
		DowntimeSeconds: row.DowntimeSeconds,
		RecipientEmail:  row.RecipientEmail,
		Breakdown:       bd,
	}, nil
}

func reportHistoryFromSQLite(row sqlitesqlc.ReportHistory) (*domain.ReportHistory, error) {
	bd, err := unmarshalBreakdown([]byte(row.ResourceBreakdown))
	if err != nil {
		return nil, err
	}
	return &domain.ReportHistory{
		Base:            domain.Base{ID: row.ID, CreatedAt: row.CreatedAt},
		Period:          row.Period,
		SentAt:          row.SentAt,
		Status:          domain.ReportStatus(row.Status),
		UptimePct:       row.UptimePct,
		IncidentCount:   int(row.IncidentCount),
		DowntimeSeconds: row.DowntimeSeconds,
		RecipientEmail:  row.RecipientEmail,
		Breakdown:       bd,
	}, nil
}
