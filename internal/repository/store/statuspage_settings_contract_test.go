package store_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
)

func TestStatusPageSettingsRepository_Contract(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		repo := store.NewStatusPageSettingsRepositorySQLC(fx.Runtime)
		runStatusPageSettingsContract(t, repo)
	})
}

func runStatusPageSettingsContract(t *testing.T, repo port.StatusPageSettingsRepository) {
	t.Helper()
	ctx := context.Background()

	t.Run("Get_Empty_Returns_Defaults", func(t *testing.T) {
		got, err := repo.Get(ctx)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "Status Page", got.Name)
		assert.True(t, got.EnableDetailsPage)
		assert.True(t, got.ShowUptimePercentage)
		assert.True(t, got.HidePausedMonitors)
		assert.True(t, got.ShowIncidentHistory)
	})

	t.Run("Upsert_Insert", func(t *testing.T) {
		s := &domain.StatusPageSettings{
			Name:                 "My Status Page",
			HomepageURL:          "https://example.com",
			EnableDetailsPage:    true,
			ShowUptimePercentage: true,
			HidePausedMonitors:   true,
			ShowIncidentHistory:  true,
		}
		require.NoError(t, repo.Upsert(ctx, s))

		got, err := repo.Get(ctx)
		require.NoError(t, err)
		assert.Equal(t, "My Status Page", got.Name)
		assert.Equal(t, "https://example.com", got.HomepageURL)
	})

	t.Run("Upsert_Update_Singleton", func(t *testing.T) {
		s := &domain.StatusPageSettings{
			Name:                 "Renamed Status Page",
			HomepageURL:          "https://renamed.example.com",
			CustomDomain:         "status.example.com",
			UmamiWebsiteID:       "72383dde-ac51-470e-991e-66d4b657adc2",
			UmamiScriptURL:       "https://cloud.umami.is/script.js",
			EnableDetailsPage:    true,
			ShowUptimePercentage: true,
			HidePausedMonitors:   true,
			ShowIncidentHistory:  true,
		}
		require.NoError(t, repo.Upsert(ctx, s))

		got, err := repo.Get(ctx)
		require.NoError(t, err)
		assert.Equal(t, "Renamed Status Page", got.Name)
		assert.Equal(t, "https://renamed.example.com", got.HomepageURL)
		assert.Equal(t, "status.example.com", got.CustomDomain)
		assert.Equal(t, "72383dde-ac51-470e-991e-66d4b657adc2", got.UmamiWebsiteID)
		assert.Equal(t, "https://cloud.umami.is/script.js", got.UmamiScriptURL)

		// Still singleton — second Get returns the SAME row.
		got2, err := repo.Get(ctx)
		require.NoError(t, err)
		assert.Equal(t, got.ID, got2.ID)
	})
}
