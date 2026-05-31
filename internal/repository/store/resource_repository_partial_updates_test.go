package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
)

// TestResourceRepository_UpdateMonitoringState_PointerSemantics verifies
// FR-005: nil preserves the column, &v writes the value, and **time.Time
// outer-non-nil + inner-nil explicitly clears the column (sets NULL).
func TestResourceRepository_UpdateMonitoringState_PointerSemantics(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		repo := store.NewResourceRepositorySQLC(fx.Runtime)
		ctx := context.Background()

		mkResource := func(id string) *domain.Resource {
			now := time.Now().UTC().Truncate(time.Second)
			return &domain.Resource{
				Base:                 domain.Base{ID: id, CreatedAt: now},
				Name:                 id,
				Type:                 domain.ResourceHTTP,
				Target:               "https://example.com",
				IsActive:             true,
				Status:               domain.StatusUp,
				FailureCount:         3,
				LastChecked:          &now,
				LastStatusTransition: &now,
				FlapStartedAt:        &now,
			}
		}

		t.Run("nil_preserves", func(t *testing.T) {
			res := mkResource("pms-nil-preserves")
			_, err := repo.Create(ctx, res)
			require.NoError(t, err)

			// All fields nil → no writes.
			require.NoError(t, repo.UpdateMonitoringState(ctx, res.ID, port.UpdateMonitoringStateRequest{}))

			loaded, err := repo.FindByID(ctx, res.ID)
			require.NoError(t, err)
			assert.Equal(t, domain.StatusUp, loaded.Status)
			assert.Equal(t, 3, loaded.FailureCount)
			require.NotNil(t, loaded.LastChecked)
			require.NotNil(t, loaded.LastStatusTransition)
			require.NotNil(t, loaded.FlapStartedAt)
		})

		t.Run("write_values", func(t *testing.T) {
			res := mkResource("pms-write")
			_, err := repo.Create(ctx, res)
			require.NoError(t, err)

			newStatus := domain.StatusDown
			newFC := 7
			require.NoError(t, repo.UpdateMonitoringState(ctx, res.ID, port.UpdateMonitoringStateRequest{
				Status:       &newStatus,
				FailureCount: &newFC,
			}))

			loaded, err := repo.FindByID(ctx, res.ID)
			require.NoError(t, err)
			assert.Equal(t, domain.StatusDown, loaded.Status)
			assert.Equal(t, 7, loaded.FailureCount)
		})

		t.Run("clear_nullable_timestamp", func(t *testing.T) {
			res := mkResource("pms-clear")
			_, err := repo.Create(ctx, res)
			require.NoError(t, err)

			// outer non-nil, inner nil → write NULL.
			var nilTime *time.Time
			require.NoError(t, repo.UpdateMonitoringState(ctx, res.ID, port.UpdateMonitoringStateRequest{
				FlapStartedAt: &nilTime,
			}))

			loaded, err := repo.FindByID(ctx, res.ID)
			require.NoError(t, err)
			assert.Nil(t, loaded.FlapStartedAt, "FlapStartedAt should have been cleared")
			// LastChecked left untouched (request did not include it).
			require.NotNil(t, loaded.LastChecked)
		})

		t.Run("nonexistent_returns_ErrNotFound", func(t *testing.T) {
			newStatus := domain.StatusDown
			err := repo.UpdateMonitoringState(ctx, "no-such-resource", port.UpdateMonitoringStateRequest{
				Status: &newStatus,
			})
			assert.Error(t, err)
		})
	})
}

// TestResourceRepository_UpdateMetadata_PointerSemantics mirrors the above
// for the SSL/domain expiry fields.
func TestResourceRepository_UpdateMetadata_PointerSemantics(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		repo := store.NewResourceRepositorySQLC(fx.Runtime)
		ctx := context.Background()

		now := time.Now().UTC().Truncate(time.Second)
		res := &domain.Resource{
			Base:     domain.Base{ID: "umd-1", CreatedAt: now},
			Name:     "umd",
			Type:     domain.ResourceHTTP,
			Target:   "https://example.com",
			IsActive: true,
			Metadata: &domain.ResourceMetaData{
				SSLExpirationDate: &now,
				SSLIssuer:         "Let's Encrypt",
			},
		}
		_, err := repo.Create(ctx, res)
		require.NoError(t, err)

		// Write SSLIssuer + clear SSLExpirationDate; leave domain fields alone.
		issuer := "DigiCert"
		var nilTime *time.Time
		require.NoError(t, repo.UpdateMetadata(ctx, res.ID, port.UpdateMetadataRequest{
			SSLIssuer:         &issuer,
			SSLExpirationDate: &nilTime,
		}))

		loaded, err := repo.FindByID(ctx, res.ID)
		require.NoError(t, err)
		require.NotNil(t, loaded.Metadata)
		assert.Equal(t, "DigiCert", loaded.Metadata.SSLIssuer)
		assert.Nil(t, loaded.Metadata.SSLExpirationDate)
	})
}
