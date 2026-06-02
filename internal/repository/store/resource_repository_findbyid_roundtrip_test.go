package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/pkg/crypto"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
)

// TestResourceRepository_FindByIDPreloads_RoundTripBound verifies the same
// 1+R-style bound on the single-row read path (FR-003 of spec 049).
// FindByID issues 1 principal SELECT + 1 IN-query per preloaded relation
// that has a non-empty FK set (Component short-circuits when nil).
func TestResourceRepository_FindByIDPreloads_RoundTripBound(t *testing.T) {
	t.Setenv("APP_SECRET_KEY", roundtripTestKey)
	crypto.SetGlobalProvider(&crypto.EnvKeyProvider{})

	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		ctx := context.Background()
		sqlcSeedRepo := store.NewResourceRepositorySQLC(fx.Runtime)
		tagsRepo := store.NewTagsRepositorySQLC(fx.Runtime)
		chRepo := store.NewNotificationChannelRepositorySQLC(fx.Runtime)
		compRepo := store.NewComponentRepositorySQLC(fx.Runtime)

		comp := &domain.Component{Base: domain.Base{ID: "fbid-comp", CreatedAt: time.Now()}, Name: "fbid-comp"}
		_, err := compRepo.Create(ctx, comp)
		require.NoError(t, err)
		require.NoError(t, tagsRepo.Create(ctx, &domain.Tags{Base: domain.Base{ID: "fbid-tag", CreatedAt: time.Now()}, Name: "fbid-tag"}))
		require.NoError(t, chRepo.Create(ctx, &domain.NotificationChannel{
			Base:   domain.Base{ID: "fbid-ch", CreatedAt: time.Now()},
			Name:   "fbid-ch",
			Type:   domain.NotificationChannelTypeSlack,
			Config: []byte(`{"webhook":"https://example.com"}`),
		}))

		compID := comp.ID
		res := &domain.Resource{
			Base:                 domain.Base{ID: "fbid-res", CreatedAt: time.Now()},
			Name:                 "fbid-res",
			Type:                 domain.ResourceHTTP,
			Target:               "https://example.com",
			IsActive:             true,
			Tags:                 []*domain.Tags{{Base: domain.Base{ID: "fbid-tag"}}},
			NotificationChannels: []*domain.NotificationChannel{{Base: domain.Base{ID: "fbid-ch"}}},
			ComponentID:          &compID,
		}
		_, err = sqlcSeedRepo.Create(ctx, res)
		require.NoError(t, err)

		sqlcRepo, counter := roundtripFactory(t, fx)
		counter.Reset()
		loaded, err := sqlcRepo.FindByID(ctx, "fbid-res")
		require.NoError(t, err)
		require.NotNil(t, loaded)

		// 1 principal SELECT + 4 preload IN-queries (Tags, Channels,
		// Component since component_id != nil, Credential check empty).
		assert.EqualValues(t, 5, counter.Snapshot(),
			"FindByID with all relations: expected 5 round-trips (1+4)")
	})
}
