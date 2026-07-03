package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository"
	pgsqlc "github.com/denisakp/ogoune/internal/repository/sqlc/pg"
	sqlitesqlc "github.com/denisakp/ogoune/internal/repository/sqlc/sqlite"
)

// AnnouncementRepositorySQLC is the sqlc-backed operator-banner store.
type AnnouncementRepositorySQLC struct {
	pgQ     *pgsqlc.Queries
	sqliteQ *sqlitesqlc.Queries
}

func NewAnnouncementRepositorySQLC(rt SqlcRuntime) port.AnnouncementRepository {
	r := &AnnouncementRepositorySQLC{}
	if pool := rt.PgxPool(); pool != nil {
		r.pgQ = pgsqlc.New(pool)
	} else if db := rt.SQLiteDB(); db != nil {
		r.sqliteQ = sqlitesqlc.New(db)
	}
	return r
}

func (r *AnnouncementRepositorySQLC) unconfigured() error {
	return fmt.Errorf("announcement_repository_sqlc: unconfigured runtime")
}

func (r *AnnouncementRepositorySQLC) Create(ctx context.Context, a *domain.Announcement) (*domain.Announcement, error) {
	a.EnsureID()
	now := time.Now()
	if a.CreatedAt.IsZero() {
		a.CreatedAt = now
	}
	a.UpdatedAt = now
	switch {
	case r.pgQ != nil:
		row, err := r.pgQ.CreateAnnouncement(ctx, pgsqlc.CreateAnnouncementParams{
			ID:          a.ID,
			Severity:    string(a.Severity),
			Title:       a.Title,
			Description: a.Description,
			Dismissible: a.Dismissible,
			Active:      a.Active,
			CreatedAt:   pgtype.Timestamptz{Time: a.CreatedAt, Valid: true},
			UpdatedAt:   pgtype.Timestamptz{Time: a.UpdatedAt, Valid: true},
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: create announcement: %w", err)
		}
		return announcementFromPG(row), nil
	case r.sqliteQ != nil:
		row, err := r.sqliteQ.CreateAnnouncement(ctx, sqlitesqlc.CreateAnnouncementParams{
			ID:          a.ID,
			Severity:    string(a.Severity),
			Title:       a.Title,
			Description: a.Description,
			Dismissible: boolToInt64(a.Dismissible),
			Active:      boolToInt64(a.Active),
			CreatedAt:   a.CreatedAt,
			UpdatedAt:   a.UpdatedAt,
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: create announcement: %w", err)
		}
		return announcementFromSQLite(row), nil
	default:
		return nil, r.unconfigured()
	}
}

func (r *AnnouncementRepositorySQLC) ListActive(ctx context.Context) ([]*domain.Announcement, error) {
	switch {
	case r.pgQ != nil:
		rows, err := r.pgQ.ListActiveAnnouncements(ctx, true)
		if err != nil {
			return nil, fmt.Errorf("sqlc: list announcements: %w", err)
		}
		out := make([]*domain.Announcement, 0, len(rows))
		for _, row := range rows {
			out = append(out, announcementFromPG(row))
		}
		return out, nil
	case r.sqliteQ != nil:
		rows, err := r.sqliteQ.ListActiveAnnouncements(ctx, 1)
		if err != nil {
			return nil, fmt.Errorf("sqlc: list announcements: %w", err)
		}
		out := make([]*domain.Announcement, 0, len(rows))
		for _, row := range rows {
			out = append(out, announcementFromSQLite(row))
		}
		return out, nil
	default:
		return nil, r.unconfigured()
	}
}

func (r *AnnouncementRepositorySQLC) Delete(ctx context.Context, id string) error {
	switch {
	case r.pgQ != nil:
		n, err := r.pgQ.DeleteAnnouncement(ctx, id)
		if err != nil {
			return fmt.Errorf("sqlc: delete announcement: %w", err)
		}
		if n == 0 {
			return repository.ErrNotFound
		}
		return nil
	case r.sqliteQ != nil:
		n, err := r.sqliteQ.DeleteAnnouncement(ctx, id)
		if err != nil {
			return fmt.Errorf("sqlc: delete announcement: %w", err)
		}
		if n == 0 {
			return repository.ErrNotFound
		}
		return nil
	default:
		return r.unconfigured()
	}
}

func announcementFromPG(row pgsqlc.Announcement) *domain.Announcement {
	return &domain.Announcement{
		Base:        domain.Base{ID: row.ID, CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time},
		Severity:    domain.AnnouncementSeverity(row.Severity),
		Title:       row.Title,
		Description: row.Description,
		Dismissible: row.Dismissible,
		Active:      row.Active,
	}
}

func announcementFromSQLite(row sqlitesqlc.Announcement) *domain.Announcement {
	return &domain.Announcement{
		Base:        domain.Base{ID: row.ID, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt},
		Severity:    domain.AnnouncementSeverity(row.Severity),
		Title:       row.Title,
		Description: row.Description,
		Dismissible: row.Dismissible != 0,
		Active:      row.Active != 0,
	}
}
