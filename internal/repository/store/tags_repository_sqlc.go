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

// SqlcRuntime is the minimal contract this repository needs from
// *database.Runtime. Defined here (not imported) to keep the store package
// free of an internal/database import — *database.Runtime satisfies it
// via duck-typing.
type SqlcRuntime interface {
	PgxPool() *pgxpool.Pool
	SQLiteDB() *sql.DB
}

// TagsRepositorySQLC implements port.TagsRepository on top of sqlc-generated
// query code. Exactly one of pgQ / sqliteQ is non-nil at construction time,
// matching the runtime's active driver (determined by whichever of
// PgxPool / SQLiteDB returns non-nil).
type TagsRepositorySQLC struct {
	pgQ     *pgsqlc.Queries
	sqliteQ *sqlitesqlc.Queries
}

func NewTagsRepositorySQLC(rt SqlcRuntime) port.TagsRepository {
	r := &TagsRepositorySQLC{}
	if pool := rt.PgxPool(); pool != nil {
		r.pgQ = pgsqlc.New(pool)
	} else if db := rt.SQLiteDB(); db != nil {
		r.sqliteQ = sqlitesqlc.New(db)
	}
	return r
}

func (r *TagsRepositorySQLC) unconfigured() error {
	return fmt.Errorf("tags_repository_sqlc: unconfigured runtime (no pgx pool or sqlite *sql.DB)")
}

// ---------- Public surface (port.TagsRepository) ----------

func (r *TagsRepositorySQLC) Create(ctx context.Context, t *domain.Tags) error {
	t.EnsureID()
	switch {
	case r.pgQ != nil:
		return r.createWithQ(ctx, r.pgQ, t)
	case r.sqliteQ != nil:
		return r.createWithSqlQ(ctx, r.sqliteQ, t)
	default:
		return r.unconfigured()
	}
}

func (r *TagsRepositorySQLC) FindByID(ctx context.Context, id string) (*domain.Tags, error) {
	switch {
	case r.pgQ != nil:
		row, err := r.pgQ.FindTagByID(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, repository.ErrNotFound
			}
			return nil, fmt.Errorf("sqlc: find tag by id: %w", err)
		}
		return tagFromPG(row), nil
	case r.sqliteQ != nil:
		row, err := r.sqliteQ.FindTagByID(ctx, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, repository.ErrNotFound
			}
			return nil, fmt.Errorf("sqlc: find tag by id: %w", err)
		}
		return tagFromSQLite(row), nil
	default:
		return nil, r.unconfigured()
	}
}

func (r *TagsRepositorySQLC) FindByIDs(ctx context.Context, ids []string) ([]*domain.Tags, error) {
	switch {
	case r.pgQ != nil:
		rows, err := r.pgQ.FindTagsByIDs(ctx, ids)
		if err != nil {
			return nil, fmt.Errorf("sqlc: find tags by ids: %w", err)
		}
		if len(rows) != len(ids) {
			return nil, repository.ErrNotFound
		}
		out := make([]*domain.Tags, len(rows))
		for i, row := range rows {
			out[i] = tagFromPG(row)
		}
		return out, nil
	case r.sqliteQ != nil:
		rows, err := r.sqliteQ.FindTagsByIDs(ctx, ids)
		if err != nil {
			return nil, fmt.Errorf("sqlc: find tags by ids: %w", err)
		}
		if len(rows) != len(ids) {
			return nil, repository.ErrNotFound
		}
		out := make([]*domain.Tags, len(rows))
		for i, row := range rows {
			out[i] = tagFromSQLite(row)
		}
		return out, nil
	default:
		return nil, r.unconfigured()
	}
}

func (r *TagsRepositorySQLC) FindByName(ctx context.Context, name string) (*domain.Tags, error) {
	switch {
	case r.pgQ != nil:
		row, err := r.pgQ.FindTagByName(ctx, name)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, repository.ErrNotFound
			}
			return nil, fmt.Errorf("sqlc: find tag by name: %w", err)
		}
		return tagFromPG(row), nil
	case r.sqliteQ != nil:
		row, err := r.sqliteQ.FindTagByName(ctx, name)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, repository.ErrNotFound
			}
			return nil, fmt.Errorf("sqlc: find tag by name: %w", err)
		}
		return tagFromSQLite(row), nil
	default:
		return nil, r.unconfigured()
	}
}

