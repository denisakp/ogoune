package store

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
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

// ErasureRepositorySQLC applies an erasure plan in one transaction (spec 095,
// FR-011): every channel rewrite, the report changes, the account deletion and
// the record succeed together or none does. It spans several tables because no
// cross-repository transaction exists; the plan is computed by the service.
type ErasureRepositorySQLC struct {
	pool    *pgxpool.Pool
	db      *sql.DB
	pgQ     *pgsqlc.Queries
	sqliteQ *sqlitesqlc.Queries
}

func NewErasureRepositorySQLC(rt SqlcRuntime) port.ErasureRepository {
	r := &ErasureRepositorySQLC{}
	if pool := rt.PgxPool(); pool != nil {
		r.pool = pool
		r.pgQ = pgsqlc.New(pool)
	} else if db := rt.SQLiteDB(); db != nil {
		r.db = db
		r.sqliteQ = sqlitesqlc.New(db)
	}
	return r
}

func (r *ErasureRepositorySQLC) unconfigured() error {
	return fmt.Errorf("erasure_repository_sqlc: unconfigured runtime")
}

// Apply writes the plan. A channel whose configuration changed since the plan
// was computed fails the whole erasure with an ErasureConflictError. Errors
// name the item kind and id, never the address.
func (r *ErasureRepositorySQLC) Apply(ctx context.Context, plan domain.ErasurePlan) error {
	switch {
	case r.pool != nil:
		return pgsqlc.WithTx(ctx, r.pool, func(q *pgsqlc.Queries) error {
			return applyErasure(ctx, pgErasureWriter{q}, plan)
		})
	case r.db != nil:
		return sqlitesqlc.WithTx(ctx, r.db, func(q *sqlitesqlc.Queries) error {
			return applyErasure(ctx, sqliteErasureWriter{q}, plan)
		})
	default:
		return r.unconfigured()
	}
}

func applyErasure(ctx context.Context, w erasureWriter, plan domain.ErasurePlan) error {
	now := time.Now().UTC()
	for _, c := range plan.Channels {
		if err := rewriteChannel(ctx, w, c, now); err != nil {
			return err
		}
	}
	if plan.ClearReportRecipient {
		if err := w.clearReportRecipient(ctx, now); err != nil {
			return fmt.Errorf("erasure: report settings: %w", err)
		}
	}
	if plan.ReportHistoryAddress != "" {
		if err := w.anonymizeReportHistory(ctx, plan.ReportHistoryAddress); err != nil {
			return fmt.Errorf("erasure: report history: %w", err)
		}
	}
	if plan.AccountID != "" {
		if err := deleteAccount(ctx, w, plan.AccountID); err != nil {
			return err
		}
	}
	rec := plan.Record
	rec.EnsureID()
	if rec.CreatedAt.IsZero() {
		rec.CreatedAt = now
	}
	changes, err := json.Marshal(rec.Changes)
	if err != nil {
		return fmt.Errorf("erasure: record: %w", err)
	}
	if err := w.createRecord(ctx, rec, changes); err != nil {
		return fmt.Errorf("erasure: record: %w", err)
	}
	return nil
}

func rewriteChannel(ctx context.Context, w erasureWriter, c domain.ChannelRewrite, now time.Time) error {
	stored, err := w.channelConfig(ctx, c.ID)
	if err != nil {
		if isErasureNoRows(err) {
			return &repository.ErasureConflictError{ChannelID: c.ID}
		}
		return channelErr(c.ID, err)
	}
	current, err := decryptChannelConfig(stored)
	if err != nil {
		return channelErr(c.ID, err)
	}
	if !bytes.Equal(current, c.ExpectConfig) {
		return &repository.ErasureConflictError{ChannelID: c.ID}
	}
	ct, err := encryptChannelConfig(c.Config)
	if err != nil {
		return channelErr(c.ID, err)
	}
	var disabledAt *time.Time
	reason := ""
	if c.Disable {
		disabledAt = &now
		reason = domain.ChannelDisabledByErasure
	}
	n, err := w.rewriteChannel(ctx, c.ID, ct, disabledAt, reason, now)
	if err != nil {
		return channelErr(c.ID, err)
	}
	if n == 0 {
		return &repository.ErasureConflictError{ChannelID: c.ID}
	}
	return nil
}

