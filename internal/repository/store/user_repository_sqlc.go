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

type UserRepositorySQLC struct {
	pgQ     *pgsqlc.Queries
	sqliteQ *sqlitesqlc.Queries
}

func NewUserRepositorySQLC(rt SqlcRuntime) port.UserRepository {
	r := &UserRepositorySQLC{}
	if pool := rt.PgxPool(); pool != nil {
		r.pgQ = pgsqlc.New(pool)
	} else if db := rt.SQLiteDB(); db != nil {
		r.sqliteQ = sqlitesqlc.New(db)
	}
	return r
}

func (r *UserRepositorySQLC) unconfigured() error {
	return fmt.Errorf("user_repository_sqlc: unconfigured runtime")
}

func (r *UserRepositorySQLC) Create(ctx context.Context, u *domain.User) (*domain.User, error) {
	u.EnsureID()
	now := time.Now()
	if u.CreatedAt.IsZero() {
		u.CreatedAt = now
	}
	if u.UpdatedAt.IsZero() {
		u.UpdatedAt = now
	}
	switch {
	case r.pgQ != nil:
		row, err := r.pgQ.CreateUser(ctx, pgsqlc.CreateUserParams{
			ID:                   u.ID,
			Email:                u.Email,
			Name:                 u.Name,
			HashedPassword:       u.HashedPassword,
			PasswordInitialized:  u.PasswordInitialized,
			ForcePasswordChange:  u.ForcePasswordChange,
			TwoFactorEnabled:     u.TwoFactorEnabled,
			TwoFactorSecret:      u.TwoFactorSecret,
			TwoFactorBackupCodes: u.TwoFactorBackupCodes,
			LastLoginAt:          pgTimestampFromPtr(u.LastLoginAt),
			CreatedAt:            pgtype.Timestamptz{Time: u.CreatedAt, Valid: true},
			UpdatedAt:            pgtype.Timestamptz{Time: u.UpdatedAt, Valid: true},
		})
		if err != nil {
			return nil, mapUserCreateErr(err)
		}
		return userFromPG(row), nil
	case r.sqliteQ != nil:
		row, err := r.sqliteQ.CreateUser(ctx, sqlitesqlc.CreateUserParams{
			ID:                   u.ID,
			Email:                u.Email,
			Name:                 u.Name,
			HashedPassword:       u.HashedPassword,
			PasswordInitialized:  boolToInt64(u.PasswordInitialized),
			ForcePasswordChange:  boolToInt64(u.ForcePasswordChange),
			TwoFactorEnabled:     boolToInt64(u.TwoFactorEnabled),
			TwoFactorSecret:      u.TwoFactorSecret,
			TwoFactorBackupCodes: u.TwoFactorBackupCodes,
			LastLoginAt:          nullTimeFromPtr(u.LastLoginAt),
			CreatedAt:            u.CreatedAt,
			UpdatedAt:            u.UpdatedAt,
		})
		if err != nil {
			return nil, mapUserCreateErr(err)
		}
		return userFromSQLite(row), nil
	default:
		return nil, r.unconfigured()
	}
}

func mapUserCreateErr(err error) error {
	if err == nil {
		return nil
	}
	low := err.Error()
	if containsAny(low, "UNIQUE", "unique", "duplicate", "Duplicate", "23505") {
		return repository.ErrDuplicate
	}
	return fmt.Errorf("sqlc: create user: %w", err)
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if len(sub) == 0 {
			continue
		}
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
	}
	return false
}

func (r *UserRepositorySQLC) FindByID(ctx context.Context, id string) (*domain.User, error) {
	switch {
	case r.pgQ != nil:
		row, err := r.pgQ.FindUserByID(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, repository.ErrNotFound
			}
			return nil, fmt.Errorf("sqlc: find user by id: %w", err)
		}
		return userFromPG(row), nil
	case r.sqliteQ != nil:
		row, err := r.sqliteQ.FindUserByID(ctx, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, repository.ErrNotFound
			}
			return nil, fmt.Errorf("sqlc: find user by id: %w", err)
		}
		return userFromSQLite(row), nil
	default:
		return nil, r.unconfigured()
	}
}

func (r *UserRepositorySQLC) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	switch {
	case r.pgQ != nil:
		row, err := r.pgQ.FindUserByEmail(ctx, email)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, repository.ErrNotFound
			}
			return nil, fmt.Errorf("sqlc: find user by email: %w", err)
		}
		return userFromPG(row), nil
	case r.sqliteQ != nil:
		row, err := r.sqliteQ.FindUserByEmail(ctx, email)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, repository.ErrNotFound
			}
			return nil, fmt.Errorf("sqlc: find user by email: %w", err)
		}
		return userFromSQLite(row), nil
	default:
		return nil, r.unconfigured()
	}
}

