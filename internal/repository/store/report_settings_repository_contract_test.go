package store_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
)

func TestReportSettingsRepository_Contract(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		repo := store.NewReportSettingsRepositorySQLC(fx.Runtime)
		ctx := context.Background()

		t.Run("Get_before_any_write_is_ErrNotFound", func(t *testing.T) {
			_, err := repo.Get(ctx)
			require.ErrorIs(t, err, repository.ErrNotFound)
		})

		t.Run("Upsert_then_Get_returns_values", func(t *testing.T) {
			saved, err := repo.Upsert(ctx, &domain.ReportSettings{
				Enabled:        true,
				RecipientEmail: "ops@example.com",
				Schedule:       domain.ReportScheduleMonthly1st,
				Scope:          domain.ReportScopeAllResources,
			})
			require.NoError(t, err)
			assert.Equal(t, domain.ReportSettingsSingletonID, saved.ID)
			assert.True(t, saved.Enabled)
			assert.Nil(t, saved.LastSentAt)

			got, err := repo.Get(ctx)
			require.NoError(t, err)
			assert.Equal(t, "ops@example.com", got.RecipientEmail)
			assert.True(t, got.Enabled)
		})

		t.Run("second_Upsert_updates_same_single_row", func(t *testing.T) {
			_, err := repo.Upsert(ctx, &domain.ReportSettings{
				Enabled:        false,
				RecipientEmail: "changed@example.com",
				Schedule:       domain.ReportScheduleMonthly1st,
				Scope:          domain.ReportScopeAllResources,
			})
			require.NoError(t, err)

			got, err := repo.Get(ctx)
			require.NoError(t, err)
			assert.False(t, got.Enabled)
			assert.Equal(t, "changed@example.com", got.RecipientEmail)
			assert.Equal(t, domain.ReportSettingsSingletonID, got.ID) // still one row, same id
		})
	})
}