// deleteAccount keeps the account's incident updates, without their author,
// and removes the account with its API keys and sessions. Explicit deletes:
// the transaction does not rely on foreign-key cascades.
func deleteAccount(ctx context.Context, w erasureWriter, userID string) error {
	steps := []struct {
		kind string
		fn   func(context.Context, string) error
	}{
		{"incident updates", w.unlinkUpdates},
		{"api keys", w.deleteAPIKeys},
		{"sessions", w.deleteSessions},
		{"account", w.deleteUser},
	}
	for _, s := range steps {
		if err := s.fn(ctx, userID); err != nil {
			return fmt.Errorf("erasure: %s of account %s: %w", s.kind, userID, err)
		}
	}
	return nil
}

func channelErr(id string, err error) error {
	return fmt.Errorf("erasure: notification channel %s: %w", id, err)
}

func isErasureNoRows(err error) bool {
	return errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows)
}

// FindRecordsByFingerprint returns the erasures of one address, newest first.
func (r *ErasureRepositorySQLC) FindRecordsByFingerprint(ctx context.Context, fp string) ([]domain.ErasureRecord, error) {
	switch {
	case r.pgQ != nil:
		rows, err := r.pgQ.ListErasureRecordsByFingerprint(ctx, fp)
		if err != nil {
			return nil, fmt.Errorf("sqlc: list erasure records: %w", err)
		}
		out := make([]domain.ErasureRecord, 0, len(rows))
		for _, row := range rows {
			out = append(out, erasureRecordFrom(row.ID, row.CreatedAt.Time, row.OperatorID, row.SubjectKind,
				row.SubjectFingerprint, row.Changes, int(row.ManualReview)))
		}
		return out, nil
	case r.sqliteQ != nil:
		rows, err := r.sqliteQ.ListErasureRecordsByFingerprint(ctx, fp)
		if err != nil {
			return nil, fmt.Errorf("sqlc: list erasure records: %w", err)
		}
		out := make([]domain.ErasureRecord, 0, len(rows))
		for _, row := range rows {
			out = append(out, erasureRecordFrom(row.ID, row.CreatedAt, row.OperatorID, row.SubjectKind,
				row.SubjectFingerprint, []byte(row.Changes), int(row.ManualReview)))
		}
		return out, nil
	default:
		return nil, r.unconfigured()
	}
}

func erasureRecordFrom(id string, at time.Time, operator, kind, fp string, changes []byte, manual int) domain.ErasureRecord {
	rec := domain.ErasureRecord{
		Base:               domain.Base{ID: id, CreatedAt: at},
		OperatorID:         operator,
		SubjectKind:        domain.ErasureSubjectKind(kind),
		SubjectFingerprint: fp,
		ManualReview:       manual,
	}
	// A record whose counts cannot be read still says an erasure happened.
	_ = json.Unmarshal(changes, &rec.Changes)
	return rec
}

// erasureWriter is the set of writes an erasure makes, bound to one
// transaction of one dialect.
type erasureWriter interface {
	channelConfig(ctx context.Context, id string) ([]byte, error)
	rewriteChannel(ctx context.Context, id string, config []byte, disabledAt *time.Time, reason string, now time.Time) (int64, error)
	clearReportRecipient(ctx context.Context, now time.Time) error
	anonymizeReportHistory(ctx context.Context, email string) error
	unlinkUpdates(ctx context.Context, userID string) error
	deleteAPIKeys(ctx context.Context, userID string) error
	deleteSessions(ctx context.Context, userID string) error
	deleteUser(ctx context.Context, userID string) error
	createRecord(ctx context.Context, rec domain.ErasureRecord, changes []byte) error
}

type pgErasureWriter struct{ q *pgsqlc.Queries }

func (w pgErasureWriter) channelConfig(ctx context.Context, id string) ([]byte, error) {
	row, err := w.q.FindNotificationChannelByID(ctx, id)
	return row.Config, err
}

func (w pgErasureWriter) rewriteChannel(ctx context.Context, id string, config []byte, disabledAt *time.Time, reason string, now time.Time) (int64, error) {
	p := pgsqlc.RewriteNotificationChannelConfigParams{
		Config: config, UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true}, ID: id,
	}
	if disabledAt != nil {
		p.DisabledAt = pgtype.Timestamptz{Time: *disabledAt, Valid: true}
		p.DisabledReason = pgtype.Text{String: reason, Valid: true}
	}
	return w.q.RewriteNotificationChannelConfig(ctx, p)
}

