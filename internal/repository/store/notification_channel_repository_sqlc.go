package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/denisakp/ogoune/pkg/crypto"
	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository"
	pgsqlc "github.com/denisakp/ogoune/internal/repository/sqlc/pg"
	sqlitesqlc "github.com/denisakp/ogoune/internal/repository/sqlc/sqlite"
)

type NotificationChannelRepositorySQLC struct {
	pgQ     *pgsqlc.Queries
	sqliteQ *sqlitesqlc.Queries
}

func NewNotificationChannelRepositorySQLC(rt SqlcRuntime) port.NotificationChannelRepository {
	r := &NotificationChannelRepositorySQLC{}
	if pool := rt.PgxPool(); pool != nil {
		r.pgQ = pgsqlc.New(pool)
	} else if db := rt.SQLiteDB(); db != nil {
		r.sqliteQ = sqlitesqlc.New(db)
	}
	return r
}

func (r *NotificationChannelRepositorySQLC) unconfigured() error {
	return fmt.Errorf("notification_channel_repository_sqlc: unconfigured runtime")
}

// encryptConfig mirrors NotificationChannel.BeforeCreate/BeforeUpdate guards.
func encryptChannelConfig(plaintext []byte) ([]byte, error) {
	if len(plaintext) == 0 {
		return plaintext, nil
	}
	ct, err := crypto.EncryptChannelConfig(string(plaintext))
	if err != nil {
		return nil, err
	}
	return []byte(ct), nil
}

// decryptConfig mirrors NotificationChannel.AfterFind guard.
// Legacy plaintext migration is OUT OF SCOPE for the sqlc wrapper (research §2).
func decryptChannelConfig(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) == 0 {
		return ciphertext, nil
	}
	pt, err := crypto.DecryptChannelConfig(string(ciphertext))
	if err != nil {
		return nil, err
	}
	return []byte(pt), nil
}

func (r *NotificationChannelRepositorySQLC) Create(ctx context.Context, ch *domain.NotificationChannel) error {
	ch.EnsureID()
	now := time.Now()
	if ch.CreatedAt.IsZero() {
		ch.CreatedAt = now
	}
	if ch.UpdatedAt.IsZero() {
		ch.UpdatedAt = now
	}
	ct, err := encryptChannelConfig(ch.Config)
	if err != nil {
		return err
	}
	switch {
	case r.pgQ != nil:
		return r.pgQ.CreateNotificationChannel(ctx, pgsqlc.CreateNotificationChannelParams{
			ID:               ch.ID,
			CreatedAt:        pgtype.Timestamptz{Time: ch.CreatedAt, Valid: true},
			UpdatedAt:        pgtype.Timestamptz{Time: ch.UpdatedAt, Valid: true},
			Name:             ch.Name,
			Type:             string(ch.Type),
			Config:           ct,
			EnabledByDefault: ch.EnabledByDefault,
		})
	case r.sqliteQ != nil:
		return r.sqliteQ.CreateNotificationChannel(ctx, sqlitesqlc.CreateNotificationChannelParams{
			ID:               ch.ID,
			CreatedAt:        ch.CreatedAt,
			UpdatedAt:        ch.UpdatedAt,
			Name:             ch.Name,
			Type:             string(ch.Type),
			Config:           ct,
			EnabledByDefault: boolToInt64(ch.EnabledByDefault),
		})
	default:
		return r.unconfigured()
	}
}

func (r *NotificationChannelRepositorySQLC) FindByID(ctx context.Context, id string) (*domain.NotificationChannel, error) {
	switch {
	case r.pgQ != nil:
		row, err := r.pgQ.FindNotificationChannelByID(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, repository.ErrNotFound
			}
			return nil, fmt.Errorf("sqlc: find notification channel by id: %w", err)
		}
		return channelFromPG(row)
	case r.sqliteQ != nil:
		row, err := r.sqliteQ.FindNotificationChannelByID(ctx, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, repository.ErrNotFound
			}
			return nil, fmt.Errorf("sqlc: find notification channel by id: %w", err)
		}
		return channelFromSQLite(row)
	default:
		return nil, r.unconfigured()
	}
}

