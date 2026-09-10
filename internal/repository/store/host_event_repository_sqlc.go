package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	pgsqlc "github.com/denisakp/ogoune/internal/repository/sqlc/pg"
	sqlitesqlc "github.com/denisakp/ogoune/internal/repository/sqlc/sqlite"
)

// HostEventRepositorySQLC persists kernel events reported by host agents
// (spec 090). Events have their own retention, longer than metrics and never
// thinned (ADR 0011).
type HostEventRepositorySQLC struct {
	pgQ     *pgsqlc.Queries
	sqliteQ *sqlitesqlc.Queries
}

func NewHostEventRepositorySQLC(rt SqlcRuntime) port.HostEventRepository {
	r := &HostEventRepositorySQLC{}
	if pool := rt.PgxPool(); pool != nil {
		r.pgQ = pgsqlc.New(pool)
	} else if db := rt.SQLiteDB(); db != nil {
		r.sqliteQ = sqlitesqlc.New(db)
	}
	return r
}

func (r *HostEventRepositorySQLC) unconfigured() error {
	return fmt.Errorf("host_event_repository_sqlc: unconfigured runtime")
}

func (r *HostEventRepositorySQLC) Create(ctx context.Context, e *domain.HostEvent) error {
	e.EnsureID()
	if e.OccurredAt.IsZero() {
		e.OccurredAt = time.Now().UTC()
	}
	if e.Occurrences < 1 {
		e.Occurrences = 1
	}

	detail, err := marshalEventDetail(e.Detail)
	if err != nil {
		return err
	}

	switch {
	case r.pgQ != nil:
		return r.pgQ.InsertHostEvent(ctx, pgsqlc.InsertHostEventParams{
			ID:          e.ID,
			HostID:      e.HostID,
			OccurredAt:  pgtype.Timestamptz{Time: e.OccurredAt, Valid: true},
			Kind:        e.Kind,
			Source:      e.Source,
			Occurrences: int32(e.Occurrences),
			Detail:      detail,
		})
	case r.sqliteQ != nil:
		return r.sqliteQ.InsertHostEvent(ctx, sqlitesqlc.InsertHostEventParams{
			ID:          e.ID,
			HostID:      e.HostID,
			OccurredAt:  e.OccurredAt,
			Kind:        e.Kind,
			Source:      e.Source,
			Occurrences: int64(e.Occurrences),
			Detail:      nullStringFromBytes(detail),
		})
	default:
		return r.unconfigured()
	}
}