func (r *TagsRepositorySQLC) List(ctx context.Context, limit, offset int) ([]*domain.Tags, error) {
	switch {
	case r.pgQ != nil:
		rows, err := r.pgQ.ListTags(ctx, pgsqlc.ListTagsParams{
			Limit:  int32(limit),
			Offset: int32(offset),
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: list tags: %w", err)
		}
		out := make([]*domain.Tags, len(rows))
		for i, row := range rows {
			out[i] = tagFromPG(row)
		}
		return out, nil
	case r.sqliteQ != nil:
		rows, err := r.sqliteQ.ListTags(ctx, sqlitesqlc.ListTagsParams{
			Limit:  int64(limit),
			Offset: int64(offset),
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: list tags: %w", err)
		}
		out := make([]*domain.Tags, len(rows))
		for i, row := range rows {
			out[i] = tagFromSQLite(row)
		}
		return out, nil
	default:
		return nil, r.unconfigured()
	}
}

func (r *TagsRepositorySQLC) Update(ctx context.Context, t *domain.Tags) error {
	now := time.Now()
	switch {
	case r.pgQ != nil:
		n, err := r.pgQ.UpdateTag(ctx, pgsqlc.UpdateTagParams{
			ID:          t.ID,
			Name:        t.Name,
			Color:       pgTextFromPtr(t.Color),
			Description: pgTextFromPtr(t.Description),
			UpdatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		})
		if err != nil {
			return fmt.Errorf("sqlc: update tag: %w", err)
		}
		if n == 0 {
			return repository.ErrNotFound
		}
		t.UpdatedAt = now
		return nil
	case r.sqliteQ != nil:
		n, err := r.sqliteQ.UpdateTag(ctx, sqlitesqlc.UpdateTagParams{
			ID:          t.ID,
			Name:        t.Name,
			Color:       nullStringFromPtr(t.Color),
			Description: nullStringFromPtr(t.Description),
			UpdatedAt:   now,
		})
		if err != nil {
			return fmt.Errorf("sqlc: update tag: %w", err)
		}
		if n == 0 {
			return repository.ErrNotFound
		}
		t.UpdatedAt = now
		return nil
	default:
		return r.unconfigured()
	}
}

func (r *TagsRepositorySQLC) Delete(ctx context.Context, id string) error {
	switch {
	case r.pgQ != nil:
		n, err := r.pgQ.DeleteTag(ctx, id)
		if err != nil {
			return fmt.Errorf("sqlc: delete tag: %w", err)
		}
		if n == 0 {
			return repository.ErrNotFound
		}
		return nil
	case r.sqliteQ != nil:
		n, err := r.sqliteQ.DeleteTag(ctx, id)
		if err != nil {
			return fmt.Errorf("sqlc: delete tag: %w", err)
		}
		if n == 0 {
			return repository.ErrNotFound
		}
		return nil
	default:
		return r.unconfigured()
	}
}

// ---------- Querier-injection template (FR-010) ----------

// createWithQ inserts a tag through the supplied pgsqlc.Queries instance,
// allowing in-package callers (typically services owning an inter-repo
// transaction) to keep the insert inside their tx. This is the template
// Wave 3 will copy for resource/tag co-insertions.
func (r *TagsRepositorySQLC) createWithQ(ctx context.Context, q *pgsqlc.Queries, t *domain.Tags) error {
	t.EnsureID()
	now := nowOr(t.CreatedAt)
	if t.UpdatedAt.IsZero() {
		t.UpdatedAt = now
	}
	row, err := q.CreateTag(ctx, pgsqlc.CreateTagParams{
		ID:          t.ID,
		CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt:   pgtype.Timestamptz{Time: t.UpdatedAt, Valid: true},
		Name:        t.Name,
		Color:       pgTextFromPtr(t.Color),
		Description: pgTextFromPtr(t.Description),
	})
	if err != nil {
		return fmt.Errorf("sqlc: create tag: %w", err)
	}
	t.CreatedAt = row.CreatedAt.Time
	t.UpdatedAt = row.UpdatedAt.Time
	return nil
}

// createWithSqlQ mirrors createWithQ on the SQLite path for shape parity.
func (r *TagsRepositorySQLC) createWithSqlQ(ctx context.Context, q *sqlitesqlc.Queries, t *domain.Tags) error {
	t.EnsureID()
	now := nowOr(t.CreatedAt)
	if t.UpdatedAt.IsZero() {
		t.UpdatedAt = now
	}
	row, err := q.CreateTag(ctx, sqlitesqlc.CreateTagParams{
		ID:          t.ID,
		CreatedAt:   now,
		UpdatedAt:   t.UpdatedAt,
		Name:        t.Name,
		Color:       nullStringFromPtr(t.Color),
		Description: nullStringFromPtr(t.Description),
	})
	if err != nil {
		return fmt.Errorf("sqlc: create tag: %w", err)
	}
	t.CreatedAt = row.CreatedAt
	t.UpdatedAt = row.UpdatedAt
	return nil
}

// ---------- Mapping helpers ----------

func tagFromPG(row pgsqlc.Tag) *domain.Tags {
	out := &domain.Tags{
		Base: domain.Base{
			ID:        row.ID,
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
		},
		Name: row.Name,
	}
	if row.Color.Valid {
		s := row.Color.String
		out.Color = &s
	}
	if row.Description.Valid {
		s := row.Description.String
		out.Description = &s
	}
	return out
}

func tagFromSQLite(row sqlitesqlc.Tag) *domain.Tags {
	out := &domain.Tags{
		Base: domain.Base{
			ID:        row.ID,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		Name: row.Name,
	}
	if row.Color.Valid {
		s := row.Color.String
		out.Color = &s
	}
	if row.Description.Valid {
		s := row.Description.String
		out.Description = &s
	}
	return out
}

func pgTextFromPtr(p *string) pgtype.Text {
	if p == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *p, Valid: true}
}

func nullStringFromPtr(p *string) sql.NullString {
	if p == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *p, Valid: true}
}

func nowOr(t time.Time) time.Time {
	if t.IsZero() {
		return time.Now()
	}
	return t
}
