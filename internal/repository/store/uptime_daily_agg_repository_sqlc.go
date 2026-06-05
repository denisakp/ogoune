package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	pgsqlc "github.com/denisakp/ogoune/internal/repository/sqlc/pg"
	sqlitesqlc "github.com/denisakp/ogoune/internal/repository/sqlc/sqlite"
)

const dayLayout = "2006-01-02"

type UptimeDailyAggRepositorySQLC struct {
	pgQ     *pgsqlc.Queries
	sqliteQ *sqlitesqlc.Queries
}

func NewUptimeDailyAggRepositorySQLC(rt SqlcRuntime) port.UptimeDailyAggRepository {
	r := &UptimeDailyAggRepositorySQLC{}
	if pool := rt.PgxPool(); pool != nil {
		r.pgQ = pgsqlc.New(pool)
	} else if db := rt.SQLiteDB(); db != nil {
		r.sqliteQ = sqlitesqlc.New(db)
	}
	return r
}

func (r *UptimeDailyAggRepositorySQLC) unconfigured() error {
	return fmt.Errorf("uptime_daily_agg_sqlc: unconfigured runtime")
}

func truncDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func numericFromFloat(v float64) pgtype.Numeric {
	var n pgtype.Numeric
	_ = n.Scan(fmt.Sprintf("%.4f", v))
	return n
}

func floatFromNumeric(n pgtype.Numeric) float64 {
	f, err := n.Float64Value()
	if err != nil || !f.Valid {
		return 0
	}
	return f.Float64
}

func (r *UptimeDailyAggRepositorySQLC) Upsert(ctx context.Context, agg *domain.UptimeDailyAgg) error {
	if agg.ComputedAt.IsZero() {
		agg.ComputedAt = time.Now()
	}
	day := truncDay(agg.Day)
	switch {
	case r.pgQ != nil:
		return r.pgQ.UpsertUptimeDailyAgg(ctx, pgsqlc.UpsertUptimeDailyAggParams{
			ResourceID:  agg.ResourceID,
			Day:         pgtype.Date{Time: day, Valid: true},
			Samples:     int32(agg.Samples),
			Up:          int32(agg.Up),
			Degraded:    int32(agg.Degraded),
			Down:        int32(agg.Down),
			UptimeRatio: numericFromFloat(agg.UptimeRatio),
			ComputedAt:  pgtype.Timestamptz{Time: agg.ComputedAt, Valid: true},
		})
	case r.sqliteQ != nil:
		return r.sqliteQ.UpsertUptimeDailyAgg(ctx, sqlitesqlc.UpsertUptimeDailyAggParams{
			ResourceID:  agg.ResourceID,
			Day:         day.Format(dayLayout),
			Samples:     int64(agg.Samples),
			Up:          int64(agg.Up),
			Degraded:    int64(agg.Degraded),
			Down:        int64(agg.Down),
			UptimeRatio: agg.UptimeRatio,
			ComputedAt:  agg.ComputedAt,
		})
	default:
		return r.unconfigured()
	}
}

func (r *UptimeDailyAggRepositorySQLC) FindRange(ctx context.Context, resourceIDs []string, from, to time.Time) ([]*domain.UptimeDailyAgg, error) {
	if len(resourceIDs) == 0 {
		return []*domain.UptimeDailyAgg{}, nil
	}
	fromDay := truncDay(from)
	toDay := truncDay(to)
	switch {
	case r.pgQ != nil:
		rows, err := r.pgQ.FindUptimeDailyAggRange(ctx, pgsqlc.FindUptimeDailyAggRangeParams{
			ResourceIds: resourceIDs,
			FromDay:     pgtype.Date{Time: fromDay, Valid: true},
			ToDay:       pgtype.Date{Time: toDay, Valid: true},
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: find uptime daily agg range: %w", err)
		}
		out := make([]*domain.UptimeDailyAgg, 0, len(rows))
		for _, row := range rows {
			out = append(out, uptimeDailyAggFromPG(row))
		}
		return out, nil
	case r.sqliteQ != nil:
		fromStr := fromDay.Format(dayLayout)
		toStr := toDay.Format(dayLayout)
		out := make([]*domain.UptimeDailyAgg, 0)
		for _, rid := range resourceIDs {
			rows, err := r.sqliteQ.FindUptimeDailyAggForResource(ctx, sqlitesqlc.FindUptimeDailyAggForResourceParams{
				ResourceID: rid,
				FromDay:    fromStr,
				ToDay:      toStr,
			})
			if err != nil {
				return nil, fmt.Errorf("sqlc: find uptime daily agg for resource: %w", err)
			}
			for _, row := range rows {
				out = append(out, uptimeDailyAggFromSQLite(row))
			}
		}
		return out, nil
	default:
		return nil, r.unconfigured()
	}
}

func (r *UptimeDailyAggRepositorySQLC) FindForResource(ctx context.Context, resourceID string, from, to time.Time) ([]*domain.UptimeDailyAgg, error) {
	fromDay := truncDay(from)
	toDay := truncDay(to)
	switch {
	case r.pgQ != nil:
		rows, err := r.pgQ.FindUptimeDailyAggForResource(ctx, pgsqlc.FindUptimeDailyAggForResourceParams{
			ResourceID: resourceID,
			Day:        pgtype.Date{Time: fromDay, Valid: true},
			Day_2:      pgtype.Date{Time: toDay, Valid: true},
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: find uptime daily agg for resource: %w", err)
		}
		out := make([]*domain.UptimeDailyAgg, 0, len(rows))
		for _, row := range rows {
			out = append(out, uptimeDailyAggFromPG(row))
		}
		return out, nil
	case r.sqliteQ != nil:
		rows, err := r.sqliteQ.FindUptimeDailyAggForResource(ctx, sqlitesqlc.FindUptimeDailyAggForResourceParams{
			ResourceID: resourceID,
			FromDay:    fromDay.Format(dayLayout),
			ToDay:      toDay.Format(dayLayout),
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: find uptime daily agg for resource: %w", err)
		}
		out := make([]*domain.UptimeDailyAgg, 0, len(rows))
		for _, row := range rows {
			out = append(out, uptimeDailyAggFromSQLite(row))
		}
		return out, nil
	default:
		return nil, r.unconfigured()
	}
}

func uptimeDailyAggFromPG(row pgsqlc.UptimeDailyAgg) *domain.UptimeDailyAgg {
	return &domain.UptimeDailyAgg{
		ResourceID:  row.ResourceID,
		Day:         row.Day.Time,
		Samples:     int(row.Samples),
		Up:          int(row.Up),
		Degraded:    int(row.Degraded),
		Down:        int(row.Down),
		UptimeRatio: floatFromNumeric(row.UptimeRatio),
		ComputedAt:  row.ComputedAt.Time,
	}
}

func uptimeDailyAggFromSQLite(row sqlitesqlc.UptimeDailyAgg) *domain.UptimeDailyAgg {
	day, _ := time.Parse(dayLayout, row.Day)
	return &domain.UptimeDailyAgg{
		ResourceID:  row.ResourceID,
		Day:         day,
		Samples:     int(row.Samples),
		Up:          int(row.Up),
		Degraded:    int(row.Degraded),
		Down:        int(row.Down),
		UptimeRatio: row.UptimeRatio,
		ComputedAt:  row.ComputedAt,
	}
}
