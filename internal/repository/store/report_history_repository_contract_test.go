package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
)

func mkReport(period string, sentAt time.Time) *domain.ReportHistory {
	return &domain.ReportHistory{
		Period:          period,
		SentAt:          sentAt,
		Status:          domain.ReportStatusDelivered,
		UptimePct:       99.5,
		IncidentCount:   2,
		DowntimeSeconds: 600,
		RecipientEmail:  "ops@example.com",
		Breakdown: []domain.ReportBreakdownLine{
			{Name: "API", UptimePct: 99.9, Incidents: 1},
			{Name: "DB", UptimePct: 99.1, Incidents: 1},
		},
	}
}

func TestReportHistoryRepository_Contract(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		repo := store.NewReportHistoryRepositorySQLC(fx.Runtime)
		ctx := context.Background()
		base := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

		t.Run("empty_history", func(t *testing.T) {
			rows, err := repo.ListRecent(ctx, 6)
			require.NoError(t, err)
			assert.Empty(t, rows)
		})

		t.Run("Create_then_FindByPeriod_roundtrips_breakdown", func(t *testing.T) {
			created, err := repo.Create(ctx, mkReport("2026-05", base.AddDate(0, -1, 0)))
			require.NoError(t, err)
			require.NotEmpty(t, created.ID)

			got, err := repo.FindByPeriod(ctx, "2026-05")
			require.NoError(t, err)
			assert.Equal(t, domain.ReportStatusDelivered, got.Status)
			assert.Equal(t, 600, int(got.DowntimeSeconds))
			require.Len(t, got.Breakdown, 2)
			assert.Equal(t, "API", got.Breakdown[0].Name)
			assert.Equal(t, 99.9, got.Breakdown[0].UptimePct)
		})

		t.Run("duplicate_period_rejected", func(t *testing.T) {
			_, err := repo.Create(ctx, mkReport("2026-05", base))
			require.Error(t, err) // UNIQUE(period)
		})

		t.Run("FindByPeriod_missing_is_ErrNotFound", func(t *testing.T) {
			_, err := repo.FindByPeriod(ctx, "1999-01")
			require.ErrorIs(t, err, repository.ErrNotFound)
		})

		t.Run("ListRecent_newest_first_and_limited", func(t *testing.T) {
			_, err := repo.Create(ctx, mkReport("2026-06", base.AddDate(0, 0, 5)))
			require.NoError(t, err)
			_, err = repo.Create(ctx, mkReport("2026-04", base.AddDate(0, -2, 0)))
			require.NoError(t, err)

			rows, err := repo.ListRecent(ctx, 2)
			require.NoError(t, err)
			require.Len(t, rows, 2)
			// newest sent_at first: 2026-06 (base+5d) then 2026-05 (base-1mo) ...
			assert.Equal(t, "2026-06", rows[0].Period)
			assert.True(t, rows[0].SentAt.After(rows[1].SentAt) || rows[0].SentAt.Equal(rows[1].SentAt))
		})
	})
}
