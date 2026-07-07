package store

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository"
	pgsqlc "github.com/denisakp/ogoune/internal/repository/sqlc/pg"
	sqlitesqlc "github.com/denisakp/ogoune/internal/repository/sqlc/sqlite"
)

type MonitoringActivityRepositorySQLC struct {
	pgQ     *pgsqlc.Queries
	sqliteQ *sqlitesqlc.Queries
}

func NewMonitoringActivityRepositorySQLC(rt SqlcRuntime) port.MonitoringActivityRepository {
	r := &MonitoringActivityRepositorySQLC{}
	if pool := rt.PgxPool(); pool != nil {
		r.pgQ = pgsqlc.New(pool)
	} else if db := rt.SQLiteDB(); db != nil {
		r.sqliteQ = sqlitesqlc.New(db)
	}
	return r
}

func (r *MonitoringActivityRepositorySQLC) unconfigured() error {
	return fmt.Errorf("monitoring_activity_repository_sqlc: unconfigured runtime")
}

func (r *MonitoringActivityRepositorySQLC) Create(ctx context.Context, a *domain.MonitoringActivity) error {
	if a == nil {
		return repository.ErrInvalidInput
	}
	a.EnsureID()
	now := time.Now()
	if a.CreatedAt.IsZero() {
		a.CreatedAt = now
	}
	if a.UpdatedAt.IsZero() {
		a.UpdatedAt = now
	}
	switch {
	case r.pgQ != nil:
		return r.pgQ.CreateMonitoringActivity(ctx, pgsqlc.CreateMonitoringActivityParams{
			ID:            a.ID,
			CreatedAt:     pgtype.Timestamptz{Time: a.CreatedAt, Valid: true},
			UpdatedAt:     pgtype.Timestamptz{Time: a.UpdatedAt, Valid: true},
			ResourceID:    a.ResourceID,
			Message:       a.Message,
			Success:       a.Success,
			ResponseTime:  int32(a.ResponseTime),
			ResponseData:  a.ResponseData,
			IsMaintenance: a.IsMaintenance,
		})
	case r.sqliteQ != nil:
		return r.sqliteQ.CreateMonitoringActivity(ctx, sqlitesqlc.CreateMonitoringActivityParams{
			ID:            a.ID,
			CreatedAt:     a.CreatedAt,
			UpdatedAt:     a.UpdatedAt,
			ResourceID:    a.ResourceID,
			Message:       a.Message,
			Success:       boolToInt64(a.Success),
			ResponseTime:  int64(a.ResponseTime),
			ResponseData:  a.ResponseData,
			IsMaintenance: boolToInt64(a.IsMaintenance),
		})
	default:
		return r.unconfigured()
	}
}