func (r *HostEventRepositorySQLC) ListByHost(ctx context.Context, hostID string, limit int) ([]*domain.HostEvent, error) {
	if limit <= 0 {
		limit = 50
	}

	switch {
	case r.pgQ != nil:
		rows, err := r.pgQ.ListHostEventsByHost(ctx, pgsqlc.ListHostEventsByHostParams{
			HostID: hostID,
			Limit:  int32(limit),
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: list host events: %w", err)
		}
		out := make([]*domain.HostEvent, 0, len(rows))
		for _, row := range rows {
			e, err := hostEventFromPG(row)
			if err != nil {
				return nil, err
			}
			out = append(out, e)
		}
		return out, nil
	case r.sqliteQ != nil:
		rows, err := r.sqliteQ.ListHostEventsByHost(ctx, sqlitesqlc.ListHostEventsByHostParams{
			HostID: hostID,
			Limit:  int64(limit),
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: list host events: %w", err)
		}
		out := make([]*domain.HostEvent, 0, len(rows))
		for _, row := range rows {
			e, err := hostEventFromSQLite(row)
			if err != nil {
				return nil, err
			}
			out = append(out, e)
		}
		return out, nil
	default:
		return nil, r.unconfigured()
	}
}

// ListInWindow returns the host's events inside [from, to), newest first
// Half-open, matching how the incident host context treats the same
// window: an event exactly on the far edge belongs to one window and not two.
func (r *HostEventRepositorySQLC) ListInWindow(ctx context.Context, hostID string, from, to time.Time, limit int) ([]*domain.HostEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	// An empty or inverted window has no events by definition, and asking the
	// database is a query spent to learn nothing.
	if !to.After(from) {
		return nil, nil
	}

	switch {
	case r.pgQ != nil:
		rows, err := r.pgQ.ListHostEventsInWindow(ctx, pgsqlc.ListHostEventsInWindowParams{
			HostID:       hostID,
			OccurredAt:   pgtype.Timestamptz{Time: from, Valid: true},
			OccurredAt_2: pgtype.Timestamptz{Time: to, Valid: true},
			Limit:        int32(limit),
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: list host events in window: %w", err)
		}
		out := make([]*domain.HostEvent, 0, len(rows))
		for _, row := range rows {
			e, err := hostEventFromPG(row)
			if err != nil {
				return nil, err
			}
			out = append(out, e)
		}
		return out, nil
	case r.sqliteQ != nil:
		rows, err := r.sqliteQ.ListHostEventsInWindow(ctx, sqlitesqlc.ListHostEventsInWindowParams{
			HostID:       hostID,
			OccurredAt:   from,
			OccurredAt_2: to,
			Limit:        int64(limit),
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: list host events in window: %w", err)
		}
		out := make([]*domain.HostEvent, 0, len(rows))
		for _, row := range rows {
			e, err := hostEventFromSQLite(row)
			if err != nil {
				return nil, err
			}
			out = append(out, e)
		}
		return out, nil
	default:
		return nil, r.unconfigured()
	}
}

func (r *HostEventRepositorySQLC) DeleteOlderThan(ctx context.Context, cutoff time.Time) (int64, error) {
	switch {
	case r.pgQ != nil:
		return r.pgQ.DeleteHostEventsOlderThan(ctx, pgtype.Timestamptz{Time: cutoff, Valid: true})
	case r.sqliteQ != nil:
		return r.sqliteQ.DeleteHostEventsOlderThan(ctx, cutoff)
	default:
		return 0, r.unconfigured()
	}
}

func (r *HostEventRepositorySQLC) DeleteByHost(ctx context.Context, hostID string) error {
	switch {
	case r.pgQ != nil:
		return r.pgQ.DeleteHostEventsByHost(ctx, hostID)
	case r.sqliteQ != nil:
		return r.sqliteQ.DeleteHostEventsByHost(ctx, hostID)
	default:
		return r.unconfigured()
	}
}

// marshalEventDetail serialises the classified detail. A nil detail stores NULL
// rather than an empty object: a kernel report that named nothing and one whose
// detail was lost must not look alike.
func marshalEventDetail(d *domain.HostEventDetail) ([]byte, error) {
	if d == nil {
		return nil, nil
	}
	b, err := json.Marshal(d)
	if err != nil {
		return nil, fmt.Errorf("marshal host event detail: %w", err)
	}
	return b, nil
}

func unmarshalEventDetail(b []byte) (*domain.HostEventDetail, error) {
	if len(b) == 0 {
		return nil, nil
	}
	var d domain.HostEventDetail
	if err := json.Unmarshal(b, &d); err != nil {
		return nil, fmt.Errorf("unmarshal host event detail: %w", err)
	}
	return &d, nil
}

func hostEventFromPG(row pgsqlc.HostEvent) (*domain.HostEvent, error) {
	d, err := unmarshalEventDetail(row.Detail)
	if err != nil {
		return nil, err
	}
	return &domain.HostEvent{
		Base:        domain.Base{ID: row.ID},
		HostID:      row.HostID,
		OccurredAt:  row.OccurredAt.Time,
		Kind:        row.Kind,
		Source:      row.Source,
		Occurrences: int(row.Occurrences),
		Detail:      d,
	}, nil
}

func hostEventFromSQLite(row sqlitesqlc.HostEvent) (*domain.HostEvent, error) {
	var raw []byte
	if row.Detail.Valid {
		raw = []byte(row.Detail.String)
	}
	d, err := unmarshalEventDetail(raw)
	if err != nil {
		return nil, err
	}
	return &domain.HostEvent{
		Base:        domain.Base{ID: row.ID},
		HostID:      row.HostID,
		OccurredAt:  row.OccurredAt,
		Kind:        row.Kind,
		Source:      row.Source,
		Occurrences: int(row.Occurrences),
		Detail:      d,
	}, nil
}
