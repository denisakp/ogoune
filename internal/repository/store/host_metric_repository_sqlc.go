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

type HostMetricRepositorySQLC struct {
	pgQ     *pgsqlc.Queries
	sqliteQ *sqlitesqlc.Queries
}

func NewHostMetricRepositorySQLC(rt SqlcRuntime) port.HostMetricsRepository {
	r := &HostMetricRepositorySQLC{}
	if pool := rt.PgxPool(); pool != nil {
		r.pgQ = pgsqlc.New(pool)
	} else if db := rt.SQLiteDB(); db != nil {
		r.sqliteQ = sqlitesqlc.New(db)
	}
	return r
}

func (r *HostMetricRepositorySQLC) unconfigured() error {
	return fmt.Errorf("host_metric_repository_sqlc: unconfigured runtime")
}

func (r *HostMetricRepositorySQLC) Insert(ctx context.Context, s *domain.HostMetricSample) error {
	s.EnsureID()
	if s.SampledAt.IsZero() {
		s.SampledAt = time.Now()
	}
	disks, err := marshalDisks(s.Disks)
	if err != nil {
		return err
	}
	switch {
	case r.pgQ != nil:
		return r.pgQ.InsertHostMetric(ctx, pgsqlc.InsertHostMetricParams{
			ID:        s.ID,
			HostID:    s.HostID,
			SampledAt: pgtype.Timestamptz{Time: s.SampledAt, Valid: true},
			CpuPct:    s.CPUPct,
			MemPct:    s.MemPct,
			NetIn:     s.NetIn,
			NetOut:    s.NetOut,
			Disks:     disks,
		})
	case r.sqliteQ != nil:
		return r.sqliteQ.InsertHostMetric(ctx, sqlitesqlc.InsertHostMetricParams{
			ID:        s.ID,
			HostID:    s.HostID,
			SampledAt: s.SampledAt,
			CpuPct:    s.CPUPct,
			MemPct:    s.MemPct,
			NetIn:     s.NetIn,
			NetOut:    s.NetOut,
			Disks:     nullStringFromBytes(disks),
		})
	default:
		return r.unconfigured()
	}
}

