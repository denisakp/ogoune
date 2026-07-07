package store_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/database"
	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
)

func TestNotificationChannelRepository_SqlcContract(t *testing.T) {
	setupChannelCryptoKey(t)
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		repo := store.NewNotificationChannelRepositorySQLC(fx.Runtime)
		runNotificationChannelContract(t, repo)
	})
}

// TestNotificationChannelRepository_SqlcEncryption_RoundTrip is the SC-006
// gate for notification_channel: the on-disk column carries ciphertext, the
// port-level read returns plaintext.
func TestNotificationChannelRepository_SqlcEncryption_RoundTrip(t *testing.T) {
	setupChannelCryptoKey(t)
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		ctx := context.Background()
		repo := store.NewNotificationChannelRepositorySQLC(fx.Runtime)

		plaintext := []byte(fmt.Sprintf(`{"webhook":"https://enc-%d"}`, time.Now().UnixNano()))
		ch := &domain.NotificationChannel{
			Name:   "encryption-roundtrip",
			Type:   domain.NotificationChannelType("webhook"),
			Config: plaintext,
		}
		require.NoError(t, repo.Create(ctx, ch))
		require.NotEmpty(t, ch.ID)

		// (a) raw read confirms ciphertext on disk.
		var rawConfig []byte
		switch fx.Runtime.Driver {
		case database.DriverPostgres:
			err := fx.Runtime.PgxPool().QueryRow(ctx,
				"SELECT config FROM notification_channels WHERE id=$1", ch.ID).Scan(&rawConfig)
			require.NoError(t, err)
		case database.DriverSQLite:
			err := fx.Runtime.SQLiteDB().QueryRowContext(ctx,
				"SELECT config FROM notification_channels WHERE id=?", ch.ID).Scan(&rawConfig)
			require.NoError(t, err)
		}
		require.NotEqual(t, plaintext, rawConfig, "row must contain ciphertext, not plaintext")
		require.NotEmpty(t, rawConfig)

		// (b) port read returns the original plaintext.
		got, err := repo.FindByID(ctx, ch.ID)
		require.NoError(t, err)
		assert.Equal(t, plaintext, got.Config)
	})
}

// silence unused if sql is only used by another file
var _ = sql.ErrNoRows
