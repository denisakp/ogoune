package store_test

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/pkg/crypto"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
)

const resourcePreloadTestKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func setupPreloadCryptoKey(t *testing.T) {
	t.Helper()
	t.Setenv("APP_SECRET_KEY", resourcePreloadTestKey)
	crypto.SetGlobalProvider(&crypto.EnvKeyProvider{})
}

// TestResourceRepository_M2M_Channels_Transitions exercises SC-006 for the
// NotificationChannels side (PR3 of US1): 0→N, N→M (overlap), N→0 on both
// dialects.
func TestResourceRepository_M2M_Channels_Transitions(t *testing.T) {
	setupPreloadCryptoKey(t)
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		repo := store.NewResourceRepositorySQLC(fx.Runtime)
		chRepo := store.NewNotificationChannelRepositorySQLC(fx.Runtime)
		ctx := context.Background()

		channelIDs := make([]string, 5)
		for i := 0; i < 5; i++ {
			ch := &domain.NotificationChannel{
				Base:   domain.Base{ID: "", CreatedAt: time.Now()},
				Name:   "ch-" + string(rune('a'+i)),
				Type:   domain.NotificationChannelTypeSlack,
				Config: []byte(`{"webhook":"https://example.com"}`),
			}
			require.NoError(t, chRepo.Create(ctx, ch))
			channelIDs[i] = ch.ID
		}

		t.Run("0_to_N", func(t *testing.T) {
			res := &domain.Resource{
				Base:     domain.Base{ID: "ch-0-to-n", CreatedAt: time.Now()},
				Name:     "0→N", Type: domain.ResourceHTTP, Target: "https://example.com",
				IsActive: true,
				NotificationChannels: []*domain.NotificationChannel{
					{Base: domain.Base{ID: channelIDs[0]}},
					{Base: domain.Base{ID: channelIDs[1]}},
				},
			}
			_, err := repo.Create(ctx, res)
			require.NoError(t, err)
			loaded, err := repo.FindByID(ctx, res.ID)
			require.NoError(t, err)
			assertChannelIDs(t, loaded.NotificationChannels, channelIDs[0], channelIDs[1])
		})

		t.Run("N_to_M_overlap", func(t *testing.T) {
			res := &domain.Resource{
				Base:     domain.Base{ID: "ch-n-to-m", CreatedAt: time.Now()},
				Name:     "N→M", Type: domain.ResourceHTTP, Target: "https://example.com",
				IsActive: true,
				NotificationChannels: []*domain.NotificationChannel{
					{Base: domain.Base{ID: channelIDs[0]}},
					{Base: domain.Base{ID: channelIDs[1]}},
					{Base: domain.Base{ID: channelIDs[2]}},
				},
			}
			_, err := repo.Create(ctx, res)
			require.NoError(t, err)

			res.NotificationChannels = []*domain.NotificationChannel{
				{Base: domain.Base{ID: channelIDs[1]}},
				{Base: domain.Base{ID: channelIDs[3]}},
				{Base: domain.Base{ID: channelIDs[4]}},
			}
			require.NoError(t, repo.Update(ctx, res))
			loaded, err := repo.FindByID(ctx, res.ID)
			require.NoError(t, err)
			assertChannelIDs(t, loaded.NotificationChannels, channelIDs[1], channelIDs[3], channelIDs[4])
		})

		t.Run("N_to_0", func(t *testing.T) {
			res := &domain.Resource{
				Base:     domain.Base{ID: "ch-n-to-0", CreatedAt: time.Now()},
				Name:     "N→0", Type: domain.ResourceHTTP, Target: "https://example.com",
				IsActive: true,
				NotificationChannels: []*domain.NotificationChannel{
					{Base: domain.Base{ID: channelIDs[0]}},
					{Base: domain.Base{ID: channelIDs[1]}},
				},
			}
			_, err := repo.Create(ctx, res)
			require.NoError(t, err)

			res.NotificationChannels = nil
			require.NoError(t, repo.Update(ctx, res))
			loaded, err := repo.FindByID(ctx, res.ID)
			require.NoError(t, err)
			assert.Empty(t, loaded.NotificationChannels)
		})
	})
}

// TestResourceRepository_Preloads_Component_Credential verifies the 1-to-1
// preloads added in PR3 of US1.
func TestResourceRepository_Preloads_Component_Credential(t *testing.T) {
	setupPreloadCryptoKey(t)
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		repo := store.NewResourceRepositorySQLC(fx.Runtime)
		componentRepo := store.NewComponentRepositorySQLC(fx.Runtime)
		credRepo := store.NewResourceCredentialRepositorySQLC(fx.Runtime)
		ctx := context.Background()

		comp := &domain.Component{
			Base: domain.Base{ID: "", CreatedAt: time.Now()},
			Name: "preload-comp",
		}
		_, err := componentRepo.Create(ctx, comp)
		require.NoError(t, err)

		res := &domain.Resource{
			Base:        domain.Base{ID: "preload-res-1", CreatedAt: time.Now()},
			Name:        "preload",
			Type:        domain.ResourceProtocol,
			Target:      "redis://example.com",
			IsActive:    true,
			ComponentID: &comp.ID,
		}
		_, err = repo.Create(ctx, res)
		require.NoError(t, err)

		cred := &domain.ResourceCredential{
			Base:       domain.Base{ID: "", CreatedAt: time.Now()},
			ResourceID: res.ID,
			Username:   "alice",
			Password:   []byte("s3cret"),
			Options:    []byte(`{"db":0}`),
		}
		require.NoError(t, credRepo.Upsert(ctx, cred))

		loaded, err := repo.FindByID(ctx, res.ID)
		require.NoError(t, err)
		require.NotNil(t, loaded.Component, "Component preload missing")
		assert.Equal(t, comp.ID, loaded.Component.ID)
		assert.Equal(t, "preload-comp", loaded.Component.Name)
		require.NotNil(t, loaded.Credential, "Credential preload missing")
		assert.Equal(t, "alice", loaded.Credential.Username)
		assert.Equal(t, []byte("s3cret"), loaded.Credential.Password)
		assert.Equal(t, []byte(`{"db":0}`), loaded.Credential.Options)
	})
}

func assertChannelIDs(t *testing.T, channels []*domain.NotificationChannel, want ...string) {
	t.Helper()
	got := make([]string, len(channels))
	for i, ch := range channels {
		got[i] = ch.ID
	}
	sort.Strings(got)
	sort.Strings(want)
	assert.Equal(t, want, got)
}
