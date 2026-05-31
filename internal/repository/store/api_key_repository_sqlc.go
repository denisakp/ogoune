package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository"
	pgsqlc "github.com/denisakp/ogoune/internal/repository/sqlc/pg"
	sqlitesqlc "github.com/denisakp/ogoune/internal/repository/sqlc/sqlite"
)

type APIKeyRepositorySQLC struct {
	pgQ     *pgsqlc.Queries
	sqliteQ *sqlitesqlc.Queries
}

func NewAPIKeyRepositorySQLC(rt SqlcRuntime) port.APIKeyRepository {
	r := &APIKeyRepositorySQLC{}
	if pool := rt.PgxPool(); pool != nil {
		r.pgQ = pgsqlc.New(pool)
	} else if db := rt.SQLiteDB(); db != nil {
		r.sqliteQ = sqlitesqlc.New(db)
	}
	return r
}

func (r *APIKeyRepositorySQLC) unconfigured() error {
	return fmt.Errorf("api_key_repository_sqlc: unconfigured runtime")
}

func (r *APIKeyRepositorySQLC) Create(ctx context.Context, k *domain.APIKey) error {
	k.EnsureID()
	now := time.Now()
	if k.CreatedAt.IsZero() {
		k.CreatedAt = now
	}
	if k.UpdatedAt.IsZero() {
		k.UpdatedAt = now
	}
	switch {
	case r.pgQ != nil:
		err := r.pgQ.CreateAPIKey(ctx, pgsqlc.CreateAPIKeyParams{
			ID:         k.ID,
			CreatedAt:  pgtype.Timestamptz{Time: k.CreatedAt, Valid: true},
			UpdatedAt:  pgtype.Timestamptz{Time: k.UpdatedAt, Valid: true},
			UserID:     k.UserID,
			Name:       k.Name,
			KeyHash:    k.KeyHash,
			KeyPrefix:  k.KeyPrefix,
			Scope:      string(k.Scope),
			ExpiresAt:  pgTimestampFromPtr(k.ExpiresAt),
			LastUsedAt: pgTimestampFromPtr(k.LastUsedAt),
			LastUsedIp: k.LastUsedIP,
			IsActive:   k.IsActive,
		})
		return mapAPIKeyCreateErr(err)
	case r.sqliteQ != nil:
		err := r.sqliteQ.CreateAPIKey(ctx, sqlitesqlc.CreateAPIKeyParams{
			ID:         k.ID,
			CreatedAt:  k.CreatedAt,
			UpdatedAt:  k.UpdatedAt,
			UserID:     k.UserID,
			Name:       k.Name,
			KeyHash:    k.KeyHash,
			KeyPrefix:  k.KeyPrefix,
			Scope:      string(k.Scope),
			ExpiresAt:  nullTimeFromPtr(k.ExpiresAt),
			LastUsedAt: nullTimeFromPtr(k.LastUsedAt),
			LastUsedIp: k.LastUsedIP,
			IsActive:   boolToInt64(k.IsActive),
		})
		return mapAPIKeyCreateErr(err)
	default:
		return r.unconfigured()
	}
}

func mapAPIKeyCreateErr(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return repository.ErrDuplicate
	}
	low := strings.ToLower(err.Error())
	if strings.Contains(low, "unique") || strings.Contains(low, "duplicate") {
		return repository.ErrDuplicate
	}
	return fmt.Errorf("sqlc: create api key: %w", err)
}

func (r *APIKeyRepositorySQLC) FindByID(ctx context.Context, id, userID string) (*domain.APIKey, error) {
	switch {
	case r.pgQ != nil:
		row, err := r.pgQ.FindAPIKeyByIDForUser(ctx, pgsqlc.FindAPIKeyByIDForUserParams{ID: id, UserID: userID})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, repository.ErrNotFound
			}
			return nil, fmt.Errorf("sqlc: find api key by id: %w", err)
		}
		return apiKeyFromPG(row), nil
	case r.sqliteQ != nil:
		row, err := r.sqliteQ.FindAPIKeyByIDForUser(ctx, sqlitesqlc.FindAPIKeyByIDForUserParams{ID: id, UserID: userID})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, repository.ErrNotFound
			}
			return nil, fmt.Errorf("sqlc: find api key by id: %w", err)
		}
		return apiKeyFromSQLite(row), nil
	default:
		return nil, r.unconfigured()
	}
}

func (r *APIKeyRepositorySQLC) FindByKeyHash(ctx context.Context, keyHash string) (*domain.APIKey, error) {
	switch {
	case r.pgQ != nil:
		row, err := r.pgQ.FindAPIKeyByKeyHash(ctx, keyHash)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, repository.ErrNotFound
			}
			return nil, fmt.Errorf("sqlc: find api key by hash: %w", err)
		}
		return apiKeyFromPG(row), nil
	case r.sqliteQ != nil:
		row, err := r.sqliteQ.FindAPIKeyByKeyHash(ctx, keyHash)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, repository.ErrNotFound
			}
			return nil, fmt.Errorf("sqlc: find api key by hash: %w", err)
		}
		return apiKeyFromSQLite(row), nil
	default:
		return nil, r.unconfigured()
	}
}