func (r *MonitoringActivityRepositorySQLC) List(ctx context.Context, limit, offset int) ([]*domain.MonitoringActivity, error) {
	switch {
	case r.pgQ != nil:
		rows, err := r.pgQ.ListMonitoringActivities(ctx, pgsqlc.ListMonitoringActivitiesParams{
			Limit: int32(limit), Offset: int32(offset),
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: list monitoring activities: %w", err)
		}
		out := make([]*domain.MonitoringActivity, len(rows))
		for i, row := range rows {
			out[i] = monitoringActivityFromPG(row)
		}
		return out, nil
	case r.sqliteQ != nil:
		rows, err := r.sqliteQ.ListMonitoringActivities(ctx, sqlitesqlc.ListMonitoringActivitiesParams{
			Limit: int64(limit), Offset: int64(offset),
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: list monitoring activities: %w", err)
		}
		out := make([]*domain.MonitoringActivity, len(rows))
		for i, row := range rows {
			out[i] = monitoringActivityFromSQLite(row)
		}
		return out, nil
	default:
		return nil, r.unconfigured()
	}
}

func (r *MonitoringActivityRepositorySQLC) FindByResourceID(ctx context.Context, resourceID string, limit, offset int) ([]*domain.MonitoringActivity, error) {
	if resourceID == "" {
		return nil, repository.ErrInvalidInput
	}
	switch {
	case r.pgQ != nil:
		rows, err := r.pgQ.FindMonitoringActivitiesByResourceID(ctx, pgsqlc.FindMonitoringActivitiesByResourceIDParams{
			ResourceID: resourceID, Limit: int32(limit), Offset: int32(offset),
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: find by resource: %w", err)
		}
		out := make([]*domain.MonitoringActivity, len(rows))
		for i, row := range rows {
			out[i] = monitoringActivityFromPG(row)
		}
		return out, nil
	case r.sqliteQ != nil:
		rows, err := r.sqliteQ.FindMonitoringActivitiesByResourceID(ctx, sqlitesqlc.FindMonitoringActivitiesByResourceIDParams{
			ResourceID: resourceID, Limit: int64(limit), Offset: int64(offset),
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: find by resource: %w", err)
		}
		out := make([]*domain.MonitoringActivity, len(rows))
		for i, row := range rows {
			out[i] = monitoringActivityFromSQLite(row)
		}
		return out, nil
	default:
		return nil, r.unconfigured()
	}
}

// CountTransitionsInWindow uses the same Go-side aggregation as the GORM impl
// to guarantee numerical parity: SELECT success rows in window, count adjacent
// changes in Go.
func (r *MonitoringActivityRepositorySQLC) CountTransitionsInWindow(ctx context.Context, resourceID string, windowStart time.Time) (int, error) {
	if resourceID == "" {
		return 0, repository.ErrInvalidInput
	}
	var successes []bool
	switch {
	case r.pgQ != nil:
		rows, err := r.pgQ.SelectMonitoringActivitySuccessInWindow(ctx, pgsqlc.SelectMonitoringActivitySuccessInWindowParams{
			ResourceID: resourceID,
			CreatedAt:  pgtype.Timestamptz{Time: windowStart, Valid: true},
		})
		if err != nil {
			return 0, fmt.Errorf("sqlc: count transitions: %w", err)
		}
		successes = rows
	case r.sqliteQ != nil:
		rows, err := r.sqliteQ.SelectMonitoringActivitySuccessInWindow(ctx, sqlitesqlc.SelectMonitoringActivitySuccessInWindowParams{
			ResourceID: resourceID,
			CreatedAt:  windowStart,
		})
		if err != nil {
			return 0, fmt.Errorf("sqlc: count transitions: %w", err)
		}
		successes = make([]bool, len(rows))
		for i, v := range rows {
			successes[i] = v != 0
		}
	default:
		return 0, r.unconfigured()
	}
	transitions := 0
	for i := 1; i < len(successes); i++ {
		if successes[i] != successes[i-1] {
			transitions++
		}
	}
	return transitions, nil
}

// GetUptimeStats does Go-side hourly bucketing — same as GORM impl, identical
// numerical results across dialects.
func (r *MonitoringActivityRepositorySQLC) GetUptimeStats(ctx context.Context, resourceID string) ([]domain.UptimeStat, error) {
	if resourceID == "" {
		return nil, repository.ErrInvalidInput
	}
	since := time.Now().Add(-24 * time.Hour)
	type pt struct {
		t time.Time
		s bool
	}
	var pts []pt
	switch {
	case r.pgQ != nil:
		rows, err := r.pgQ.SelectMonitoringActivityHourlyAggregateInputs(ctx, pgsqlc.SelectMonitoringActivityHourlyAggregateInputsParams{
			ResourceID: resourceID,
			CreatedAt:  pgtype.Timestamptz{Time: since, Valid: true},
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: get uptime stats: %w", err)
		}
		pts = make([]pt, len(rows))
		for i, r := range rows {
			pts[i] = pt{t: r.CreatedAt.Time, s: r.Success}
		}
	case r.sqliteQ != nil:
		rows, err := r.sqliteQ.SelectMonitoringActivityHourlyAggregateInputs(ctx, sqlitesqlc.SelectMonitoringActivityHourlyAggregateInputsParams{
			ResourceID: resourceID,
			CreatedAt:  since,
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: get uptime stats: %w", err)
		}
		pts = make([]pt, len(rows))
		for i, r := range rows {
			pts[i] = pt{t: r.CreatedAt, s: r.Success != 0}
		}
	default:
		return nil, r.unconfigured()
	}
	type agg struct{ s, total int }
	byHour := map[time.Time]agg{}
	for _, p := range pts {
		h := p.t.Truncate(time.Hour)
		a := byHour[h]
		if p.s {
			a.s++
		}
		a.total++
		byHour[h] = a
	}
	hours := make([]time.Time, 0, len(byHour))
	for h := range byHour {
		hours = append(hours, h)
	}
	sort.Slice(hours, func(i, j int) bool { return hours[i].Before(hours[j]) })
	out := make([]domain.UptimeStat, 0, len(hours))
	for _, h := range hours {
		v := byHour[h]
		uptime := 0.0
		if v.total > 0 {
			uptime = math.Round((float64(v.s)/float64(v.total)*100)*100) / 100
		}
		out = append(out, domain.UptimeStat{
			Hour: h, UptimePercent: uptime, SuccessfulCount: v.s, TotalCount: v.total,
		})
	}
	return out, nil
}

func (r *MonitoringActivityRepositorySQLC) GetRecentResponseTimes(ctx context.Context, resourceID string, limit int) ([]domain.ResponseTimePoint, error) {
	if resourceID == "" {
		return nil, repository.ErrInvalidInput
	}
	if limit <= 0 {
		limit = 50
	}
	var out []domain.ResponseTimePoint
	switch {
	case r.pgQ != nil:
		rows, err := r.pgQ.GetRecentResponseTimes(ctx, pgsqlc.GetRecentResponseTimesParams{
			ResourceID: resourceID, Limit: int32(limit),
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: recent response times: %w", err)
		}
		out = make([]domain.ResponseTimePoint, len(rows))
		for i, r := range rows {
			out[i] = domain.ResponseTimePoint{Timestamp: r.CreatedAt.Time, ResponseTime: int(r.ResponseTime)}
		}
	case r.sqliteQ != nil:
		rows, err := r.sqliteQ.GetRecentResponseTimes(ctx, sqlitesqlc.GetRecentResponseTimesParams{
			ResourceID: resourceID, Limit: int64(limit),
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: recent response times: %w", err)
		}
		out = make([]domain.ResponseTimePoint, len(rows))
		for i, r := range rows {
			out[i] = domain.ResponseTimePoint{Timestamp: r.CreatedAt, ResponseTime: int(r.ResponseTime)}
		}
	default:
		return nil, r.unconfigured()
	}
	// Mirror GORM impl: reverse to chronological order.
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out, nil
}

func (r *MonitoringActivityRepositorySQLC) GetGlobalUptimeStats(ctx context.Context, hours int) (float64, error) {
	since := time.Now().Add(-time.Duration(hours) * time.Hour)
	switch {
	case r.pgQ != nil:
		total, err := r.pgQ.CountMonitoringActivitySinceTotal(ctx, pgtype.Timestamptz{Time: since, Valid: true})
		if err != nil {
			return 0, fmt.Errorf("sqlc: global uptime total: %w", err)
		}
		if total == 0 {
			return 0, nil
		}
		successful, err := r.pgQ.CountMonitoringActivitySinceSuccess(ctx, pgtype.Timestamptz{Time: since, Valid: true})
		if err != nil {
			return 0, fmt.Errorf("sqlc: global uptime success: %w", err)
		}
		return math.Round((float64(successful)/float64(total)*100)*100) / 100, nil
	case r.sqliteQ != nil:
		total, err := r.sqliteQ.CountMonitoringActivitySinceTotal(ctx, since)
		if err != nil {
			return 0, fmt.Errorf("sqlc: global uptime total: %w", err)
		}
		if total == 0 {
			return 0, nil
		}
		successful, err := r.sqliteQ.CountMonitoringActivitySinceSuccess(ctx, since)
		if err != nil {
			return 0, fmt.Errorf("sqlc: global uptime success: %w", err)
		}
		return math.Round((float64(successful)/float64(total)*100)*100) / 100, nil
	default:
		return 0, r.unconfigured()
	}
}

func (r *MonitoringActivityRepositorySQLC) GetUptimeByWindow(ctx context.Context, resourceID string, hours int) (*float64, error) {
	if resourceID == "" {
		return nil, repository.ErrInvalidInput
	}
	since := time.Now().Add(-time.Duration(hours) * time.Hour)
	switch {
	case r.pgQ != nil:
		total, err := r.pgQ.CountMonitoringActivityByResourceTotal(ctx, pgsqlc.CountMonitoringActivityByResourceTotalParams{
			ResourceID: resourceID, CreatedAt: pgtype.Timestamptz{Time: since, Valid: true},
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: uptime window total: %w", err)
		}
		if total == 0 {
			return nil, nil
		}
		successful, err := r.pgQ.CountMonitoringActivityByResourceSuccess(ctx, pgsqlc.CountMonitoringActivityByResourceSuccessParams{
			ResourceID: resourceID, CreatedAt: pgtype.Timestamptz{Time: since, Valid: true},
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: uptime window success: %w", err)
		}
		uptime := math.Round((float64(successful)/float64(total)*100)*100) / 100
		return &uptime, nil
	case r.sqliteQ != nil:
		total, err := r.sqliteQ.CountMonitoringActivityByResourceTotal(ctx, sqlitesqlc.CountMonitoringActivityByResourceTotalParams{
			ResourceID: resourceID, CreatedAt: since,
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: uptime window total: %w", err)
		}
		if total == 0 {
			return nil, nil
		}
		successful, err := r.sqliteQ.CountMonitoringActivityByResourceSuccess(ctx, sqlitesqlc.CountMonitoringActivityByResourceSuccessParams{
			ResourceID: resourceID, CreatedAt: since,
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: uptime window success: %w", err)
		}
		uptime := math.Round((float64(successful)/float64(total)*100)*100) / 100
		return &uptime, nil
	default:
		return nil, r.unconfigured()
	}
}

func (r *MonitoringActivityRepositorySQLC) GetAvgResponseTimeByWindow(ctx context.Context, resourceID string, hours int) (*int, error) {
	if resourceID == "" {
		return nil, repository.ErrInvalidInput
	}
	since := time.Now().Add(-time.Duration(hours) * time.Hour)
	switch {
	case r.pgQ != nil:
		// Check COUNT first so we don't scan NULL from AVG on empty window.
		successful, err := r.pgQ.CountMonitoringActivityByResourceSuccess(ctx, pgsqlc.CountMonitoringActivityByResourceSuccessParams{
			ResourceID: resourceID, CreatedAt: pgtype.Timestamptz{Time: since, Valid: true},
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: avg response time count: %w", err)
		}
		if successful == 0 {
			return nil, nil
		}
		avg, err := r.pgQ.AvgResponseTimeByResourceInWindow(ctx, pgsqlc.AvgResponseTimeByResourceInWindowParams{
			ResourceID: resourceID, CreatedAt: pgtype.Timestamptz{Time: since, Valid: true},
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: avg response time: %w", err)
		}
		v := int(math.Round(avg))
		return &v, nil
	case r.sqliteQ != nil:
		avg, err := r.sqliteQ.AvgResponseTimeByResourceInWindow(ctx, sqlitesqlc.AvgResponseTimeByResourceInWindowParams{
			ResourceID: resourceID, CreatedAt: since,
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: avg response time: %w", err)
		}
		if !avg.Valid {
			return nil, nil
		}
		v := int(math.Round(avg.Float64))
		return &v, nil
	default:
		return nil, r.unconfigured()
	}
}

// ---------- mapping helpers ----------

func monitoringActivityFromPG(row pgsqlc.MonitoringActivity) *domain.MonitoringActivity {
	return &domain.MonitoringActivity{
		Base: domain.Base{
			ID:        row.ID,
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
		},
		ResourceID:    row.ResourceID,
		Message:       row.Message,
		Success:       row.Success,
		ResponseTime:  int(row.ResponseTime),
		ResponseData:  row.ResponseData,
		IsMaintenance: row.IsMaintenance,
	}
}

func monitoringActivityFromSQLite(row sqlitesqlc.MonitoringActivity) *domain.MonitoringActivity {
	return &domain.MonitoringActivity{
		Base: domain.Base{
			ID:        row.ID,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		ResourceID:    row.ResourceID,
		Message:       row.Message,
		Success:       row.Success != 0,
		ResponseTime:  int(row.ResponseTime),
		ResponseData:  row.ResponseData,
		IsMaintenance: row.IsMaintenance != 0,
	}
}
