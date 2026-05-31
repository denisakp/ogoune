package store

import (
	"context"
	"database/sql"
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

type IncidentDiagnosticsRepositorySQLC struct {
	pgQ     *pgsqlc.Queries
	sqliteQ *sqlitesqlc.Queries
}

func NewIncidentDiagnosticsRepositorySQLC(rt SqlcRuntime) port.IncidentDiagnosticsRepository {
	r := &IncidentDiagnosticsRepositorySQLC{}
	if pool := rt.PgxPool(); pool != nil {
		r.pgQ = pgsqlc.New(pool)
	} else if db := rt.SQLiteDB(); db != nil {
		r.sqliteQ = sqlitesqlc.New(db)
	}
	return r
}

func (r *IncidentDiagnosticsRepositorySQLC) unconfigured() error {
	return fmt.Errorf("incident_diagnostics_sqlc: unconfigured runtime")
}

func headersToJSON(h map[string]string) (string, error) {
	if h == nil {
		return "{}", nil
	}
	b, err := json.Marshal(h)
	if err != nil {
		return "", fmt.Errorf("marshal headers: %w", err)
	}
	return string(b), nil
}

func headersFromJSON(s string) (map[string]string, error) {
	if s == "" {
		return map[string]string{}, nil
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return nil, fmt.Errorf("unmarshal headers: %w", err)
	}
	return m, nil
}

func (r *IncidentDiagnosticsRepositorySQLC) Create(ctx context.Context, d *domain.IncidentDiagnostics) (*domain.IncidentDiagnostics, error) {
	if d == nil {
		return nil, fmt.Errorf("incident diagnostics cannot be nil")
	}
	d.EnsureID()
	now := time.Now()
	if d.CreatedAt.IsZero() {
		d.CreatedAt = now
	}
	if d.UpdatedAt.IsZero() {
		d.UpdatedAt = now
	}
	reqHeaders, err := headersToJSON(d.RequestHeaders)
	if err != nil {
		return nil, err
	}
	respHeaders, err := headersToJSON(d.ResponseHeaders)
	if err != nil {
		return nil, err
	}
	switch {
	case r.pgQ != nil:
		row, err := r.pgQ.CreateIncidentDiagnostics(ctx, pgsqlc.CreateIncidentDiagnosticsParams{
			ID:                d.ID,
			CreatedAt:         pgtype.Timestamptz{Time: d.CreatedAt, Valid: true},
			UpdatedAt:         pgtype.Timestamptz{Time: d.UpdatedAt, Valid: true},
			IncidentID:        d.IncidentID,
			RequestMethod:     d.RequestMethod,
			RequestUrl:        d.RequestURL,
			RequestHeaders:    reqHeaders,
			RequestTimeout:    int32(d.RequestTimeout),
			HttpStatusCode:    int32(d.HTTPStatusCode),
			ResponseHeaders:   respHeaders,
			ResponseBody:      d.ResponseBody,
			ResponseSize:      int32(d.ResponseSize),
			FailureType:       d.FailureType,
			ErrorMessage:      d.ErrorMessage,
			ErrorSummary:      d.ErrorSummary,
			TotalDuration:     int32(d.TotalDuration),
			DnsDuration:       int32(d.DNSDuration),
			TlsDuration:       int32(d.TLSDuration),
			FirstByteDuration: int32(d.FirstByteDuration),
			BodyTruncated:     d.BodyTruncated,
			BodyEncoded:       d.BodyEncoded,
			Keyword:           pgTextFromPtr(d.Keyword),
			KeywordMode:       pgTextFromPtr(d.KeywordMode),
			KeywordFound:      pgBoolFromPtr(d.KeywordFound),
			IcmpAvailable:     pgBoolFromPtr(d.ICMPAvailable),
			IcmpReachable:     pgBoolFromPtr(d.ICMPReachable),
			IcmpRttMs:         pgInt4FromPtr(d.ICMPRttMs),
			RootCauseHint:     d.RootCauseHint,
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: create incident diagnostics: %w", err)
		}
		return incidentDiagnosticsFromPG(row)
	case r.sqliteQ != nil:
		row, err := r.sqliteQ.CreateIncidentDiagnostics(ctx, sqlitesqlc.CreateIncidentDiagnosticsParams{
			ID:                d.ID,
			CreatedAt:         d.CreatedAt,
			UpdatedAt:         d.UpdatedAt,
			IncidentID:        d.IncidentID,
			RequestMethod:     d.RequestMethod,
			RequestUrl:        d.RequestURL,
			RequestHeaders:    reqHeaders,
			RequestTimeout:    int64(d.RequestTimeout),
			HttpStatusCode:    int64(d.HTTPStatusCode),
			ResponseHeaders:   respHeaders,
			ResponseBody:      d.ResponseBody,
			ResponseSize:      int64(d.ResponseSize),
			FailureType:       d.FailureType,
			ErrorMessage:      d.ErrorMessage,
			ErrorSummary:      d.ErrorSummary,
			TotalDuration:     int64(d.TotalDuration),
			DnsDuration:       int64(d.DNSDuration),
			TlsDuration:       int64(d.TLSDuration),
			FirstByteDuration: int64(d.FirstByteDuration),
			BodyTruncated:     boolToInt64(d.BodyTruncated),
			BodyEncoded:       boolToInt64(d.BodyEncoded),
			Keyword:           nullStringFromPtr(d.Keyword),
			KeywordMode:       nullStringFromPtr(d.KeywordMode),
			KeywordFound:      nullBoolFromPtr(d.KeywordFound),
			IcmpAvailable:     nullBoolFromPtrAsInt64(d.ICMPAvailable),
			IcmpReachable:     nullBoolFromPtrAsInt64(d.ICMPReachable),
			IcmpRttMs:         nullIntFromPtr(d.ICMPRttMs),
			RootCauseHint:     d.RootCauseHint,
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: create incident diagnostics: %w", err)
		}
		return incidentDiagnosticsFromSQLite(row)
	default:
		return nil, r.unconfigured()
	}
}

func (r *IncidentDiagnosticsRepositorySQLC) FindByIncidentID(ctx context.Context, incidentID string) (*domain.IncidentDiagnostics, error) {
	switch {
	case r.pgQ != nil:
		row, err := r.pgQ.FindIncidentDiagnosticsByIncidentID(ctx, incidentID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, repository.ErrNotFound
			}
			return nil, fmt.Errorf("sqlc: find diagnostics by incident id: %w", err)
		}
		return incidentDiagnosticsFromPG(row)
	case r.sqliteQ != nil:
		row, err := r.sqliteQ.FindIncidentDiagnosticsByIncidentID(ctx, incidentID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, repository.ErrNotFound
			}
			return nil, fmt.Errorf("sqlc: find diagnostics by incident id: %w", err)
		}
		return incidentDiagnosticsFromSQLite(row)
	default:
		return nil, r.unconfigured()
	}
}

func (r *IncidentDiagnosticsRepositorySQLC) Update(ctx context.Context, d *domain.IncidentDiagnostics) error {
	if d == nil {
		return fmt.Errorf("incident diagnostics cannot be nil")
	}
	d.UpdatedAt = time.Now()
	reqHeaders, err := headersToJSON(d.RequestHeaders)
	if err != nil {
		return err
	}
	respHeaders, err := headersToJSON(d.ResponseHeaders)
	if err != nil {
		return err
	}
	switch {
	case r.pgQ != nil:
		n, err := r.pgQ.UpdateIncidentDiagnostics(ctx, pgsqlc.UpdateIncidentDiagnosticsParams{
			ID:                d.ID,
			RequestMethod:     d.RequestMethod,
			RequestUrl:        d.RequestURL,
			RequestHeaders:    reqHeaders,
			RequestTimeout:    int32(d.RequestTimeout),
			HttpStatusCode:    int32(d.HTTPStatusCode),
			ResponseHeaders:   respHeaders,
			ResponseBody:      d.ResponseBody,
			ResponseSize:      int32(d.ResponseSize),
			FailureType:       d.FailureType,
			ErrorMessage:      d.ErrorMessage,
			ErrorSummary:      d.ErrorSummary,
			TotalDuration:     int32(d.TotalDuration),
			DnsDuration:       int32(d.DNSDuration),
			TlsDuration:       int32(d.TLSDuration),
			FirstByteDuration: int32(d.FirstByteDuration),
			BodyTruncated:     d.BodyTruncated,
			BodyEncoded:       d.BodyEncoded,
			Keyword:           pgTextFromPtr(d.Keyword),
			KeywordMode:       pgTextFromPtr(d.KeywordMode),
			KeywordFound:      pgBoolFromPtr(d.KeywordFound),
			IcmpAvailable:     pgBoolFromPtr(d.ICMPAvailable),
			IcmpReachable:     pgBoolFromPtr(d.ICMPReachable),
			IcmpRttMs:         pgInt4FromPtr(d.ICMPRttMs),
			RootCauseHint:     d.RootCauseHint,
			UpdatedAt:         pgtype.Timestamptz{Time: d.UpdatedAt, Valid: true},
		})
		if err != nil {
			return fmt.Errorf("sqlc: update incident diagnostics: %w", err)
		}
		if n == 0 {
			return repository.ErrNotFound
		}
		return nil
	case r.sqliteQ != nil:
		n, err := r.sqliteQ.UpdateIncidentDiagnostics(ctx, sqlitesqlc.UpdateIncidentDiagnosticsParams{
			ID:                d.ID,
			RequestMethod:     d.RequestMethod,
			RequestUrl:        d.RequestURL,
			RequestHeaders:    reqHeaders,
			RequestTimeout:    int64(d.RequestTimeout),
			HttpStatusCode:    int64(d.HTTPStatusCode),
			ResponseHeaders:   respHeaders,
			ResponseBody:      d.ResponseBody,
			ResponseSize:      int64(d.ResponseSize),
			FailureType:       d.FailureType,
			ErrorMessage:      d.ErrorMessage,
			ErrorSummary:      d.ErrorSummary,
			TotalDuration:     int64(d.TotalDuration),
			DnsDuration:       int64(d.DNSDuration),
			TlsDuration:       int64(d.TLSDuration),
			FirstByteDuration: int64(d.FirstByteDuration),
			BodyTruncated:     boolToInt64(d.BodyTruncated),
			BodyEncoded:       boolToInt64(d.BodyEncoded),
			Keyword:           nullStringFromPtr(d.Keyword),
			KeywordMode:       nullStringFromPtr(d.KeywordMode),
			KeywordFound:      nullBoolFromPtr(d.KeywordFound),
			IcmpAvailable:     nullBoolFromPtrAsInt64(d.ICMPAvailable),
			IcmpReachable:     nullBoolFromPtrAsInt64(d.ICMPReachable),
			IcmpRttMs:         nullIntFromPtr(d.ICMPRttMs),
			RootCauseHint:     d.RootCauseHint,
			UpdatedAt:         d.UpdatedAt,
		})
		if err != nil {
			return fmt.Errorf("sqlc: update incident diagnostics: %w", err)
		}
		if n == 0 {
			return repository.ErrNotFound
		}
		return nil
	default:
		return r.unconfigured()
	}
}

func (r *IncidentDiagnosticsRepositorySQLC) Delete(ctx context.Context, id string) error {
	switch {
	case r.pgQ != nil:
		n, err := r.pgQ.DeleteIncidentDiagnostics(ctx, id)
		if err != nil {
			return fmt.Errorf("sqlc: delete incident diagnostics: %w", err)
		}
		if n == 0 {
			return repository.ErrNotFound
		}
		return nil
	case r.sqliteQ != nil:
		n, err := r.sqliteQ.DeleteIncidentDiagnostics(ctx, id)
		if err != nil {
			return fmt.Errorf("sqlc: delete incident diagnostics: %w", err)
		}
		if n == 0 {
			return repository.ErrNotFound
		}
		return nil
	default:
		return r.unconfigured()
	}
}

func incidentDiagnosticsFromPG(row pgsqlc.IncidentDiagnostic) (*domain.IncidentDiagnostics, error) {
	reqHeaders, err := headersFromJSON(row.RequestHeaders)
	if err != nil {
		return nil, err
	}
	respHeaders, err := headersFromJSON(row.ResponseHeaders)
	if err != nil {
		return nil, err
	}
	out := &domain.IncidentDiagnostics{
		Base: domain.Base{
			ID:        row.ID,
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
		},
		IncidentID:        row.IncidentID,
		RequestMethod:     row.RequestMethod,
		RequestURL:        row.RequestUrl,
		RequestHeaders:    reqHeaders,
		RequestTimeout:    int(row.RequestTimeout),
		HTTPStatusCode:    int(row.HttpStatusCode),
		ResponseHeaders:   respHeaders,
		ResponseBody:      row.ResponseBody,
		ResponseSize:      int(row.ResponseSize),
		FailureType:       row.FailureType,
		ErrorMessage:      row.ErrorMessage,
		ErrorSummary:      row.ErrorSummary,
		TotalDuration:     int(row.TotalDuration),
		DNSDuration:       int(row.DnsDuration),
		TLSDuration:       int(row.TlsDuration),
		FirstByteDuration: int(row.FirstByteDuration),
		BodyTruncated:     row.BodyTruncated,
		BodyEncoded:       row.BodyEncoded,
		RootCauseHint:     row.RootCauseHint,
	}
	if row.Keyword.Valid {
		s := row.Keyword.String
		out.Keyword = &s
	}
	if row.KeywordMode.Valid {
		s := row.KeywordMode.String
		out.KeywordMode = &s
	}
	if row.KeywordFound.Valid {
		b := row.KeywordFound.Bool
		out.KeywordFound = &b
	}
	if row.IcmpAvailable.Valid {
		b := row.IcmpAvailable.Bool
		out.ICMPAvailable = &b
	}
	if row.IcmpReachable.Valid {
		b := row.IcmpReachable.Bool
		out.ICMPReachable = &b
	}
	if row.IcmpRttMs.Valid {
		i := int(row.IcmpRttMs.Int32)
		out.ICMPRttMs = &i
	}
	return out, nil
}

func incidentDiagnosticsFromSQLite(row sqlitesqlc.IncidentDiagnostic) (*domain.IncidentDiagnostics, error) {
	reqHeaders, err := headersFromJSON(row.RequestHeaders)
	if err != nil {
		return nil, err
	}
	respHeaders, err := headersFromJSON(row.ResponseHeaders)
	if err != nil {
		return nil, err
	}
	out := &domain.IncidentDiagnostics{
		Base: domain.Base{
			ID:        row.ID,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		IncidentID:        row.IncidentID,
		RequestMethod:     row.RequestMethod,
		RequestURL:        row.RequestUrl,
		RequestHeaders:    reqHeaders,
		RequestTimeout:    int(row.RequestTimeout),
		HTTPStatusCode:    int(row.HttpStatusCode),
		ResponseHeaders:   respHeaders,
		ResponseBody:      row.ResponseBody,
		ResponseSize:      int(row.ResponseSize),
		FailureType:       row.FailureType,
		ErrorMessage:      row.ErrorMessage,
		ErrorSummary:      row.ErrorSummary,
		TotalDuration:     int(row.TotalDuration),
		DNSDuration:       int(row.DnsDuration),
		TLSDuration:       int(row.TlsDuration),
		FirstByteDuration: int(row.FirstByteDuration),
		BodyTruncated:     row.BodyTruncated != 0,
		BodyEncoded:       row.BodyEncoded != 0,
		RootCauseHint:     row.RootCauseHint,
	}
	if row.Keyword.Valid {
		s := row.Keyword.String
		out.Keyword = &s
	}
	if row.KeywordMode.Valid {
		s := row.KeywordMode.String
		out.KeywordMode = &s
	}
	if row.KeywordFound.Valid {
		b := row.KeywordFound.Bool
		out.KeywordFound = &b
	}
	// SQLite migration 0009 used INTEGER for icmp_* (sqlc → NullInt64); 0011 used BOOLEAN for keyword_found (sqlc → NullBool).
	if row.IcmpAvailable.Valid {
		b := row.IcmpAvailable.Int64 != 0
		out.ICMPAvailable = &b
	}
	if row.IcmpReachable.Valid {
		b := row.IcmpReachable.Int64 != 0
		out.ICMPReachable = &b
	}
	if row.IcmpRttMs.Valid {
		i := int(row.IcmpRttMs.Int64)
		out.ICMPRttMs = &i
	}
	return out, nil
}

// ---------- nullable pointer helpers (used by multiple Wave-1 wrappers) ----------

func pgBoolFromPtr(p *bool) pgtype.Bool {
	if p == nil {
		return pgtype.Bool{}
	}
	return pgtype.Bool{Bool: *p, Valid: true}
}

func pgInt4FromPtr(p *int) pgtype.Int4 {
	if p == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: int32(*p), Valid: true}
}

func nullBoolFromPtrAsInt64(p *bool) sql.NullInt64 {
	if p == nil {
		return sql.NullInt64{}
	}
	v := int64(0)
	if *p {
		v = 1
	}
	return sql.NullInt64{Int64: v, Valid: true}
}

func nullBoolFromPtr(p *bool) sql.NullBool {
	if p == nil {
		return sql.NullBool{}
	}
	return sql.NullBool{Bool: *p, Valid: true}
}

func nullIntFromPtr(p *int) sql.NullInt64 {
	if p == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(*p), Valid: true}
}
