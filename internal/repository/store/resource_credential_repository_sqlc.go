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
	"github.com/denisakp/ogoune/pkg/crypto"
)

type ResourceCredentialRepositorySQLC struct {
	pgQ     *pgsqlc.Queries
	sqliteQ *sqlitesqlc.Queries
}

func NewResourceCredentialRepositorySQLC(rt SqlcRuntime) port.ResourceCredentialRepository {
	r := &ResourceCredentialRepositorySQLC{}
	if pool := rt.PgxPool(); pool != nil {
		r.pgQ = pgsqlc.New(pool)
	} else if db := rt.SQLiteDB(); db != nil {
		r.sqliteQ = sqlitesqlc.New(db)
	}
	return r
}

func (r *ResourceCredentialRepositorySQLC) unconfigured() error {
	return fmt.Errorf("resource_credential_repository_sqlc: unconfigured runtime")
}

// encryptCredentialFields mirrors ResourceCredential.encryptSecrets guards.
func encryptCredentialPassword(plaintext []byte) ([]byte, error) {
	if len(plaintext) == 0 {
		return plaintext, nil
	}
	ct, err := crypto.EncryptCredentialPassword(string(plaintext))
	if err != nil {
		return nil, err
	}
	return []byte(ct), nil
}

func encryptCredentialOptions(plaintext []byte) ([]byte, error) {
	if len(plaintext) == 0 {
		return plaintext, nil
	}
	ct, err := crypto.EncryptCredentialOptions(string(plaintext))
	if err != nil {
		return nil, err
	}
	return []byte(ct), nil
}

func decryptCredentialPassword(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) == 0 {
		return ciphertext, nil
	}
	pt, err := crypto.DecryptCredentialPassword(string(ciphertext))
	if err != nil {
		return nil, domain.ErrCredentialDecryption
	}
	return []byte(pt), nil
}

func decryptCredentialOptions(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) == 0 {
		return ciphertext, nil
	}
	pt, err := crypto.DecryptCredentialOptions(string(ciphertext))
	if err != nil {
		return nil, domain.ErrCredentialDecryption
	}
	return []byte(pt), nil
}

func (r *ResourceCredentialRepositorySQLC) Get(ctx context.Context, resourceID string) (*domain.ResourceCredential, error) {
	switch {
	case r.pgQ != nil:
		row, err := r.pgQ.GetResourceCredentialByResourceID(ctx, resourceID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, repository.ErrCredentialNotFound
			}
			return nil, fmt.Errorf("sqlc: get resource credential: %w", err)
		}
		return credentialFromPG(row)
	case r.sqliteQ != nil:
		row, err := r.sqliteQ.GetResourceCredentialByResourceID(ctx, resourceID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, repository.ErrCredentialNotFound
			}
			return nil, fmt.Errorf("sqlc: get resource credential: %w", err)
		}
		return credentialFromSQLite(row)
	default:
		return nil, r.unconfigured()
	}
}

func (r *ResourceCredentialRepositorySQLC) Upsert(ctx context.Context, c *domain.ResourceCredential) error {
	c.EnsureID()
	now := time.Now()
	if c.CreatedAt.IsZero() {
		c.CreatedAt = now
	}
	c.UpdatedAt = now
	encPwd, err := encryptCredentialPassword(c.Password)
	if err != nil {
		return err
	}
	encOpts, err := encryptCredentialOptions(c.Options)
	if err != nil {
		return err
	}
	switch {
	case r.pgQ != nil:
		return r.pgQ.UpsertResourceCredential(ctx, pgsqlc.UpsertResourceCredentialParams{
			ID:         c.ID,
			ResourceID: c.ResourceID,
			Username:   pgtype.Text{String: c.Username, Valid: c.Username != ""},
			Password:   encPwd,
			Options:    encOpts,
			CreatedAt:  pgtype.Timestamptz{Time: c.CreatedAt, Valid: true},
			UpdatedAt:  pgtype.Timestamptz{Time: c.UpdatedAt, Valid: true},
		})
	case r.sqliteQ != nil:
		return r.sqliteQ.UpsertResourceCredential(ctx, sqlitesqlc.UpsertResourceCredentialParams{
			ID:         c.ID,
			ResourceID: c.ResourceID,
			Username:   sql.NullString{String: c.Username, Valid: c.Username != ""},
			Password:   encPwd,
			Options:    encOpts,
			CreatedAt:  c.CreatedAt,
			UpdatedAt:  c.UpdatedAt,
		})
	default:
		return r.unconfigured()
	}
}

func (r *ResourceCredentialRepositorySQLC) Delete(ctx context.Context, resourceID string) error {
	switch {
	case r.pgQ != nil:
		n, err := r.pgQ.DeleteResourceCredentialByResourceID(ctx, resourceID)
		if err != nil {
			return fmt.Errorf("sqlc: delete resource credential: %w", err)
		}
		if n == 0 {
			return repository.ErrCredentialNotFound
		}
		return nil
	case r.sqliteQ != nil:
		n, err := r.sqliteQ.DeleteResourceCredentialByResourceID(ctx, resourceID)
		if err != nil {
			return fmt.Errorf("sqlc: delete resource credential: %w", err)
		}
		if n == 0 {
			return repository.ErrCredentialNotFound
		}
		return nil
	default:
		return r.unconfigured()
	}
}

func (r *ResourceCredentialRepositorySQLC) Exists(ctx context.Context, resourceID string) (bool, error) {
	switch {
	case r.pgQ != nil:
		n, err := r.pgQ.ResourceCredentialExists(ctx, resourceID)
		if err != nil {
			return false, fmt.Errorf("sqlc: resource credential exists: %w", err)
		}
		return n > 0, nil
	case r.sqliteQ != nil:
		n, err := r.sqliteQ.ResourceCredentialExists(ctx, resourceID)
		if err != nil {
			return false, fmt.Errorf("sqlc: resource credential exists: %w", err)
		}
		return n > 0, nil
	default:
		return false, r.unconfigured()
	}
}

// ---------- mapping helpers ----------

func credentialFromPG(row pgsqlc.ResourceCredential) (*domain.ResourceCredential, error) {
	pwd, err := decryptCredentialPassword(row.Password)
	if err != nil {
		return nil, err
	}
	opts, err := decryptCredentialOptions(row.Options)
	if err != nil {
		return nil, err
	}
	username := ""
	if row.Username.Valid {
		username = row.Username.String
	}
	return &domain.ResourceCredential{
		Base: domain.Base{
			ID:        row.ID,
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
		},
		ResourceID: row.ResourceID,
		Username:   username,
		Password:   pwd,
		Options:    opts,
	}, nil
}

func credentialFromSQLite(row sqlitesqlc.ResourceCredential) (*domain.ResourceCredential, error) {
	pwd, err := decryptCredentialPassword(row.Password)
	if err != nil {
		return nil, err
	}
	opts, err := decryptCredentialOptions(row.Options)
	if err != nil {
		return nil, err
	}
	username := ""
	if row.Username.Valid {
		username = row.Username.String
	}
	return &domain.ResourceCredential{
		Base: domain.Base{
			ID:        row.ID,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		ResourceID: row.ResourceID,
		Username:   username,
		Password:   pwd,
		Options:    opts,
	}, nil
}
