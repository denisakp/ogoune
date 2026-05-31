package store_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
	"github.com/denisakp/ogoune/pkg/crypto"
)

const notifChannelTestKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func setupChannelCryptoKey(t *testing.T) {
	t.Helper()
	t.Setenv("APP_SECRET_KEY", notifChannelTestKey)
	crypto.SetGlobalProvider(&crypto.EnvKeyProvider{})
	t.Cleanup(func() {
		crypto.SetGlobalProvider(&crypto.EnvKeyProvider{})
	})
}

func TestNotificationChannelRepository_Contract(t *testing.T) {
	setupChannelCryptoKey(t)
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		repo := store.NewNotificationChannelRepository(fx.Runtime.GormDB())
		runNotificationChannelContract(t, repo)
	})
}

func runNotificationChannelContract(t *testing.T, repo port.NotificationChannelRepository) {
	t.Helper()
	ctx := context.Background()
	tag := fmt.Sprintf("%d", time.Now().UnixNano())

	t.Run("Create_and_FindByID_round_trip_plaintext", func(t *testing.T) {
		plaintext := []byte(`{"webhook":"https://example.invalid/hook"}`)
		ch := &domain.NotificationChannel{
			Base:   domain.Base{ID: "01NC" + tag + "001"},
			Name:   "Webhook " + tag,
			Type:   domain.NotificationChannelType("webhook"),
			Config: plaintext,
		}
		require.NoError(t, repo.Create(ctx, ch))

		got, err := repo.FindByID(ctx, ch.ID)
		require.NoError(t, err)
		assert.Equal(t, "Webhook "+tag, got.Name)
		assert.Equal(t, plaintext, got.Config, "Config must round-trip as plaintext via the port")
	})

	t.Run("FindByID_NotFound", func(t *testing.T) {
		_, err := repo.FindByID(ctx, "nope-channel-"+tag)
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})

	t.Run("Update_round_trip", func(t *testing.T) {
		ch := &domain.NotificationChannel{
			Base:   domain.Base{ID: "01NC" + tag + "002"},
			Name:   "Original",
			Type:   domain.NotificationChannelType("webhook"),
			Config: []byte(`{"v":1}`),
		}
		require.NoError(t, repo.Create(ctx, ch))

		ch.Name = "Updated"
		ch.Config = []byte(`{"v":2}`)
		require.NoError(t, repo.Update(ctx, ch))

		got, err := repo.FindByID(ctx, ch.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated", got.Name)
		assert.Equal(t, []byte(`{"v":2}`), got.Config)
	})

	// NOTE: Update_NotFound is intentionally not asserted at contract level —
	// GORM's Save() upserts; sqlc impl returns ErrNotFound via :execrows.
	// Behavior divergence is GORM-specific and not required by the port contract.

	t.Run("Delete_and_Delete_NotFound", func(t *testing.T) {
		ch := &domain.NotificationChannel{
			Base:   domain.Base{ID: "01NC" + tag + "003"},
			Name:   "ToDelete",
			Type:   domain.NotificationChannelType("webhook"),
			Config: []byte(`{}`),
		}
		require.NoError(t, repo.Create(ctx, ch))
		require.NoError(t, repo.Delete(ctx, ch.ID))

		_, err := repo.FindByID(ctx, ch.ID)
		assert.ErrorIs(t, err, repository.ErrNotFound)

		err = repo.Delete(ctx, ch.ID)
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})

	t.Run("List_pagination", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			require.NoError(t, repo.Create(ctx, &domain.NotificationChannel{
				Base:   domain.Base{ID: fmt.Sprintf("01NC%sLIST%02d", tag, i)},
				Name:   fmt.Sprintf("list-%d", i),
				Type:   domain.NotificationChannelType("webhook"),
				Config: []byte(`{}`),
			}))
		}
		page, err := repo.List(ctx, 2, 0)
		require.NoError(t, err)
		assert.Len(t, page, 2)
	})

	t.Run("FindByType", func(t *testing.T) {
		uniqueType := domain.NotificationChannelType("type-" + tag)
		ch := &domain.NotificationChannel{
			Base:   domain.Base{ID: "01NC" + tag + "TYPE"},
			Name:   "by-type",
			Type:   uniqueType,
			Config: []byte(`{}`),
		}
		require.NoError(t, repo.Create(ctx, ch))

		got, err := repo.FindByType(ctx, uniqueType)
		require.NoError(t, err)
		assert.Len(t, got, 1)
		assert.Equal(t, ch.ID, got[0].ID)
	})

	t.Run("FindDefaultChannels", func(t *testing.T) {
		ch := &domain.NotificationChannel{
			Base:             domain.Base{ID: "01NC" + tag + "DEF"},
			Name:             "default-on",
			Type:             domain.NotificationChannelType("webhook"),
			Config:           []byte(`{}`),
			EnabledByDefault: true,
		}
		require.NoError(t, repo.Create(ctx, ch))

		got, err := repo.FindDefaultChannels(ctx)
		require.NoError(t, err)
		found := false
		for _, c := range got {
			if c.ID == ch.ID {
				found = true
				break
			}
		}
		assert.True(t, found, "default channel must appear in FindDefaultChannels")
	})

	// FindByResourceID + FindByComponentID return empty slice when no junction rows;
	// per-junction integration tested in service layer. Contract just exercises the
	// happy-path "no joined rows → empty slice".
	t.Run("FindByResourceID_empty", func(t *testing.T) {
		got, err := repo.FindByResourceID(ctx, "nonexistent-res-"+tag)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("FindByComponentID_empty", func(t *testing.T) {
		got, err := repo.FindByComponentID(ctx, "nonexistent-comp-"+tag)
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}