func (r *APIKeyRepositorySQLC) ListByUserID(ctx context.Context, userID string) ([]domain.APIKey, error) {
	switch {
	case r.pgQ != nil:
		rows, err := r.pgQ.ListAPIKeysByUserID(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("sqlc: list api keys: %w", err)
		}
		out := make([]domain.APIKey, len(rows))
		for i, row := range rows {
			out[i] = *apiKeyFromPG(row)
		}
		return out, nil
	case r.sqliteQ != nil:
		rows, err := r.sqliteQ.ListAPIKeysByUserID(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("sqlc: list api keys: %w", err)
		}
		out := make([]domain.APIKey, len(rows))
		for i, row := range rows {
			out[i] = *apiKeyFromSQLite(row)
		}
		return out, nil
	default:
		return nil, r.unconfigured()
	}
}

func (r *APIKeyRepositorySQLC) UpdateLastUsed(ctx context.Context, id string, at time.Time, ip string) error {
	switch {
	case r.pgQ != nil:
		n, err := r.pgQ.UpdateAPIKeyLastUsed(ctx, pgsqlc.UpdateAPIKeyLastUsedParams{
			ID:         id,
			LastUsedAt: pgtype.Timestamptz{Time: at, Valid: true},
			LastUsedIp: ip,
		})
		if err != nil {
			return fmt.Errorf("sqlc: update last used: %w", err)
		}
		if n == 0 {
			return repository.ErrNotFound
		}
		return nil
	case r.sqliteQ != nil:
		n, err := r.sqliteQ.UpdateAPIKeyLastUsed(ctx, sqlitesqlc.UpdateAPIKeyLastUsedParams{
			ID:         id,
			LastUsedAt: sql.NullTime{Time: at, Valid: true},
			LastUsedIp: ip,
		})
		if err != nil {
			return fmt.Errorf("sqlc: update last used: %w", err)
		}
		if n == 0 {
			return repository.ErrNotFound
		}
		return nil
	default:
		return r.unconfigured()
	}
}

func (r *APIKeyRepositorySQLC) Revoke(ctx context.Context, id, userID string) error {
	switch {
	case r.pgQ != nil:
		n, err := r.pgQ.RevokeAPIKey(ctx, pgsqlc.RevokeAPIKeyParams{ID: id, UserID: userID})
		if err != nil {
			return fmt.Errorf("sqlc: revoke api key: %w", err)
		}
		if n == 0 {
			return repository.ErrNotFound
		}
		return nil
	case r.sqliteQ != nil:
		n, err := r.sqliteQ.RevokeAPIKey(ctx, sqlitesqlc.RevokeAPIKeyParams{ID: id, UserID: userID})
		if err != nil {
			return fmt.Errorf("sqlc: revoke api key: %w", err)
		}
		if n == 0 {
			return repository.ErrNotFound
		}
		return nil
	default:
		return r.unconfigured()
	}
}

func (r *APIKeyRepositorySQLC) CountByUserID(ctx context.Context, userID string) (int64, error) {
	switch {
	case r.pgQ != nil:
		n, err := r.pgQ.CountAPIKeysByUserID(ctx, userID)
		if err != nil {
			return 0, fmt.Errorf("sqlc: count api keys: %w", err)
		}
		return n, nil
	case r.sqliteQ != nil:
		n, err := r.sqliteQ.CountAPIKeysByUserID(ctx, userID)
		if err != nil {
			return 0, fmt.Errorf("sqlc: count api keys: %w", err)
		}
		return n, nil
	default:
		return 0, r.unconfigured()
	}
}

// ---------- mapping helpers ----------

func apiKeyFromPG(row pgsqlc.ApiKey) *domain.APIKey {
	out := &domain.APIKey{
		Base: domain.Base{
			ID:        row.ID,
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
		},
		UserID:     row.UserID,
		Name:       row.Name,
		KeyHash:    row.KeyHash,
		KeyPrefix:  row.KeyPrefix,
		Scope:      domain.APIKeyScope(row.Scope),
		LastUsedIP: row.LastUsedIp,
		IsActive:   row.IsActive,
	}
	if row.ExpiresAt.Valid {
		t := row.ExpiresAt.Time
		out.ExpiresAt = &t
	}
	if row.LastUsedAt.Valid {
		t := row.LastUsedAt.Time
		out.LastUsedAt = &t
	}
	return out
}

func apiKeyFromSQLite(row sqlitesqlc.ApiKey) *domain.APIKey {
	out := &domain.APIKey{
		Base: domain.Base{
			ID:        row.ID,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		UserID:     row.UserID,
		Name:       row.Name,
		KeyHash:    row.KeyHash,
		KeyPrefix:  row.KeyPrefix,
		Scope:      domain.APIKeyScope(row.Scope),
		LastUsedIP: row.LastUsedIp,
		IsActive:   row.IsActive != 0,
	}
	if row.ExpiresAt.Valid {
		t := row.ExpiresAt.Time
		out.ExpiresAt = &t
	}
	if row.LastUsedAt.Valid {
		t := row.LastUsedAt.Time
		out.LastUsedAt = &t
	}
	return out
}

// ---------- shared helpers (used by other Wave-1 wrappers) ----------

func pgTimestampFromPtr(p *time.Time) pgtype.Timestamptz {
	if p == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *p, Valid: true}
}

func nullTimeFromPtr(p *time.Time) sql.NullTime {
	if p == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *p, Valid: true}
}

func boolToInt64(b bool) int64 {
	if b {
		return 1
	}
	return 0
}