func (r *HostMetricRepositorySQLC) ListInRange(ctx context.Context, hostID string, from, to time.Time) ([]*domain.HostMetricSample, error) {
	switch {
	case r.pgQ != nil:
		rows, err := r.pgQ.ListHostMetricsInRange(ctx, pgsqlc.ListHostMetricsInRangeParams{
			HostID:      hostID,
			SampledAt:   pgtype.Timestamptz{Time: from, Valid: true},
			SampledAt_2: pgtype.Timestamptz{Time: to, Valid: true},
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: list host metrics: %w", err)
		}
		out := make([]*domain.HostMetricSample, 0, len(rows))
		for _, row := range rows {
			s, err := hostMetricFromPG(row)
			if err != nil {
				return nil, err
			}
			out = append(out, s)
		}
		return out, nil
	case r.sqliteQ != nil:
		rows, err := r.sqliteQ.ListHostMetricsInRange(ctx, sqlitesqlc.ListHostMetricsInRangeParams{
			HostID:      hostID,
			SampledAt:   from,
			SampledAt_2: to,
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: list host metrics: %w", err)
		}
		out := make([]*domain.HostMetricSample, 0, len(rows))
		for _, row := range rows {
			s, err := hostMetricFromSQLite(row)
			if err != nil {
				return nil, err
			}
			out = append(out, s)
		}
		return out, nil
	default:
		return nil, r.unconfigured()
	}
}

// AggregateWindow reduces a bounded correlation window to its peaks and sample
// count. One query, one round trip: the database computes the numeric peaks via
// window functions so no numeric column is reduced in Go (spec 089, FR-021), and
// the disk documents ride along on the same rows because nothing in this codebase
// reaches inside stored JSON from SQL on either dialect (FR-021a). The aggregates
// repeat on every row; a few duplicated bytes are cheaper than a second round
// trip on a per-view read.
//
// Returns nil, nil when the window holds no samples. A zero-filled aggregate
// would be indistinguishable from an idle host, so absence is expressed by
// absence (spec 089, FR-010).
func (r *HostMetricRepositorySQLC) AggregateWindow(ctx context.Context, hostID string, from, to time.Time) (*domain.HostMetricsWindowAggregate, error) {
	switch {
	case r.pgQ != nil:
		rows, err := r.pgQ.AggregateHostMetricsInWindow(ctx, pgsqlc.AggregateHostMetricsInWindowParams{
			HostID:      hostID,
			SampledAt:   pgtype.Timestamptz{Time: from, Valid: true},
			SampledAt_2: pgtype.Timestamptz{Time: to, Valid: true},
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: aggregate host metrics: %w", err)
		}
		if len(rows) == 0 {
			return nil, nil
		}
		docs := make([][]byte, 0, len(rows))
		for _, row := range rows {
			if len(row.Disks) > 0 {
				docs = append(docs, row.Disks)
			}
		}
		return buildWindowAggregate(rows[0].PeakCpuPct, rows[0].PeakMemPct, rows[0].SampleCount, docs)
	case r.sqliteQ != nil:
		rows, err := r.sqliteQ.AggregateHostMetricsInWindow(ctx, sqlitesqlc.AggregateHostMetricsInWindowParams{
			HostID:      hostID,
			SampledAt:   from,
			SampledAt_2: to,
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: aggregate host metrics: %w", err)
		}
		if len(rows) == 0 {
			return nil, nil
		}
		// SQLite stores the document as TEXT, so it comes back as a nullable
		// string rather than the JSONB bytes Postgres returns.
		docs := make([][]byte, 0, len(rows))
		for _, row := range rows {
			if row.Disks.Valid && row.Disks.String != "" {
				docs = append(docs, []byte(row.Disks.String))
			}
		}
		return buildWindowAggregate(rows[0].PeakCpuPct, rows[0].PeakMemPct, rows[0].SampleCount, docs)
	default:
		return nil, r.unconfigured()
	}
}

// buildWindowAggregate decodes the disk documents and assembles the aggregate. A
// document that will not decode is skipped rather than failing the whole window:
// one malformed sample must not cost the operator the CPU and memory peaks that
// decoded fine (spec 089, FR-012).
func buildWindowAggregate(peakCPU, peakMem float64, count int64, raw [][]byte) (*domain.HostMetricsWindowAggregate, error) {
	disks := make([][]domain.DiskUsage, 0, len(raw))
	for _, b := range raw {
		d, err := unmarshalDisks(b)
		if err != nil {
			continue
		}
		if len(d) > 0 {
			disks = append(disks, d)
		}
	}
	return &domain.HostMetricsWindowAggregate{
		PeakCPUPct:  peakCPU,
		PeakMemPct:  peakMem,
		SampleCount: int(count),
		Disks:       disks,
	}, nil
}

func (r *HostMetricRepositorySQLC) DeleteOlderThan(ctx context.Context, cutoff time.Time) (int64, error) {
	switch {
	case r.pgQ != nil:
		return r.pgQ.DeleteHostMetricsOlderThan(ctx, pgtype.Timestamptz{Time: cutoff, Valid: true})
	case r.sqliteQ != nil:
		return r.sqliteQ.DeleteHostMetricsOlderThan(ctx, cutoff)
	default:
		return 0, r.unconfigured()
	}
}

func (r *HostMetricRepositorySQLC) DeleteByHost(ctx context.Context, hostID string) error {
	switch {
	case r.pgQ != nil:
		return r.pgQ.DeleteHostMetricsByHost(ctx, hostID)
	case r.sqliteQ != nil:
		return r.sqliteQ.DeleteHostMetricsByHost(ctx, hostID)
	default:
		return r.unconfigured()
	}
}

func (r *HostMetricRepositorySQLC) Decimate(ctx context.Context, cutoff time.Time) (int64, error) {
	switch {
	case r.pgQ != nil:
		return r.pgQ.DecimateHostMetrics(ctx, pgtype.Timestamptz{Time: cutoff, Valid: true})
	case r.sqliteQ != nil:
		return r.sqliteQ.DecimateHostMetrics(ctx, cutoff)
	default:
		return 0, r.unconfigured()
	}
}

// ---------- mappers ----------

func hostMetricFromPG(row pgsqlc.HostMetric) (*domain.HostMetricSample, error) {
	disks, err := unmarshalDisks(row.Disks)
	if err != nil {
		return nil, err
	}
	return &domain.HostMetricSample{
		Base:      domain.Base{ID: row.ID},
		HostID:    row.HostID,
		SampledAt: row.SampledAt.Time,
		CPUPct:    row.CpuPct,
		MemPct:    row.MemPct,
		NetIn:     row.NetIn,
		NetOut:    row.NetOut,
		Disks:     disks,
	}, nil
}

func hostMetricFromSQLite(row sqlitesqlc.HostMetric) (*domain.HostMetricSample, error) {
	disks, err := unmarshalDisks([]byte(row.Disks.String))
	if err != nil {
		return nil, err
	}
	return &domain.HostMetricSample{
		Base:      domain.Base{ID: row.ID},
		HostID:    row.HostID,
		SampledAt: row.SampledAt,
		CPUPct:    row.CpuPct,
		MemPct:    row.MemPct,
		NetIn:     row.NetIn,
		NetOut:    row.NetOut,
		Disks:     disks,
	}, nil
}