func (r *UserRepositorySQLC) Update(ctx context.Context, u *domain.User) error {
	u.UpdatedAt = time.Now()
	switch {
	case r.pgQ != nil:
		return r.pgQ.UpdateUser(ctx, pgsqlc.UpdateUserParams{
			ID:                   u.ID,
			Email:                u.Email,
			Name:                 u.Name,
			HashedPassword:       u.HashedPassword,
			PasswordInitialized:  u.PasswordInitialized,
			ForcePasswordChange:  u.ForcePasswordChange,
			TwoFactorEnabled:     u.TwoFactorEnabled,
			TwoFactorSecret:      u.TwoFactorSecret,
			TwoFactorBackupCodes: u.TwoFactorBackupCodes,
			LastLoginAt:          pgTimestampFromPtr(u.LastLoginAt),
			UpdatedAt:            pgtype.Timestamptz{Time: u.UpdatedAt, Valid: true},
		})
	case r.sqliteQ != nil:
		return r.sqliteQ.UpdateUser(ctx, sqlitesqlc.UpdateUserParams{
			ID:                   u.ID,
			Email:                u.Email,
			Name:                 u.Name,
			HashedPassword:       u.HashedPassword,
			PasswordInitialized:  boolToInt64(u.PasswordInitialized),
			ForcePasswordChange:  boolToInt64(u.ForcePasswordChange),
			TwoFactorEnabled:     boolToInt64(u.TwoFactorEnabled),
			TwoFactorSecret:      u.TwoFactorSecret,
			TwoFactorBackupCodes: u.TwoFactorBackupCodes,
			LastLoginAt:          nullTimeFromPtr(u.LastLoginAt),
			UpdatedAt:            u.UpdatedAt,
		})
	default:
		return r.unconfigured()
	}
}

func (r *UserRepositorySQLC) Delete(ctx context.Context, id string) error {
	switch {
	case r.pgQ != nil:
		return r.pgQ.DeleteUser(ctx, id)
	case r.sqliteQ != nil:
		return r.sqliteQ.DeleteUser(ctx, id)
	default:
		return r.unconfigured()
	}
}

func (r *UserRepositorySQLC) UpdatePassword(ctx context.Context, userID string, hashedPassword string) error {
	switch {
	case r.pgQ != nil:
		return r.pgQ.UpdateUserPassword(ctx, pgsqlc.UpdateUserPasswordParams{ID: userID, HashedPassword: hashedPassword})
	case r.sqliteQ != nil:
		return r.sqliteQ.UpdateUserPassword(ctx, sqlitesqlc.UpdateUserPasswordParams{ID: userID, HashedPassword: hashedPassword})
	default:
		return r.unconfigured()
	}
}

func (r *UserRepositorySQLC) UpdateLastLogin(ctx context.Context, userID string) error {
	switch {
	case r.pgQ != nil:
		return r.pgQ.UpdateUserLastLogin(ctx, userID)
	case r.sqliteQ != nil:
		return r.sqliteQ.UpdateUserLastLogin(ctx, userID)
	default:
		return r.unconfigured()
	}
}

func (r *UserRepositorySQLC) UpdateTwoFactorSecret(ctx context.Context, userID string, secret string, enabled bool) error {
	switch {
	case r.pgQ != nil:
		return r.pgQ.UpdateUserTwoFactorSecret(ctx, pgsqlc.UpdateUserTwoFactorSecretParams{
			ID:               userID,
			TwoFactorSecret:  secret,
			TwoFactorEnabled: enabled,
		})
	case r.sqliteQ != nil:
		return r.sqliteQ.UpdateUserTwoFactorSecret(ctx, sqlitesqlc.UpdateUserTwoFactorSecretParams{
			ID:               userID,
			TwoFactorSecret:  secret,
			TwoFactorEnabled: boolToInt64(enabled),
		})
	default:
		return r.unconfigured()
	}
}

// ---------- mapping helpers ----------

func userFromPG(row pgsqlc.User) *domain.User {
	out := &domain.User{
		Base: domain.Base{
			ID:        row.ID,
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
		},
		Email:                row.Email,
		Name:                 row.Name,
		HashedPassword:       row.HashedPassword,
		PasswordInitialized:  row.PasswordInitialized,
		ForcePasswordChange:  row.ForcePasswordChange,
		TwoFactorEnabled:     row.TwoFactorEnabled,
		TwoFactorSecret:      row.TwoFactorSecret,
		TwoFactorBackupCodes: row.TwoFactorBackupCodes,
		CreatedAt:            row.CreatedAt.Time,
		UpdatedAt:            row.UpdatedAt.Time,
	}
	if row.LastLoginAt.Valid {
		t := row.LastLoginAt.Time
		out.LastLoginAt = &t
	}
	return out
}

func userFromSQLite(row sqlitesqlc.User) *domain.User {
	out := &domain.User{
		Base: domain.Base{
			ID:        row.ID,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		Email:                row.Email,
		Name:                 row.Name,
		HashedPassword:       row.HashedPassword,
		PasswordInitialized:  row.PasswordInitialized != 0,
		ForcePasswordChange:  row.ForcePasswordChange != 0,
		TwoFactorEnabled:     row.TwoFactorEnabled != 0,
		TwoFactorSecret:      row.TwoFactorSecret,
		TwoFactorBackupCodes: row.TwoFactorBackupCodes,
		CreatedAt:            row.CreatedAt,
		UpdatedAt:            row.UpdatedAt,
	}
	if row.LastLoginAt.Valid {
		t := row.LastLoginAt.Time
		out.LastLoginAt = &t
	}
	return out
}