func (w pgErasureWriter) clearReportRecipient(ctx context.Context, now time.Time) error {
	_, err := w.q.ClearReportRecipient(ctx, pgtype.Timestamptz{Time: now, Valid: true})
	return err
}

func (w pgErasureWriter) anonymizeReportHistory(ctx context.Context, email string) error {
	_, err := w.q.AnonymizeReportHistoryRecipient(ctx, email)
	return err
}

func (w pgErasureWriter) unlinkUpdates(ctx context.Context, userID string) error {
	_, err := w.q.UnlinkIncidentUpdatesAuthor(ctx, userID)
	return err
}

func (w pgErasureWriter) deleteAPIKeys(ctx context.Context, userID string) error {
	_, err := w.q.DeleteAPIKeysByUser(ctx, userID)
	return err
}

func (w pgErasureWriter) deleteSessions(ctx context.Context, userID string) error {
	_, err := w.q.DeleteSessionsByUser(ctx, userID)
	return err
}

func (w pgErasureWriter) deleteUser(ctx context.Context, userID string) error {
	return w.q.DeleteUser(ctx, userID)
}

func (w pgErasureWriter) createRecord(ctx context.Context, rec domain.ErasureRecord, changes []byte) error {
	return w.q.CreateErasureRecord(ctx, pgsqlc.CreateErasureRecordParams{
		ID:                 rec.ID,
		CreatedAt:          pgtype.Timestamptz{Time: rec.CreatedAt, Valid: true},
		OperatorID:         rec.OperatorID,
		SubjectKind:        string(rec.SubjectKind),
		SubjectFingerprint: rec.SubjectFingerprint,
		Changes:            changes,
		ManualReview:       int32(rec.ManualReview),
	})
}

type sqliteErasureWriter struct{ q *sqlitesqlc.Queries }

func (w sqliteErasureWriter) channelConfig(ctx context.Context, id string) ([]byte, error) {
	row, err := w.q.FindNotificationChannelByID(ctx, id)
	return row.Config, err
}

func (w sqliteErasureWriter) rewriteChannel(ctx context.Context, id string, config []byte, disabledAt *time.Time, reason string, now time.Time) (int64, error) {
	p := sqlitesqlc.RewriteNotificationChannelConfigParams{Config: config, UpdatedAt: now, ID: id}
	if disabledAt != nil {
		p.DisabledAt = sql.NullTime{Time: *disabledAt, Valid: true}
		p.DisabledReason = sql.NullString{String: reason, Valid: true}
	}
	return w.q.RewriteNotificationChannelConfig(ctx, p)
}

func (w sqliteErasureWriter) clearReportRecipient(ctx context.Context, now time.Time) error {
	_, err := w.q.ClearReportRecipient(ctx, now)
	return err
}

func (w sqliteErasureWriter) anonymizeReportHistory(ctx context.Context, email string) error {
	_, err := w.q.AnonymizeReportHistoryRecipient(ctx, email)
	return err
}

func (w sqliteErasureWriter) unlinkUpdates(ctx context.Context, userID string) error {
	_, err := w.q.UnlinkIncidentUpdatesAuthor(ctx, userID)
	return err
}

func (w sqliteErasureWriter) deleteAPIKeys(ctx context.Context, userID string) error {
	_, err := w.q.DeleteAPIKeysByUser(ctx, userID)
	return err
}

func (w sqliteErasureWriter) deleteSessions(ctx context.Context, userID string) error {
	_, err := w.q.DeleteSessionsByUser(ctx, userID)
	return err
}

func (w sqliteErasureWriter) deleteUser(ctx context.Context, userID string) error {
	return w.q.DeleteUser(ctx, userID)
}

func (w sqliteErasureWriter) createRecord(ctx context.Context, rec domain.ErasureRecord, changes []byte) error {
	return w.q.CreateErasureRecord(ctx, sqlitesqlc.CreateErasureRecordParams{
		ID:                 rec.ID,
		CreatedAt:          rec.CreatedAt,
		OperatorID:         rec.OperatorID,
		SubjectKind:        string(rec.SubjectKind),
		SubjectFingerprint: rec.SubjectFingerprint,
		Changes:            string(changes),
		ManualReview:       int64(rec.ManualReview),
	})
}