func (r *NotificationChannelRepositorySQLC) List(ctx context.Context, limit, offset int) ([]*domain.NotificationChannel, error) {
	switch {
	case r.pgQ != nil:
		rows, err := r.pgQ.ListNotificationChannels(ctx, pgsqlc.ListNotificationChannelsParams{
			Limit:  int32(limit),
			Offset: int32(offset),
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: list notification channels: %w", err)
		}
		return channelsFromPG(rows)
	case r.sqliteQ != nil:
		rows, err := r.sqliteQ.ListNotificationChannels(ctx, sqlitesqlc.ListNotificationChannelsParams{
			Limit:  int64(limit),
			Offset: int64(offset),
		})
		if err != nil {
			return nil, fmt.Errorf("sqlc: list notification channels: %w", err)
		}
		return channelsFromSQLite(rows)
	default:
		return nil, r.unconfigured()
	}
}

func (r *NotificationChannelRepositorySQLC) Update(ctx context.Context, ch *domain.NotificationChannel) error {
	ch.UpdatedAt = time.Now()
	ct, err := encryptChannelConfig(ch.Config)
	if err != nil {
		return err
	}
	switch {
	case r.pgQ != nil:
		n, err := r.pgQ.UpdateNotificationChannel(ctx, pgsqlc.UpdateNotificationChannelParams{
			ID:               ch.ID,
			Name:             ch.Name,
			Type:             string(ch.Type),
			Config:           ct,
			EnabledByDefault: ch.EnabledByDefault,
			UpdatedAt:        pgtype.Timestamptz{Time: ch.UpdatedAt, Valid: true},
		})
		if err != nil {
			return fmt.Errorf("sqlc: update notification channel: %w", err)
		}
		if n == 0 {
			return repository.ErrNotFound
		}
		return nil
	case r.sqliteQ != nil:
		n, err := r.sqliteQ.UpdateNotificationChannel(ctx, sqlitesqlc.UpdateNotificationChannelParams{
			ID:               ch.ID,
			Name:             ch.Name,
			Type:             string(ch.Type),
			Config:           ct,
			EnabledByDefault: boolToInt64(ch.EnabledByDefault),
			UpdatedAt:        ch.UpdatedAt,
		})
		if err != nil {
			return fmt.Errorf("sqlc: update notification channel: %w", err)
		}
		if n == 0 {
			return repository.ErrNotFound
		}
		return nil
	default:
		return r.unconfigured()
	}
}

func (r *NotificationChannelRepositorySQLC) Delete(ctx context.Context, id string) error {
	switch {
	case r.pgQ != nil:
		n, err := r.pgQ.DeleteNotificationChannel(ctx, id)
		if err != nil {
			return fmt.Errorf("sqlc: delete notification channel: %w", err)
		}
		if n == 0 {
			return repository.ErrNotFound
		}
		return nil
	case r.sqliteQ != nil:
		n, err := r.sqliteQ.DeleteNotificationChannel(ctx, id)
		if err != nil {
			return fmt.Errorf("sqlc: delete notification channel: %w", err)
		}
		if n == 0 {
			return repository.ErrNotFound
		}
		return nil
	default:
		return r.unconfigured()
	}
}

func (r *NotificationChannelRepositorySQLC) FindByType(ctx context.Context, channelType domain.NotificationChannelType) ([]*domain.NotificationChannel, error) {
	switch {
	case r.pgQ != nil:
		rows, err := r.pgQ.FindNotificationChannelsByType(ctx, string(channelType))
		if err != nil {
			return nil, fmt.Errorf("sqlc: find by type: %w", err)
		}
		return channelsFromPG(rows)
	case r.sqliteQ != nil:
		rows, err := r.sqliteQ.FindNotificationChannelsByType(ctx, string(channelType))
		if err != nil {
			return nil, fmt.Errorf("sqlc: find by type: %w", err)
		}
		return channelsFromSQLite(rows)
	default:
		return nil, r.unconfigured()
	}
}

func (r *NotificationChannelRepositorySQLC) FindDefaultChannels(ctx context.Context) ([]*domain.NotificationChannel, error) {
	switch {
	case r.pgQ != nil:
		rows, err := r.pgQ.FindDefaultNotificationChannels(ctx)
		if err != nil {
			return nil, fmt.Errorf("sqlc: find default channels: %w", err)
		}
		return channelsFromPG(rows)
	case r.sqliteQ != nil:
		rows, err := r.sqliteQ.FindDefaultNotificationChannels(ctx)
		if err != nil {
			return nil, fmt.Errorf("sqlc: find default channels: %w", err)
		}
		return channelsFromSQLite(rows)
	default:
		return nil, r.unconfigured()
	}
}

func (r *NotificationChannelRepositorySQLC) FindByResourceID(ctx context.Context, resourceID string) ([]*domain.NotificationChannel, error) {
	switch {
	case r.pgQ != nil:
		rows, err := r.pgQ.FindNotificationChannelsByResourceID(ctx, resourceID)
		if err != nil {
			return nil, fmt.Errorf("sqlc: find by resource id: %w", err)
		}
		return channelsFromPG(rows)
	case r.sqliteQ != nil:
		rows, err := r.sqliteQ.FindNotificationChannelsByResourceID(ctx, resourceID)
		if err != nil {
			return nil, fmt.Errorf("sqlc: find by resource id: %w", err)
		}
		return channelsFromSQLite(rows)
	default:
		return nil, r.unconfigured()
	}
}

func (r *NotificationChannelRepositorySQLC) FindByComponentID(ctx context.Context, componentID string) ([]*domain.NotificationChannel, error) {
	switch {
	case r.pgQ != nil:
		rows, err := r.pgQ.FindNotificationChannelsByComponentID(ctx, componentID)
		if err != nil {
			return nil, fmt.Errorf("sqlc: find by component id: %w", err)
		}
		return channelsFromPG(rows)
	case r.sqliteQ != nil:
		rows, err := r.sqliteQ.FindNotificationChannelsByComponentID(ctx, componentID)
		if err != nil {
			return nil, fmt.Errorf("sqlc: find by component id: %w", err)
		}
		return channelsFromSQLite(rows)
	default:
		return nil, r.unconfigured()
	}
}

// ---------- mapping helpers ----------

func channelFromPG(row pgsqlc.NotificationChannel) (*domain.NotificationChannel, error) {
	pt, err := decryptChannelConfig(row.Config)
	if err != nil {
		return nil, err
	}
	return &domain.NotificationChannel{
		Base: domain.Base{
			ID:        row.ID,
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
		},
		Name:             row.Name,
		Type:             domain.NotificationChannelType(row.Type),
		Config:           pt,
		EnabledByDefault: row.EnabledByDefault,
	}, nil
}

func channelsFromPG(rows []pgsqlc.NotificationChannel) ([]*domain.NotificationChannel, error) {
	out := make([]*domain.NotificationChannel, len(rows))
	for i, row := range rows {
		ch, err := channelFromPG(row)
		if err != nil {
			return nil, err
		}
		out[i] = ch
	}
	return out, nil
}

func channelFromSQLite(row sqlitesqlc.NotificationChannel) (*domain.NotificationChannel, error) {
	pt, err := decryptChannelConfig(row.Config)
	if err != nil {
		return nil, err
	}
	return &domain.NotificationChannel{
		Base: domain.Base{
			ID:        row.ID,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		Name:             row.Name,
		Type:             domain.NotificationChannelType(row.Type),
		Config:           pt,
		EnabledByDefault: row.EnabledByDefault != 0,
	}, nil
}

func channelsFromSQLite(rows []sqlitesqlc.NotificationChannel) ([]*domain.NotificationChannel, error) {
	out := make([]*domain.NotificationChannel, len(rows))
	for i, row := range rows {
		ch, err := channelFromSQLite(row)
		if err != nil {
			return nil, err
		}
		out[i] = ch
	}
	return out, nil
}
