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

type TwoFactorResetTokenRepositorySQLC struct {
	pgQ     *pgsqlc.Queries
	sqliteQ *sqlitesqlc.Queries
}

func NewTwoFactorResetTokenRepositorySQLC(rt SqlcRuntime) port.TwoFactorResetTokenRepository {
	r := &TwoFactorResetTokenRepositorySQLC{}
	if pool := rt.PgxPool(); pool != nil {
		r.pgQ = pgsqlc.New(pool)
	} else if db := rt.SQLiteDB(); db != nil {
		r.sqliteQ = sqlitesqlc.New(db)
	}
	return r
}

func (r *TwoFactorResetTokenRepositorySQLC) unconfigured() error {
	return fmt.Errorf("two_factor_reset_token_repository_sqlc: unconfigured runtime")
}

func (r *TwoFactorResetTokenRepositorySQLC) Create(ctx context.Context, t *domain.TwoFactorResetToken) error {
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	switch {
	case r.pgQ != nil:
		return r.pgQ.CreateTwoFactorResetToken(ctx, pgsqlc.CreateTwoFactorResetTokenParams{
			TokenHash: t.TokenHash,
			UserID:    t.UserID,
			ExpiresAt: pgtype.Timestamptz{Time: t.ExpiresAt, Valid: true},
			CreatedAt: pgtype.Timestamptz{Time: t.CreatedAt, Valid: true},
		})
	case r.sqliteQ != nil:
		return r.sqliteQ.CreateTwoFactorResetToken(ctx, sqlitesqlc.CreateTwoFactorResetTokenParams{
			TokenHash: t.TokenHash,
			UserID:    t.UserID,
			ExpiresAt: t.ExpiresAt,
			CreatedAt: t.CreatedAt,
		})
	default:
		return r.unconfigured()
	}
}

// ConsumeByHash atomically finds + marks the token as used.
// Returns the token on success; repository.ErrNotFound if absent / expired / used.
func (r *TwoFactorResetTokenRepositorySQLC) ConsumeByHash(ctx context.Context, tokenHash string, at time.Time) (*domain.TwoFactorResetToken, error) {
	switch {
	case r.pgQ != nil:
		row, err := r.pgQ.FindActiveTwoFactorResetToken(ctx, pgsqlc.FindActiveTwoFactorResetTokenParams{
			TokenHash: tokenHash,
			ExpiresAt: pgtype.Timestamptz{Time: at, Valid: true},
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, repository.ErrNotFound
			}
			return nil, fmt.Errorf("sqlc: find reset token: %w", err)
		}
		n, err := r.pgQ.MarkTwoFactorResetTokenUsed(ctx, pgsqlc.MarkTwoFactorResetTokenUsedParams{
			TokenHash: tokenHash,
			UsedAt:    pgtype.Timestamptz{Time: at, Valid: true},
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: mark used: %w", err)
		}
		if n == 0 {
			return nil, repository.ErrNotFound
		}
		out := &domain.TwoFactorResetToken{
			TokenHash: row.TokenHash,
			UserID:    row.UserID,
			ExpiresAt: row.ExpiresAt.Time,
			CreatedAt: row.CreatedAt.Time,
		}
		usedCopy := at
		out.UsedAt = &usedCopy
		return out, nil
	case r.sqliteQ != nil:
		row, err := r.sqliteQ.FindActiveTwoFactorResetToken(ctx, sqlitesqlc.FindActiveTwoFactorResetTokenParams{
			TokenHash: tokenHash,
			ExpiresAt: at,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, repository.ErrNotFound
			}
			return nil, fmt.Errorf("sqlc: find reset token: %w", err)
		}
		n, err := r.sqliteQ.MarkTwoFactorResetTokenUsed(ctx, sqlitesqlc.MarkTwoFactorResetTokenUsedParams{
			TokenHash: tokenHash,
			UsedAt:    sql.NullTime{Time: at, Valid: true},
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: mark used: %w", err)
		}
		if n == 0 {
			return nil, repository.ErrNotFound
		}
		out := &domain.TwoFactorResetToken{
			TokenHash: row.TokenHash,
			UserID:    row.UserID,
			ExpiresAt: row.ExpiresAt,
			CreatedAt: row.CreatedAt,
		}
		usedCopy := at
		out.UsedAt = &usedCopy
		return out, nil
	default:
		return nil, r.unconfigured()
	}
}

func (r *TwoFactorResetTokenRepositorySQLC) CountRecentByUser(ctx context.Context, userID string, since time.Time) (int64, error) {
	switch {
	case r.pgQ != nil:
		return r.pgQ.CountRecentTwoFactorResetTokensByUser(ctx, pgsqlc.CountRecentTwoFactorResetTokensByUserParams{
			UserID:    userID,
			CreatedAt: pgtype.Timestamptz{Time: since, Valid: true},
		})
	case r.sqliteQ != nil:
		return r.sqliteQ.CountRecentTwoFactorResetTokensByUser(ctx, sqlitesqlc.CountRecentTwoFactorResetTokensByUserParams{
			UserID:    userID,
			CreatedAt: since,
		})
	default:
		return 0, r.unconfigured()
	}
}

func (r *TwoFactorResetTokenRepositorySQLC) DeleteExpired(ctx context.Context, cutoff time.Time) error {
	switch {
	case r.pgQ != nil:
		return r.pgQ.DeleteExpiredTwoFactorResetTokens(ctx, pgtype.Timestamptz{Time: cutoff, Valid: true})
	case r.sqliteQ != nil:
		return r.sqliteQ.DeleteExpiredTwoFactorResetTokens(ctx, cutoff)
	default:
		return r.unconfigured()
	}
}
